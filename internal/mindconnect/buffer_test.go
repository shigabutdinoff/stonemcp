package mindconnect

import (
	"bytes"
	"errors"
	"slices"
	"testing"
)

func TestResponseReaderKeepsOnlyTheTail(t *testing.T) {
	const values, chunk = 1000, 7
	resp := responseOf(values, func(int) string { return "abc" })
	r := NewResponseReader(1, values+1)
	for part := range slices.Chunk(resp, chunk) {
		if _, err := r.Feed(part); err != nil {
			t.Fatal(err)
		}
		if len(r.buf) > chunk+len("True") {
			t.Fatalf("в буфере %d байт, ожидался только недочитанный хвост", len(r.buf))
		}
	}
	if _, got, err := r.Response(); err != nil || len(got) != values+1 || got[0] != "abc" || got[values] != "True" {
		t.Fatalf("значений %d, err = %v", len(got), err)
	}
}

func TestResponseReaderOverflow(t *testing.T) {
	r := NewResponseReader(1, 2)
	chunk := bytes.Repeat([]byte("x"), 4096)
	for range MaxResponse/len(chunk) + 1 {
		done, err := r.Feed(chunk)
		if err != nil {
			if !errors.Is(err, ErrMalformedResponse) {
				t.Errorf("err = %v, ожидалось ErrMalformedResponse", err)
			}
			return
		}
		if done {
			t.Fatal("мусор не должен завершать ответ")
		}
	}
	t.Fatalf("нет ошибки после %d байт, лимит %d", len(r.buf), MaxResponse)
}

func TestResponseReaderOverflowInOneChunk(t *testing.T) {
	r := NewResponseReader(1, 2)
	chunk := bytes.Repeat([]byte("x"), MaxResponse+1)
	if done, err := r.Feed(chunk); !done || !errors.Is(err, ErrMalformedResponse) {
		t.Fatalf("done=%v err=%v, ожидалось, что кусок отвергнут как искажённый", done, err)
	}
	if len(r.buf) != 0 {
		t.Errorf("в буфере %d байт, ожидалось, что лимит их не пропустит", len(r.buf))
	}
}
