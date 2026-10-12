package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestPrintHoldsHaveIndependentPositions(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	at := &Point{X: 0, Y: 1}
	for _, pos := range []*Point{nil, at} {
		cmd, err := Print("old", pos, "", time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := e.Do(testContext(t), cmd, 0); err != nil {
			t.Fatal(err)
		}
	}
	cmd, err := Print("new", at, "green", 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.Do(testContext(t), cmd, 0); err != nil {
		t.Fatal(err)
	}
	frame := skipFrames(t, g, 3, "кадры после смены текста")
	defaultKept := false
	for _, c := range g.CommandsNamed(">") {
		if c.Frame <= frame {
			continue
		}
		if c.Arg != "old" {
			t.Fatalf("повтор вытесненного текста %q", c.Arg)
		}
		defaultKept = true
	}
	if !defaultKept {
		t.Fatal("текст другой позиции перестал удерживаться")
	}
}

func TestPrintHoldSlotIgnoresCoordinateSpelling(t *testing.T) {
	padded := Command{Name: ">", Arg: "`01,02,old", Hold: time.Second}
	plain := Command{Name: ">", Arg: "`1,2,new", Hold: time.Second}
	if padded.slot() != plain.slot() {
		t.Fatalf("слоты позиции = %q и %q", padded.slot(), plain.slot())
	}
	holds := holdTable{}
	holds.accept(padded, time.Now())
	holds.accept(plain, time.Now())
	if len(holds) != 1 {
		t.Fatalf("удержаний = %d, ожидалось 1", len(holds))
	}
	if held := holds[plain.slot()]; held.cmd.Arg != plain.Arg {
		t.Fatalf("удержан текст %q", held.cmd.Arg)
	}
}
