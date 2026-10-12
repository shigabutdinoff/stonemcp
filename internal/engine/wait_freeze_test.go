package engine

import (
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestWaitFreezeStopsOnEventFrame(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	done := waitAsync(testContext(t), e, []Condition{{Kind: CondFoePresent}}, time.Minute)
	eventually(t, "регистрация ожидания", func() bool { return waiterCount(e) == 1 })
	at := g.Frame() + 50
	g.SwitchAt("foe", at, "", "wasp Wasp")
	out := receive(t, done, "враг не замечен")
	if out.err != nil || out.result.Reason != CondFoePresent || out.result.Snapshot.Time != at {
		t.Fatalf("заморозка на появлении врага: %+v, %v, ожидался кадр %d", out.result, out.err, at)
	}
	standsOn(t, g, at, 100*time.Millisecond)
}

func TestWaitFreezeElementChangesInARow(t *testing.T) {
	g := fakegame.Start(t)
	g.SetElementalist("Ice")
	e := startEngine(t, g, Config{})
	before := "Ice"
	for _, element := range []string{"AEther", "Fire"} {
		done := waitAsync(testContext(t), e, []Condition{{Kind: CondFoeElementChanged}}, time.Minute)
		eventually(t, "регистрация ожидания", func() bool { return waiterCount(e) == 1 })
		at := g.Frame() + 50
		g.SwitchAt("foe", at, fakegame.ElementalistFoe(before), fakegame.ElementalistFoe(element))
		out := receive(t, done, "смена стихии не замечена")
		if out.err != nil || out.result.Reason != CondFoeElementChanged || out.result.Snapshot.Foe.Element != element || out.result.Snapshot.Time != at {
			t.Fatalf("смена на %s: %+v, %v, ожидался кадр %d", element, out.result, out.err, at)
		}
		standsOn(t, g, at, 50*time.Millisecond)
		before = element
	}
}

func TestWaitFreezeTimeoutAndLongLimitLeaveGameRunning(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	res, err := e.WaitFreeze(testContext(t), []Condition{{Kind: CondFoePresent}}, 50*time.Millisecond, time.Minute)
	if err != nil || res.Reason != ReasonTimeout {
		t.Fatalf("таймаут ожидания с заморозкой: %+v, %v", res, err)
	}
	skipFrames(t, g, 3, "игра идёт после таймаута")
	if _, err := e.WaitFreeze(testContext(t), []Condition{{Kind: CondFoePresent}}, time.Second, MaxFreeze+time.Second); err == nil {
		t.Fatal("лимит больше 2 минут принят")
	}
	skipFrames(t, g, 3, "игра идёт после отказа")
}

func TestWaitFreezeOnMetConditionAndWaitFromFreeze(t *testing.T) {
	g := fakegame.Start(t)
	g.SetFoe("wasp", 20)
	e := startEngine(t, g, Config{})
	res, err := e.WaitFreeze(testContext(t), []Condition{{Kind: CondFoePresent}}, time.Second, 0)
	if err != nil || res.Reason != CondFoePresent || res.Snapshot.Time != g.Frame() || frozenLimit(e) != DefaultFreeze {
		t.Fatalf("заморозка на выполненном условии: %+v, %v, лимит %v", res, err, frozenLimit(e))
	}
	again, err := e.WaitFor(testContext(t), []Condition{{Kind: CondFoePresent}}, time.Second)
	if err != nil || again.Snapshot.Frame != res.Snapshot.Frame {
		t.Fatalf("выполненное условие в заморозке: %+v, %v", again, err)
	}
	standsOn(t, g, res.Snapshot.Time, 50*time.Millisecond)
	next, err := e.WaitFor(testContext(t), []Condition{{Kind: CondFrames, Value: 2}}, time.Second)
	if err != nil || next.Snapshot.Frame < res.Snapshot.Frame+2 {
		t.Fatalf("ожидание из заморозки не отпустило игру: %+v, %v", next, err)
	}
}

func TestFreezeAndWaitShareFrame(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{})
	g.SetFoe("wasp", 20)
	frozen := freezeAsync(e)
	done := waitAsync(testContext(t), e, []Condition{{Kind: CondFoePresent}}, MaxFreeze)
	eventually(t, "оба запроса", func() bool { return freezeCallCount(e) == 1 && waiterCount(e) == 1 })
	g.Resume()
	s := receive(t, frozen, "заморозка не наступила").snapshot
	out := receive(t, done, "ожидание не завершилось")
	if s == nil || out.err != nil || out.result.Snapshot.Frame != s.Frame || frozenLimit(e) != MaxFreeze {
		t.Fatalf("общая заморозка: %+v, %v, заморозка %+v, лимит %v", out.result, out.err, s, frozenLimit(e))
	}
}
