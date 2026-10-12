package engine

import (
	"context"
	"errors"
	"slices"
	"testing"
	"testing/synctest"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func heldSlots(e *Engine) []string {
	_, ready, _ := e.takeCommands()
	var slots []string
	for _, h := range ready {
		slots = append(slots, h.slot)
	}
	return slots
}

func TestFreezePausesHolds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := New(Config{})
		e.snapshot = &Snapshot{Frame: 1}
		check := func(when string, want ...string) {
			t.Helper()
			if got := heldSlots(e); !slices.Equal(got, want) {
				t.Fatalf("удержания %s = %v, ожидались %v", when, got, want)
			}
		}
		e.commands.holds.accept(EquipRight("shield", 3*time.Second), time.Now())
		time.Sleep(time.Second)
		freezeFor(e, 10*time.Second)
		job := newJob(context.Background(), EquipLeft("wand", 5*time.Second))
		e.applyResults(nil, []*commandJob{job}, []mc.Value{"cmd_done"})
		time.Sleep(time.Minute)
		check("в заморозке", "equipL", "equipR")
		e.finishFreeze()
		time.Sleep(time.Second)
		check("через 1 с после отпускания", "equipL", "equipR")
		time.Sleep(2 * time.Second)
		check("через 3 с после отпускания", "equipL")
		time.Sleep(2 * time.Second)
		check("через 5 с после отпускания")
	})
}

func TestTickingExcludesFreeze(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := frozenFor(time.Minute)
		e.finishFreeze()
		if err := e.Ticking(10 * time.Second); err != nil {
			t.Fatalf("кадр сразу после отпускания отвергнут: %v", err)
		}
		time.Sleep(11 * time.Second)
		if err := e.Ticking(10 * time.Second); !errors.Is(err, ErrNotTicking) {
			t.Fatalf("старый кадр после отпускания err = %v, ожидалось %v", err, ErrNotTicking)
		}
	})
}

func TestTickingDuringRelease(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := frozenFor(time.Minute)
		e.Release()
		if err := e.Ticking(10 * time.Second); err != nil {
			t.Fatalf("отпускание принято за остановку игры: %v", err)
		}
	})
}
