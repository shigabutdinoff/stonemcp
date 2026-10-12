package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const argsEnv = "STATICLINT_TEST_ARGS"

func TestMain(m *testing.M) {
	if args, ok := os.LookupEnv(argsEnv); ok {
		os.Args = append([]string{"staticlint"}, strings.Fields(args)...)
		main()
	}
	os.Exit(m.Run())
}

func TestMainRunsAnalyzers(t *testing.T) {
	tests := []struct {
		name string
		args string
		code int
		want []string
	}{
		{"справка", "help", 0, []string{"printf", "nilerr", "osexitcheck"}},
		{"os.Exit в main", "./osexitcheck/testdata/src/exitinmain", 3, []string{"прямой вызов os.Exit в функции main запрещён"}},
		{"stonemcp", "../stonemcp", 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command(os.Args[0], "-test.run=^$")
			cmd.Env = append(os.Environ(), argsEnv+"="+tt.args)
			out, err := cmd.CombinedOutput()
			code := 0
			if exit, ok := errors.AsType[*exec.ExitError](err); ok {
				code = exit.ExitCode()
			} else if err != nil {
				t.Fatalf("staticlint не запущен: %v", err)
			}
			if code != tt.code {
				t.Errorf("код выхода %d, ждали %d\n%s", code, tt.code, out)
			}
			for _, want := range tt.want {
				if !strings.Contains(string(out), want) {
					t.Errorf("в выводе нет %q\n%s", want, out)
				}
			}
		})
	}
}
