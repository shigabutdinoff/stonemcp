package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestQueuedCommandGuards(t *testing.T) {
	for _, tt := range []interruptCase{
		{"location", 0, func(g *fakegame.Game, _ context.CancelFunc) { g.Set("loc.id", "caves"); g.Resume() }, ErrLocationChanged},
		{"expired", 0, func(g *fakegame.Game, _ context.CancelFunc) { time.Sleep(60 * time.Millisecond); g.Resume() }, ErrExpired},
		{"canceled", 0, cancelWait, context.Canceled},
		{"deadline", 40 * time.Millisecond, keepWaiting, ErrNotTicking},
		{"timeout", 0, keepWaiting, ErrNotTicking},
		{"disconnect", 0, dropLink, ErrDisconnected},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, g := pausedCommandEngine(t, Config{CommandTTL: 30 * time.Millisecond, CommandTimeout: 100 * time.Millisecond})
			ctx, cancel := commandContext(t, tt.deadline)
			done := doAsync(ctx, e, EquipLeft("sword", 0), 0)
			eventually(t, "команда в очереди", func() bool { return queuedCount(e) == 1 })
			tt.act(g, cancel)
			if err := receive(t, done, "ожидание не завершилось").err; !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, ожидалось %v", err, tt.want)
			}
			g.Resume()
			skipFrames(t, g, 2, "возобновление кадров")
			if len(g.Commands()) != 0 {
				t.Fatal("брошенная команда отправлена")
			}
		})
	}
}

func TestCommandQueueLimit(t *testing.T) {
	e, _ := pausedCommandEngine(t, Config{QueueLimit: 3})
	ctx, cancel := commandContext(t, 0)
	var dones []<-chan commandOutcome
	for range 3 {
		dones = append(dones, doAsync(ctx, e, Activate("potion"), 0))
	}
	eventually(t, "заполнение очереди", func() bool { return queuedCount(e) == 3 })
	_, err := e.Do(testContext(t), Activate("potion"), 0)
	if !errors.Is(err, ErrQueueFull) {
		t.Fatalf("err = %v, ожидалось переполнение", err)
	}
	cancel()
	for _, done := range dones {
		<-done
	}
	if queuedCount(e) != 0 {
		t.Fatal("отменённые команды занимают очередь")
	}
}

func TestEnqueueReportsConnectionBeforeDeadline(t *testing.T) {
	e := New(Config{})
	if _, err := e.Do(expiredContext(t), Activate("potion"), 0); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("err = %v, ожидалось отсутствие связи", err)
	}
	if _, err := e.Do(canceledContext(t), Activate("potion"), 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, ожидалась отмена", err)
	}
}

func TestQueueDropsStoppedCommand(t *testing.T) {
	for _, tt := range []struct {
		name string
		make func(*testing.T) context.Context
		want error
	}{
		{"deadline", expiredContext, ErrNotTicking},
		{"canceled", canceledContext, context.Canceled},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := New(Config{})
			e.snapshot, e.failure = &Snapshot{}, nil
			e.status.Connected = true
			ctx := tt.make(t)
			job := newJob(ctx, Activate("potion"))
			e.commands.jobs = []*commandJob{job}
			e.takeCommands()
			if _, ok := tryReceive(job.done); ok || len(e.commands.jobs) != 0 {
				t.Fatalf("брошенное задание получило ответ или осталось в очереди: %d", len(e.commands.jobs))
			}
			if _, err := e.awaitCommand(ctx, job); !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, ожидалось %v", err, tt.want)
			}
		})
	}
}

func TestMissingSnapshotStopsCommands(t *testing.T) {
	e := New(Config{})
	queued := newJob(context.Background(), Activate("potion"))
	e.commands.jobs = []*commandJob{queued}
	e.takeCommands()
	if err := (<-queued.done).err; !errors.Is(err, ErrDisconnected) {
		t.Fatalf("ошибка охраны без снимка = %v, ожидалось %v", err, ErrDisconnected)
	}
	sent := newJob(context.Background(), Activate("potion"))
	e.applyResults(nil, []*commandJob{sent}, []mc.Value{"cmd_done"})
	if err := (<-sent.done).err; !errors.Is(err, ErrDisconnected) {
		t.Fatalf("ошибка ответа без снимка = %v, ожидалось %v", err, ErrDisconnected)
	}
}
