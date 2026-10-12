package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestCommandsInSameFrameReplaceHoldsInOrder(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{})
	var dones []<-chan commandOutcome
	for i, cmd := range []Command{EquipRight("shield", time.Minute), Equip("bardiche", 0)} {
		dones = append(dones, doAsync(testContext(t), e, cmd, 0))
		eventually(t, "команда в очереди", func() bool { return queuedCount(e) == i+1 })
	}
	g.Resume()
	for _, done := range dones {
		if err := (<-done).err; err != nil {
			t.Fatal(err)
		}
	}
	skipFrames(t, g, 3, "последующие кадры")
	cmds := g.Commands()
	if len(cmds) != 2 || cmds[0].Name != "equipR" || cmds[1].Name != "equip" || cmds[0].Frame != cmds[1].Frame {
		t.Fatalf("нарушен порядок команд или сохранено вытесненное удержание: %+v", cmds)
	}
}

func TestRejectedRepeatStopsHold(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	if _, err := e.Do(testContext(t), Loadout(2, time.Minute), 0); err != nil {
		t.Fatal(err)
	}
	g.Reject("loadout", true)
	skipFrames(t, g, 3, "отказ повтора")
	before := len(g.CommandsNamed("loadout"))
	skipFrames(t, g, 3, "кадры после отказа")
	if len(g.CommandsNamed("loadout")) != before {
		t.Fatal("отвергнутое удержание продолжилось")
	}
}

func TestDisconnectDoesNotReplayHolds(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	if _, err := e.Do(testContext(t), EquipLeft("sword", time.Minute), 0); err != nil {
		t.Fatal(err)
	}
	if err := g.PauseAndWait(testContext(t)); err != nil {
		t.Fatal(err)
	}
	g.DropConn()
	eventually(t, "разрыв связи", func() bool { return !e.Status().Connected })
	before := len(g.Commands())
	g.Resume()
	eventually(t, "восстановление связи", func() bool { return e.Status().Connected })
	skipFrames(t, g, 3, "кадры после восстановления")
	if len(g.Commands()) != before {
		t.Fatal("удержание пережило разрыв связи")
	}
}

func TestDoWithoutSnapshotOrWithCanceledContext(t *testing.T) {
	e := New(Config{})
	if _, err := e.Do(context.Background(), Activate("potion"), 0); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("err = %v", err)
	}
	e.connected("0.3")
	if _, err := e.Do(context.Background(), Activate("potion"), 0); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("err = %v", err)
	}
	if _, err := e.Do(canceledContext(t), Activate("potion"), 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	if _, err := e.Do(expiredContext(t), Activate("potion"), 0); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("err = %v", err)
	}
}

func TestDisconnectReturnsDeliveredResponse(t *testing.T) {
	e := New(Config{})
	job := sentJob(Activate("potion"))
	e.commands.jobs = []*commandJob{job}
	e.disconnect(ErrDisconnected)
	outcome := <-job.done
	if !errors.Is(outcome.err, ErrDisconnected) {
		t.Fatalf("ошибка после обрыва = %v, ожидалось %v", outcome.err, ErrDisconnected)
	}
	if outcome.result.Result != "cmd_done" || outcome.result.Snapshot != nil {
		t.Fatalf("результат после обрыва = %+v, ожидался ответ игры без кадра", outcome.result)
	}
}
