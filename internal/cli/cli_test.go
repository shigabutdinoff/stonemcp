package cli

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func execute(args ...string) (out, errOut string, err error) {
	var stdout, stderr bytes.Buffer
	root := New("27.0", "27A41", "2026-09-28", "3f2a9c1", &stdout, &stderr)
	root.SetArgs(append([]string{}, args...))
	err = root.Execute()
	return stdout.String(), stderr.String(), err
}

func mustExecute(t *testing.T, args ...string) string {
	t.Helper()
	out, errOut, err := execute(args...)
	if err != nil || errOut != "" {
		t.Fatalf("команда %q: err = %v, stderr = %q", args, err, errOut)
	}
	return out
}

func paths() [][]string {
	var all [][]string
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		all = append(all, strings.Fields(c.CommandPath())[1:])
		for _, sub := range c.Commands() {
			walk(sub)
		}
	}
	walk(New("", "", "", "", io.Discard, io.Discard))
	return all
}

func TestVersion(t *testing.T) {
	want := "Build version: 27.0\nBuild number: 27A41\nBuild date: 2026-09-28\nBuild commit: 3f2a9c1\n"
	for _, flag := range []string{"--version", "-v"} {
		if got := mustExecute(t, flag); got != want {
			t.Errorf("%s = %q, ожидалось %q", flag, got, want)
		}
	}
}

func TestVersionWithoutBuildInfo(t *testing.T) {
	var out bytes.Buffer
	root := New("", "", "", "", &out, io.Discard)
	root.SetArgs([]string{"--version"})
	want := "Build version: N/A\nBuild number: N/A\nBuild date: N/A\nBuild commit: N/A\n"
	if err := root.Execute(); err != nil || out.String() != want {
		t.Errorf("версия без сведений = %q, err = %v, ожидалось %q", out.String(), err, want)
	}
}

func TestCompletionScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		out := mustExecute(t, "completion", shell)
		if !strings.Contains(out, shell+" completion") || !strings.Contains(out, "stonemcp") {
			t.Errorf("completion %s: скрипт не похож на автодополнение stonemcp: %.80q", shell, out)
		}
	}
}

func TestMousetrap(t *testing.T) {
	want := "Это консольная программа.\n\nЗапустите её из командной строки cmd.exe или PowerShell.\n"
	if cobra.MousetrapHelpText != want {
		t.Errorf("текст для двойного щелчка в Windows = %q", cobra.MousetrapHelpText)
	}
}
