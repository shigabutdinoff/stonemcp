package engine

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestEngineSnapshotTracksGame(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	g.Set("hp", "41")
	g.SetFoe("scarab", 23)
	eventually(t, "состояние забега", func() bool {
		s, err := e.Snapshot()
		return err == nil && s.Player.HP == 41 && s.Foe != nil && s.Foe.ID == "scarab" && s.Foe.HP == 23 && s.Foe.Distance == 12
	})
	g.SetFoe("", 0)
	eventually(t, "исчезновение врага", func() bool { s, _ := e.Snapshot(); return s.Foe == nil })
	if s := e.Status(); !s.Connected || s.LocID != "rocky_plateau" || s.ProtocolVersion != "0.3" {
		t.Fatalf("неверный статус: %+v", s)
	}
}

func TestEvaluateSameFrameAndErrors(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFuncArgs("item.GetCooldown", []string{"bardiche"}, "42")
	e := startEngine(t, g, Config{MaxPacket: 100})
	queries := []Query{{Name: "item.GetCooldown", Args: []mc.Arg{mc.Str("bardiche")}}, {Name: "loc.id"}, {Name: "missing"}, {Name: "time"}}
	for range 20 {
		queries = append(queries, Query{Name: "time"})
	}
	got, err := e.Evaluate(testContext(t), queries)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(got[:3], []mc.Value{"42", "rocky_plateau", "fnc_err3"}) {
		t.Fatalf("значения: %v", got)
	}
	for _, value := range got[3:] {
		if value != got[3] {
			t.Fatalf("значения разных кадров: %v", got)
		}
	}
}

func TestEvaluateExceedsBudgetInOneFrame(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{FrameBudget: 35, MaxPacket: 100})
	queries := make([]Query, 80)
	for i := range queries {
		queries[i] = Query{Name: "time"}
	}
	got, err := e.Evaluate(testContext(t), queries)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range got {
		if value != got[0] {
			t.Fatalf("задание разделено между кадрами: %v", got)
		}
	}
	n, _ := strconv.Atoi(string(got[0]))
	eventually(t, "продолжение кадров", func() bool { return g.Frame() > n+2 })
}

func TestPauseKeepsStateAndConnection(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	before, _ := e.Snapshot()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := e.Evaluate(ctx, []Query{{Name: "hp"}})
	after, snapshotErr := e.Snapshot()
	if !errors.Is(err, ErrNotTicking) || snapshotErr != nil || !e.Status().Connected || after.AgeMS <= before.AgeMS {
		t.Fatalf("пауза: err=%v, снимок=%v, статус=%+v", err, snapshotErr, e.Status())
	}
	g.Resume()
	eventually(t, "возобновление", func() bool { s, _ := e.Snapshot(); return s.Frame > after.Frame })
}
