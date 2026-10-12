package main

import (
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/shigabutdinoff/stonemcp/internal/cli"
)

func parseFlags(args []string, env lookup) (options, error) {
	root := cli.New("", "", "", "", io.Discard, io.Discard)
	settle := bindFlags(root, env)
	var o options
	root.RunE = func(*cobra.Command, []string) (err error) {
		o, err = settle()
		return err
	}
	root.SetArgs(append([]string{}, args...))
	return o, root.Execute()
}

func TestFlagDefaults(t *testing.T) {
	o, err := parseFlags(nil, noEnv)
	if want := (options{settings: settings{Mode: "sync", Addr: "127.0.0.1:64649"}}); err != nil || o != want {
		t.Errorf("флаги = %+v, err = %v, ожидалось %+v", o, err, want)
	}
}

func TestStartErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"--mode", "async"}, "--mode async: режим не поддержан в stonemcp 27.0"},
		{[]string{"--mode", "exclusive"}, "--mode exclusive: режим не поддержан в stonemcp 27.0"},
		{[]string{"--mode", "turbo"}, "--mode turbo: неизвестный режим"},
		{[]string{"--nosuch"}, `неизвестный флаг "--nosuch"`},
		{[]string{"-addr", "127.0.0.1:7000"}, `неизвестный флаг "-a"`},
		{[]string{"--addr"}, `флагу "--addr" нужно значение`},
		{[]string{"nosuch"}, `неизвестная команда "nosuch"`},
		{[]string{"stdio", "--addr", "127.0.0.1:7000"}, `неизвестная команда "stdio"`},
		{[]string{"--addr", "127.0.0.1", "7000"}, `неизвестная команда "7000"`},
		{[]string{"completion", "zsh", "x"}, `лишний аргумент "x"`},
	} {
		var stdout buffer
		var stderr strings.Builder
		code := run(tc.args, noEnv, nil, &stdout, &stderr)
		if want := "Ошибка: " + tc.want + "\n"; code != 2 || stdout.Len() != 0 || stderr.String() != want {
			t.Errorf("%v: код = %d, stdout = %q, stderr = %q, ожидались 2 и %q", tc.args, code, stdout.String(), stderr.String(), want)
		}
	}
}

func TestHelpOnStdout(t *testing.T) {
	var want string
	for _, args := range [][]string{{"--help"}, {"-h"}, {"help"}} {
		var stdout buffer
		var stderr strings.Builder
		if code := run(args, noEnv, nil, &stdout, &stderr); code != 0 || stderr.Len() != 0 {
			t.Errorf("%v: код = %d, stderr = %q, ожидались 0 и пустой stderr", args, code, stderr.String())
		}
		if want == "" {
			want = stdout.String()
		}
		if stdout.String() != want {
			t.Errorf("%v: справка отличается от --help:\n%s", args, stdout.String())
		}
	}
	for _, part := range []string{
		"Использование:\n  stonemcp [флаги]\n  stonemcp [команда]\n",
		"      --addr       адрес игры, по умолчанию 127.0.0.1:64649\n",
		"  -c, --config     файл конфигурации JSON с ключами addr, mode, log_file\n",
		"      --log-file   файл журнала, без флага журнал пишется в stderr\n",
		"      --mode       режим MindConnect, поддержан только sync\n",
	} {
		if !strings.Contains(want, part) {
			t.Errorf("в справке нет %q:\n%s", part, want)
		}
	}
	for _, leftover := range []string{"string", "default", "Usage:", "Flags:"} {
		if strings.Contains(want, leftover) {
			t.Errorf("в справке осталось %q:\n%s", leftover, want)
		}
	}
}

func TestCompletionOnStdout(t *testing.T) {
	var stdout buffer
	var stderr strings.Builder
	code := run([]string{"completion", "zsh"}, noEnv, nil, &stdout, &stderr)
	if code != 0 || stderr.Len() != 0 || !strings.HasPrefix(stdout.String(), "#compdef stonemcp\n") {
		t.Errorf("completion zsh: код = %d, stderr = %q, stdout = %.40q", code, stderr.String(), stdout.String())
	}
}

func TestVersion(t *testing.T) {
	for _, tc := range []struct {
		build [4]string
		want  string
	}{
		{[4]string{"27.0", "27A41", "2026-09-28", "3f2a9c1"}, "Build version: 27.0\nBuild number: 27A41\nBuild date: 2026-09-28\nBuild commit: 3f2a9c1\n"},
		{[4]string{"27.1", "27B5012b", "2026-10-13", "c4e1d20"}, "Build version: 27.1\nBuild number: 27B5012b\nBuild date: 2026-10-13\nBuild commit: c4e1d20\n"},
		{[4]string{"dev", "dev", "2026-10-13", "c4e1d20"}, "Build version: dev\nBuild number: dev\nBuild date: 2026-10-13\nBuild commit: c4e1d20\n"},
		{[4]string{}, "Build version: N/A\nBuild number: N/A\nBuild date: N/A\nBuild commit: N/A\n"},
	} {
		setBuild(t, tc.build[0], tc.build[1], tc.build[2], tc.build[3])
		var stdout buffer
		if code := run([]string{"--version"}, noEnv, nil, &stdout, io.Discard); code != 0 || stdout.String() != tc.want {
			t.Errorf("--version = %q, код = %d, ожидалось %q", stdout.String(), code, tc.want)
		}
	}
}
