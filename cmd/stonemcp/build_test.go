package main

import (
	"regexp"
	"strings"
	"testing"
)

var builtDate = regexp.MustCompile(`-X main\.buildDate=\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`)

func TestBuildDateFromMake(t *testing.T) {
	t.Parallel()
	if out := newRepo(t).make("build"); !builtDate.MatchString(out) {
		t.Errorf("make build = %q, ожидалась дата сборки в UTC", out)
	}
}

func setBuildDate(t *testing.T, d string) {
	t.Helper()
	old := buildDate
	buildDate = d
	t.Cleanup(func() { buildDate = old })
}

func TestStartRecordHasBuild(t *testing.T) {
	for _, tc := range []struct{ version, date, commit, want string }{
		{"27.0-beta.2", "2026-09-30T12:00:00Z", "c4e1d20", `версия=27.0-beta.2 "дата сборки"=2026-09-30T12:00:00Z коммит=c4e1d20`},
		{"dev", "", "", `версия=dev "дата сборки"=N/A коммит=N/A`},
		{"", "", "", `версия=N/A "дата сборки"=N/A коммит=N/A`},
	} {
		setVersion(t, tc.version, tc.commit)
		setBuildDate(t, tc.date)
		var stderr strings.Builder
		code := run([]string{"--addr", noGame}, noEnv, brokenInput(), &buffer{}, &stderr)
		if code != 1 || !strings.Contains(stderr.String(), `msg="stonemcp запущен" `+tc.want) {
			t.Errorf("код = %d, журнал = %q, ожидалась запись запуска с %q", code, stderr.String(), tc.want)
		}
	}
}
