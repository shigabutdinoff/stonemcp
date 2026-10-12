package engine

import (
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func (e *Engine) takeCommands() (jobs []*commandJob, ready, deferred []heldCommand) {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	return e.commands.take(func(job *commandJob) error { return e.commandGuardLocked(job, now) })
}

func (e *Engine) commandGuardLocked(job *commandJob, now time.Time) error {
	if e.snapshot == nil {
		return e.contextErrorLocked()
	}
	if e.snapshot.Loc.ID != job.locID {
		return ErrLocationChanged
	}
	if !now.Before(job.expires) {
		return ErrExpired
	}
	return nil
}

func (e *Engine) post(s *session) error {
	jobs, ready, deferred := e.takeCommands()
	return e.sendCommands(s, e.bannerRequests(), jobs, ready, deferred)
}

func (e *Engine) sendCommands(s *session, lead []mc.Request, jobs []*commandJob, ready, deferred []heldCommand) error {
	reqs := append(lead, holdRequests(ready)...)
	for _, job := range jobs {
		reqs = append(reqs, job.cmd.request())
	}
	if len(reqs) == 0 {
		return nil
	}
	values, err := s.exchange(reqs)
	if err != nil {
		return err
	}
	e.applyResults(ready, jobs, values[len(lead):])
	return e.restoreHolds(s, deferred)
}

func (e *Engine) restoreHolds(s *session, deferred []heldCommand) error {
	if len(deferred) == 0 {
		return nil
	}
	e.mu.Lock()
	live := e.commands.liveHolds(deferred)
	e.mu.Unlock()
	if len(live) == 0 {
		return nil
	}
	values, err := s.exchange(holdRequests(live))
	if err != nil {
		return err
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.commands.holds.dropRejected(live, values)
	return nil
}

func holdRequests(held []heldCommand) []mc.Request {
	reqs := make([]mc.Request, 0, len(held))
	for _, h := range held {
		reqs = append(reqs, h.cmd.request())
	}
	return reqs
}

func (e *Engine) applyResults(ready []heldCommand, jobs []*commandJob, values []mc.Value) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.commands.holds.dropRejected(ready, values[:len(ready)])
	for i, job := range jobs {
		e.commandResponseLocked(job, values[len(ready)+i])
	}
	e.answerFrozenLocked(time.Now())
	e.commands.sweep()
}

func (e *Engine) commandResponseLocked(job *commandJob, value mc.Value) {
	result := CommandResult{Result: value.String()}
	if value.IsError() {
		err := &GameError{Code: value.String(), Command: job.cmd.request().String()}
		job.finish(commandOutcome{result: result, err: err})
		return
	}
	if e.snapshot == nil {
		job.finish(commandOutcome{result: result, err: e.contextErrorLocked()})
		return
	}
	e.commands.delivered(job, result.Result, e.snapshot.Frame)
}
