package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func prepareHoldRecovery(t *testing.T, e *Engine, until time.Time) {
	t.Helper()
	e.mu.Lock()
	defer e.mu.Unlock()
	e.commands.holds.accept(EquipRight("shield", time.Minute), time.Now())
	h := e.commands.holds["equipR"]
	h.until = until
	e.commands.holds["equipR"] = h
}

func expireQueued(e *Engine) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.commands.jobs[0].expires = time.Now().Add(-time.Second)
}

func assertRecoveredHold(t *testing.T, g *fakegame.Game, frame, want int) {
	t.Helper()
	eventually(t, "кадр после замены", func() bool { return g.Frame() > frame })
	count := 0
	for _, cmd := range g.CommandsNamed("equipR") {
		if cmd.Frame == frame && cmd.Arg == "shield" {
			count++
		}
	}
	if count != want {
		t.Fatalf("повторов удержания = %d, ожидалось %d, команды = %+v", count, want, g.Commands())
	}
}
