package engine

import (
	"context"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestWaitForFoeAndImmediateCondition(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	done := make(chan WaitResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := e.WaitFor(testContext(t), []Condition{{Kind: CondFoePresent}, {Kind: CondLocChanged}}, time.Second)
		done <- res
		errCh <- err
	}()
	eventually(t, "регистрация ожидания", func() bool { return waiterCount(e) == 1 })
	g.SetFoe("wasp", 20)
	res := <-done
	if err := <-errCh; err != nil || res.Reason != CondFoePresent || res.Snapshot.Foe.ID != "wasp" {
		t.Fatalf("ожидание врага: %+v, %v", res, err)
	}
	g.Pause()
	res, err := e.WaitFor(testContext(t), []Condition{{Kind: CondFoePresent}}, time.Second)
	if err != nil || res.Reason != CondFoePresent {
		t.Fatalf("выполненное условие: %+v, %v", res, err)
	}
}

func TestWaitingDoesNotStopFrames(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	before := g.Frame()
	res, err := e.WaitFor(testContext(t), []Condition{{Kind: CondLocChanged}}, 200*time.Millisecond)
	if err != nil || res.Reason != ReasonTimeout || g.Frame()-before < 20 || res.Snapshot.AgeMS > 100 {
		t.Fatalf("ожидание задержало игру: кадров=%d, результат=%+v, err=%v", g.Frame()-before, res, err)
	}
}

func TestWaitFramesSecondsAndHP(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	before, _ := e.Snapshot()
	res, err := e.WaitFor(testContext(t), []Condition{{Kind: CondFrames, Value: 5}}, time.Second)
	if err != nil || res.Reason != CondFrames || res.Snapshot.Frame < before.Frame+5 {
		t.Fatalf("ожидание кадров: %+v, %v", res, err)
	}
	g.Pause()
	start := time.Now()
	res, err = e.WaitFor(testContext(t), []Condition{{Kind: CondSeconds, Value: 0.03}}, time.Second)
	if err != nil || res.Reason != CondSeconds || time.Since(start) < 30*time.Millisecond {
		t.Fatalf("ожидание секунд: %+v, %v", res, err)
	}
	g.Set("hp", "39")
	g.Resume()
	res, err = e.WaitFor(testContext(t), []Condition{{Kind: CondHPBelow, Value: 0.5}}, time.Second)
	if err != nil || res.Reason != CondHPBelow || res.Snapshot.Player.HP != 39 {
		t.Fatalf("ожидание здоровья: %+v, %v", res, err)
	}
}

func TestWaitTimeoutAndCancellation(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	res, err := e.WaitFor(testContext(t), []Condition{{Kind: CondLocChanged}}, 40*time.Millisecond)
	if err != nil || res.Reason != ReasonTimeout || res.Snapshot.Loc.ID != "rocky_plateau" {
		t.Fatalf("таймаут: %+v, %v", res, err)
	}
	if _, err := e.WaitFor(canceledContext(t), []Condition{{Kind: CondLocChanged}}, time.Second); err != context.Canceled {
		t.Fatalf("отмена ожидания: %v", err)
	}
}
