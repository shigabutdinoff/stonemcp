package main

import (
	"strings"
	"testing"
)

type vars map[string]string

func (v vars) lookup(name string) (string, bool) {
	s, ok := v[name]
	return s, ok
}

func noEnv(string) (string, bool) { return "", false }

func TestEnvOverridesFlags(t *testing.T) {
	flags := []string{"--addr", "127.0.0.1:7000", "--mode", "turbo", "--log-file", "f.log"}
	env := vars{"STONEMCP_ADDR": "127.0.0.1:7001", "STONEMCP_MODE": "sync", "STONEMCP_LOG_FILE": "e.log"}
	for _, tc := range []struct {
		name string
		args []string
		env  vars
		want settings
	}{
		{"умолчания", nil, nil, settings{Addr: "127.0.0.1:64649", Mode: "sync"}},
		{"флаги", flags[:2], nil, settings{Addr: "127.0.0.1:7000", Mode: "sync"}},
		{"окружение", nil, env, settings{Addr: "127.0.0.1:7001", Mode: "sync", LogFile: "e.log"}},
		{"окружение главнее флагов", flags, env, settings{Addr: "127.0.0.1:7001", Mode: "sync", LogFile: "e.log"}},
		{"пустая переменная", flags[4:], vars{"STONEMCP_LOG_FILE": ""}, settings{Addr: "127.0.0.1:64649", Mode: "sync"}},
		{"пустой адрес", flags[:2], vars{"STONEMCP_ADDR": ""}, settings{Mode: "sync"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, err := parseFlags(tc.args, tc.env.lookup)
			if err != nil || o.settings != tc.want {
				t.Errorf("настройки = %+v, err = %v, ожидалось %+v", o.settings, err, tc.want)
			}
		})
	}
}

func TestEnvModeErrors(t *testing.T) {
	for _, tc := range []struct{ mode, want string }{
		{"async", "Ошибка: STONEMCP_MODE async: режим не поддержан в stonemcp 27.0\n"},
		{"exclusive", "Ошибка: STONEMCP_MODE exclusive: режим не поддержан в stonemcp 27.0\n"},
		{"turbo", "Ошибка: STONEMCP_MODE turbo: неизвестный режим\n"},
		{"", "Ошибка: STONEMCP_MODE : неизвестный режим\n"},
	} {
		var stdout buffer
		var stderr strings.Builder
		code := run(nil, vars{"STONEMCP_MODE": tc.mode}.lookup, nil, &stdout, &stderr)
		if code != 2 || stdout.Len() != 0 || stderr.String() != tc.want {
			t.Errorf("%s: код = %d, stdout = %q, stderr = %q, ожидались 2 и %q", tc.mode, code, stdout.String(), stderr.String(), tc.want)
		}
	}
}

func TestEnvAddrInJournal(t *testing.T) {
	t.Parallel()
	var stderr strings.Builder
	code := run([]string{"--addr", "127.0.0.1:2"}, vars{"STONEMCP_ADDR": noGame}.lookup, brokenInput(), &buffer{}, &stderr)
	if code != 1 || !strings.Contains(stderr.String(), "адрес="+noGame+"\n") {
		t.Errorf("код = %d, журнал = %q, ожидался адрес %s из окружения", code, stderr.String(), noGame)
	}
}
