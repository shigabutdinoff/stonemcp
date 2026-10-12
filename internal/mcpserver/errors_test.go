package mcpserver

import (
	"errors"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

func TestHintStartsLowercaseWithoutPeriod(t *testing.T) {
	first, _ := utf8.DecodeRuneInString(Hint)
	if !unicode.IsLower(first) || strings.HasSuffix(Hint, ".") {
		t.Errorf("Hint = %q, ожидалась строчная буква в начале и нет точки в конце", Hint)
	}
}

func TestExplainDisconnectedAddsHint(t *testing.T) {
	err := explain(engine.ErrDisconnected)
	want := engine.ErrDisconnected.Error() + ": " + Hint
	if !errors.Is(err, engine.ErrDisconnected) || err.Error() != want {
		t.Errorf("explain = %q, ожидалось %q", err, want)
	}
}
