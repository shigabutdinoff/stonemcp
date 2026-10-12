package mindconnect

import (
	"errors"
	"testing"
)

func TestResponseReaderStalled(t *testing.T) {
	r := NewResponseReader(7, 2)
	if done, err := r.Feed([]byte("7\x1f80\x1fmaybe")); done || err != nil {
		t.Fatalf("необычный ответ сторожевого запроса: done=%v err=%v, ожидалось, что читатель ждёт", done, err)
	}
	if err := r.Stalled(); err != nil {
		t.Fatalf("Stalled: err=%v, ожидалось, что ответ принят", err)
	}
	if _, values, err := r.Response(); err != nil || len(values) != 2 || values[1] != "maybe" {
		t.Errorf("values=%q err=%v", values, err)
	}

	r = NewResponseReader(7, 2)
	if _, err := r.Feed([]byte("7\x1f80\x1f")); err != nil {
		t.Fatal(err)
	}
	if err := r.Stalled(); !errors.Is(err, ErrTruncatedResponse) {
		t.Fatalf("пустой ответ сторожевого запроса: err=%v, ожидалось ErrTruncatedResponse: ни один его байт не пришёл", err)
	}

	r = NewResponseReader(7, 3)
	if _, err := r.Feed([]byte("7\x1f80")); err != nil {
		t.Fatal(err)
	}
	if err := r.Stalled(); !errors.Is(err, ErrTruncatedResponse) {
		t.Fatalf("оборван: err=%v, ожидалось ErrTruncatedResponse", err)
	}
}

func TestResponseReaderRejectedPacket(t *testing.T) {
	for _, tt := range []struct {
		name string
		id   int
		want int
		in   string
	}{
		{"single value", 5, 1, "5\x1fprt_err"},
		{"first of several", 5, 3, "5\x1fprt_err"},
		{"after values", 5, 3, "5\x1f80\x1fprt_err"},
		{"in place of the sentinel", 5, 3, "5\x1f80\x1frocky_plateau\x1fprt_err"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResponseReader(tt.id, tt.want)
			if done, err := r.Feed([]byte(tt.in)); done || err != nil {
				t.Fatalf("done=%v err=%v, ожидалось, что читатель ждёт новых значений", done, err)
			}
			if err := r.Stalled(); !errors.Is(err, ErrPacketRejected) {
				t.Fatalf("Stalled: err=%v, ожидалось ErrPacketRejected", err)
			}
			if _, _, err := r.Response(); !errors.Is(err, ErrPacketRejected) {
				t.Errorf("Response err = %v, ожидалось ErrPacketRejected", err)
			}
		})
	}
}

func TestResponseReaderBarePrtErrIsTruncation(t *testing.T) {
	r := NewResponseReader(5, 3)
	if done, err := r.Feed([]byte("prt_err")); done || err != nil {
		t.Fatalf("done=%v err=%v, ожидалось, что читатель ждёт", done, err)
	}
	if err := r.Stalled(); !errors.Is(err, ErrTruncatedResponse) {
		t.Fatalf("Stalled: err=%v, ожидался обрыв: игра всегда присылает id", err)
	}
}

func TestResponseReaderIncompleteValue(t *testing.T) {
	r := NewResponseReader(7, 2)
	if done, err := r.Feed([]byte("7\x1f80\x1fTru")); done || err != nil {
		t.Fatalf("done=%v err=%v, ожидалось, что читатель ждёт", done, err)
	}
	if _, _, err := r.Response(); !errors.Is(err, ErrIncompleteResponse) {
		t.Errorf("оборвано последнее значение: err = %v, ожидалось ErrIncompleteResponse", err)
	}
	if err := r.Stalled(); !errors.Is(err, ErrTruncatedResponse) {
		t.Fatalf("Stalled: err=%v, ожидалось, что обрезанный ответ сторожевого запроса признан оборванным", err)
	}
}
