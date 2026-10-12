package engine

import (
	"context"
	"fmt"
	"time"
)

type commandOutcome struct {
	result CommandResult
	err    error
}

type commandJob struct {
	stop     <-chan struct{}
	cmd      Command
	slot     string
	locID    string
	expires  time.Time
	frame    int64
	response string
	inFlight bool
	sent     bool
	finished bool
	done     chan commandOutcome
}

// Do отправляет команду и ждёт первый кадр после исполнения.
// waitTimeout увеличивает защитное ожидание, ноль выбирает защитный предел.
// Отказ игры даёт GameError, прочие отказы возвращают ошибки пакета.
func (e *Engine) Do(ctx context.Context, cmd Command, waitTimeout time.Duration) (CommandResult, error) {
	started := time.Now()
	if waitTimeout < 0 {
		return CommandResult{}, fmt.Errorf("ожидание не может быть отрицательным")
	}
	if err := cmd.validate(e.cfg.MaxPacket); err != nil {
		return CommandResult{}, err
	}
	waitCtx, cancel := context.WithDeadline(ctx, started.Add(max(e.cfg.CommandTimeout, waitTimeout)))
	defer cancel()
	job := &commandJob{stop: waitCtx.Done(), cmd: cmd, slot: cmd.slot(), expires: started.Add(e.cfg.CommandTTL), done: make(chan commandOutcome, 1)}
	if err := e.enqueueCommand(waitCtx, job); err != nil {
		return CommandResult{}, err
	}
	return e.awaitCommand(waitCtx, job)
}

func (e *Engine) awaitCommand(ctx context.Context, job *commandJob) (CommandResult, error) {
	var outcome commandOutcome
	select {
	case outcome = <-job.done:
	case <-ctx.Done():
		outcome.err = e.contextError(ctx)
	}
	if late, ok := e.removeCommand(job); ok {
		outcome = late
	}
	return outcome.result, outcome.err
}

func (e *Engine) enqueueCommand(ctx context.Context, job *commandJob) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.jobErrorLocked(ctx); err != nil {
		return err
	}
	if err := e.apiErrorLocked(); err != nil {
		return err
	}
	if e.snapshot == nil {
		return ErrNotTicking
	}
	job.locID = e.snapshot.Loc.ID
	if err := e.commands.add(job); err != nil {
		return err
	}
	e.kick()
	return nil
}

func (e *Engine) removeCommand(job *commandJob) (commandOutcome, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.commands.remove(job)
}

func (j *commandJob) finish(outcome commandOutcome) {
	if j.finished {
		return
	}
	j.finished = true
	j.done <- outcome
}
