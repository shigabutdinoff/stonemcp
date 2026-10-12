package engine

import (
	"context"
	"fmt"
	"slices"
	"time"
)

// WaitFor ждёт первое выполненное условие, не задерживая игру.
// Таймаут возвращает результат с timeout, отмена ctx возвращает его ошибку.
// Вызов в заморозке отпускает игру, если условие ещё не выполнено.
func (e *Engine) WaitFor(ctx context.Context, conditions []Condition, timeout time.Duration) (WaitResult, error) {
	return e.wait(ctx, conditions, timeout, 0)
}

// WaitFreeze ждёт как WaitFor и замораживает игру на кадре события.
// Лимит заморозки как у Freeze, таймаут игру не замораживает, дедлайн
// ctx при живом соединении возвращает ErrNotTicking и снимает заморозку.
func (e *Engine) WaitFreeze(ctx context.Context, conditions []Condition, timeout, limit time.Duration) (WaitResult, error) {
	limit, err := freezeLimit(limit)
	if err != nil {
		return WaitResult{}, err
	}
	return e.wait(ctx, conditions, timeout, limit)
}

func (e *Engine) wait(ctx context.Context, conditions []Condition, timeout, limit time.Duration) (WaitResult, error) {
	if ctx.Err() != nil {
		return WaitResult{}, ctx.Err()
	}
	delay, err := waitDelay(conditions, timeout)
	if err != nil {
		return WaitResult{}, err
	}
	w := &waiter{conditions: slices.Clone(conditions), freezeLimit: limit, started: time.Now(), done: make(chan waitOutcome, 1)}
	if outcome, done := e.thawAndRegister(w); done {
		return e.settle(ctx, outcome)
	}
	timer := time.NewTimer(max(0, delay-time.Since(w.started)))
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return e.settle(ctx, e.abandonWaiter(w))
	case outcome := <-w.done:
		return e.settle(ctx, outcome)
	case <-timer.C:
		return e.settle(ctx, e.timedOutcome(w))
	}
}

func waitDelay(conditions []Condition, timeout time.Duration) (time.Duration, error) {
	if len(conditions) == 0 || timeout < 0 {
		return 0, fmt.Errorf("нужны условия и неотрицательный таймаут")
	}
	delay := timeout
	for _, c := range conditions {
		if err := c.validate(); err != nil {
			return 0, err
		}
		if c.Kind == CondSeconds {
			delay = min(delay, time.Duration(c.Value*float64(time.Second)))
		}
	}
	return delay, nil
}
