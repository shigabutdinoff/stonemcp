package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) repo {
	t.Helper()
	r := repo{t, t.TempDir()}
	r.git("init", "-q")
	r.write("a", "1")
	r.git("add", ".")
	r.git("commit", "-qm", "первый")
	return r
}

func (r repo) tag(names ...string) {
	r.t.Helper()
	for _, name := range names {
		r.run([]string{"GIT_COMMITTER_DATE=1790000000 +0000"}, "git", "tag", "-a", name, "-m", name)
	}
}

func (r repo) untracked() { r.write("b", "1") }

func (r repo) write(name, text string) {
	r.t.Helper()
	if err := os.WriteFile(filepath.Join(r.dir, name), []byte(text), 0o644); err != nil {
		r.t.Fatalf("файл %s не записан: %v", name, err)
	}
}

func (r repo) git(args ...string) {
	r.t.Helper()
	r.run(nil, "git", args...)
}
func (r repo) make(target string) string {
	r.t.Helper()
	makefile, err := filepath.Abs("../../Makefile")
	if err != nil {
		r.t.Fatalf("путь к Makefile не получен: %v", err)
	}
	return strings.TrimSpace(r.run(nil, "make", "-s", "-f", makefile, target, "GORELEASER=echo"))
}

func (r repo) run(env []string, name string, args ...string) string {
	r.t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = r.dir
	cmd.Env = slices.DeleteFunc(os.Environ(), func(kv string) bool { return strings.HasPrefix(kv, "GIT_") })
	cmd.Env = append(cmd.Env, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	cmd.Env = append(cmd.Env, env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return string(out)
}
