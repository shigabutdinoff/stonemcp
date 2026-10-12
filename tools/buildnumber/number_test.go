package main

import (
	"regexp"
	"testing"
)

var formatCases = []struct {
	version string
	count   int
	want    string
}{
	{"27.0", 41, "27A41"}, {"27.0.1", 44, "27A44"}, {"27.0.0", 41, "27A41"},
	{"27.25", 3, "27Z3"}, {"27.1-beta", 7, "27B5007a"}, {"27.1-beta.1", 7, "27B5007a"},
	{"27.1-beta.2", 12, "27B5012b"}, {"27.1-beta.26", 1234, "27B51234z"},
	{"27.1-rc", 22, "27B22"}, {"27.1-rc.2", 23, "27B23"}, {"27.1", 23, "27B23"},
	{"28.0-beta", 9, "28A5009a"},
}

var parseErrors = []struct{ version, want string }{
	{"27.1-alpha.1", `неверная версия "27.1-alpha.1"`},
	{"27.26", "минор 26 больше 25, для него нет буквы"},
	{"27.1-beta.27", "бета 27 больше 26, для неё нет буквы"},
	{"27.1-beta.0", `неверная версия "27.1-beta.0"`},
	{"v27.0", `неверная версия "v27.0"`},
	{"27", `неверная версия "27"`},
	{"", `неверная версия ""`},
	{"27.01", `неверная версия "27.01"`},
	{"27.0+meta", `неверная версия "27.0+meta"`},
	{"99999999999999999999.0", `неверная версия "99999999999999999999.0"`},
}

func TestFormat(t *testing.T) {
	for _, tc := range formatCases {
		v, err := parse(tc.version)
		if err != nil {
			t.Errorf("версия %q не разобрана: %v", tc.version, err)
			continue
		}
		if got := format(v, tc.count); got != tc.want {
			t.Errorf("номер %q с %d коммитами = %q, ожидался %q", tc.version, tc.count, got, tc.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, tc := range parseErrors {
		if _, err := parse(tc.version); err == nil || err.Error() != tc.want {
			t.Errorf("разбор %q: err = %v, ожидалась ошибка %q", tc.version, err, tc.want)
		}
	}
}

func TestBase(t *testing.T) {
	tags := []string{"27.0.1", "27.0.0", "27.0-rc", "26.3", "27.0", "v27.1", "27.1-beta", "27.1", "27.2", "nightly", "99999999999999999999.0"}
	for _, tc := range []struct{ current, want string }{
		{"26.0", ""}, {"27.0", "26.3"}, {"27.1", "27.0"}, {"27.2", "27.1"}, {"28.0", "27.2"},
	} {
		v, err := parse(tc.current)
		if err != nil {
			t.Fatalf("версия %q не разобрана: %v", tc.current, err)
		}
		if got := base(tags, v); got != tc.want {
			t.Errorf("база для %q = %q, ожидалась %q", tc.current, got, tc.want)
		}
	}
	if got := base(nil, version{major: 27, minor: 1}); got != "" {
		t.Errorf("база без тегов = %q, ожидалась пустая", got)
	}
}

func FuzzFormat(f *testing.F) {
	for _, tc := range formatCases {
		f.Add(tc.version, uint16(tc.count))
	}
	number := regexp.MustCompile(`^\d+[A-Z](\d+|5\d{3,}[a-z])$`)
	russian := regexp.MustCompile(`^(неверная версия|минор \d+ больше 25|бета \d+ больше 26)`)
	f.Fuzz(func(t *testing.T, s string, count uint16) {
		v, err := parse(s)
		if err != nil {
			if !russian.MatchString(err.Error()) {
				t.Errorf("ошибка разбора %q не на русском: %v", s, err)
			}
			return
		}
		if got := format(v, int(count)); !number.MatchString(got) {
			t.Errorf("номер сборки %q для %q не в форме Apple", got, s)
		}
	})
}
