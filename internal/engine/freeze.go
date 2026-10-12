package engine

import (
	"context"
	"fmt"
	"time"
)

// MaxFreeze ограничивает длительность одной заморозки игры.
const MaxFreeze = 2 * time.Minute

// DefaultFreeze задаёт лимит заморозки без явного значения.
const DefaultFreeze = 30 * time.Second

// ErrFreezeLimit означает лимит заморозки меньше нуля или больше MaxFreeze.
var ErrFreezeLimit = fmt.Errorf("лимит заморозки должен быть от 0 до %v секунд", MaxFreeze.Seconds())

// Freeze останавливает игру на ближайшем кадре и возвращает его состояние.
// Лимит от нуля до MaxFreeze, ноль выбирает DefaultFreeze.
// Повторный вызов в заморозке возвращает тот же кадр и не продлевает лимит.
func (e *Engine) Freeze(ctx context.Context, limit time.Duration) (*Snapshot, error) {
	limit, err := freezeLimit(limit)
	if err != nil {
		return nil, err
	}
	e.mu.Lock()
	s, err := e.snapshotLocked()
	var call *freezeCall
	if err == nil {
		call = e.freezer.want(limit)
	}
	e.mu.Unlock()
	if call == nil {
		return s, err
	}
	return e.awaitFreeze(ctx, call)
}

// Release отпускает замороженную игру, она продолжает со следующего кадра.
// Возвращает кадр заморозки или false, если игра не заморожена.
func (e *Engine) Release() (int64, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.releaseLocked()
}

// TakeNotice возвращает и сбрасывает сообщение о снятии заморозки по лимиту.
// Пустая строка означает, что сообщения нет.
func (e *Engine) TakeNotice() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.freezer.takeNotice()
}

func freezeLimit(limit time.Duration) (time.Duration, error) {
	if limit < 0 || limit > MaxFreeze {
		return 0, ErrFreezeLimit
	}
	if limit == 0 {
		return DefaultFreeze, nil
	}
	return limit, nil
}

func (e *Engine) releaseLocked() (int64, bool) {
	frame, ok := e.freezer.release()
	if ok {
		e.kick()
	}
	return frame, ok
}

func (e *Engine) awaitFreeze(ctx context.Context, call *freezeCall) (*Snapshot, error) {
	select {
	case outcome := <-call.done:
		return outcome.snapshot, outcome.err
	case <-ctx.Done():
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if outcome, ok := e.freezer.drop(call); ok {
		return outcome.snapshot, outcome.err
	}
	return nil, e.jobErrorLocked(ctx)
}
