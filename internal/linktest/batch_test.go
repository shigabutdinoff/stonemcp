package linktest

import (
	"strings"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	"github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestRoundTripSplitBatches(t *testing.T) {
	const w, h = 90, 40
	lines := make([]string, h)
	for y := range lines {
		row := make([]byte, w)
		for x := range row {
			row[x] = byte('a' + (x+y)%26)
		}
		lines[y] = string(row)
	}
	g := fakegame.Start(t)
	g.SetScreen(lines)
	conn := g.Dial(t)
	signal(t, conn)

	var reqs []mindconnect.Request
	for y := range h {
		for x := range w {
			reqs = append(reqs, mindconnect.Var("draw.GetSymbol", mindconnect.Int(x), mindconnect.Int(y)))
		}
	}
	groups, err := mindconnect.SplitBatches(reqs, mindconnect.SafePacket)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) < 2 {
		t.Fatalf("получено групп: %d, ожидалось, что экран разобьётся на несколько пакетов", len(groups))
	}

	seen := 0
	for i, group := range groups {
		values := ask(t, conn, i+1, group)
		if len(values) != len(group)+1 {
			t.Fatalf("на группу из %d запросов пришло %d значений", len(group), len(values))
		}
		for i := range group {
			x, y := seen%w, seen/w
			if want := string(lines[y][x]); values[i].String() != want {
				t.Fatalf("символ в %d,%d = %q, ожидалось %q", x, y, values[i], want)
			}
			seen++
		}
	}
	if seen != len(reqs) {
		t.Errorf("прочитано символов: %d, ожидалось %d", seen, len(reqs))
	}
}

func TestRoundTripLongCommand(t *testing.T) {
	g := fakegame.Start(t)
	conn := g.Dial(t)
	signal(t, conn)

	const size = 8165
	long := mindconnect.Cmd(">", strings.Repeat("x", size))
	groups, err := mindconnect.SplitBatches([]mindconnect.Request{long}, mindconnect.SafePacket)
	if err != nil {
		t.Fatalf("SplitBatches = %v, ожидалась команда, которая ещё помещается в пакет", err)
	}
	if len(groups) != 1 {
		t.Fatalf("groups = %d, ожидалось 1", len(groups))
	}
	values := ask(t, conn, mindconnect.MaxID, groups[0])
	if values[0] != mindconnect.CmdDone {
		t.Errorf("values = %q, ожидалось, что команда выполнена", values)
	}
	if got := g.CommandsNamed(">"); len(got) != 1 || len(got[0].Arg) != size {
		t.Errorf("commands = %d, ожидалась одна с целым аргументом", len(got))
	}
}
