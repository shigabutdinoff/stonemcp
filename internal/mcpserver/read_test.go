package mcpserver

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestStatusAndSnapshot(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)

	st := decode[StatusOut](t, call(t, cs, "stone_status", nil))
	if !st.Connected || st.ServerVersion != "27.0" || st.Mode != "sync" || st.ProtocolVersion != "0.3" || st.Hint != "" {
		t.Errorf("статус = %+v", st)
	}
	g.SetFoe("scarab", 30)
	var snap engine.Snapshot
	if !eventually(t, time.Second, func() bool {
		snap = decode[engine.Snapshot](t, call(t, cs, "stone_snapshot", nil))
		return snap.Foe != nil
	}) {
		t.Fatal("враг не появился в состоянии")
	}
	if snap.Loc.ID != "rocky_plateau" || snap.Player.HP != 80 || snap.Foe.ID != "scarab" {
		t.Errorf("состояние = %+v", snap)
	}
}

func TestScreenshot(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen([]string{"#--#", "|o |", "#--#"})
	cs := online(t, g)
	out := decode[ScreenshotOut](t, call(t, cs, "stone_screenshot", nil))
	if out.Text != "#--#\n|o |\n#--#" || out.Width != 4 || out.Height != 3 || out.FramesSpanned < 1 {
		t.Errorf("экран = %+v", out)
	}
}

func TestOfflineGame(t *testing.T) {
	cs := connect(t, "127.0.0.1:1", Options{Version: "27.0"})

	st := decode[StatusOut](t, call(t, cs, "stone_status", nil))
	if st.Connected || st.Hint != Hint {
		t.Errorf("статус = %+v", st)
	}
	for tool, args := range map[string]any{
		"stone_snapshot":   nil,
		"stone_screenshot": nil,
		"stone_equip":      map[string]any{"item": "sword"},
	} {
		wantFastError(t, cs, tool, args, "//MindConnect: sync")
	}
}
