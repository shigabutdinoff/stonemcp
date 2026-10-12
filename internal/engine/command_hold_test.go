package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestHoldRepeatsOncePerFrameAndExpires(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	started := time.Now()
	if _, err := e.Do(testContext(t), EquipRight("shield", 150*time.Millisecond), 0); err != nil {
		t.Fatal(err)
	}
	if time.Since(started) >= 150*time.Millisecond {
		t.Fatal("ответ ждал конца удержания")
	}
	eventually(t, "повторы удержания", func() bool { return len(g.CommandsNamed("equipR")) >= 3 })
	time.Sleep(180 * time.Millisecond)
	before := len(g.CommandsNamed("equipR"))
	skipFrames(t, g, 3, "кадры после истечения")
	cmds := g.CommandsNamed("equipR")
	if len(cmds) != before {
		t.Fatal("удержание продолжилось после срока")
	}
	for i, cmd := range cmds {
		if cmd.Phase != "post" || (i > 0 && cmd.Frame != cmds[i-1].Frame+1) {
			t.Fatalf("неверный повтор в кадре %+v", cmd)
		}
	}
}

func TestHoldReplacement(t *testing.T) {
	for _, tt := range []struct {
		name     string
		reject   bool
		wantSame int
	}{
		{"accepted", false, 0},
		{"rejected", true, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := fakegame.Start(t)
			e := startEngine(t, g, Config{})
			if _, err := e.Do(testContext(t), EquipRight("shield", time.Second), 0); err != nil {
				t.Fatal(err)
			}
			if tt.reject {
				g.Reject("equip", true)
			}
			_, err := e.Do(testContext(t), Equip("bardiche", 0), 0)
			if (err != nil) != tt.reject {
				t.Fatalf("err = %v", err)
			}
			frame := g.CommandsNamed("equip")[0].Frame
			eventually(t, "кадры после замены", func() bool { return g.Frame() > frame+4 })
			same, after := 0, 0
			for _, cmd := range g.CommandsNamed("equipR") {
				switch {
				case cmd.Frame == frame:
					same++
				case cmd.Frame > frame:
					after++
				}
			}
			if same != tt.wantSame {
				t.Fatalf("повторов удержания в кадре замены = %d, ожидалось %d", same, tt.wantSame)
			}
			if (after > 0) != tt.reject {
				t.Fatalf("повторов после замены = %d", after)
			}
			if len(g.CommandsNamed("equip")) != 1 {
				t.Fatal("замена отправлена повторно")
			}
		})
	}
}

func TestHoldEvictionSlots(t *testing.T) {
	for _, tt := range []struct {
		name      string
		old, next Command
		keep      bool
	}{
		{"same", EquipLeft("sword", time.Second), EquipLeft("axe", 0), false},
		{"both_left", Equip("bardiche", time.Second), EquipLeft("sword", 0), false},
		{"both_right", Equip("bardiche", time.Second), EquipRight("shield", 0), false},
		{"left_both", EquipLeft("sword", time.Second), Equip("bardiche", 0), false},
		{"other_hand", EquipLeft("sword", time.Second), EquipRight("shield", 0), true},
		{"loadout_both", Loadout(2, time.Second), Equip("bardiche", 0), false},
		{"loadout_hand", Loadout(2, time.Second), EquipRight("shield", 0), false},
		{"hand_loadout", EquipLeft("sword", time.Second), Loadout(2, 0), false},
		{"activate", EquipLeft("sword", time.Second), Activate("potion"), true},
		{"brew", EquipLeft("sword", time.Second), Brew([]string{"stone"}), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := fakegame.Start(t)
			e := startEngine(t, g, Config{})
			if _, err := e.Do(testContext(t), tt.old, 0); err != nil {
				t.Fatal(err)
			}
			if _, err := e.Do(testContext(t), tt.next, 0); err != nil {
				t.Fatal(err)
			}
			frame := skipFrames(t, g, 3, "кадры после команды")
			found := false
			for _, cmd := range g.Commands() {
				if cmd.Frame > frame && cmd.Name == tt.old.Name && cmd.Arg == tt.old.Arg {
					found = true
				}
			}
			if found != tt.keep {
				t.Fatalf("сохранение удержания = %v, ожидалось %v", found, tt.keep)
			}
		})
	}
}
