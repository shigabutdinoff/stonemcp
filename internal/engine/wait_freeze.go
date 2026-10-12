package engine

import (
	"context"
	"slices"
	"time"
)

func (e *Engine) thawAndRegister(w *waiter) (waitOutcome, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	base, err := e.snapshotLocked()
	if err != nil {
		return waitOutcome{err: err}, true
	}
	w.base = base
	if outcome, ok := e.completeWaitLocked(w, base, time.Now()); ok {
		return outcome, true
	}
	e.releaseLocked()
	e.waiters[w] = struct{}{}
	return waitOutcome{}, false
}

func (e *Engine) completeWaitLocked(w *waiter, s *Snapshot, now time.Time) (waitOutcome, bool) {
	result, ok := w.result(e.waitConditionsLocked(w, s), s, now)
	if !ok {
		return waitOutcome{}, false
	}
	return waitOutcome{result: result, freeze: e.freezer.want(w.freezeLimit)}, true
}

func (e *Engine) waitConditionsLocked(w *waiter, s *Snapshot) []Condition {
	if w.freezeLimit == 0 || !e.freezer.missed(s.Frame) {
		return w.conditions
	}
	return slices.DeleteFunc(slices.Clone(w.conditions), func(c Condition) bool { return c.Kind != CondSeconds })
}

func (e *Engine) dropWaiterLocked(w *waiter) (waitOutcome, bool) {
	delete(e.waiters, w)
	return tryReceive(w.done)
}

func (e *Engine) abandonWaiter(w *waiter) waitOutcome {
	e.mu.Lock()
	defer e.mu.Unlock()
	outcome, _ := e.dropWaiterLocked(w)
	return outcome
}

func (e *Engine) settle(ctx context.Context, outcome waitOutcome) (WaitResult, error) {
	if outcome.freeze != nil {
		s, err := e.awaitFreeze(ctx, outcome.freeze)
		if err != nil {
			return WaitResult{}, err
		}
		return WaitResult{Reason: outcome.result.Reason, Snapshot: s}, nil
	}
	if err := ctx.Err(); err != nil {
		return WaitResult{}, err
	}
	return outcome.result, outcome.err
}

func (e *Engine) timedOutcome(w *waiter) waitOutcome {
	e.mu.Lock()
	defer e.mu.Unlock()
	if outcome, ok := e.dropWaiterLocked(w); ok {
		return outcome
	}
	s, err := e.snapshotLocked()
	if err != nil {
		return waitOutcome{err: err}
	}
	if outcome, ok := e.completeWaitLocked(w, s, time.Now()); ok {
		return outcome
	}
	return waitOutcome{result: WaitResult{Reason: ReasonTimeout, Snapshot: s}}
}
