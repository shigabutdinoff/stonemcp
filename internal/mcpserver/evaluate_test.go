package mcpserver

import (
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func evaluate(queries ...map[string]any) map[string]any {
	return map[string]any{"queries": queries}
}

func TestEvaluateKeepsOrder(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFuncArgs("item.GetCooldown", []string{"bardiche"}, "42:bardiche")
	g.SetScreen([]string{"ab", "cd"})
	cs := online(t, g)
	out := decode[EvaluateOut](t, call(t, cs, "stone_evaluate", evaluate(
		map[string]any{"name": "item.GetCooldown", "args": []any{"bardiche"}},
		map[string]any{"name": "draw.GetSymbol", "args": []any{1, 1}},
		map[string]any{"name": "loc.id"},
	)))
	if len(out.Values) != 3 || out.Values[0] != "42:bardiche" || out.Values[1] != "d" || out.Values[2] != "rocky_plateau" {
		t.Errorf("значения = %q", out.Values)
	}
}

func TestEvaluateUnknownNameShowsGameCode(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	res := call(t, cs, "stone_evaluate", evaluate(map[string]any{"name": "no.such.thing", "args": []any{"x", 7}}))
	wantToolError(t, res, "fnc_err3")
	wantToolError(t, res, "no.such.thing(x, 7)")
}

func TestEvaluateRejectsBadArguments(t *testing.T) {
	g := fakegame.Start(t)
	calls := countCalls(g, "f")
	cs := online(t, g)
	for _, arg := range []any{1.5, true, nil, 1 << 31, -(1 << 31) - 1, []any{1}} {
		wantToolError(t, call(t, cs, "stone_evaluate", evaluate(map[string]any{"name": "f", "args": []any{arg}})), "аргумент")
	}
	wantToolError(t, call(t, cs, "stone_evaluate", evaluate()), "запрос")
	if n := calls(); n != 0 {
		t.Errorf("игра получила %d вызовов, ожидалось ни одного", n)
	}
}

func TestToArgAcceptsInt32Bounds(t *testing.T) {
	for _, v := range []any{float64(1<<31 - 1), float64(-(1 << 31)), "s"} {
		if _, err := toArg(v); err != nil {
			t.Errorf("%v отвергнут: %v", v, err)
		}
	}
}
