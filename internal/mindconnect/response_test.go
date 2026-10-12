package mindconnect

import (
	"errors"
	"slices"
	"testing"
)

func feedCut(t *testing.T, r *ResponseReader, full string, cut int) {
	t.Helper()
	if done, err := r.Feed([]byte(full[:cut])); done || err != nil {
		t.Fatalf("разрез %d: done=%v err=%v после первой части %q", cut, done, err, full[:cut])
	}
	if done, err := r.Feed([]byte(full[cut:])); !done || err != nil {
		t.Fatalf("разрез %d: done=%v err=%v после второй части", cut, done, err)
	}
}

func TestResponseReaderSingleRead(t *testing.T) {
	r := NewResponseReader(1, 8)
	done, err := r.Feed([]byte("1\x1f1\x1f80\x1f80\x1frocky_plateau\x1f\x1f9999\x1f-2\x1fTrue"))
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("ответ не завершён")
	}
	id, values, err := r.Response()
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 || len(values) != 8 {
		t.Fatalf("id=%d len=%d", id, len(values))
	}
	if values[3] != "rocky_plateau" || values[4] != "" || values[7] != "True" {
		t.Errorf("values = %q", values)
	}
	again, values2, err := r.Response()
	if err != nil || again != id || len(values2) != len(values) || values2[3] != values[3] {
		t.Errorf("повторный Response: id=%d values=%q err=%v", again, values2, err)
	}
}

func TestResponseReaderSplitAnywhere(t *testing.T) {
	for _, tt := range []struct {
		name     string
		id, want int
		full     string
		values   []Value
	}{
		{"plain values", 17, 3, "17\x1fcmd_done\x1f80\x1fFalse", []Value{CmdDone, "80", "False"}},
		{"error in the sentinel", 7, 2, "7\x1f80\x1ffnc_err1", []Value{"80", "fnc_err1"}},
		{"prt_err as a value", 7, 3, "7\x1f80\x1fprt_err\x1fTrue", []Value{"80", PrtErr, "True"}},
		{"prt_err prefix in a value", 7, 2, "7\x1fprt_err_zone\x1fTrue", []Value{"prt_err_zone", "True"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for cut := 1; cut < len(tt.full); cut++ {
				r := NewResponseReader(tt.id, tt.want)
				feedCut(t, r, tt.full, cut)
				id, values, err := r.Response()
				if err != nil || id != tt.id || !slices.Equal(values, tt.values) {
					t.Fatalf("разрез %d: id=%d values=%q err=%v", cut, id, values, err)
				}
			}
		})
	}
}

func TestResponseReaderErrors(t *testing.T) {
	for _, tt := range []struct {
		name     string
		id, want int
		in       string
	}{
		{"overlong response", 5, 1, "5\x1fTrue\x1fextra"},
		{"bad id", 5, 1, "x\x1fTrue"},
		{"foreign id", 7, 2, "8\x1f80"},
		{"garbage after the sentinel", 7, 2, "7\x1f80\x1fTrue\x06pre0.3"},
		{"control byte inside a value", 7, 2, "7\x1f8\x030\x1fTrue"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResponseReader(tt.id, tt.want)
			if done, err := r.Feed([]byte(tt.in)); !done || !errors.Is(err, ErrMalformedResponse) {
				t.Fatalf("done=%v err=%v, ожидалось true и ErrMalformedResponse", done, err)
			}
			if _, _, err := r.Response(); err == nil {
				t.Error("Response: ожидалось, что ошибка сохранится")
			}
		})
	}
}

func TestNewResponseReaderRejectsBadArguments(t *testing.T) {
	for _, tt := range []struct{ id, want int }{{5, 0}, {5, -1}, {0, 2}, {-1, 2}, {MaxID + 1, 2}} {
		r := NewResponseReader(tt.id, tt.want)
		if done, err := r.Feed([]byte("True")); !done || !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("NewResponseReader(%d, %d): Feed done=%v err=%v, ожидалось ErrInvalidRequest", tt.id, tt.want, done, err)
		}
		if err := r.Stalled(); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("NewResponseReader(%d, %d): Stalled err=%v, ожидалось ErrInvalidRequest", tt.id, tt.want, err)
		}
		if _, _, err := r.Response(); !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("NewResponseReader(%d, %d): Response err=%v, ожидалось ErrInvalidRequest", tt.id, tt.want, err)
		}
	}
}

func TestResponseReaderResponseIsACopy(t *testing.T) {
	r := NewResponseReader(7, 2)
	if done, err := r.Feed([]byte("7\x1f80\x1fTrue")); !done || err != nil {
		t.Fatalf("done=%v err=%v", done, err)
	}
	_, got, _ := r.Response()
	got[0] = "changed"
	_ = append(got[:1], "changed")
	if _, again, _ := r.Response(); again[0] != "80" || again[1] != "True" {
		t.Errorf("Response после записи вызывающего в свой срез = %q, ожидалось, что читатель не изменился", again)
	}
}
