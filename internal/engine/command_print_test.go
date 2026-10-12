package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestCommandConstructorsSendArguments(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	text, err := Print("hi", &Point{X: 0, Y: 1}, "green", 0)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		cmd       Command
		name, arg string
	}{
		{Equip("bardiche", 0), "equip", "bardiche"},
		{EquipRight("shield", 0), "equipR", "shield"},
		{Loadout(2, 0), "loadout", "2"},
		{Activate("potion"), "activate", "potion"},
		{Brew([]string{"stone", "wood", "tar", "bronze"}), "brew", "stone + wood + tar + bronze"},
		{text, ">", "`0,1,#green,hi"},
	}
	for _, tt := range cases {
		if _, err := e.Do(testContext(t), tt.cmd, 0); err != nil {
			t.Fatal(err)
		}
		cmds := g.CommandsNamed(tt.name)
		if len(cmds) != 1 || cmds[0].Arg != tt.arg {
			t.Fatalf("команда %q = %+v, ожидался аргумент %q", tt.name, cmds, tt.arg)
		}
	}
}

func TestPrintCoordinates(t *testing.T) {
	at := &Point{X: 0, Y: 1}
	if _, err := Print("hi", nil, "green", 0); err == nil {
		t.Error("принят цвет без позиции текста")
	}
	if _, err := Print("`5,5,hi", nil, "", 0); err == nil {
		t.Error("принят текст с позицией без координат")
	}
	if _, err := Print("#1 Player", at, "", 0); err == nil {
		t.Error("принят текст с решёткой без цвета, игра примет его за цвет")
	}
	if cmd, err := Print("#1 Player", nil, "", 0); err != nil || cmd.Arg != "#1 Player" {
		t.Errorf("текст с решёткой без координат = %+v, err = %v", cmd, err)
	}
	for _, color := range []string{"red,99,99", "`9,9"} {
		if _, err := Print("hi", at, color, 0); err == nil {
			t.Errorf("принят цвет %q, ломающий разбор полей", color)
		}
	}
	if cmd, err := Print("`5,5,hi", at, "", 0); err != nil || cmd.Arg != "`0,1,`5,5,hi" {
		t.Errorf("текст с координатами = %+v, err = %v", cmd, err)
	}
	for _, tt := range []struct {
		at          *Point
		color, want string
	}{
		{nil, "", "hi"}, {at, "", "`0,1,hi"}, {at, "#green", "`0,1,#green,hi"},
	} {
		cmd, err := Print("hi", tt.at, tt.color, 0)
		if err != nil || cmd.Arg != tt.want {
			t.Errorf("текст = %+v, err = %v", cmd, err)
		}
	}
}

func TestPrintAcceptsBlankText(t *testing.T) {
	blank, err := Print(" ", nil, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := blank.validate(mc.SafePacket); err != nil {
		t.Fatalf("текст из пробелов отвергнут, err = %v", err)
	}
	empty, err := Print("", nil, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := empty.validate(mc.SafePacket); err == nil {
		t.Fatal("принят пустой текст")
	}
}

func TestEquipRejectsTruncatedArgument(t *testing.T) {
	for _, cmd := range []Command{Equip("sword//sharp", 0), EquipLeft("sword//sharp", 0), EquipRight("shield//round", time.Minute)} {
		if err := cmd.validate(0); err == nil {
			t.Errorf("принят аргумент %q, который игра обрежет", cmd.Arg)
		}
	}
	if err := Equip("sword", 0).validate(0); err != nil {
		t.Errorf("обычный предмет отвергнут, err = %v", err)
	}
}
