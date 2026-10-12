package engine

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestDisconnectFailsWaitAndReconnects(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{ReconnectDelays: []time.Duration{100 * time.Millisecond}})
	ctx := testContext(t)
	done := make(chan error, 1)
	go func() { _, err := e.WaitFor(ctx, []Condition{{Kind: CondLocChanged}}, time.Second); done <- err }()
	eventually(t, "регистрация ожидания", func() bool { return waiterCount(e) == 1 })
	g.DropConn()
	if err := <-done; !errors.Is(err, ErrDisconnected) {
		t.Fatalf("разрыв ожидания: %v", err)
	}
	if _, err := e.Snapshot(); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("снимок после разрыва: %v", err)
	}
	eventually(t, "повторное подключение", func() bool { _, err := e.Snapshot(); return err == nil })
}

func TestAbsentGameReturnsImmediately(t *testing.T) {
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := listener.Addr().String()
	listener.Close()
	e, _ := launchEngine(t, Config{Addr: addr})
	eventually(t, "причина подключения", func() bool { return e.Status().LastError != "" })
	if _, err := e.Snapshot(); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("состояние без игры: %v", err)
	}
	if _, err := e.Evaluate(context.Background(), []Query{{Name: "hp"}}); !errors.Is(err, ErrDisconnected) {
		t.Fatalf("чтение без игры: %v", err)
	}
}

func TestProtocolVersionIsInformational(t *testing.T) {
	g := fakegame.Start(t)
	g.SetVersion("0.4")
	e := startEngine(t, g, Config{})
	if status := e.Status(); status.ProtocolVersion != "0.4" || !status.Connected {
		t.Fatalf("неверный статус: %+v", status)
	}
}

func TestModeChangeAndRecovery(t *testing.T) {
	for _, mode := range []string{"async", "exclusive"} {
		t.Run(mode, func(t *testing.T) {
			g := fakegame.Start(t)
			e := startEngine(t, g, Config{})
			g.SetMode(mode)
			eventually(t, "неподдержанный режим", func() bool { return e.Status().LastError == ErrUnsupportedMode.Error() })
			if _, err := e.Snapshot(); !errors.Is(err, ErrUnsupportedMode) {
				t.Fatalf("снимок в другом режиме: %v", err)
			}
			if _, err := e.Evaluate(testContext(t), []Query{{Name: "hp"}}); !errors.Is(err, ErrUnsupportedMode) {
				t.Fatalf("чтение в другом режиме: %v", err)
			}
			g.SetMode("sync")
			eventually(t, "возвращение sync", func() bool { _, err := e.Snapshot(); return err == nil })
		})
	}
}
