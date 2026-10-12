package engine

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// Engine обслуживает чтение игры и допускает одновременные вызовы API.
// Создаётся через New, Run вызывается не более одного раза одновременно.
type Engine struct {
	cfg        Config
	running    atomic.Bool
	mu         sync.Mutex
	status     Status
	snapshot   *Snapshot
	failure    error
	evals      []*evalJob
	shots      []*shotJob
	waiters    map[*waiter]struct{}
	commands   commandQueue
	frame      int64
	startFrame int64
	started    time.Time
	freezer    freezer
	wake       chan struct{}
}

// New создаёт движок без подключения, соединение обслуживает Run.
func New(cfg Config) *Engine {
	cfg = cfg.defaults()
	e := &Engine{
		cfg: cfg, failure: ErrDisconnected, status: Status{Mode: ModeSync},
		waiters: make(map[*waiter]struct{}), commands: newCommandQueue(cfg), wake: make(chan struct{}, 1),
	}
	e.commands.clock = func() time.Time { return e.freezer.clock(time.Now()) }
	return e
}

func (e *Engine) publish(s *Snapshot) {
	e.mu.Lock()
	defer e.mu.Unlock()
	prev := e.snapshot
	e.snapshot = s
	e.commands.extend(prev, s, e.cfg.ReadTimeout)
	e.commands.publish(s, time.Now(), false)
	e.checkWaitersLocked(s)
	e.status.Frame, e.status.LocID = s.Frame, s.Loc.ID
	if elapsed := e.freezer.clock(time.Now()).Sub(e.started).Seconds(); elapsed > 0 {
		e.status.FPS = float64(s.Frame-e.startFrame) / elapsed
	}
}

func (e *Engine) disconnect(cause error) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if errors.Is(cause, ErrUnsupportedMode) || !errors.Is(e.failure, ErrUnsupportedMode) {
		e.failure = cause
	}
	e.status.Connected = false
	e.status.LastError = e.failure.Error()
	e.snapshot = nil
	apiErr := e.apiErrorLocked()
	e.commands.abort(apiErr)
	for _, job := range e.evals {
		job.done <- evalOutcome{err: apiErr}
	}
	e.evals = nil
	for _, job := range e.shots {
		job.done <- shotOutcome{err: apiErr}
	}
	e.shots = nil
	for w := range e.waiters {
		w.done <- waitOutcome{err: apiErr}
		delete(e.waiters, w)
	}
	e.freezer.fail(apiErr)
	return e.failure
}

func (e *Engine) apiErrorLocked() error {
	if errors.Is(e.failure, ErrUnsupportedMode) {
		return ErrUnsupportedMode
	}
	if !e.status.Connected {
		return ErrDisconnected
	}
	return nil
}

func (e *Engine) kick() {
	select {
	case e.wake <- struct{}{}:
	default:
	}
}
