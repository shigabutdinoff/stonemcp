package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestInvalidCommandLimits(t *testing.T) {
	for _, cfg := range []Config{{QueueLimit: -1}, {HoldLimit: -1}, {CommandTTL: -1}, {CommandTimeout: -1}} {
		if err := New(cfg).Run(context.Background()); err == nil {
			t.Fatalf("приняты неверные лимиты %+v", cfg)
		}
	}
}

func TestRequestedWaitCannotShortenProtection(t *testing.T) {
	g := fakegame.Start(t)
	e := startEngine(t, g, Config{CommandTimeout: time.Second})
	paused := g.PauseAfterCommand("activate")
	done := doAsync(testContext(t), e, Activate("potion"), 20*time.Millisecond)
	receive(t, paused, "не дождались паузы после активации")
	expectPending(t, done, 50*time.Millisecond, "запрошенное время сократило защиту")
	g.Resume()
	if err := (<-done).err; err != nil {
		t.Fatal(err)
	}
}

func TestFrameBudgetDefersCommands(t *testing.T) {
	q := newCommandQueue(Config{FrameBudget: 3}.defaults())
	q.holds.accept(EquipLeft("sword", time.Minute), time.Now())
	for range 5 {
		q.jobs = append(q.jobs, newJob(context.Background(), Activate("potion")))
	}
	if jobs, _, _ := q.take(admitAll); len(jobs) != 2 {
		t.Fatalf("команд в кадре = %d, ожидался бюджет 2", len(jobs))
	}
	if queuedIn(&q) != 3 {
		t.Fatalf("осталось в очереди = %d, ожидалось 3", queuedIn(&q))
	}
}

func TestFrameBudgetMustExceedHoldLimit(t *testing.T) {
	err := (Config{FrameBudget: 40, HoldLimit: 64}).defaults().validate()
	if err == nil {
		t.Fatal("принят бюджет кадра меньше предела удержаний")
	}
	if got := err.Error(); got != "бюджет кадра должен превышать предел удержаний" {
		t.Fatalf("сообщение о конфликте лимитов = %q", got)
	}
}

func TestQueueLimitCountsSentCommands(t *testing.T) {
	q := newCommandQueue(Config{QueueLimit: 1}.defaults())
	if err := addJob(&q, Activate("potion")); err != nil {
		t.Fatal(err)
	}
	q.take(admitAll)
	if err := addJob(&q, Activate("potion")); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("ошибка очереди = %v, ожидалось %v", err, ErrQueueFull)
	}
}

func TestDeferredCommandsStayGuarded(t *testing.T) {
	q := newCommandQueue(Config{FrameBudget: 1}.defaults())
	now := time.Now()
	fresh, expired := newJob(context.Background(), Activate("potion")), newJob(context.Background(), Activate("potion"))
	expired.expires = now.Add(-time.Second)
	q.jobs = []*commandJob{fresh, expired}
	guard := func(job *commandJob) error {
		if !now.Before(job.expires) {
			return ErrExpired
		}
		return nil
	}
	if jobs, _, _ := q.take(guard); len(jobs) != 1 {
		t.Fatalf("команд в кадре = %d, ожидалась одна", len(jobs))
	}
	if err := (<-expired.done).err; !errors.Is(err, ErrExpired) {
		t.Fatalf("ошибка отложенной команды = %v, ожидалось %v", err, ErrExpired)
	}
	if queuedIn(&q) != 0 {
		t.Fatalf("осталось в очереди = %d, ожидалось 0", queuedIn(&q))
	}
}

func TestExpiredHoldFreesSlotWithoutFrames(t *testing.T) {
	q := newCommandQueue(Config{HoldLimit: 1}.defaults())
	q.holds.accept(EquipLeft("sword", time.Minute), time.Now().Add(-2*time.Minute))
	if err := addJob(&q, EquipRight("shield", time.Minute)); err != nil {
		t.Fatalf("просроченное удержание заняло слот, err = %v", err)
	}
	if len(q.holds) != 0 {
		t.Fatalf("просроченное удержание осталось: %+v", q.holds)
	}
}
