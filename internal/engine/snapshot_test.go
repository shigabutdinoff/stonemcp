package engine

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestSnapshotStateAndAge(t *testing.T) {
	values := map[string]mc.Value{
		"hp": "41", "maxhp": "80", "armor": "5", "pos.x": "17",
		"buffs.string": "vigor", "debuffs.string": "poison",
		"loc.id": "rocky_plateau", "foe": "scarab foe", "foe.id": "scarab",
		"foe.hp": "23", "foe.distance": "12", "ai.enabled": "True",
	}
	v := make([]mc.Value, len(baseVars))
	for i, name := range baseVars {
		v[i] = values[name]
	}
	now := time.Now()
	s := buildSnapshot(v, 7, now)
	if s.Player.HP != 41 || s.Player.MaxHP != 80 || s.Player.Armor != 5 || s.Player.X != 17 ||
		s.Player.Buffs != "vigor" || s.Player.Debuffs != "poison" || !s.AI.Enabled ||
		s.Loc.ID != "rocky_plateau" || s.Foe == nil || s.Foe.ID != "scarab" ||
		s.Foe.HP != 23 || s.Foe.Distance != 12 || s.Frame != 7 {
		t.Fatalf("неверное состояние: %+v", s)
	}
	copy := s.aged(now.Add(50 * time.Millisecond))
	copy.Foe.HP = 0
	if copy.AgeMS != 50 || s.AgeMS != 0 || s.Foe.HP != 23 {
		t.Fatalf("возраст или независимость снимка нарушены: %+v", copy)
	}
}

func TestSnapshotNoFoe(t *testing.T) {
	s := buildSnapshot(nil, 1, time.Now())
	b, err := json.Marshal(s)
	if err != nil || !strings.Contains(string(b), `"foe":null`) {
		t.Fatalf("без врага ожидался null: %s, %v", b, err)
	}
}

func TestTickingByFrameAge(t *testing.T) {
	e := New(Config{})
	if err := e.Ticking(time.Second); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("без связи err = %v, ожидалось %v", err, ErrDisconnected)
	}
	e.connected("0.3")
	if err := e.Ticking(time.Second); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("без кадра err = %v, ожидалось %v", err, ErrNotTicking)
	}
	e.snapshot = buildSnapshot(nil, 1, time.Now().Add(-time.Minute))
	if err := e.Ticking(time.Hour); err != nil {
		t.Fatalf("свежий кадр отвергнут: %v", err)
	}
	if err := e.Ticking(time.Second); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("старый кадр err = %v, ожидалось %v", err, ErrNotTicking)
	}
}
