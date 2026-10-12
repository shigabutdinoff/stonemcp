package engine

import (
	"context"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestActivationWaitsForLongAnimation(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{CommandTTL: 30 * time.Millisecond, CommandTimeout: 1500 * time.Millisecond})
	paused := g.PauseAfterCommand("activate")
	done := doAsync(context.Background(), e, Activate("potion"), 0)
	receive(t, paused, "не дождались паузы после активации")
	expectPending(t, done, 750*time.Millisecond, "ответ до конца анимации")
	if !e.Status().Connected {
		t.Fatal("анимация разорвала соединение")
	}
	g.Set("hp", "63")
	g.Resume()
	got := receive(t, done, "нет результата после анимации")
	cmds := g.CommandsNamed("activate")
	if got.err != nil || got.result.Snapshot == nil {
		t.Fatalf("результат анимации = %+v", got)
	}
	if len(cmds) != 1 || got.result.Snapshot.Frame != int64(cmds[0].Frame+1) || got.result.Snapshot.Player.HP != 63 {
		t.Fatalf("неверный первый кадр после анимации = %+v, команды = %+v", got.result.Snapshot, cmds)
	}
}

func readWithin(t *testing.T, e *Engine, why string) *Snapshot {
	t.Helper()
	read := make(chan error, 1)
	var snapshot *Snapshot
	go func() {
		var err error
		snapshot, err = e.Snapshot()
		e.Status()
		read <- err
	}()
	select {
	case err := <-read:
		if err != nil || snapshot == nil {
			t.Fatalf("сохранённое состояние недоступно, err = %v", err)
		}
		return snapshot
	case <-time.After(time.Second):
		t.Fatal(why)
		return nil
	}
}

func TestRequestedWaitExtendsProtectionWithoutBlockingReaders(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{CommandTimeout: 40 * time.Millisecond})
	paused := g.PauseAfterCommand("activate")
	done := doAsync(testContext(t), e, Activate("potion"), time.Second)
	receive(t, paused, "не дождались паузы после активации")
	read := make(chan error, 1)
	go func() { _, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}}); read <- err }()
	before := readWithin(t, e, "ожидание команды заблокировало данные")
	expectPending(t, done, 100*time.Millisecond, "короткая защита прервала запрос")
	after := readWithin(t, e, "ожидание команды заблокировало сохранённое состояние")
	if after.Frame != before.Frame || after.AgeMS <= before.AgeMS {
		t.Fatalf("состояние без кадров не устарело: до %+v, после %+v", before, after)
	}
	select {
	case err := <-read:
		t.Fatalf("чтение завершилось без кадров, err = %v", err)
	default:
	}
	g.Resume()
	if got := <-done; got.err != nil {
		t.Fatal(got.err)
	}
	if err := <-read; err != nil {
		t.Fatal(err)
	}
	if len(g.CommandsNamed("activate")) != 1 {
		t.Fatal("команда повторилась при ожидании")
	}
}

func TestDefaultProtectionCoversLongestAnimation(t *testing.T) {
	const longestAnimation = 15 * time.Second
	if got := New(Config{}).cfg.CommandTimeout; got < longestAnimation {
		t.Fatalf("защита по умолчанию = %v, короче анимации %v", got, longestAnimation)
	}
}

func TestStalledGameDoesNotSpendCommandTTL(t *testing.T) {
	e, g := pausedCommandEngine(t, Config{ReadTimeout: 100 * time.Millisecond, CommandTTL: 100 * time.Millisecond, CommandTimeout: 3 * time.Second})
	done := doAsync(testContext(t), e, Activate("potion"), 0)
	eventually(t, "команда в очереди", func() bool { return queuedCount(e) == 1 })
	time.Sleep(400 * time.Millisecond)
	g.Resume()
	if got := receive(t, done, "нет результата после остановки игры"); got.err != nil || got.result.Result != "cmd_done" {
		t.Fatalf("результат после остановки игры = %+v", got)
	}
	if len(g.CommandsNamed("activate")) != 1 {
		t.Fatalf("команд activate = %d, ожидалась одна", len(g.CommandsNamed("activate")))
	}
}
