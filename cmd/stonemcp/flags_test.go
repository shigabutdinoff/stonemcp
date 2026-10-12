package main

import (
	"io"
	"strings"
	"testing"
)

func TestFlagDefaults(t *testing.T) {
	o, err := parseFlags(nil, noEnv, io.Discard)
	if want := (options{settings: settings{Mode: "sync", Addr: "127.0.0.1:64649"}}); err != nil || o != want {
		t.Errorf("флаги = %+v, err = %v, ожидалось %+v", o, err, want)
	}
}

func TestFlagErrorsAndHelp(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
		want []string
	}{
		{[]string{"--mode", "async"}, 2, []string{"stonemcp: --mode async: режим не поддержан в stonemcp 27.0"}},
		{[]string{"--mode", "exclusive"}, 2, []string{"stonemcp: --mode exclusive: режим не поддержан в stonemcp 27.0"}},
		{[]string{"--mode", "turbo"}, 2, []string{"stonemcp: --mode turbo: неизвестный режим", "Использование"}},
		{[]string{"--bogus"}, 2, []string{"-bogus", "Использование"}},
		{[]string{"stdio", "--addr", "127.0.0.1:7000"}, 2, []string{`stonemcp: лишний аргумент "stdio"`, "Использование"}},
		{[]string{"--addr", "127.0.0.1", "7000"}, 2, []string{`stonemcp: лишний аргумент "7000"`, "Использование"}},
		{[]string{"-h"}, 0, []string{"Использование"}},
	} {
		var stdout buffer
		var stderr strings.Builder
		code := run(tc.args, noEnv, nil, &stdout, &stderr)
		if code != tc.code || stdout.Len() != 0 {
			t.Errorf("%v: код = %d, stdout = %q, ожидался код %d и пустой stdout", tc.args, code, stdout.String(), tc.code)
		}
		for _, want := range tc.want {
			if !strings.Contains(stderr.String(), want) {
				t.Errorf("%v: stderr = %q, ожидалось %q", tc.args, stderr.String(), want)
			}
		}
	}
}

func TestVersion(t *testing.T) {
	for _, tc := range []struct{ version, commit, want string }{
		{version, commit, "dev\n"},
		{"27.0", "", "27.0\n"},
		{"27.0", "abc1234", "27.0 (abc1234)\n"},
		{"27.0-beta.2", "c4e1d20", "27.0-beta.2 (c4e1d20)\n"},
	} {
		setVersion(t, tc.version, tc.commit)
		var stdout buffer
		if code := run([]string{"--version"}, noEnv, nil, &stdout, io.Discard); code != 0 || stdout.String() != tc.want {
			t.Errorf("--version = %q, код = %d, ожидалось %q", stdout.String(), code, tc.want)
		}
	}
}
