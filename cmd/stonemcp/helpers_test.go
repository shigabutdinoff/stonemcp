package main

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	patience = 5 * time.Second
	noGame   = "127.0.0.1:1"
)

func init() {
	journalGrace = 200 * time.Millisecond
}

type buffer struct{ strings.Builder }

func (*buffer) Close() error { return nil }

func brokenInput() io.ReadCloser {
	return io.NopCloser(strings.NewReader("{"))
}

type outcome struct {
	done chan struct{}
	code int
}

func runAsync(args []string, stdin io.ReadCloser, stdout io.WriteCloser, stderr io.Writer) *outcome {
	e := &outcome{done: make(chan struct{})}
	go func() {
		defer close(e.done)
		e.code = run(args, noEnv, stdin, stdout, stderr)
	}()
	return e
}

func (e *outcome) want(t *testing.T, code int) {
	t.Helper()
	within(t, e.done, "stonemcp не завершился")
	if e.code != code {
		t.Errorf("код выхода = %d, ожидался %d", e.code, code)
	}
}

type launch struct {
	*mcp.ClientSession
	*outcome
	stdin  io.Closer
	stderr strings.Builder
}

func start(t *testing.T, args ...string) *launch {
	t.Helper()
	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	l := &launch{stdin: inW}
	l.outcome = runAsync(args, inR, outW, &l.stderr)
	t.Cleanup(func() {
		inW.Close()
		within(t, l.done, "stonemcp не завершился после закрытия stdin")
	})
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(context.Background(), &mcp.IOTransport{Reader: outR, Writer: inW}, nil)
	if err != nil {
		t.Fatalf("клиент не подключился: %v", err)
	}
	l.ClientSession = cs
	return l
}

func within(t *testing.T, done <-chan struct{}, msg string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(patience):
		t.Fatal(msg)
	}
}

func eventually(t *testing.T, why string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(patience)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("не дождались: %s", why)
		}
		time.Sleep(time.Millisecond)
	}
}

func setBuild(t *testing.T, version, number, date, commit string) {
	t.Helper()
	old := [4]string{buildVersion, buildNumber, buildDate, buildCommit}
	buildVersion, buildNumber, buildDate, buildCommit = version, number, date, commit
	t.Cleanup(func() { buildVersion, buildNumber, buildDate, buildCommit = old[0], old[1], old[2], old[3] })
}

func listen(t *testing.T) *net.TCPListener {
	t.Helper()
	ln, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	return ln
}

func accept(t *testing.T, ln *net.TCPListener) net.Conn {
	t.Helper()
	ln.SetDeadline(time.Now().Add(patience))
	conn, err := ln.Accept()
	if err != nil {
		t.Fatalf("stonemcp не подключился к игре: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func (e *outcome) wantReleased(t *testing.T, conn net.Conn) {
	t.Helper()
	e.want(t, 0)
	conn.SetReadDeadline(time.Now().Add(patience))
	if _, err := conn.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Errorf("соединение с игрой не закрыто, err = %v", err)
	}
}
