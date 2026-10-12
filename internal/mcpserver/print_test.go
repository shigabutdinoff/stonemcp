package mcpserver

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestPrintRepeatsEveryFrame(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	out := decode[PrintOut](t, call(t, cs, "stone_print", map[string]any{"text": "hello", "x": 0, "y": 1, "color": "green", "hold_seconds": 0.05}))
	if out.Result != "cmd_done" {
		t.Errorf("ответ = %+v", out)
	}
	eventually(t, time.Second, func() bool { return len(g.CommandsNamed(">")) >= 3 })
	cmds := g.CommandsNamed(">")
	if len(cmds) < 3 || cmds[0].Arg != "`0,1,#green,hello" {
		t.Errorf("текст отправлен %d раз, первая команда %+v", len(cmds), cmds)
	}
}

func TestPrintHoldsThreeSecondsByDefault(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	decode[PrintOut](t, call(t, cs, "stone_print", map[string]any{"text": "hi"}))
	time.Sleep(50 * time.Millisecond)
	before := len(g.CommandsNamed(">"))
	if !eventually(t, time.Second, func() bool { return len(g.CommandsNamed(">")) > before }) {
		t.Errorf("текст перестал повторяться через 0,05 с, отправок %d", before)
	}
	if hold, _ := printHold(PrintIn{}); hold != 3*time.Second {
		t.Errorf("удержание по умолчанию %v", hold)
	}
}

func TestPrintRejectsBadInput(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	for _, tt := range []struct {
		args map[string]any
		want string
	}{
		{map[string]any{"text": "hi", "color": "red"}, "цвет требует позиции"},
		{map[string]any{"text": "hi", "x": 1}, "x и y"},
		{map[string]any{"text": "hi", "y": 1}, "x и y"},
		{map[string]any{"text": "@hp@"}, "@"},
		{map[string]any{"text": "hi", "hold_seconds": 601}, "удержание"},
	} {
		wantToolError(t, call(t, cs, "stone_print", tt.args), tt.want)
	}
	if cmds := g.CommandsNamed(">"); len(cmds) != 0 {
		t.Errorf("игра получила %+v, текст не ожидался", cmds)
	}
}
