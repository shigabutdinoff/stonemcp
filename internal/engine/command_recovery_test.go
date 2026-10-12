package engine

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSkippedReplacementRestoresHold(t *testing.T) {
	for _, tt := range []struct {
		name        string
		hold        time.Duration
		act         func(e *Engine, cancel context.CancelFunc, until time.Time)
		wantErr     error
		wantRepeats int
	}{
		{"canceled", time.Minute, func(_ *Engine, cancel context.CancelFunc, _ time.Time) { cancel() }, context.Canceled, 1},
		{"expired", time.Minute, func(e *Engine, _ context.CancelFunc, _ time.Time) { expireQueued(e) }, ErrExpired, 1},
		{"hold_expired", 100 * time.Millisecond, func(_ *Engine, cancel context.CancelFunc, until time.Time) {
			cancel()
			<-time.After(time.Until(until))
		}, context.Canceled, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, g := pausedCommandEngine(t, Config{})
			until := time.Now().Add(tt.hold)
			prepareHoldRecovery(t, e, until)
			ctx, cancel := commandContext(t, 0)
			done := doAsync(ctx, e, Equip("bardiche", 0), 0)
			eventually(t, "команда в очереди", func() bool { return queuedCount(e) == 1 })
			tt.act(e, cancel, until)
			frame := g.Frame() + 1
			g.Resume()
			if err := receive(t, done, "ожидание замены не завершилось").err; !errors.Is(err, tt.wantErr) {
				t.Fatalf("ошибка замены = %v, ожидалась %v", err, tt.wantErr)
			}
			assertRecoveredHold(t, g, frame, tt.wantRepeats)
			if len(g.CommandsNamed("equip")) != 0 {
				t.Fatalf("отменённая замена дошла до игры, команды = %+v", g.Commands())
			}
		})
	}
}

func TestSecondReplacementDecidesHoldRecovery(t *testing.T) {
	for _, tt := range []struct {
		name        string
		reject      bool
		wantRepeats int
	}{
		{"accepted", false, 0},
		{"rejected", true, 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, g := pausedCommandEngine(t, Config{})
			prepareHoldRecovery(t, e, time.Now().Add(time.Minute))
			ctx, cancel := commandContext(t, 0)
			done := doAsync(ctx, e, Equip("bardiche", 0), 0)
			eventually(t, "команда в очереди", func() bool { return queuedCount(e) == 1 })
			if tt.reject {
				g.Reject("equipR", true)
			}
			second := doAsync(testContext(t), e, EquipRight("axe", 0), 0)
			eventually(t, "вторая команда в очереди", func() bool { return queuedCount(e) == 2 })
			cancel()
			frame := g.Frame() + 1
			g.Resume()
			if err := receive(t, done, "ожидание замены не завершилось").err; !errors.Is(err, context.Canceled) {
				t.Fatalf("ошибка замены = %v, ожидалась отмена", err)
			}
			err := receive(t, second, "ожидание второй замены не завершилось").err
			var gameErr *GameError
			if (!tt.reject && err != nil) || (tt.reject && !errors.As(err, &gameErr)) {
				t.Fatalf("ошибка второй замены = %v", err)
			}
			assertRecoveredHold(t, g, frame, tt.wantRepeats)
		})
	}
}
