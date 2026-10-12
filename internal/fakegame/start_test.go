package fakegame

import (
	"strings"
	"testing"
	"time"
)

func TestGameAnswersThroughBinary(t *testing.T) {
	g := Start(t)
	g.SetFunc("probe", "7")
	got := exchange(t, g.Dial(t), strings.Join([]string{"1", "2", "0", "probe", "0", "ai.enabled"}, sep))
	if want := strings.Join([]string{"1", "7", "True"}, sep); got != want {
		t.Fatalf("ответ = %q, ожидалось %q", got, want)
	}
	if calls := g.Calls("probe"); len(calls) != 1 || calls[0].Frame != 1 {
		t.Errorf("calls = %+v, ожидался один вызов в кадре 1", calls)
	}
	if len(g.Commands()) != 0 || len(g.Violations()) != 0 {
		t.Errorf("команды %v и нарушения %v, ожидалось пусто", g.Commands(), g.Violations())
	}
}

func TestViolationsFailTheTestAtStop(t *testing.T) {
	for _, allow := range []bool{false, true} {
		rec := &recorder{TB: t}
		g := Start(rec)
		if allow {
			g.AllowViolations()
		}
		exchange(t, g.Dial(t), strings.Join([]string{"1", "1", "0", "hp"}, sep))
		g.stop()
		errs := rec.list()
		if failed := len(errs) == 1 && strings.Contains(errs[0], "клиент нарушил протокол"); failed == allow {
			t.Errorf("AllowViolations=%v: ошибки %q", allow, errs)
		}
	}
}

func TestLostGameIsReported(t *testing.T) {
	rec := &recorder{TB: t}
	g := Start(rec)
	if err := g.cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(rec.list()) == 0 {
		if time.Now().After(deadline) {
			t.Fatal("потеря игры не замечена")
		}
		g.Frame()
	}
	if errs := rec.list(); !strings.Contains(errs[0], "связь с игрой потеряна") {
		t.Errorf("ошибки = %q, ожидалась потеря связи", errs)
	}
}

func TestStopEndsGameWithPendingCommand(t *testing.T) {
	g := Start(t)
	g.PauseAfterCommand("equip")
	done := make(chan struct{})
	go func() { g.stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("игра не завершилась при незавершённой команде")
	}
	if state := g.cmd.ProcessState; state == nil || !state.Exited() {
		t.Errorf("состояние процесса игры = %v, ожидался обычный выход", state)
	}
}
