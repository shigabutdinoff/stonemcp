package engine

import "time"

// WaitResult содержит причину завершения ожидания и состояние игры.
type WaitResult struct {
	// Reason содержит вид выполненного условия или timeout.
	Reason string `json:"reason"`
	// Snapshot содержит состояние при завершении ожидания.
	Snapshot *Snapshot `json:"snapshot"`
}

type waitOutcome struct {
	result WaitResult
	err    error
	freeze *freezeCall
}
type waiter struct {
	conditions  []Condition
	base        *Snapshot
	started     time.Time
	freezeLimit time.Duration
	done        chan waitOutcome
}

func (w *waiter) result(conds []Condition, cur *Snapshot, now time.Time) (WaitResult, bool) {
	for _, c := range conds {
		if c.met(w.base, cur, now.Sub(w.started)) {
			return WaitResult{Reason: c.Kind, Snapshot: cur.aged(now)}, true
		}
	}
	return WaitResult{}, false
}

func (e *Engine) checkWaitersLocked(s *Snapshot) {
	now := time.Now()
	for w := range e.waiters {
		if outcome, ok := e.completeWaitLocked(w, s, now); ok {
			w.done <- outcome
			delete(e.waiters, w)
		}
	}
}
