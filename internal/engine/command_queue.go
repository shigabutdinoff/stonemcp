package engine

import (
	"slices"
	"time"
)

type commandQueue struct {
	jobs       []*commandJob
	holds      holdTable
	queueLimit int
	holdLimit  int
	budget     int
	clock      func() time.Time
}

func newCommandQueue(cfg Config) commandQueue {
	return commandQueue{holds: make(holdTable), queueLimit: cfg.QueueLimit, holdLimit: cfg.HoldLimit, budget: cfg.FrameBudget, clock: time.Now}
}

func (q *commandQueue) add(job *commandJob) error {
	if len(q.jobs) >= q.queueLimit {
		return ErrQueueFull
	}
	q.holds.prune(q.clock())
	if job.cmd.Hold > 0 && q.holdsFull(job.cmd) {
		return ErrHoldsFull
	}
	q.jobs = append(q.jobs, job)
	return nil
}

func (q *commandQueue) remove(job *commandJob) (commandOutcome, bool) {
	job.finished = true
	q.jobs = removeJob(q.jobs, job)
	return tryReceive(job.done)
}

func (q *commandQueue) sweep() {
	q.jobs = slices.DeleteFunc(q.jobs, func(job *commandJob) bool { return job.finished })
}

func (q *commandQueue) extend(prev, next *Snapshot, stall time.Duration) {
	if prev == nil {
		return
	}
	gap := next.taken.Sub(prev.taken)
	if gap <= stall {
		return
	}
	for _, job := range q.jobs {
		if !job.sent {
			job.expires = job.expires.Add(gap)
		}
	}
}

func (q *commandQueue) publish(snapshot *Snapshot, now time.Time, frozen bool) {
	for _, job := range q.jobs {
		if job.sent && (frozen || snapshot.Frame > job.frame) {
			job.finish(commandOutcome{result: CommandResult{Result: job.response, Snapshot: snapshot.aged(now), Frozen: frozen}})
		}
	}
	q.sweep()
}

func (q *commandQueue) liveHolds(deferred []heldCommand) []heldCommand {
	return q.holds.live(deferred, q.clock())
}

func (q *commandQueue) abort(err error) {
	for _, job := range q.jobs {
		job.finish(commandOutcome{result: CommandResult{Result: job.response}, err: err})
	}
	q.jobs = nil
	clear(q.holds)
}

func (q *commandQueue) take(guard func(*commandJob) error) (jobs []*commandJob, ready, deferred []heldCommand) {
	q.holds.prune(q.clock())
	budget := max(q.budget-len(q.holds), 0)
	for _, job := range q.jobs {
		if job.inFlight {
			continue
		}
		if stopped(job.stop) {
			job.finished = true
			continue
		}
		if err := guard(job); err != nil {
			job.finish(commandOutcome{err: err})
			continue
		}
		if len(jobs) >= budget {
			continue
		}
		job.inFlight, jobs = true, append(jobs, job)
	}
	q.sweep()
	ready, deferred = q.holds.split(jobs)
	return jobs, ready, deferred
}

func (q *commandQueue) delivered(job *commandJob, response string, frame int64) {
	if !stopped(job.stop) {
		q.holds.accept(job.cmd, q.clock())
	} else {
		q.holds.evict(job.cmd)
	}
	job.response, job.frame, job.sent = response, frame, true
}
