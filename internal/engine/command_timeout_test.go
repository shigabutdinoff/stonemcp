package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestCommandWaitExpiresWithoutFrames(t *testing.T) {
	for _, tt := range []struct {
		name                       string
		protection, wait, deadline time.Duration
		want                       time.Duration
	}{
		{"extended", 40 * time.Millisecond, 200 * time.Millisecond, 0, 200 * time.Millisecond},
		{"shortened", 200 * time.Millisecond, 40 * time.Millisecond, 0, 200 * time.Millisecond},
		{"deadline", time.Second, 2 * time.Second, 100 * time.Millisecond, 100 * time.Millisecond},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := fakegame.Start(t)
			e := startEngine(t, g, Config{CommandTimeout: tt.protection})
			paused := g.PauseAfterCommand("activate")
			started := time.Now()
			ctx, _ := commandContext(t, tt.deadline)
			done := doAsync(ctx, e, Activate("potion"), tt.wait)
			receive(t, paused, "не дождались паузы после активации")
			limit := time.NewTimer(time.Until(started.Add(tt.want + 500*time.Millisecond)))
			defer limit.Stop()
			select {
			case got := <-done:
				if !errors.Is(got.err, ErrNotTicking) || got.result.Snapshot != nil {
					t.Fatalf("результат ожидания без кадров = %+v", got)
				}
				if elapsed := time.Since(started); elapsed < tt.want {
					t.Fatalf("ожидание завершилось через %v, раньше %v", elapsed, tt.want)
				}
			case <-limit.C:
				t.Fatal("ожидание превысило выбранный предел")
			}
			g.Resume()
			skipFrames(t, g, 2, "кадры после истечения ожидания")
			if len(g.CommandsNamed("activate")) != 1 {
				t.Fatal("истечение ожидания вызвало повторную активацию")
			}
		})
	}
}

func TestDeliveredOutcomeSurvivesExpiredWait(t *testing.T) {
	for name, wait := range map[string]func(*testing.T) context.Context{
		"deadline": expiredContext,
		"canceled": canceledContext,
	} {
		t.Run(name, func(t *testing.T) {
			waitCtx := wait(t)
			job := newJob(waitCtx, Command{})
			want := CommandResult{Result: "cmd_done", Snapshot: &Snapshot{Frame: 7}}
			job.done <- commandOutcome{result: want}
			got, err := New(Config{}).awaitCommand(waitCtx, job)
			if err != nil {
				t.Fatalf("доставленный результат потерян, err = %v", err)
			}
			if got.Result != want.Result || got.Snapshot != want.Snapshot {
				t.Fatalf("результат = %+v, ожидалось %+v", got, want)
			}
		})
	}
}

func TestRemovedCommandReturnsLateOutcome(t *testing.T) {
	q := newCommandQueue(Config{}.defaults())
	job := newJob(context.Background(), Activate("potion"))
	q.jobs = []*commandJob{job}
	want := CommandResult{Result: "cmd_done", Snapshot: &Snapshot{Frame: 9}}
	job.done <- commandOutcome{result: want}
	outcome, ok := q.remove(job)
	if !ok || outcome.result.Snapshot != want.Snapshot {
		t.Fatalf("поздний результат потерян = %+v, ok = %v", outcome, ok)
	}
	if len(q.jobs) != 0 {
		t.Fatal("задание осталось в очереди")
	}
}
