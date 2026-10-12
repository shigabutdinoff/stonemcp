//go:build unix

package main

import (
	"io"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const (
	stuckEnv = "STONEMCP_TEST_STUCK"
	addrEnv  = "STONEMCP_TEST_ADDR"
)

func TestMain(m *testing.M) {
	if os.Getenv(stuckEnv) == "1" {
		os.Exit(run([]string{"--addr", noGame}, noEnv, os.Stdin, stuckWriter{}, io.Discard))
	}
	if addr := os.Getenv(addrEnv); addr != "" {
		os.Exit(run([]string{"--addr", addr}, noEnv, os.Stdin, os.Stdout, os.Stderr))
	}
	os.Exit(m.Run())
}

type stuckWriter struct{}

func (stuckWriter) Write([]byte) (int, error) {
	os.Stdout.Write([]byte{'!'})
	select {}
}

func (stuckWriter) Close() error { return nil }

func child(env string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = slices.DeleteFunc(os.Environ(), func(kv string) bool {
		return strings.HasPrefix(kv, "STONEMCP_") && !strings.HasPrefix(kv, "STONEMCP_TEST_")
	})
	cmd.Env = append(cmd.Env, env)
	return cmd
}

func spawn(t *testing.T, cmd *exec.Cmd) (io.WriteCloser, <-chan struct{}) {
	t.Helper()
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("stonemcp не запущен: %v", err)
	}
	exited := make(chan struct{})
	go func() {
		defer close(exited)
		cmd.Wait()
	}()
	t.Cleanup(func() {
		cmd.Process.Kill()
		<-exited
	})
	return stdin, exited
}

func TestChildWithoutUserSettings(t *testing.T) {
	t.Setenv("STONEMCP_ADDR", "127.0.0.1:7")
	t.Setenv("STONEMCP_MODE", "async")
	cmd := child(addrEnv + "=" + noGame)
	for _, kv := range cmd.Env {
		if strings.HasPrefix(kv, "STONEMCP_") && !strings.HasPrefix(kv, "STONEMCP_TEST_") {
			t.Errorf("дочернему процессу передана настройка пользователя %s", kv)
		}
	}
	if want := "STONEMCP_TEST_ADDR=" + noGame; !slices.Contains(cmd.Env, want) {
		t.Errorf("в окружении дочернего процесса нет %s", want)
	}
}
