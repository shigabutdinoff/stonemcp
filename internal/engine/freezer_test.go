package engine

import (
	"testing"
	"time"
)

func TestFreezerWantWhileFrozen(t *testing.T) {
	f := freezer{frozen: &frozenState{}}
	if call := f.want(time.Minute); call != nil || len(f.calls) != 0 {
		t.Fatalf("просьба в заморозке завела вызов %+v, вызовов %d", call, len(f.calls))
	}
	f.release()
	if call := f.want(time.Minute); call == nil || len(f.calls) != 1 {
		t.Fatalf("просьба после отпускания = %+v, вызовов %d", call, len(f.calls))
	}
}

func TestExpireNoticeText(t *testing.T) {
	f := freezer{frozen: &frozenState{limit: 300 * time.Millisecond}, stepFrame: 7}
	f.expire()
	want := "игра продолжила сама: заморозка на кадре 7 снята по лимиту 0.3 с"
	if got := f.takeNotice(); got != want {
		t.Errorf("сообщение о лимите = %q, ожидалось %q", got, want)
	}
}
