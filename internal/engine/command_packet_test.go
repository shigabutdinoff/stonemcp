package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestPostSendsFrameInOnePacket(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{})
	var dones []<-chan commandOutcome
	for _, cmd := range []Command{EquipLeft("sword", 150*time.Millisecond), EquipRight("shield", 150*time.Millisecond)} {
		dones = append(dones, doAsync(testContext(t), e, cmd, 0))
	}
	eventually(t, "команды в очереди", func() bool { return queuedCount(e) == 2 })
	g.Resume()
	for _, done := range dones {
		if err := receive(t, done, "команда не завершилась").err; err != nil {
			t.Fatal(err)
		}
	}
	eventually(t, "повторы удержаний", func() bool { return len(g.CommandsNamed("equipL")) >= 3 })
	frames := map[int][]fakegame.Command{}
	for _, cmd := range g.Commands() {
		frames[cmd.Frame] = append(frames[cmd.Frame], cmd)
	}
	shared := 0
	for _, cmds := range frames {
		if len(cmds) < 2 {
			continue
		}
		shared++
		for _, cmd := range cmds {
			if cmd.Packet != cmds[0].Packet {
				t.Fatalf("кадр %d отправлен несколькими пакетами: %+v", cmd.Frame, cmds)
			}
		}
	}
	if shared < 2 {
		t.Fatalf("кадров с двумя командами = %d, команды = %+v", shared, g.Commands())
	}
}

func TestPostMixesHoldsAndCommandsInOnePacket(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	if _, err := e.Do(testContext(t), EquipLeft("sword", time.Second), 0); err != nil {
		t.Fatal(err)
	}
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	done := doAsync(testContext(t), e, Activate("potion"), 0)
	eventually(t, "команда в очереди", func() bool { return queuedCount(e) == 1 })
	frame := g.Frame() + 1
	g.Resume()
	if err := receive(t, done, "команда не завершилась").err; err != nil {
		t.Fatal(err)
	}
	var mixed []fakegame.Command
	for _, cmd := range g.Commands() {
		if cmd.Frame == frame {
			mixed = append(mixed, cmd)
		}
	}
	if len(mixed) != 2 || mixed[0].Name != "equipL" || mixed[1].Name != "activate" {
		t.Fatalf("кадр %d = %+v, ожидались повтор удержания и новая команда", frame, mixed)
	}
	if mixed[0].Packet != mixed[1].Packet {
		t.Fatalf("повтор удержания и команда ушли разными пакетами: %+v", mixed)
	}
}
