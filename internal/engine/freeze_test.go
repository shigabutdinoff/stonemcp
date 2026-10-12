package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestFreezeStopsGameUntilRelease(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	s := freezeNow(t, e, time.Minute)
	standsOn(t, g, s.Time, 100*time.Millisecond)
	if err := e.Ticking(10 * time.Millisecond); err != nil {
		t.Fatalf("заморозка принята за остановку игры: %v", err)
	}
	if frame, ok := e.Release(); !ok || frame != s.Frame {
		t.Fatalf("отпускание = %d, %v, ожидался кадр %d", frame, ok, s.Frame)
	}
	res, err := e.WaitFor(testContext(t), []Condition{{Kind: CondFrames, Value: 3}}, time.Second)
	if err != nil || res.Snapshot.Frame-s.Frame != int64(res.Snapshot.Time-s.Time) {
		t.Fatalf("после отпускания пропущены кадры: было %+v, стало %+v, %v", s, res.Snapshot, err)
	}
	if _, ok := e.Release(); ok {
		t.Error("отпускание без заморозки вернуло true")
	}
}

func TestFreezeWithoutFramesFailsAtOnce(t *testing.T) {
	e := New(Config{})
	if _, err := e.Freeze(testContext(t), 0); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("заморозка без связи: %v", err)
	}
	e.connected("0.3")
	if _, err := e.Freeze(testContext(t), 0); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("заморозка до первого кадра: %v", err)
	}
}

func TestPendingFreezeCanceledOrDropped(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	if _, err := e.Freeze(ctx, time.Minute); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("заморозка на паузе: %v", err)
	}
	g.Resume()
	skipFrames(t, g, 3, "снятый запрос не должен замораживать игру")
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	frozen := freezeAsync(e)
	eventually(t, "запрос заморозки", func() bool { return freezeCallCount(e) == 1 })
	g.DropConn()
	if out := receive(t, frozen, "заморозка не завершилась при разрыве"); !errors.Is(out.err, ErrDisconnected) {
		t.Fatalf("заморозка при разрыве: %v", out.err)
	}
}

func TestStopDuringFreezeLetsGameGo(t *testing.T) {
	g := fakegame.Start(t)
	e, stop := startStoppable(t, g, Config{})
	s := freezeNow(t, e, time.Minute)
	stop()
	startEngine(t, g, Config{})
	eventually(t, "игра идёт после отключения в заморозке", func() bool { return g.Frame() > s.Time+3 })
}
