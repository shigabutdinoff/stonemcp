//go:build unix

package main

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestClosedJournalDoesNotKill(t *testing.T) {
	t.Parallel()
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd := child(addrEnv + "=" + noGame)
	cmd.Stderr = errW
	stdin, exited := spawn(t, cmd)
	errW.Close()
	errR.SetReadDeadline(time.Now().Add(patience))
	line, err := bufio.NewReader(errR).ReadString('\n')
	if !strings.Contains(line, "stonemcp запущен") {
		t.Fatalf("первая строка журнала = %q, err = %v", line, err)
	}
	errR.Close()
	io.WriteString(stdin, "{")
	stdin.Close()
	within(t, exited, "stonemcp не завершился после неверного ввода")
	if code := cmd.ProcessState.ExitCode(); code != 1 {
		t.Errorf("stonemcp завершился так: %v, ожидался код 1", cmd.ProcessState)
	}
}
