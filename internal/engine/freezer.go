package engine

import (
	"fmt"
	"time"
)

type freezeOutcome struct {
	snapshot *Snapshot
	err      error
}

type freezeCall struct {
	limit time.Duration
	done  chan freezeOutcome
}

type frozenState struct {
	started  time.Time
	limit    time.Duration
	released bool
}

type freezer struct {
	calls     []*freezeCall
	frozen    *frozenState
	stepFrame int64
	thawed    time.Time
	paused    time.Duration
	notice    string
}

func (f *freezer) want(limit time.Duration) *freezeCall {
	if limit == 0 || f.active() {
		return nil
	}
	call := &freezeCall{limit: limit, done: make(chan freezeOutcome, 1)}
	f.calls = append(f.calls, call)
	return call
}

func (f *freezer) drop(call *freezeCall) (freezeOutcome, bool) {
	f.calls = removeJob(f.calls, call)
	return tryReceive(call.done)
}

func (f *freezer) active() bool { return f.frozen != nil && !f.frozen.released }

func (f *freezer) missed(frame int64) bool { return !f.active() && f.stepFrame >= frame }

func (f *freezer) start(s *Snapshot, now time.Time) time.Duration {
	f.stepFrame = s.Frame
	if len(f.calls) == 0 {
		return 0
	}
	f.frozen = &frozenState{started: now}
	for _, call := range f.calls {
		f.frozen.limit = max(f.frozen.limit, call.limit)
		call.done <- freezeOutcome{snapshot: s.aged(now)}
	}
	f.calls = nil
	return f.frozen.limit
}

func (f *freezer) release() (int64, bool) {
	if !f.active() {
		return 0, false
	}
	f.frozen.released = true
	return f.stepFrame, true
}

func (f *freezer) expire() {
	if frame, ok := f.release(); ok {
		f.notice = fmt.Sprintf("игра продолжила сама: заморозка на кадре %d снята по лимиту %v с", frame, f.frozen.limit.Seconds())
	}
}

func (f *freezer) finish(now time.Time) {
	f.paused += now.Sub(f.frozen.started)
	f.thawed, f.frozen = now, nil
}

func (f *freezer) fail(err error) {
	for _, call := range f.calls {
		call.done <- freezeOutcome{err: err}
	}
	f.calls = nil
}

func (f *freezer) takeNotice() string {
	notice := f.notice
	f.notice = ""
	return notice
}

func (f *freezer) clock(now time.Time) time.Time {
	if f.frozen != nil {
		now = f.frozen.started
	}
	return now.Add(-f.paused)
}

func (f *freezer) age(taken, now time.Time) time.Duration {
	if f.frozen != nil {
		return 0
	}
	if f.thawed.After(taken) {
		taken = f.thawed
	}
	return now.Sub(taken)
}
