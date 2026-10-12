package mcpserver

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestActionsReturnResultAndSnapshot(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	tests := []struct {
		tool      string
		args      map[string]any
		name, arg string
	}{
		{"stone_equip", map[string]any{"item": "bardiche"}, "equip", "bardiche"},
		{"stone_equip_left", map[string]any{"item": "poison sword"}, "equipL", "poison sword"},
		{"stone_equip_right", map[string]any{"item": "vigor shield"}, "equipR", "vigor shield"},
		{"stone_activate", map[string]any{"target": "potion"}, "activate", "potion"},
		{"stone_loadout", map[string]any{"number": 2}, "loadout", "2"},
		{"stone_brew", map[string]any{"ingredients": []string{"bronze", "tar"}}, "brew", "bronze + tar"},
	}
	for _, tt := range tests {
		out := decode[ActionOut](t, call(t, cs, tt.tool, tt.args))
		if out.Result != "cmd_done" || out.Snapshot == nil || out.Snapshot.Loc.ID != "rocky_plateau" {
			t.Errorf("%s: ответ = %+v", tt.tool, out)
		}
		cmds := g.CommandsNamed(tt.name)
		if len(cmds) != 1 || cmds[0].Arg != tt.arg {
			t.Errorf("%s: игра получила %+v, ожидалась одна команда %q", tt.tool, cmds, tt.arg)
		}
	}
}

func TestHoldSecondsKeepsSending(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	decode[ActionOut](t, call(t, cs, "stone_equip_right", map[string]any{"item": "shield", "hold_seconds": 0.1}))
	if !eventually(t, time.Second, func() bool { return len(g.CommandsNamed("equipR")) >= 3 }) {
		t.Errorf("удержание отправлено %d раз", len(g.CommandsNamed("equipR")))
	}
	wantToolError(t, call(t, cs, "stone_equip_right", map[string]any{"item": "shield", "hold_seconds": 601}), "удержание")
}

func TestRejectedCommandIsToolError(t *testing.T) {
	g := fakegame.Start(t)
	g.Reject("equipL", true)
	cs := online(t, g)
	res := call(t, cs, "stone_equip_left", map[string]any{"item": "sword"})
	wantToolError(t, res, "cmd_err")
	wantToolError(t, res, "sword")
}

func TestBrewRejectsUnknownIngredients(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	for _, ingredients := range [][]string{{"fire_elemental"}, {"empty"}, {"stone", "gold"}, {}} {
		wantToolError(t, call(t, cs, "stone_brew", map[string]any{"ingredients": ingredients}), "stone, wood, tar, bronze")
	}
	if cmds := g.CommandsNamed("brew"); len(cmds) != 0 {
		t.Errorf("игра получила %+v, зелье не ожидалось", cmds)
	}
}

func TestAtSignIsRefused(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	for tool, args := range map[string]map[string]any{
		"stone_equip":       {"item": "@hp@"},
		"stone_equip_left":  {"item": "sword @hp@"},
		"stone_equip_right": {"item": "@"},
		"stone_activate":    {"target": "@item.left@"},
		"stone_brew":        {"ingredients": []string{"@x@"}},
	} {
		wantToolError(t, call(t, cs, tool, args), "@")
	}
	if cmds := g.Commands(); len(cmds) != 0 {
		t.Errorf("игра получила %+v, команды не ожидались", cmds)
	}
}
