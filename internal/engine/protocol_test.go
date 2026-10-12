package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestSplitResponses(t *testing.T) {
	g := fakegame.Start(t)
	g.SplitResponses(true)
	e := startEngine(t, g, Config{MaxPacket: 100})
	got, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}, {Name: "loc.id"}})
	if err != nil || len(got) != 2 || got[0] != "80" || got[1] != "rocky_plateau" {
		t.Fatalf("разделённый ответ: %v, %v", got, err)
	}
}

func TestRejectedPacketEndsSession(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{ReadTimeout: 30 * time.Millisecond, ReconnectDelays: []time.Duration{100 * time.Millisecond}})
	g.RejectPackets(true)
	eventually(t, "ошибка разбора пакета", func() bool { return strings.Contains(e.Status().LastError, "prt_err") })
	if _, err := e.Snapshot(); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("снимок повреждённой сессии: %v", err)
	}
	g.RejectPackets(false)
	eventually(t, "восстановление после отказа", func() bool { _, err := e.Snapshot(); return err == nil })
}

func TestCancelDuringExchangePreservesProtocol(t *testing.T) {
	g := fakegame.Start(t)
	g.HoldFunc("slow", "late")
	defer g.ReleaseFunc("slow")
	e := startEngine(t, g, Config{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { _, err := e.Evaluate(ctx, []Query{{Name: "slow"}}); done <- err }()
	eventually(t, "начало чтения", func() bool { return len(g.Calls("slow")) == 1 })
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("отмена во время обмена: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("отмена ждёт ответа игры")
	}
	g.ReleaseFunc("slow")
	got, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}})
	if err != nil || got[0] != "80" {
		t.Fatalf("обмен после отмены: %v, %v", got, err)
	}
}

func TestUndottedProtocolVersion(t *testing.T) {
	g := fakegame.Start(t)
	g.SetVersion("1")
	e := startEngine(t, g, Config{ReadTimeout: 10 * time.Millisecond})
	if e.Status().ProtocolVersion != "1" {
		t.Fatalf("версия не сохранена: %+v", e.Status())
	}
}

func TestSmallPacketScreenshotDoesNotDisconnect(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{MaxPacket: 40, ReconnectDelays: []time.Duration{time.Second}})
	_, err := e.Screenshot(testContext(t))
	if !errors.Is(err, mc.ErrTooLarge) {
		t.Fatalf("ожидалась ошибка слишком большого запроса клетки: %v", err)
	}
	if !e.Status().Connected {
		t.Fatalf("локальная ошибка запроса разорвала сессию: %v", err)
	}
}
