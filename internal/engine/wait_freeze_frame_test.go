package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func pausedWithFoe(t *testing.T) (*Engine, *fakegame.Game) {
	t.Helper()
	g := fakegame.Start(t)
	g.SetFoe("wasp", 20)
	return pausedEngine(t, g, Config{}), g
}

func TestWaitFreezeAfterStepWaitsForEventAgain(t *testing.T) {
	e, g := pausedWithFoe(t)
	at := g.Frame() + 20
	g.SwitchAt("foe", at, "", "wasp Wasp")
	done := waitAsync(testContext(t), e, []Condition{{Kind: CondFoePresent}}, time.Minute)
	eventually(t, "ожидание без заморозки", func() bool { return waiterCount(e) == 1 && freezeCallCount(e) == 0 })
	g.Resume()
	out := receive(t, done, "враг не вернулся")
	if out.err != nil || out.result.Reason != CondFoePresent || out.result.Snapshot.Foe == nil || out.result.Snapshot.Time != at {
		t.Fatalf("заморозка после исчезнувшего врага: %+v, %v, ожидался кадр %d", out.result, out.err, at)
	}
}

func TestWaitFreezeTimeoutAfterStepLeavesGameRunning(t *testing.T) {
	e, g := pausedWithFoe(t)
	res, err := e.WaitFreeze(testContext(t), []Condition{{Kind: CondFoePresent}}, 0, time.Minute)
	if err != nil || res.Reason != ReasonTimeout || freezeCallCount(e) != 0 {
		t.Fatalf("таймаут после шага заморозки: %+v, %v, вызовов заморозки %d", res, err, freezeCallCount(e))
	}
	g.Resume()
	skipFrames(t, g, 3, "игра идёт после таймаута")
}
