package main

import (
	"regexp"
	"strings"
	"testing"
)

var builtVersion = regexp.MustCompile(`main\.version=(\S+)`)

func TestVersionFromTag(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		tags []string
		then func(repo)
		want string
	}{
		{"без тега", nil, nil, "dev"},
		{"альфа", []string{"27.0-alpha.1"}, nil, "27.0-alpha.1"},
		{"бета", []string{"27.0-beta.2"}, nil, "27.0-beta.2"},
		{"кандидат", []string{"27.0-rc.1"}, nil, "27.0-rc.1"},
		{"релиз", []string{"27.0"}, nil, "27.0"},
		{"бета обновления", []string{"27.1-beta.1"}, nil, "27.1-beta.1"},
		{"кандидат исправлений", []string{"27.0.1-rc.1"}, nil, "27.0.1-rc.1"},
		{"префикс v", []string{"v27.0"}, nil, "dev"},
		{"нулевой patch", []string{"27.0.0"}, nil, "dev"},
		{"нулевой номер", []string{"27.0-beta.0"}, nil, "dev"},
		{"чужая ступень", []string{"27.0-gamma.1"}, nil, "dev"},
		{"без номера", []string{"27.0-beta"}, nil, "dev"},
		{"следующая ступень", []string{"27.0-alpha.3", "27.0-beta.1"}, nil, "27.0-beta.1"},
		{"кандидат после беты", []string{"27.0-rc.1", "27.0-beta.2"}, nil, "27.0-rc.1"},
		{"двузначный номер", []string{"27.0-beta.9", "27.0-beta.10"}, nil, "27.0-beta.10"},
		{"релиз после кандидата", []string{"27.0-rc.2", "27.0"}, nil, "27.0"},
		{"новый коммит", []string{"27.0-beta.2"}, repo.advance, "dev"},
		{"правка", []string{"27.0-beta.2"}, repo.edit, "dev"},
		{"новый файл", []string{"27.0-beta.2"}, repo.untracked, "dev"},
		{"чужой порядок ступеней", []string{"27.0-alpha.3", "27.0-beta.1", "27.0-rc.1"}, repo.foreignOrder, "27.0-rc.1"},
		{"скрытый новый файл", []string{"27.0-beta.2"}, repo.hiddenUntracked, "dev"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := newRepo(t)
			r.tag(tc.tags...)
			if tc.then != nil {
				tc.then(r)
			}
			if got := r.version(); got != tc.want {
				t.Errorf("версия сборки = %q, ожидалась %q", got, tc.want)
			}
		})
	}
}

func TestReleaseFileNames(t *testing.T) {
	t.Parallel()
	r := newRepo(t)
	r.tag("27.0-beta.2")
	if out := r.make("release"); !strings.Contains(out, "dist/stonemcp-27.0-beta.2-$os-$arch$ext") {
		t.Errorf("make release = %q, ожидались файлы stonemcp-27.0-beta.2-<os>-<arch>", out)
	}
}

func (r repo) version() string {
	r.t.Helper()
	m := builtVersion.FindStringSubmatch(r.make("build"))
	if m == nil {
		r.t.Fatal("в команде сборки нет версии")
	}
	return m[1]
}
