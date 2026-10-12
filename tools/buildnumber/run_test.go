package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func workDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(dir))
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, name := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(name, "test")
	}
	for _, name := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(name, "test@example.com")
	}
	return dir
}

func gitDo(t *testing.T, args ...string) {
	t.Helper()
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func release(t *testing.T, msg, version, want string) {
	t.Helper()
	gitDo(t, "commit", "-q", "--allow-empty", "-m", msg)
	gitDo(t, "tag", "-a", version, "-m", version)
	if err := run([]string{"out", version}); err != nil {
		t.Fatalf("номер сборки %s не посчитан: %v", version, err)
	}
	if got, _ := os.ReadFile("out"); string(got) != want+"\n" {
		t.Errorf("номер сборки %s = %q, ожидался %q", version, got, want)
	}
}

func TestNumberFromHistory(t *testing.T) {
	workDir(t)
	gitDo(t, "init", "-q")
	gitDo(t, "commit", "-q", "--allow-empty", "-m", "c1")
	release(t, "c2", "27.0", "27A2")
	gitDo(t, "tag", "nightly")
	gitDo(t, "tag", "v27.0")
	release(t, "c3", "27.1-beta", "27B5001a")
	gitDo(t, "checkout", "-q", "-b", "hotfix", "27.0")
	release(t, "h1", "27.0.1", "27A3")
	gitDo(t, "checkout", "-q", "-")
	gitDo(t, "merge", "-q", "--no-ff", "-m", "m1", "hotfix")
	release(t, "c4", "27.1-beta.2", "27B5004b")
	release(t, "c5", "27.1-rc", "27B5")
	release(t, "c6", "27.1", "27B6")
	release(t, "c7", "28.0-beta", "28A5001a")
}

func TestDevNumber(t *testing.T) {
	workDir(t)
	if err := run([]string{"out", "dev"}); err != nil {
		t.Fatalf("номер рабочей сборки не записан: %v", err)
	}
	if got, _ := os.ReadFile("out"); string(got) != "dev\n" {
		t.Errorf("номер рабочей сборки = %q, ожидался dev", got)
	}
}

func TestRunErrors(t *testing.T) {
	workDir(t)
	for _, tc := range []struct {
		args []string
		want string
	}{
		{nil, "нужны два аргумента, файл и версия"},
		{[]string{"-x", "dev"}, `неизвестный флаг "-x"`},
		{[]string{"out", "27.1-alpha.1"}, `неверная версия "27.1-alpha.1"`},
		{[]string{"out", "27.26"}, "минор 26 больше 25, для него нет буквы"},
		{[]string{"out", "27.1"}, `команда "git tag --list": exit status 128`},
		{[]string{"nodir/out", "dev"}, "запись номера сборки: "},
	} {
		if err := run(tc.args); err == nil || !strings.HasPrefix(err.Error(), tc.want) {
			t.Errorf("%v: err = %v, ожидалась ошибка %q", tc.args, err, tc.want)
		}
	}
	for _, name := range []string{"out", "-x"} {
		if _, err := os.Stat(name); err == nil {
			t.Errorf("после ошибки остался файл %s", name)
		}
	}
}

func TestShallowHistory(t *testing.T) {
	dir := workDir(t)
	gitDo(t, "init", "-q", "src")
	gitDo(t, "-C", "src", "commit", "-q", "--allow-empty", "-m", "c1")
	gitDo(t, "-C", "src", "commit", "-q", "--allow-empty", "-m", "c2")
	gitDo(t, "-C", "src", "tag", "-a", "27.0", "-m", "27.0")
	gitDo(t, "clone", "-q", "--depth", "1", "file://"+filepath.Join(dir, "src"), "dst")
	t.Chdir("dst")
	want := "история репозитория неполная, номер сборки не посчитать"
	if err := run([]string{"out", "27.0"}); err == nil || err.Error() != want {
		t.Errorf("неполная история: err = %v, ожидалась ошибка %q", err, want)
	}
}
