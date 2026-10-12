package engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestConcurrentReaders(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{FrameBudget: 40})
	ctx := testContext(t)
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			values, err := e.Evaluate(ctx, []Query{{Name: "hp"}, {Name: "time"}})
			if err != nil || len(values) != 2 || values[0] != "80" {
				t.Errorf("параллельное чтение: %v, %v", values, err)
			}
			shot, err := e.Screenshot(ctx)
			if err != nil || shot.Text != "#--#\n|..|\n#--#" {
				t.Errorf("параллельный экран: %+v, %v", shot, err)
			}
			res, err := e.WaitFor(ctx, []Condition{{Kind: CondFrames, Value: 2}}, time.Second)
			if err != nil || res.Reason != CondFrames {
				t.Errorf("параллельное ожидание: %+v, %v", res, err)
			}
		})
	}
	wg.Wait()
}

func TestCancelledQueuedReadNeverReachesGame(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFunc("probe", "1")
	e := startEngine(t, g, Config{})
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := e.Evaluate(ctx, []Query{{Name: "probe"}}); done <- err }()
	eventually(t, "задание в очереди", func() bool { return evalCount(e) == 1 })
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("отмена чтения: %v", err)
	}
	g.Resume()
	if _, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}}); err != nil {
		t.Fatal(err)
	}
	if len(g.Calls("probe")) != 0 {
		t.Fatal("отменённое задание отправлено игре")
	}
}

func TestDisconnectDiscardsQueuedWork(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFunc("probe", "1")
	e := startEngine(t, g, Config{ReconnectDelays: []time.Duration{100 * time.Millisecond}})
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	ctx := testContext(t)
	done := make(chan error, 2)
	go func() { _, err := e.Evaluate(ctx, []Query{{Name: "probe"}}); done <- err }()
	go func() { _, err := e.Screenshot(ctx); done <- err }()
	eventually(t, "очереди чтения", func() bool { e.mu.Lock(); defer e.mu.Unlock(); return len(e.evals) == 1 && len(e.shots) == 1 })
	g.DropConn()
	for range 2 {
		if err := <-done; !errors.Is(err, ErrDisconnected) {
			t.Fatalf("сброс задания: %v", err)
		}
	}
	g.Resume()
	eventually(t, "новая сессия", func() bool { _, err := e.Snapshot(); return err == nil })
	if len(g.Calls("probe")) != 0 {
		t.Fatal("старое задание перенесено в новую сессию")
	}
}

func TestCommandsInFrameGetIndependentSnapshots(t *testing.T) {
	q := newCommandQueue(Config{}.defaults())
	left, right := sentJob(EquipLeft("sword", 0)), sentJob(EquipRight("shield", 0))
	q.jobs = []*commandJob{left, right}
	q.publish(&Snapshot{Frame: 2, Foe: &Foe{HP: 10}}, time.Now(), false)
	first, second := (<-left.done).result.Snapshot, (<-right.done).result.Snapshot
	if first == nil || second == nil || first == second || first.Foe == second.Foe {
		t.Fatalf("команды кадра получили общий снимок: %p и %p", first, second)
	}
	first.Player.HP, first.Foe.HP = 1, 1
	if second.Player.HP == 1 || second.Foe.HP == 1 {
		t.Fatalf("правка снимка одной команды видна другой: %+v", second)
	}
}
