package engine

import (
	"testing"
	"time"
)

func TestConflictsFollowCommandHands(t *testing.T) {
	text, err := Print("готов", nil, "", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	placed, err := Print("готов", &Point{X: 0, Y: 1}, "", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		a, b Command
		want bool
	}{
		{"same_hand", EquipLeft("sword", 0), EquipLeft("axe", 0), true},
		{"both_hands", Equip("bardiche", 0), EquipRight("shield", 0), true},
		{"loadout", Loadout(2, 0), EquipLeft("sword", 0), true},
		{"other_hand", EquipLeft("sword", 0), EquipRight("shield", 0), false},
		{"same_text", text, text, true},
		{"text_positions", text, placed, false},
		{"text_hand", placed, EquipLeft("sword", 0), false},
		{"activate", Activate("potion"), Equip("bardiche", 0), false},
	} {
		if got := slotsConflict(tt.a, tt.a.slot(), tt.b, tt.b.slot()); got != tt.want {
			t.Errorf("%s: конфликт = %v, ожидалось %v", tt.name, got, tt.want)
		}
	}
}
