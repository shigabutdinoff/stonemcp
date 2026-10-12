package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCanceledWaitLeavesNoFreeze(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	done := waitAsync(ctx, e, []Condition{{Kind: CondFoePresent}}, time.Minute)
	eventually(t, "регистрация ожидания", func() bool { return waiterCount(e) == 1 })
	e.mu.Lock()
	cancel()
	e.checkWaitersLocked(&Snapshot{Foe: &Foe{ID: "wasp"}})
	e.mu.Unlock()
	out := receive(t, done, "отмена не замечена")
	if !errors.Is(out.err, context.Canceled) || freezeCallCount(e) != 0 {
		t.Fatalf("отменённое ожидание с заморозкой: %+v, %v, вызовов заморозки %d", out.result, out.err, freezeCallCount(e))
	}
	g.Resume()
	skipFrames(t, g, 3, "игра идёт после отмены")
}

func TestTimedOutWaitIsUnregistered(t *testing.T) {
	e, _ := pausedCommandEngine(t, Config{})
	w := &waiter{conditions: []Condition{{Kind: CondFoePresent}}, started: time.Now(), done: make(chan waitOutcome, 1)}
	if _, done := e.thawAndRegister(w); done {
		t.Fatal("условие оказалось выполненным сразу")
	}
	outcome := e.timedOutcome(w)
	if outcome.result.Reason != ReasonTimeout || waiterCount(e) != 0 {
		t.Fatalf("незавершённое ожидание после таймаута: %+v, ждущих %d", outcome.result, waiterCount(e))
	}
}
