package engine

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestInvalidReadDoesNotBreakSession(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	for _, queries := range [][]Query{
		nil, {{Name: "equip"}}, {{Name: "hp\n"}}, make([]Query, 257),
		{{Name: "hp", Args: []mc.Arg{mc.Str(strings.Repeat("x", 9000))}}},
		{{Name: "hp", Guard: &Guard{Check: Query{Name: "hp"}}}},
	} {
		if _, err := e.Evaluate(testContext(t), queries); err == nil {
			t.Errorf("неверное чтение принято: %v", queries)
		}
	}
	if got, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}}); err != nil || got[0] != "80" {
		t.Fatalf("сессия повреждена: %v, %v", got, err)
	}
}

func TestInvalidWaitConditions(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	for _, condition := range []Condition{
		{Kind: "unknown"}, {Kind: CondFrames, Value: 1.5}, {Kind: CondFrames, Value: -1},
		{Kind: CondSeconds, Value: math.NaN()}, {Kind: CondSeconds, Value: math.Inf(1)},
		{Kind: CondSeconds, Value: 1e20}, {Kind: CondHPBelow, Value: 1.1},
	} {
		if _, err := e.WaitFor(testContext(t), []Condition{condition}, time.Second); err == nil {
			t.Errorf("неверное условие принято: %+v", condition)
		}
	}
}

func TestLocationAndFoeChanges(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	for _, change := range []struct {
		kind  string
		apply func()
	}{
		{CondLocChanged, func() { g.Set("loc.id", "deadwood") }},
		{CondFoeChanged, func() { g.SetFoe("wasp", 20) }},
	} {
		ctx := testContext(t)
		done := make(chan waitOutcome, 1)
		go func() {
			res, err := e.WaitFor(ctx, []Condition{{Kind: change.kind}}, time.Second)
			done <- waitOutcome{result: res, err: err}
		}()
		eventually(t, "регистрация условия", func() bool { return waiterCount(e) == 1 })
		change.apply()
		outcome := <-done
		if outcome.err != nil || outcome.result.Reason != change.kind {
			t.Fatalf("изменение не обнаружено: %+v", outcome)
		}
	}
}
