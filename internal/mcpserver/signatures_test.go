package mcpserver

import (
	"slices"
	"strings"
	"testing"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func argsOf(t *testing.T, raw []any) []mc.Arg {
	t.Helper()
	args, err := toArgs(raw)
	if err != nil {
		t.Fatal(err)
	}
	return args
}

func TestSignatureTableCoversTheGame(t *testing.T) {
	if len(functions)+len(changing) != 44 {
		t.Errorf("функций %d, в MindConnect 0.3 их 44", len(functions)+len(changing))
	}
	if len(changing) != 11 {
		t.Errorf("функций с побочным эффектом %d, ожидалось 11", len(changing))
	}
	for name, sig := range functions {
		if slices.Contains(changing, name) {
			t.Errorf("%s одновременно читает и меняет игру", name)
		}
		if !strings.HasPrefix(sig.usage, name+"(") {
			t.Errorf("%s: подсказка %q должна начинаться с вызова", name, sig.usage)
		}
		if sig.check != nil && len(sig.args) == 0 {
			t.Errorf("%s: проверке нужен обязательный аргумент", name)
		}
	}
	if len(gameStates) != 57 {
		t.Errorf("состояний игры %d, ожидалось 57", len(gameStates))
	}
}

func TestCheckQueryAccepts(t *testing.T) {
	for _, tt := range []struct {
		name string
		raw  []any
	}{
		{"draw.GetSymbol", []any{1.0, 2.0}},
		{"app.GetStateNumber", []any{"Playing"}},
		{"event.GetObjectiveId", []any{0.0}},
		{"summon.GetId", nil},
		{"summon.GetId", []any{1.0}},
		{"summon.GetVar", []any{"ignition"}},
		{"summon.GetVar", []any{"ignition", 1.0}},
		{"item.CanActivate", nil},
		{"item.CanActivate", []any{"potion"}},
		{"item.GetTreasureCount", nil},
		{"var.get", []any{"phase"}},
		{"var.has", []any{7.0}},
		{"no.such.thing", []any{"x"}},
		{"hp", nil},
	} {
		if _, err := checkQuery(tt.name, tt.raw, argsOf(t, tt.raw)); err != nil {
			t.Errorf("%s%v: %v", tt.name, tt.raw, err)
		}
	}
}

func TestCheckQueryRejects(t *testing.T) {
	for _, tt := range []struct {
		name string
		raw  []any
		want string
	}{
		{"draw.GetSymbol", []any{1.0}, "draw.GetSymbol(x int, y int)"},
		{"draw.GetSymbol", []any{"1", "2"}, "draw.GetSymbol(x int, y int)"},
		{"app.GetStateNumber", []any{"Nope"}, "неизвестное состояние"},
		{"app.GetStateNumber", []any{"39"}, "неизвестное состояние"},
		{"app.GetStateNumber", []any{"playing"}, "неизвестное состояние"},
		{"event.GetObjectiveGoal", []any{-1.0}, "отрицательный"},
		{"event.GetObjectiveId", []any{-1e6}, "отрицательный"},
		{"summon.GetVar", []any{"fire"}, "неизвестная переменная призыва"},
		{"summon.GetVar", []any{1.0}, "summon.GetVar("},
		{"item.GetCooldown", nil, "item.GetCooldown(id string)"},
		{"item.GetTreasureCount", []any{1.0}, "item.GetTreasureCount()"},
		{"buffs.GetCount", []any{1.0}, "buffs.GetCount(id string)"},
		{"var.get", []any{"state.phase"}, "составные имена"},
		{"var.has", nil, "var.has(name"},
		{"loc.Leave", nil, "только читает"},
		{"var.set", []any{"x", 1.0}, "только читает"},
		{"player.ShowScaredFace", nil, "только читает"},
	} {
		_, err := checkQuery(tt.name, tt.raw, argsOf(t, tt.raw))
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s%v: ошибка %v, ожидалось упоминание %q", tt.name, tt.raw, err, tt.want)
		}
	}
}

func TestCheckQueryRejectsNonNumberForInt(t *testing.T) {
	for _, raw := range [][]any{{true}, {nil}, {[]any{1.0}}} {
		if _, err := checkQuery("event.GetObjectiveId", raw, nil); err == nil {
			t.Errorf("аргумент %v принят как целое число", raw)
		}
	}
}
