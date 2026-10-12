package engine

import (
	"net"
	"strings"
	"testing"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestPartialSignalHasBoundedRead(t *testing.T) {
	client, game := net.Pipe()
	defer client.Close()
	defer game.Close()
	s := &session{conn: client, buf: make([]byte, 128), cfg: Config{ReadTimeout: 10 * time.Millisecond}}
	done := make(chan error, 1)
	go func() { _, _, err := s.signal(); done <- err }()
	if _, err := game.Write([]byte("\x06pre0.")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "оборван") {
			t.Fatalf("оборванный сигнал: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("неполный сигнал ждёт бесконечно")
	}
}

func TestSignalWaitsThroughPauseAndReadsFragments(t *testing.T) {
	client, game := net.Pipe()
	defer client.Close()
	defer game.Close()
	s := &session{conn: client, buf: make([]byte, 128), cfg: Config{ReadTimeout: 50 * time.Millisecond}}
	done := make(chan error, 1)
	go func() {
		phase, version, err := s.signal()
		if err == nil && (phase != mc.PhasePre || version != "0.4") {
			t.Errorf("сигнал: %v, %q", phase, version)
		}
		done <- err
	}()
	time.Sleep(70 * time.Millisecond)
	for _, part := range []string{"\x06", "p", "re", "0.", "4"} {
		game.SetWriteDeadline(time.Now().Add(time.Second))
		if _, err := game.Write([]byte(part)); err != nil {
			t.Fatal(err)
		}
	}
	if err := <-done; err != nil {
		t.Fatalf("сигнал после паузы: %v", err)
	}
}
