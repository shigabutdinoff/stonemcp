package osexitcheck_test

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/shigabutdinoff/stonemcp/cmd/staticlint/osexitcheck"
)

const report = "прямой вызов os.Exit в функции main запрещён"

var wantLines = map[string][]int{
	"exitaliased":   {6},
	"exitdotimport": {6},
	"exitinclosure": {7},
	"exitinfunc":    nil,
	"exitinlib":     nil,
	"exitinmain":    {6},
	"exitparen":     {6},
	"exitviavar":    nil,
	"shadowedos":    nil,
}

type recorder struct {
	messages []string
}

func (r *recorder) Errorf(format string, args ...any) {
	r.messages = append(r.messages, fmt.Sprintf(format, args...))
}

func TestAnalyzer(t *testing.T) {
	var rec recorder
	patterns := slices.Sorted(maps.Keys(wantLines))
	results := analysistest.Run(&rec, analysistest.TestData(), osexitcheck.Analyzer, patterns...)
	if len(results) != len(wantLines) {
		t.Fatalf("проверено пакетов %d, ждали %d, сообщения %q", len(results), len(wantLines), rec.messages)
	}
	found := 0
	for _, res := range results {
		pkg := res.Action.Package
		var lines []int
		for _, d := range res.Action.Diagnostics {
			lines = append(lines, pkg.Fset.Position(d.Pos).Line)
			if d.Message != report {
				t.Errorf("пакет %s, сообщение %q, ждали %q", pkg.PkgPath, d.Message, report)
			}
		}
		if !slices.Equal(lines, wantLines[pkg.PkgPath]) {
			t.Errorf("пакет %s, находки в строках %v, ждали %v", pkg.PkgPath, lines, wantLines[pkg.PkgPath])
		}
		found += len(lines)
	}
	for _, msg := range rec.messages {
		if !strings.HasSuffix(msg, ": unexpected diagnostic: "+report) {
			t.Errorf("analysistest сообщил %q", msg)
		}
	}
	if len(rec.messages) != found {
		t.Errorf("analysistest сообщил %d находок, ждали %d", len(rec.messages), found)
	}
}
