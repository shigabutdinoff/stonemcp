package engine

import (
	"context"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func waitAsync(ctx context.Context, e *Engine, conditions []Condition, limit time.Duration) <-chan waitOutcome {
	done := make(chan waitOutcome, 1)
	go func() {
		res, err := e.wait(ctx, conditions, time.Second, limit)
		done <- waitOutcome{result: res, err: err}
	}()
	return done
}

func TestElementOfFoeText(t *testing.T) {
	for _, tt := range []struct{ text, id, name, want string }{
		{fakegame.ElementalistFoe("AEther"), "dysangelos_elementalist", "One With the Elements", "AEther"},
		{fakegame.ElementalistFoe("Ice"), "dysangelos_elementalist", "One With the Elements", "Ice"},
		{"dysangelos_bearer Dysangelos, Bearer of Stones boss unpushable phase1 ranged immune_to_stun", "dysangelos_bearer", "Dysangelos, Bearer of Stones", ""},
		{"golem Golem stone", "golem", "Golem", "Stone"},
		{"ice_pillar Ice Pillar Poison", "ice_pillar", "Ice Pillar", "Poison"},
		{"ice_pillar Ice Pillar", "ice_pillar", "Ice Pillar", ""},
		{"wasp Wasp immune_to_Fire", "wasp", "Wasp", ""},
		{"", "", "", ""},
	} {
		if got := elementOf(tt.text, tt.id, tt.name); got != tt.want {
			t.Errorf("стихия %q = %q, ожидалась %q", tt.text, got, tt.want)
		}
	}
}

func TestFoeElementChangedCondition(t *testing.T) {
	ice, aether := &Snapshot{Foe: &Foe{Element: "Ice"}}, &Snapshot{Foe: &Foe{Element: "AEther"}}
	none, plain := &Snapshot{}, &Snapshot{Foe: &Foe{}}
	c := Condition{Kind: CondFoeElementChanged}
	if err := c.validate(); err != nil {
		t.Fatalf("условие отвергнуто: %v", err)
	}
	for _, tt := range []struct {
		base, cur *Snapshot
		want      bool
	}{
		{ice, aether, true}, {none, ice, true}, {ice, ice, false},
		{ice, none, false}, {ice, plain, false}, {plain, plain, false},
	} {
		if got := c.met(tt.base, tt.cur, 0); got != tt.want {
			t.Errorf("смена стихии с %+v на %+v = %v, ожидалось %v", tt.base.Foe, tt.cur.Foe, got, tt.want)
		}
	}
}

func TestWaitForElementChangeKeepsGameRunning(t *testing.T) {
	g := fakegame.Start(t)
	g.SetElementalist("Ice")
	e := startEngine(t, g, Config{})
	done := waitAsync(testContext(t), e, []Condition{{Kind: CondFoeElementChanged}}, 0)
	eventually(t, "регистрация ожидания", func() bool { return waiterCount(e) == 1 })
	g.Set("foe", fakegame.ElementalistFoe("AEther"))
	out := receive(t, done, "смена стихии не замечена")
	if out.err != nil || out.result.Reason != CondFoeElementChanged || out.result.Snapshot.Foe.Element != "AEther" {
		t.Fatalf("смена стихии: %+v, %v", out.result, out.err)
	}
	skipFrames(t, g, 3, "игра идёт после ожидания без заморозки")
}
