package mcpserver

import (
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func countCalls(g *fakegame.Game, names ...string) func() int {
	for _, name := range names {
		g.SetFunc(name, "")
	}
	return func() int {
		n := 0
		for _, name := range names {
			n += len(g.Calls(name))
		}
		return n
	}
}

func TestEvaluateRefusesGameChanges(t *testing.T) {
	g := fakegame.Start(t)
	calls := countCalls(g, changing...)
	cs := online(t, g)
	for _, name := range changing {
		wantToolError(t, call(t, cs, "stone_evaluate", evaluate(map[string]any{"name": name})), "только читает")
	}
	if n := calls(); n != 0 {
		t.Errorf("игра получила %d вызовов с побочным эффектом", n)
	}
}

func TestEvaluateChecksArguments(t *testing.T) {
	g := fakegame.Start(t)
	calls := countCalls(g, "draw.GetSymbol", "var.get")
	cs := online(t, g)
	for _, tt := range []struct {
		query map[string]any
		want  string
	}{
		{map[string]any{"name": "draw.GetSymbol", "args": []any{1}}, "draw.GetSymbol(x int, y int)"},
		{map[string]any{"name": "var.get", "args": []any{"state.phase"}}, "составные имена переменных скрипта не поддерживаются"},
	} {
		wantToolError(t, call(t, cs, "stone_evaluate", evaluate(tt.query)), tt.want)
	}
	if n := calls(); n != 0 {
		t.Errorf("игра получила %d неверных вызовов", n)
	}
}
