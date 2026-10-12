package engine

import (
	"maps"
	"slices"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

type heldCommand struct {
	cmd   Command
	slot  string
	until time.Time
}

type holdTable map[string]heldCommand

func (c Command) slot() string {
	if slot := commandSpecs[c.Name].slot; slot != nil {
		return slot(c)
	}
	return ""
}

func slotsConflict(a Command, slotA string, b Command, slotB string) bool {
	if slotA == "" || slotB == "" {
		return false
	}
	return slotA == slotB || commandSpecs[a.Name].hands&commandSpecs[b.Name].hands != 0
}

func (t holdTable) evict(cmd Command) {
	slot := cmd.slot()
	for key, held := range t {
		if slotsConflict(cmd, slot, held.cmd, held.slot) {
			delete(t, key)
		}
	}
}

func (t holdTable) accept(cmd Command, now time.Time) {
	t.evict(cmd)
	if slot := cmd.slot(); slot != "" && cmd.Hold > 0 {
		t[slot] = heldCommand{cmd: cmd, slot: slot, until: now.Add(cmd.Hold)}
	}
}

func (t holdTable) prune(now time.Time) {
	for slot, held := range t {
		if !now.Before(held.until) {
			delete(t, slot)
		}
	}
}

func (t holdTable) current(h heldCommand) bool { return t[h.slot] == h }

func (t holdTable) split(jobs []*commandJob) (ready, deferred []heldCommand) {
	for _, slot := range slices.Sorted(maps.Keys(t)) {
		held := t[slot]
		if slices.ContainsFunc(jobs, func(job *commandJob) bool { return slotsConflict(held.cmd, held.slot, job.cmd, job.slot) }) {
			deferred = append(deferred, held)
		} else {
			ready = append(ready, held)
		}
	}
	return ready, deferred
}

func (t holdTable) live(deferred []heldCommand, now time.Time) []heldCommand {
	t.prune(now)
	var live []heldCommand
	for _, h := range deferred {
		if t.current(h) {
			live = append(live, h)
		}
	}
	return live
}

func (t holdTable) dropRejected(held []heldCommand, values []mc.Value) {
	for i, h := range held {
		if values[i].IsError() && t.current(h) {
			delete(t, h.slot)
		}
	}
}

func (q *commandQueue) holdsFull(cmd Command) bool {
	slots := make(map[string]Command, len(q.holds))
	for slot, held := range q.holds {
		slots[slot] = held.cmd
	}
	for _, job := range q.jobs {
		if job.cmd.Hold > 0 {
			slots[job.slot] = job.cmd
		}
	}
	if len(slots) < q.holdLimit {
		return false
	}
	slot := cmd.slot()
	for takenSlot, taken := range slots {
		if slotsConflict(taken, takenSlot, cmd, slot) {
			return false
		}
	}
	return true
}
