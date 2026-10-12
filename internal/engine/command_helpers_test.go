package engine

import (
	"context"
	"testing"
	"time"
)

func newJob(ctx context.Context, cmd Command) *commandJob {
	return &commandJob{stop: ctx.Done(), cmd: cmd, slot: cmd.slot(), expires: time.Now().Add(time.Minute), done: make(chan commandOutcome, 1)}
}

func sentJob(cmd Command) *commandJob {
	job := newJob(context.Background(), cmd)
	job.response, job.sent = "cmd_done", true
	return job
}

func expiredContext(t *testing.T) context.Context {
	ctx, stop := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	t.Cleanup(stop)
	return ctx
}

func canceledContext(*testing.T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func commandContext(t *testing.T, deadline time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	if deadline > 0 {
		var stop context.CancelFunc
		ctx, stop = context.WithTimeout(ctx, deadline)
		t.Cleanup(stop)
	}
	return ctx, cancel
}

func admitAll(*commandJob) error { return nil }

func addJob(q *commandQueue, cmd Command) error {
	return q.add(newJob(context.Background(), cmd))
}

func queuedIn(q *commandQueue) int {
	queued := 0
	for _, job := range q.jobs {
		if !job.inFlight {
			queued++
		}
	}
	return queued
}

func queuedCount(e *Engine) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return queuedIn(&e.commands)
}

func waiterCount(e *Engine) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.waiters)
}

func evalCount(e *Engine) int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.evals)
}

func doAsync(ctx context.Context, e *Engine, cmd Command, wait time.Duration) <-chan commandOutcome {
	done := make(chan commandOutcome, 1)
	go func() {
		result, err := e.Do(ctx, cmd, wait)
		done <- commandOutcome{result, err}
	}()
	return done
}

func receive[T any](t *testing.T, ch <-chan T, why string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(2 * time.Second):
		t.Fatal(why)
		var zero T
		return zero
	}
}

func expectPending(t *testing.T, done <-chan commandOutcome, d time.Duration, why string) {
	t.Helper()
	select {
	case got := <-done:
		t.Fatalf("%s = %+v", why, got)
	case <-time.After(d):
	}
}
