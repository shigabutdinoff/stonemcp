package engine

import (
	"errors"
	"testing"
	"time"
)

func TestHoldLimitRejectsNewSlots(t *testing.T) {
	q := newCommandQueue(Config{HoldLimit: 2}.defaults())
	q.holds.accept(EquipLeft("sword", time.Minute), time.Now())
	q.holds.accept(EquipRight("shield", time.Minute), time.Now())
	if err := addJob(&q, Command{Name: ">", Arg: "готов", Hold: time.Minute}); !errors.Is(err, ErrHoldsFull) {
		t.Fatalf("ошибка нового удержания = %v, ожидалось %v", err, ErrHoldsFull)
	}
	if err := addJob(&q, EquipLeft("axe", time.Minute)); err != nil {
		t.Fatalf("занятый слот отвергнут, err = %v", err)
	}
	if err := addJob(&q, Loadout(2, time.Minute)); err != nil {
		t.Fatalf("вытесняющее удержание отвергнуто, err = %v", err)
	}
	if err := addJob(&q, Activate("potion")); err != nil {
		t.Fatalf("команда без удержания отвергнута, err = %v", err)
	}
}

func TestHoldLimitCountsQueuedCommands(t *testing.T) {
	q := newCommandQueue(Config{HoldLimit: 2}.defaults())
	if err := addJob(&q, EquipLeft("sword", time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := addJob(&q, EquipRight("shield", time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := addJob(&q, Command{Name: ">", Arg: "готов", Hold: time.Minute}); !errors.Is(err, ErrHoldsFull) {
		t.Fatalf("ошибка нового удержания = %v, ожидалось %v", err, ErrHoldsFull)
	}
	if err := addJob(&q, Loadout(2, time.Minute)); err != nil {
		t.Fatalf("вытесняющее удержание отвергнуто, err = %v", err)
	}
}

func TestHoldLimitCountsSentCommands(t *testing.T) {
	q := newCommandQueue(Config{HoldLimit: 1}.defaults())
	if err := addJob(&q, EquipLeft("sword", time.Minute)); err != nil {
		t.Fatal(err)
	}
	q.take(admitAll)
	if err := addJob(&q, EquipRight("shield", time.Minute)); !errors.Is(err, ErrHoldsFull) {
		t.Fatalf("ошибка нового удержания = %v, ожидалось %v", err, ErrHoldsFull)
	}
}
