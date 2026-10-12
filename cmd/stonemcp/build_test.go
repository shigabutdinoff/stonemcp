package main

import (
	"strings"
	"testing"
)

func TestVersionLine(t *testing.T) {
	for _, tc := range []struct {
		build [4]string
		want  string
	}{
		{[4]string{"27.0", "27A41", "2026-09-28", "3f2a9c1"}, "27.0 (27A41)"},
		{[4]string{"27.1", "27B5012b", "2026-10-13", "c4e1d20"}, "27.1 (27B5012b)"},
		{[4]string{"dev", "dev", "2026-10-13", "c4e1d20"}, "dev (c4e1d20)"},
		{[4]string{"dev", "dev", "", ""}, "dev"},
		{[4]string{"27.0", "", "", "abc1234"}, "27.0 (abc1234)"},
		{[4]string{"27.0", "", "", ""}, "27.0"},
		{[4]string{}, "N/A"},
	} {
		setBuild(t, tc.build[0], tc.build[1], tc.build[2], tc.build[3])
		if got := versionLine(); got != tc.want {
			t.Errorf("строка версии для %v = %q, ожидалось %q", tc.build, got, tc.want)
		}
	}
}

func TestStartRecordHasBuild(t *testing.T) {
	for _, tc := range []struct {
		build [4]string
		want  string
	}{
		{[4]string{"27.1", "27B5012b", "2026-09-30", "c4e1d20"}, `версия=27.1 "номер сборки"=27B5012b "дата сборки"=2026-09-30 коммит=c4e1d20`},
		{[4]string{"dev", "dev", "", ""}, `версия=dev "номер сборки"=dev "дата сборки"=N/A коммит=N/A`},
		{[4]string{}, `версия=N/A "номер сборки"=N/A "дата сборки"=N/A коммит=N/A`},
	} {
		setBuild(t, tc.build[0], tc.build[1], tc.build[2], tc.build[3])
		var stderr strings.Builder
		code := run([]string{"--addr", noGame}, noEnv, brokenInput(), &buffer{}, &stderr)
		if code != 1 || !strings.Contains(stderr.String(), `msg="stonemcp запущен" `+tc.want) {
			t.Errorf("код = %d, журнал = %q, ожидалась запись запуска с %q", code, stderr.String(), tc.want)
		}
	}
}
