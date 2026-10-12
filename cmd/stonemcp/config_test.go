package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "stonemcp.json")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("файл конфигурации не записан: %v", err)
	}
	return path
}

func TestConfigFileLayers(t *testing.T) {
	full := writeConfig(t, `{"addr": "127.0.0.1:7100", "mode": "turbo", "log_file": "c.log"}`)
	short := writeConfig(t, `{"addr": "127.0.0.1:7200"}`)
	fromShort := settings{Addr: "127.0.0.1:7200", Mode: "sync"}
	for _, tc := range []struct {
		name string
		args []string
		env  vars
		want settings
	}{
		{"флаг c", []string{"-c", short}, nil, fromShort},
		{"флаг config", []string{"--config", short}, nil, fromShort},
		{"флаги главнее файла", []string{"-c", full, "--mode", "sync", "--log-file", "f.log"}, nil,
			settings{Addr: "127.0.0.1:7100", Mode: "sync", LogFile: "f.log"}},
		{"окружение главнее флагов", []string{"-c", full, "--addr", "127.0.0.1:7000"},
			vars{"STONEMCP_ADDR": "127.0.0.1:7001", "STONEMCP_MODE": "sync"},
			settings{Addr: "127.0.0.1:7001", Mode: "sync", LogFile: "c.log"}},
		{"STONEMCP_CONFIG главнее флага", []string{"-c", full}, vars{"STONEMCP_CONFIG": short}, fromShort},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, err := parseFlags(tc.args, tc.env.lookup)
			if err != nil || o.settings != tc.want {
				t.Errorf("настройки = %+v, err = %v, ожидалось %+v", o.settings, err, tc.want)
			}
		})
	}
}

func TestConfigFileErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "нет.json")
	unknown := writeConfig(t, `{"port": 1}`)
	broken := writeConfig(t, `{"addr":`)
	async := writeConfig(t, `{"mode": "async"}`)
	for _, tc := range []struct {
		name string
		args []string
		env  vars
		path string
		want string
	}{
		{"нет файла", []string{"-c", missing}, nil, missing, "no such file or directory"},
		{"нет файла из окружения", nil, vars{"STONEMCP_CONFIG": missing}, missing, "no such file or directory"},
		{"лишний ключ", []string{"-c", unknown}, nil, unknown, `json: unknown field "port"`},
		{"оборванный JSON", []string{"-c", broken}, nil, broken, "unexpected EOF"},
		{"режим из файла", []string{"-c", async}, nil, async, "mode async: режим не поддержан в stonemcp 27.0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout buffer
			var stderr strings.Builder
			code := run(tc.args, tc.env.lookup, brokenInput(), &stdout, &stderr)
			got := stderr.String()
			prefix := "Ошибка: файл конфигурации " + tc.path + ": "
			if code != 2 || stdout.Len() != 0 || !strings.HasPrefix(got, prefix) || !strings.HasSuffix(got, tc.want+"\n") {
				t.Errorf("код = %d, stderr = %q, ожидались 2 и %q…%q", code, got, prefix, tc.want)
			}
			if strings.Contains(got, "Использование") {
				t.Errorf("stderr = %q, ожидалась ошибка без справки", got)
			}
		})
	}
}
