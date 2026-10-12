package engine

import (
	"slices"
	"strconv"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestGuardChecksBeforeReadingInSameFrame(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFunc("check", "True")
	g.SetFunc("guarded", "ok")
	e := startEngine(t, g, Config{FrameBudget: 35, MaxPacket: 100})
	got, err := e.Evaluate(testContext(t), []Query{
		{Name: "guarded", Guard: &Guard{Check: Query{Name: "check"}, OK: func(v mc.Value) bool { return v == "True" }}},
		{Name: "time"},
	})
	checks, reads := g.Calls("check"), g.Calls("guarded")
	if err != nil || len(got) != 2 || got[0] != "ok" || len(checks) == 0 || len(reads) == 0 {
		t.Fatalf("защищённое чтение: %v, %v, проверок %d, чтений %d", got, err, len(checks), len(reads))
	}
	check, read := checks[len(checks)-1].Frame, reads[len(reads)-1].Frame
	if check != read || got[1] != mc.Value(strconv.Itoa(read)) {
		t.Fatalf("проверка и чтение разных кадров: %d и %d, time %v", check, read, got[1])
	}
}

func TestGuardSkipsUnsafeRead(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFunc("unsafe", "не должно читаться")
	e := startEngine(t, g, Config{})
	got, err := e.Evaluate(testContext(t), []Query{
		{Name: "hp"},
		{Name: "unsafe", Guard: &Guard{Check: Query{Name: "foe"}, OK: func(v mc.Value) bool { return v != "" }, Else: "fnc_err3"}},
	})
	if err != nil || !slices.Equal(got, []mc.Value{"80", "fnc_err3"}) || len(g.Calls("unsafe")) != 0 {
		t.Fatalf("защищённое чтение: %v, %v", got, err)
	}
}
