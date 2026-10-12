package mcpserver

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func wantRead(t *testing.T, cs *mcp.ClientSession, query map[string]any, want string, reads func() int, n int) {
	t.Helper()
	out := decode[EvaluateOut](t, call(t, cs, "stone_evaluate", query))
	if len(out.Values) != 1 || out.Values[0] != want || reads() != n {
		t.Errorf("%v: значения %q и чтений %d, ожидалось %q и %d", query, out.Values, reads(), want, n)
	}
}

func TestEvaluateOldestWithoutEffects(t *testing.T) {
	for _, list := range []string{"buffs", "debuffs"} {
		g := fakegame.Start(t)
		g.SetFunc(list+".oldest", "haste")
		reads := func() int { return len(g.Calls(list + ".oldest")) }
		g.Set(list+".count", "0")
		cs := online(t, g)
		query := evaluate(map[string]any{"name": list + ".oldest"})
		wantRead(t, cs, query, "", reads, 0)
		g.Set(list+".count", "2")
		wantRead(t, cs, query, "haste", reads, 1)
	}
}

func TestEvaluateSummonVarOnlyOnCinderwisp(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFunc("summon.GetId", "voidweaver")
	g.SetFunc("summon.GetVar", "3")
	reads := func() int { return len(g.Calls("summon.GetVar")) }
	cs := online(t, g)
	query := evaluate(map[string]any{"name": "summon.GetVar", "args": []any{"ignition"}})
	wantRead(t, cs, query, "", reads, 0)
	g.SetFunc("summon.GetId", "cinderwisp")
	wantRead(t, cs, query, "3", reads, 1)
}

func TestCheckQueryGuards(t *testing.T) {
	for _, name := range []string{"buffs.oldest", "debuffs.oldest"} {
		g, err := checkQuery(name, nil, nil)
		count := strings.TrimSuffix(name, ".oldest") + ".count"
		if err != nil || g == nil || g.Check.Name != count || g.OK("0") || g.OK("") || !g.OK("2") || g.Else != "" {
			t.Errorf("%s: страж %+v, ошибка %v", name, g, err)
		}
	}
	raw := []any{"ignition", 1.0}
	g, err := checkQuery("summon.GetVar", raw, argsOf(t, raw))
	if err != nil || g == nil || g.Check.Name != "summon.GetId" || len(g.Check.Args) != 1 || g.Check.Args[0].String() != "1" {
		t.Fatalf("summon.GetVar: страж %+v, ошибка %v", g, err)
	}
	if !g.OK("cinderwisp") || g.OK("voidweaver") || g.OK("") || g.Else != "" {
		t.Error("страж summon.GetVar должен пропускать только cinderwisp")
	}
	if g, _ := checkQuery("hp", nil, nil); g != nil {
		t.Errorf("hp: страж %+v не ожидался", g)
	}
	first, _ := checkQuery("buffs.oldest", nil, nil)
	second, _ := checkQuery("buffs.oldest", nil, nil)
	if first == second {
		t.Error("buffs.oldest: два запроса получили общий страж")
	}
}
