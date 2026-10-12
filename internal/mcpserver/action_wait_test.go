package mcpserver

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestActivationWaitsForAnimation(t *testing.T) {
	g := fakegame.Start(t)
	cs := onlineWith(t, g, Options{Version: "27.0", Timeout: 50 * time.Millisecond})
	paused := g.PauseAfterCommand("activate")
	go func() {
		<-paused
		time.Sleep(100 * time.Millisecond)
		g.Set("hp", "63")
		g.Resume()
	}()
	out := decode[ActionOut](t, call(t, cs, "stone_activate", map[string]any{"target": "potion"}))
	cmds := g.CommandsNamed("activate")
	if out.Result != "cmd_done" || out.Snapshot == nil || out.Snapshot.Player.HP != 63 {
		t.Fatalf("ответ после анимации = %+v", out)
	}
	if len(cmds) != 1 || out.Snapshot.Frame != int64(cmds[0].Frame+1) {
		t.Errorf("кадр ответа %d, команды %+v", out.Snapshot.Frame, cmds)
	}
}

func TestPausedGameTimesOut(t *testing.T) {
	calls := map[string]map[string]any{
		"stone_equip_left": {"item": "sword"},
		"stone_activate":   {"target": "potion"},
	}
	for _, tt := range []struct {
		stale time.Duration
		tools []string
	}{
		{0, []string{"stone_equip_left"}},
		{100 * time.Millisecond, []string{"stone_equip_left", "stone_activate"}},
	} {
		g := fakegame.Start(t)
		cs := onlineWith(t, g, Options{Version: "27.0", Timeout: 50 * time.Millisecond, StaleAfter: 50 * time.Millisecond})
		if err := g.PauseAndWait(t.Context()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(tt.stale)
		for _, tool := range tt.tools {
			wantFastError(t, cs, tool, calls[tool], "игра не тикает")
		}
		if st := decode[StatusOut](t, call(t, cs, "stone_status", nil)); !st.Connected {
			t.Error("пауза не должна разрывать связь")
		}
		g.Resume()
	}
}
