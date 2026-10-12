package engine

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestDoRunsInPostAndReturnsNextFrame(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	res, err := e.Do(testContext(t), EquipLeft("sword", 0), 0)
	if err != nil || res.Result != "cmd_done" || res.Snapshot == nil {
		t.Fatalf("результат команды = %+v, err = %v", res, err)
	}
	cmds := g.CommandsNamed("equipL")
	if len(cmds) != 1 || cmds[0].Phase != "post" || cmds[0].Arg != "sword" {
		t.Fatalf("команды = %+v", cmds)
	}
	if res.Snapshot.Frame != int64(cmds[0].Frame+1) {
		t.Fatalf("получен кадр %d после команды кадра %d", res.Snapshot.Frame, cmds[0].Frame)
	}
}

func TestDoReportsGameError(t *testing.T) {
	g := fakegame.Start(t)
	g.Reject("equipL", true)
	e := startEngine(t, g, Config{})
	g.PauseAfterCommand("equipL")
	res, err := e.Do(testContext(t), EquipLeft("sword", time.Second), 0)
	var gameErr *GameError
	if !errors.As(err, &gameErr) || gameErr.Code != "cmd_err" || gameErr.Command != "equipL sword" {
		t.Fatalf("ошибка игры = %v", err)
	}
	if res.Result != "cmd_err" || !strings.Contains(err.Error(), "equipL sword") {
		t.Fatalf("ответ отказа = %+v, err = %v", res, err)
	}
	g.Resume()
	skipFrames(t, g, 3, "следующие кадры")
	if len(g.CommandsNamed("equipL")) != 1 {
		t.Fatal("отвергнутая команда повторилась")
	}
}

func TestDoRejectsInvalidCommandsBeforeSending(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	commands := []Command{
		Equip("", 0), Equip("sword", -1), Equip("sword", 601*time.Second),
		{Name: "activate", Arg: "potion", Hold: time.Second},
		{Name: "brew", Arg: "stone", Hold: time.Second},
		Brew(nil), Brew([]string{"fire_elemental"}), Brew([]string{"empty"}),
		EquipLeft("@hp@", 0), EquipRight("@hp@", 0), Activate("@hp@"),
		{Name: "loadout", Arg: "@hp@"}, {Name: ">", Arg: "@hp@"},
		{Name: ">", Arg: "`abc,def,hi"}, {Name: ">", Arg: "`5,hi"},
		{Name: ">", Arg: "`5,5"},
		Brew([]string{"stone", "@hp@"}), {Name: "profile", Arg: "Mindstone.main"},
	}
	for _, cmd := range commands {
		if _, err := e.Do(testContext(t), cmd, 0); err == nil {
			t.Errorf("принята неверная команда %+v", cmd)
		}
	}
	if _, err := e.Do(testContext(t), Equip("sword", 0), -1); err == nil {
		t.Error("принято отрицательное ожидание")
	}
	if len(g.Commands()) != 0 {
		t.Fatal("неверные команды дошли до игры")
	}
}
