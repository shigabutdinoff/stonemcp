package engine

import (
	"errors"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func framed(cmds []fakegame.Command, frame int) []fakegame.Command {
	return slices.DeleteFunc(cmds, func(c fakegame.Command) bool { return c.Frame != frame })
}

func TestReadsDuringFreezeStayOnFrame(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen([]string{"12345678", "abcdefgh", "ABCDEFGH"})
	e := startEngine(t, g, Config{FrameBudget: 40})
	s := freezeNow(t, e, time.Minute)
	got, err := e.Evaluate(testContext(t), []Query{{Name: "time"}})
	if err != nil || got[0].String() != strconv.Itoa(s.Time) {
		t.Fatalf("чтение в заморозке: %v, %v, ожидался кадр %d", got, err, s.Time)
	}
	shot, err := e.Screenshot(testContext(t))
	if err != nil || shot.Text != "12345678\nabcdefgh\nABCDEFGH" || shot.FramesSpanned != 1 {
		t.Fatalf("экран в заморозке: %+v, %v", shot, err)
	}
	if now, err := e.Snapshot(); err != nil || now.Frame != s.Frame {
		t.Fatalf("снимок в заморозке: %+v, %v", now, err)
	}
	standsOn(t, g, s.Time, 50*time.Millisecond)
}

func TestCommandDuringFreezeAnswersAtOnce(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	if _, err := e.Do(testContext(t), EquipRight("shield", time.Minute), 0); err != nil {
		t.Fatal(err)
	}
	s := freezeNow(t, e, time.Minute)
	res, err := e.Do(testContext(t), EquipLeft("wand", 0), 0)
	if err != nil || res.Result != "cmd_done" || !res.Frozen || res.Snapshot == nil || res.Snapshot.Frame != s.Frame {
		t.Fatalf("команда в заморозке: %+v, %v", res, err)
	}
	left, right := framed(g.CommandsNamed("equipL"), s.Time), framed(g.CommandsNamed("equipR"), s.Time)
	if len(left) != 1 || left[0].Phase != "post" || len(right) != 1 {
		t.Fatalf("команды кадра %d: левая %+v, правая %+v", s.Time, left, right)
	}
	standsOn(t, g, s.Time, 50*time.Millisecond)
	e.Release()
	skipFrames(t, g, 2, "игра после отпускания")
	if n := len(g.CommandsNamed("equipL")); n != 1 {
		t.Errorf("однократная команда ушла %d раз", n)
	}
}

func TestCommandSentBeforeFreezeAnswersWithFrozenFrame(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{})
	frozen := freezeAsync(e)
	eventually(t, "запрос заморозки", func() bool { return freezeCallCount(e) == 1 })
	done := doAsync(testContext(t), e, EquipLeft("wand", 0), 0)
	eventually(t, "команда в очереди", func() bool { return queuedCount(e) == 1 })
	g.Resume()
	s := receive(t, frozen, "заморозка не наступила").snapshot
	out := receive(t, done, "команда ждёт отпускания")
	if s == nil || out.err != nil || !out.result.Frozen || out.result.Snapshot.Frame != s.Frame {
		t.Fatalf("команда кадра заморозки: %+v, %v, заморозка %+v", out.result, out.err, s)
	}
}

func TestDropDuringFreezeFailsNextRead(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	freezeNow(t, e, time.Minute)
	g.DropConn()
	if _, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}}); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("чтение после разрыва в заморозке: %v", err)
	}
}
