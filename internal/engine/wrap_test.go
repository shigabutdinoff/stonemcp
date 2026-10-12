package engine

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
)

func pipeSession(t *testing.T) (*session, net.Conn) {
	t.Helper()
	conn, game := net.Pipe()
	t.Cleanup(func() { conn.Close(); game.Close() })
	return &session{conn: conn, buf: make([]byte, 256), cfg: Config{}.defaults()}, game
}

func wantWrapped(t *testing.T, err, target error, want string) {
	t.Helper()
	if !errors.Is(err, target) || err.Error() != target.Error()+": "+want {
		t.Errorf("err = %v, ожидалось %q после причины %v", err, want, target)
	}
}

func TestConnectErrorNamesAddr(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()
	err = New(Config{Addr: addr}).connect(context.Background())
	var opErr *net.OpError
	if !errors.As(err, &opErr) || err.Error() != opErr.Error()+": подключение к игре" || !strings.Contains(err.Error(), addr) {
		t.Errorf("err = %v, ожидалась ошибка подключения к %s", err, addr)
	}
}

func TestSignalReadErrorNamesPhase(t *testing.T) {
	s, game := pipeSession(t)
	game.Close()
	_, _, err := s.signal()
	wantWrapped(t, err, io.EOF, "чтение сигнала игры")
}

func TestPacketWriteErrorNamesPacket(t *testing.T) {
	s, game := pipeSession(t)
	game.Close()
	_, err := s.exchange(baseRequests)
	wantWrapped(t, err, io.ErrClosedPipe, "отправка пакета 1")
}

func TestResponseReadErrorNamesPacket(t *testing.T) {
	s, game := pipeSession(t)
	go func() {
		game.Read(make([]byte, 4096))
		game.Close()
	}()
	_, err := s.exchange(baseRequests)
	wantWrapped(t, err, io.EOF, "чтение ответа на пакет 1")
}
