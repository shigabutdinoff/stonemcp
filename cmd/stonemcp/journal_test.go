package main

import (
	"errors"
	"io"
	"testing"
)

func TestJournalUnreadDoesNotBlock(t *testing.T) {
	t.Parallel()
	r, w := io.Pipe()
	defer r.Close()
	j := newJournal(w)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 1000 {
			j.Write([]byte("строка\n"))
		}
		j.shutdown()
	}()
	within(t, done, "журнал заблокировал запись, пока его никто не читает")
}

func TestJournalWriteAfterClose(t *testing.T) {
	j := newJournal(io.Discard)
	j.shutdown()
	if _, err := j.Write([]byte("строка\n")); err != nil {
		t.Errorf("запись после закрытия журнала вернула %v", err)
	}
	j.shutdown()
}

func runUnread(t *testing.T, stdin io.ReadCloser, args ...string) *outcome {
	t.Helper()
	errR, errW := io.Pipe()
	e := runAsync(args, stdin, &buffer{}, errW)
	t.Cleanup(func() {
		errR.Close()
		<-e.done
	})
	return e
}

func TestServeWithUnreadJournal(t *testing.T) {
	t.Parallel()
	ln := listen(t)
	inR, inW := io.Pipe()
	defer inW.Close()
	e := runUnread(t, inR, "--addr", ln.Addr().String())
	conn := accept(t, ln)
	inW.Close()
	e.wantReleased(t, conn)
}

func TestServeErrorWithUnreadJournal(t *testing.T) {
	t.Parallel()
	runUnread(t, brokenInput(), "--addr", noGame).want(t, 1)
}

func TestFlagErrorWithUnreadJournal(t *testing.T) {
	t.Parallel()
	runUnread(t, nil, "--mode", "turbo").want(t, 2)
}

type failingWriter struct{ calls int }

func (w *failingWriter) Write([]byte) (int, error) {
	w.calls++
	return 0, errors.New("вывод закрыт")
}

func TestJournalStopsAfterWriteError(t *testing.T) {
	w := &failingWriter{}
	j := newJournal(w)
	for range 3 {
		j.Write([]byte("строка\n"))
	}
	j.shutdown()
	within(t, j.done, "журнал не дочитал строки после ошибки вывода")
	if w.calls != 1 {
		t.Errorf("запись в вывод после ошибки: вызовов %d, ожидался 1", w.calls)
	}
}
