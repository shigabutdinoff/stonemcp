package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func eventually(t *testing.T, why string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if ok() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("не дождались: %s", why)
}

func launchEngine(t *testing.T, cfg Config) (*Engine, func()) {
	t.Helper()
	if cfg.ReconnectDelays == nil {
		cfg.ReconnectDelays = []time.Duration{5 * time.Millisecond, 20 * time.Millisecond}
	}
	e := New(cfg)
	stop := runEngine(t, e)
	t.Cleanup(stop)
	return e, stop
}

func runEngine(t *testing.T, e *Engine) (stop func()) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- e.Run(ctx) }()
	var once sync.Once
	return func() {
		once.Do(func() {
			cancel()
			select {
			case err := <-done:
				if err != nil {
					t.Errorf("остановка движка: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Error("движок не остановился")
			}
		})
	}
}

func startEngine(t *testing.T, g *fakegame.Game, cfg Config) *Engine {
	t.Helper()
	e, _ := startStoppable(t, g, cfg)
	return e
}

func startStoppable(t *testing.T, g *fakegame.Game, cfg Config) (*Engine, func()) {
	t.Helper()
	cfg.Addr = g.Addr()
	e, stop := launchEngine(t, cfg)
	eventually(t, "первый снимок", func() bool { _, err := e.Snapshot(); return err == nil })
	return e, stop
}
