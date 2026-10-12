package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestSentCommandCancellation(t *testing.T) {
	for _, tt := range []interruptCase{
		{"cancel", 0, cancelWait, context.Canceled},
		{"deadline", 40 * time.Millisecond, keepWaiting, ErrNotTicking},
		{"protection", 0, keepWaiting, ErrNotTicking},
		{"disconnect", 0, dropLink, ErrDisconnected},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := fakegame.Start(t)
			e := startEngine(t, g, Config{CommandTimeout: 100 * time.Millisecond})
			paused := g.PauseAfterCommand("activate")
			ctx, cancel := commandContext(t, tt.deadline)
			done := doAsync(ctx, e, Activate("potion"), 0)
			receive(t, paused, "не дождались паузы после активации")
			tt.act(g, cancel)
			if err := receive(t, done, "ожидание не прервано").err; !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, ожидалось %v", err, tt.want)
			}
			g.Resume()
			skipFrames(t, g, 2, "продолжение кадров")
			if len(g.CommandsNamed("activate")) != 1 {
				t.Fatal("отменённая команда отправлена повторно")
			}
		})
	}
}

func TestShutdownReleasesCommandsAndHolds(t *testing.T) {
	g := fakegame.Start(t)
	e := New(Config{Addr: g.Addr()})
	stop := runEngine(t, e)
	t.Cleanup(stop)
	eventually(t, "первый кадр", func() bool { _, err := e.Snapshot(); return err == nil })
	if _, err := e.Do(testContext(t), EquipLeft("sword", time.Minute), 0); err != nil {
		t.Fatal(err)
	}
	paused := g.PauseAfterCommand("activate")
	sent := doAsync(context.Background(), e, Activate("potion"), 0)
	receive(t, paused, "не дождались паузы после активации")
	queued := doAsync(context.Background(), e, EquipRight("shield", 0), 0)
	eventually(t, "ожидающая команда", func() bool { return queuedCount(e) == 1 })
	stop()
	for _, done := range []<-chan commandOutcome{sent, queued} {
		if err := receive(t, done, "shutdown не прервал ожидание").err; !errors.Is(err, ErrDisconnected) {
			t.Fatalf("err = %v", err)
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.commands.holds) != 0 || len(e.commands.jobs) != 0 {
		t.Fatal("shutdown оставил команды или удержания")
	}
	if len(g.CommandsNamed("equipR")) != 0 {
		t.Fatal("shutdown отправил ожидающую команду")
	}
}
