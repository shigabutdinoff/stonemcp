//go:build unix

package main

import (
	"io"
	"os"
	"syscall"
	"testing"
	"time"
)

const initialize = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`

func TestStopOnSignal(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT} {
		t.Run(sig.String(), func(t *testing.T) {
			ln := listen(t)
			l := start(t, "--addr", ln.Addr().String())
			conn := accept(t, ln)
			if err := syscall.Kill(syscall.Getpid(), sig); err != nil {
				t.Fatalf("сигнал %v не отправлен: %v", sig, err)
			}
			l.wantReleased(t, conn)
		})
	}
}

func TestSecondSignalStops(t *testing.T) {
	t.Parallel()
	out, marker, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	cmd := child(stuckEnv + "=1")
	cmd.Stdout = marker
	stdin, exited := spawn(t, cmd)
	marker.Close()
	io.WriteString(stdin, initialize+"\n")
	out.SetReadDeadline(time.Now().Add(patience))
	if _, err := out.Read(make([]byte, 1)); err != nil {
		t.Fatalf("stonemcp не начал отвечать: %v", err)
	}
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	deadline := time.After(patience)
	for {
		cmd.Process.Signal(syscall.SIGTERM)
		select {
		case <-exited:
			if ws := cmd.ProcessState.Sys().(syscall.WaitStatus); !ws.Signaled() || ws.Signal() != syscall.SIGTERM {
				t.Errorf("stonemcp завершился так: %v, ожидалась остановка по SIGTERM", cmd.ProcessState)
			}
			return
		case <-deadline:
			t.Fatal("stonemcp не завершился после повторных SIGTERM")
		case <-tick.C:
		}
	}
}
