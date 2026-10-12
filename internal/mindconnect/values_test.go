package mindconnect

import (
	"errors"
	"slices"
	"testing"
)

func TestResponseReaderAppendValues(t *testing.T) {
	r := NewResponseReader(7, 3)
	if done, err := r.Feed([]byte("7\x1f40\x1f50\x1fTrue")); !done || err != nil {
		t.Fatalf("done=%v err=%v", done, err)
	}
	got, err := r.AppendValues([]Value{"old"})
	if err != nil || !slices.Equal(got, []Value{"old", "40", "50"}) {
		t.Fatalf("AppendValues = %q, %v, ожидалось [old 40 50] без ошибки", got, err)
	}
	got[1] = "changed"
	if _, again, _ := r.Response(); !slices.Equal(again, []Value{"40", "50", "True"}) {
		t.Errorf("Response после записи в срез AppendValues = %q, ожидалось, что читатель не изменился", again)
	}
}

func TestResponseReaderAppendValuesErrors(t *testing.T) {
	tests := []struct {
		name  string
		id    int
		input string
		want  error
	}{
		{"ответ не дочитан", 7, "7\x1f40", ErrIncompleteResponse},
		{"чужой id", 7, "8\x1f40\x1f50\x1fTrue", ErrMalformedResponse},
		{"негодный id", 0, "", ErrInvalidRequest},
	}
	for _, tt := range tests {
		r := NewResponseReader(tt.id, 3)
		r.Feed([]byte(tt.input))
		dst := []Value{"old"}
		got, err := r.AppendValues(dst)
		if !errors.Is(err, tt.want) || !slices.Equal(got, dst) {
			t.Errorf("%s: AppendValues = %q, %v, ожидалось [old] и %v", tt.name, got, err, tt.want)
		}
	}
}
