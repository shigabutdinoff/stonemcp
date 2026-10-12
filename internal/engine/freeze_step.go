package engine

import (
	"context"
	"math"
	"time"
)

func (e *Engine) freezeStep(ctx context.Context, s *session) error {
	snapshot, limit := e.startFreeze()
	if limit == 0 {
		return nil
	}
	defer e.finishFreeze()
	timer := time.NewTimer(limit)
	defer timer.Stop()
	check := time.NewTicker(e.cfg.ProbeInterval)
	defer check.Stop()
	for e.frozen() {
		if err := e.serveFrozen(s, snapshot); err != nil {
			return err
		}
		if err := e.awaitFrozen(ctx, s, timer.C, check.C); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) awaitFrozen(ctx context.Context, s *session, limit, check <-chan time.Time) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.wake:
		return nil
	case <-check:
		return s.probe()
	case <-limit:
	}
	if err := s.probe(); err != nil {
		return err
	}
	e.expireFreeze()
	return nil
}

func (e *Engine) startFreeze() (*Snapshot, time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.snapshot == nil {
		return nil, 0
	}
	now := time.Now()
	limit := e.freezer.start(e.snapshot, now)
	e.answerFrozenLocked(now)
	return e.snapshot, limit
}

func (e *Engine) answerFrozenLocked(now time.Time) {
	if e.freezer.active() {
		e.commands.publish(e.snapshot, now, true)
	}
}

func (e *Engine) frozen() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.freezer.active()
}

func (e *Engine) expireFreeze() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.freezer.expire()
}

func (e *Engine) finishFreeze() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.freezer.finish(time.Now())
}

func (e *Engine) serveFrozen(s *session, snapshot *Snapshot) error {
	jobs, _ := e.takeEvaluations(math.MaxInt, nil, false)
	if _, err := e.runEvaluations(s, jobs); err != nil {
		return err
	}
	for job := e.nextShot(); job != nil; job = e.nextShot() {
		if err := e.readScreen(s, snapshot, job, math.MaxInt); err != nil {
			return err
		}
	}
	commands, _, deferred := e.takeCommands()
	return e.sendCommands(s, commands, nil, deferred)
}
