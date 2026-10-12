package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestCanceledCommandLeavesNoHold(t *testing.T) {
	q := newCommandQueue(Config{}.defaults())
	q.delivered(newJob(canceledContext(t), EquipLeft("sword", time.Minute)), "cmd_done", 1)
	if len(q.holds) != 0 {
		t.Fatalf("удержание принято для отменённой команды: %+v", q.holds)
	}
}

func TestDeliveredCommandDropsConflictingHold(t *testing.T) {
	q := newCommandQueue(Config{}.defaults())
	q.holds.accept(EquipRight("shield", time.Minute), time.Now())
	q.delivered(newJob(canceledContext(t), Equip("bardiche", time.Minute)), "cmd_done", 1)
	if len(q.holds) != 0 {
		t.Fatalf("удержания после исполненной команды: %+v", q.holds)
	}
}

func TestMissingSnapshotDropsHold(t *testing.T) {
	e := New(Config{})
	job := newJob(context.Background(), EquipLeft("sword", time.Minute))
	e.applyResults(nil, []*commandJob{job}, []mc.Value{"cmd_done"})
	if err := (<-job.done).err; !errors.Is(err, ErrDisconnected) {
		t.Fatalf("ошибка ответа без снимка = %v, ожидалось %v", err, ErrDisconnected)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.commands.holds) != 0 {
		t.Fatalf("удержание принято без снимка: %+v", e.commands.holds)
	}
}

func TestGameErrorKeepsReplacedHold(t *testing.T) {
	now := time.Now()
	holds := holdTable{}
	holds.accept(EquipLeft("sword", time.Minute), now)
	stale := heldCommand{cmd: EquipLeft("bow", time.Minute), slot: "equipL", until: now.Add(time.Minute)}
	holds.dropRejected([]heldCommand{stale}, []mc.Value{"cmd_err"})
	if held := holds["equipL"]; held.cmd.Arg != "sword" {
		t.Fatalf("удержание слота equipL = %+v, ожидался sword", held)
	}
}

func TestExpiredHoldIsNotRepeated(t *testing.T) {
	q := newCommandQueue(Config{}.defaults())
	q.holds.accept(EquipRight("shield", time.Minute), time.Now().Add(-2*time.Minute))
	_, ready, deferred := q.take(admitAll)
	if len(ready) != 0 || len(deferred) != 0 {
		t.Fatalf("просроченное удержание повторено: ready = %+v, deferred = %+v", ready, deferred)
	}
	if len(q.holds) != 0 {
		t.Fatalf("просроченное удержание осталось: %+v", q.holds)
	}
}
