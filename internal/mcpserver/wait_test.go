package mcpserver

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestWaitForFoe(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	go func() {
		time.Sleep(30 * time.Millisecond)
		g.SetFoe("wasp", 12)
	}()
	out := decode[WaitOut](t, call(t, cs, "stone_wait_for", map[string]any{
		"conditions": []map[string]any{{"kind": "foe_present"}}, "timeout_seconds": 2,
	}))
	if out.Reason != "foe_present" || out.Snapshot == nil || out.Snapshot.Foe == nil || out.Snapshot.Foe.ID != "wasp" {
		t.Errorf("ответ = %+v", out)
	}
}

func TestWaitForTimeout(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	out := decode[WaitOut](t, call(t, cs, "stone_wait_for", map[string]any{
		"conditions": []map[string]any{{"kind": "loc_changed"}}, "timeout_seconds": 0.05,
	}))
	if out.Reason != "timeout" || out.Snapshot == nil {
		t.Errorf("ответ = %+v", out)
	}
}

func TestWaitForRejectsBadInput(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	for _, tt := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"conditions": []map[string]any{{"kind": "loc_changed"}}, "timeout_seconds": 121}, "120"},
		{map[string]any{"conditions": []map[string]any{{"kind": "loc_changed"}}, "timeout_seconds": -1}, "120"},
		{map[string]any{"conditions": []map[string]any{{"kind": "foe_dead"}}}, "foe_dead"},
		{map[string]any{"conditions": []map[string]any{}}, "услови"},
	} {
		wantToolError(t, call(t, cs, "stone_wait_for", tt.args), tt.want)
	}
}
