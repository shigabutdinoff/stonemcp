package mcpserver

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestFreezeToolsKeepFrame(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	s := decode[engine.Snapshot](t, call(t, cs, "stone_freeze", map[string]any{"limit_seconds": 60}))
	ev := decode[EvaluateOut](t, call(t, cs, "stone_evaluate", map[string]any{"queries": []map[string]any{{"name": "time"}}}))
	act := decode[ActionOut](t, call(t, cs, "stone_equip_left", map[string]any{"item": "wand"}))
	if ev.Values[0] != strconv.Itoa(s.Time) || act.Result != "cmd_done" || !act.Frozen || act.Snapshot.Frame != s.Frame || g.Frame() != s.Time {
		t.Fatalf("заморозка %+v, чтение %+v, команда %+v, кадр игры %d", s, ev, act, g.Frame())
	}
	if rel := decode[ReleaseOut](t, call(t, cs, "stone_release", nil)); !rel.Released || rel.Frame != s.Frame {
		t.Fatalf("отпускание = %+v, ожидался кадр %d", rel, s.Frame)
	}
	if again := decode[ReleaseOut](t, call(t, cs, "stone_release", nil)); again.Released {
		t.Errorf("отпускание без заморозки = %+v", again)
	}
}

func TestFreezeLimitNotice(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	thawByLimit(t, cs, g)
	out := decode[StatusOut](t, call(t, cs, "stone_status", nil))
	if !strings.Contains(out.Notice, "снята по лимиту") || !out.Connected {
		t.Fatalf("ответ после лимита: %+v", out)
	}
	if next := decode[StatusOut](t, call(t, cs, "stone_status", nil)); next.Notice != "" {
		t.Errorf("сообщение о лимите повторилось: %q", next.Notice)
	}
}

var noticeThenError = regexp.MustCompile(`^лимит заморозки должен быть от 0 до 120 секунд: игра продолжила сама: заморозка на кадре \d+ снята по лимиту 0\.2 с$`)

func TestFreezeLimitNoticeOnError(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	thawByLimit(t, cs, g)
	res := call(t, cs, "stone_freeze", map[string]any{"limit_seconds": 300})
	got := text(res)
	if !res.IsError || !noticeThenError.MatchString(got) {
		t.Fatalf("ошибка после снятия заморозки по лимиту: %q", got)
	}
}

func TestFreezeRejectsBadLimits(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	cond := []map[string]any{{"kind": "loc_changed"}}
	wantToolError(t, call(t, cs, "stone_freeze", map[string]any{"limit_seconds": 300}), "120 секунд")
	wantToolError(t, call(t, cs, "stone_wait_for", map[string]any{"conditions": cond, "freeze": true, "freeze_limit_seconds": 300}), "120 секунд")
	wantToolError(t, call(t, cs, "stone_wait_for", map[string]any{"conditions": cond, "freeze_limit_seconds": 5}), "freeze")
	keepsRunning(t, g, "игра замерла после отказа")
}

func TestWaitForFreezeHasNoUnboundedBlock(t *testing.T) {
	g := fakegame.Start(t)
	cs := onlineWith(t, g, Options{Version: "27.0", Timeout: 50 * time.Millisecond})
	if err := g.PauseAndWait(t.Context()); err != nil {
		t.Fatal(err)
	}
	args := map[string]any{
		"conditions":      []map[string]any{{"kind": "seconds", "value": 0.05}},
		"timeout_seconds": 0.1, "freeze": true,
	}
	wantFastError(t, cs, "stone_wait_for", args, "игра не тикает")
	g.Resume()
	keepsRunning(t, g, "игра не продолжила после отказа заморозки")
}

func TestWaitForElementWithFreeze(t *testing.T) {
	g := fakegame.Start(t)
	g.SetElementalist("Ice")
	cs := online(t, g)
	at := g.Frame() + 50
	g.SwitchAt("foe", at, fakegame.ElementalistFoe("Ice"), fakegame.ElementalistFoe("AEther"))
	out := decode[WaitOut](t, call(t, cs, "stone_wait_for", map[string]any{
		"conditions": []map[string]any{{"kind": "foe_element_changed"}}, "timeout_seconds": 2, "freeze": true,
	}))
	if out.Reason != "foe_element_changed" || out.Snapshot.Foe.Element != "AEther" || g.Frame() != out.Snapshot.Time || out.Snapshot.Time != at {
		t.Fatalf("ожидание смены стихии = %+v, кадр игры %d", out, g.Frame())
	}
	call(t, cs, "stone_release", nil)
}

func thawByLimit(t *testing.T, cs *mcp.ClientSession, g *fakegame.Game) {
	t.Helper()
	s := decode[engine.Snapshot](t, call(t, cs, "stone_freeze", map[string]any{"limit_seconds": 0.2}))
	if !eventually(t, 2*time.Second, func() bool { return g.Frame() > s.Time }) {
		t.Fatal("игра не продолжила по лимиту")
	}
}

func keepsRunning(t *testing.T, g *fakegame.Game, why string) {
	t.Helper()
	before := g.Frame()
	if !eventually(t, time.Second, func() bool { return g.Frame() > before+3 }) {
		t.Error(why)
	}
}
