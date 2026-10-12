package engine

import (
	"context"
	"errors"
	"slices"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

type evalOutcome struct {
	values []mc.Value
	err    error
}
type evalJob struct {
	stop    <-chan struct{}
	queries []Query
	cost    int
	done    chan evalOutcome
}

// Evaluate читает значения в одном кадре, сохраняя порядок и коды ошибок игры.
// Дедлайн ctx при живом соединении возвращает ErrNotTicking.
func (e *Engine) Evaluate(ctx context.Context, queries []Query) ([]mc.Value, error) {
	copy, cost, err := prepareQueries(queries, e.cfg.MaxPacket)
	if err != nil {
		return nil, err
	}
	job := &evalJob{stop: ctx.Done(), queries: copy, cost: cost, done: make(chan evalOutcome, 1)}
	e.mu.Lock()
	err = e.apiErrorLocked()
	if err == nil && ctx.Err() == nil {
		e.evals = append(e.evals, job)
		e.kick()
	}
	e.mu.Unlock()
	if err != nil {
		return nil, err
	}
	defer e.removeEval(job)
	result, err := awaitJob(ctx, e, job.done)
	if err != nil {
		return nil, err
	}
	return result.values, result.err
}

func awaitJob[T any](ctx context.Context, e *Engine, done <-chan T) (T, error) {
	select {
	case <-ctx.Done():
	case result := <-done:
		if ctx.Err() == nil {
			return result, nil
		}
	}
	var zero T
	return zero, e.contextError(ctx)
}

func tryReceive[T any](done <-chan T) (T, bool) {
	select {
	case v := <-done:
		return v, true
	default:
		var zero T
		return zero, false
	}
}

func stopped(stop <-chan struct{}) bool {
	select {
	case <-stop:
		return true
	default:
		return false
	}
}

func removeJob[T comparable](jobs []T, job T) []T {
	return slices.DeleteFunc(jobs, func(j T) bool { return j == job })
}

func (e *Engine) removeEval(job *evalJob) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.evals = removeJob(e.evals, job)
}

func (e *Engine) finishEval(job *evalJob, outcome evalOutcome) {
	e.removeEval(job)
	job.done <- outcome
}

func (e *Engine) contextError(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.jobErrorLocked(ctx)
}

func (e *Engine) jobErrorLocked(ctx context.Context) error {
	err := ctx.Err()
	if err == nil || errors.Is(err, context.Canceled) {
		return err
	}
	return e.contextErrorLocked()
}

func (e *Engine) contextErrorLocked() error {
	if err := e.apiErrorLocked(); err != nil {
		return err
	}
	return ErrNotTicking
}
