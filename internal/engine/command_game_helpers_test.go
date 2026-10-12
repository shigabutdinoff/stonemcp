package engine

import (
	"context"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

type interruptCase struct {
	name     string
	deadline time.Duration
	act      func(g *fakegame.Game, cancel context.CancelFunc)
	want     error
}

func keepWaiting(*fakegame.Game, context.CancelFunc) {}

func cancelWait(_ *fakegame.Game, cancel context.CancelFunc) { cancel() }

func dropLink(g *fakegame.Game, _ context.CancelFunc) { g.DropConn() }

func pausedCommandEngine(t *testing.T, cfg Config) (*Engine, *fakegame.Game) {
	t.Helper()
	g := fakegame.Start(t)
	return pausedEngine(t, g, cfg), g
}

func pausedEngine(t *testing.T, g *fakegame.Game, cfg Config) *Engine {
	t.Helper()
	e := startEngine(t, g, cfg)
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	return e
}

func skipFrames(t *testing.T, g *fakegame.Game, n int, why string) int {
	t.Helper()
	frame := g.Frame()
	eventually(t, why, func() bool { return g.Frame() > frame+n })
	return frame
}
