package engine

import (
	"context"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func freezeNow(t *testing.T, e *Engine, limit time.Duration) *Snapshot {
	t.Helper()
	s, err := e.Freeze(testContext(t), limit)
	if err != nil {
		t.Fatalf("заморозка: %v", err)
	}
	return s
}

func freezeAsync(e *Engine) <-chan freezeOutcome {
	frozen := make(chan freezeOutcome, 1)
	go func() {
		s, err := e.Freeze(context.Background(), time.Minute)
		frozen <- freezeOutcome{snapshot: s, err: err}
	}()
	return frozen
}

func standsOn(t *testing.T, g *fakegame.Game, frame int, d time.Duration) {
	t.Helper()
	time.Sleep(d)
	if got := g.Frame(); got != frame {
		t.Fatalf("игра ушла с кадра %d на %d", frame, got)
	}
}

func freezeCallCount(e *Engine) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.freezer.calls)
}

func frozenLimit(e *Engine) time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.freezer.frozen == nil {
		return 0
	}
	return e.freezer.frozen.limit
}

func frozenFor(d time.Duration) *Engine {
	e := New(Config{})
	e.connected(mc.Version)
	e.snapshot = &Snapshot{Frame: 1, taken: time.Now()}
	freezeFor(e, d)
	return e
}

func freezeFor(e *Engine, d time.Duration) {
	e.freezer.frozen = &frozenState{started: time.Now()}
	time.Sleep(d)
}

func ranFor(t *testing.T, d time.Duration) *Engine {
	t.Helper()
	e := New(Config{})
	if err := e.Run(canceledContext(t)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(d)
	return e
}

func wantFPS20(t *testing.T, e *Engine, when string) {
	t.Helper()
	e.frame += 20
	e.publish(&Snapshot{Frame: e.frame})
	if fps := e.Status().FPS; fps != 20 {
		t.Errorf("FPS %s = %v, ожидался 20", when, fps)
	}
}
