package engine

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func frozenPipe(e *Engine, limit time.Duration) (*session, net.Conn) {
	conn, game := net.Pipe()
	e.snapshot = &Snapshot{Frame: 1}
	e.freezer.want(limit)
	return &session{conn: conn, buf: make([]byte, 1), cfg: e.cfg}, game
}

func TestDropDuringFreezeReconnects(t *testing.T) {
	g := fakegame.Start(t)
	cfg := Config{ProbeInterval: 20 * time.Millisecond, ReconnectDelays: []time.Duration{300 * time.Millisecond}}
	e := startEngine(t, g, cfg)
	s := freezeNow(t, e, time.Minute)
	standsOn(t, g, s.Time, 100*time.Millisecond)
	g.DropConn()
	eventually(t, "обрыв в заморозке", func() bool {
		_, err := e.Snapshot()
		return errors.Is(err, ErrDisconnected)
	})
	eventually(t, "переподключение после обрыва", func() bool {
		cur, err := e.Snapshot()
		return err == nil && cur.Frame > s.Frame
	})
}

func TestFPSExcludesFreezeLostWithGame(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := ranFor(t, time.Second)
		s, game := frozenPipe(e, time.Minute)
		game.Close()
		if err := e.freezeStep(context.Background(), s); !errors.Is(err, io.EOF) || !strings.HasSuffix(err.Error(), ": проверка связи в заморозке") {
			t.Fatalf("заморозка без игры вернула %v, ожидался обёрнутый io.EOF", err)
		}
		wantFPS20(t, e, "после обрыва в заморозке")
	})
}

func TestFreezeLimitWithoutGameGivesNoNotice(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := New(Config{ProbeInterval: time.Hour})
		s, game := frozenPipe(e, time.Second)
		game.Close()
		if err := e.freezeStep(context.Background(), s); !errors.Is(err, io.EOF) || !strings.HasSuffix(err.Error(), ": проверка связи в заморозке") {
			t.Fatalf("лимит без игры вернул %v, ожидался обёрнутый io.EOF", err)
		}
		if notice := e.TakeNotice(); notice != "" {
			t.Errorf("сообщение о лимите без игры = %q", notice)
		}
	})
}

func TestDataDuringFreezeBreaksSession(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := New(Config{})
		s, game := frozenPipe(e, time.Minute)
		go game.Write([]byte("x"))
		if err := e.freezeStep(context.Background(), s); err == nil || !strings.Contains(err.Error(), "данные") {
			t.Fatalf("данные от игры в заморозке: %v", err)
		}
	})
}

func TestNegativeProbeIntervalRejected(t *testing.T) {
	if err := (Config{ProbeInterval: -time.Second}).defaults().validate(); err == nil {
		t.Fatal("принят отрицательный период проверки связи")
	}
}
