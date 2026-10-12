package engine

import (
	"strconv"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestFreezeLimitReleasesGameWithNotice(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	s := freezeNow(t, e, 300*time.Millisecond)
	standsOn(t, g, s.Time, 100*time.Millisecond)
	if notice := e.TakeNotice(); notice != "" {
		t.Fatalf("сообщение до лимита: %q", notice)
	}
	eventually(t, "игра продолжила по лимиту", func() bool { return g.Frame() > s.Time })
	notice := e.TakeNotice()
	if !strings.Contains(notice, "по лимиту 0.3 с") || !strings.Contains(notice, strconv.FormatInt(s.Frame, 10)) {
		t.Fatalf("сообщение о лимите: %q", notice)
	}
	if again := e.TakeNotice(); again != "" {
		t.Fatalf("сообщение не сброшено: %q", again)
	}
}

func TestRepeatedFreezeKeepsFrameAndLimit(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	first := freezeNow(t, e, 300*time.Millisecond)
	second := freezeNow(t, e, MaxFreeze)
	if second.Frame != first.Frame || frozenLimit(e) != 300*time.Millisecond {
		t.Fatalf("повторная заморозка: кадр %d вместо %d, лимит %v", second.Frame, first.Frame, frozenLimit(e))
	}
	eventually(t, "лимит первой заморозки", func() bool { return g.Frame() > first.Time })
}

func TestFreezeLimitBoundsAndDefault(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{})
	for _, limit := range []time.Duration{-time.Second, MaxFreeze + time.Millisecond} {
		if _, err := e.Freeze(testContext(t), limit); err == nil || !strings.Contains(err.Error(), "120 секунд") {
			t.Errorf("лимит %v: %v", limit, err)
		}
	}
	skipFrames(t, g, 3, "игра идёт после отказа в заморозке")
	freezeNow(t, e, 0)
	if got := frozenLimit(e); got != DefaultFreeze {
		t.Errorf("лимит по умолчанию %v, ожидался %v", got, DefaultFreeze)
	}
}

func TestFPSExcludesFreeze(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := ranFor(t, time.Second)
		freezeFor(e, time.Minute)
		e.finishFreeze()
		wantFPS20(t, e, "с заморозкой")
	})
}

func TestExpireFreezeSkipsNoticeAfterRelease(t *testing.T) {
	e := New(Config{})
	e.freezer.frozen = &frozenState{limit: time.Second, released: true}
	e.expireFreeze()
	if notice := e.TakeNotice(); notice != "" {
		t.Errorf("сообщение о лимите после отпускания = %q, ожидалась пустая строка", notice)
	}
}
