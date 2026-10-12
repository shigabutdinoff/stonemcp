package main

import (
	"bytes"
	"io"
	"sync"
	"time"
)

const journalLines = 64

var journalGrace = time.Second

type journal struct {
	mu     sync.Mutex
	closed bool
	lines  chan []byte
	done   chan struct{}
}

func newJournal(w io.Writer) *journal {
	j := &journal{lines: make(chan []byte, journalLines), done: make(chan struct{})}
	go func() {
		defer close(j.done)
		var failed error
		for line := range j.lines {
			if failed == nil {
				_, failed = w.Write(line)
			}
		}
	}()
	return j
}

func (j *journal) Write(p []byte) (int, error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if !j.closed && len(j.lines) < cap(j.lines) {
		j.lines <- bytes.Clone(p)
	}
	return len(p), nil
}

func (j *journal) shutdown() {
	j.mu.Lock()
	if !j.closed {
		j.closed = true
		close(j.lines)
	}
	j.mu.Unlock()
	select {
	case <-j.done:
	case <-time.After(journalGrace):
	}
}
