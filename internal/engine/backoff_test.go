package engine

import (
	"context"
	"slices"
	"testing"
	"time"
)

func pauses(b *backoff, n int) []time.Duration {
	got := make([]time.Duration, n)
	for i := range got {
		got[i] = b.wait()
	}
	return got
}

func TestBackoffRepeatsLastDelay(t *testing.T) {
	b := backoff{delays: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}}
	want := []time.Duration{time.Second, 3 * time.Second, 5 * time.Second, 5 * time.Second, 5 * time.Second}
	if got := pauses(&b, 5); !slices.Equal(got, want) {
		t.Fatalf("паузы = %v, ожидались %v", got, want)
	}
	b.reset()
	if got := b.wait(); got != time.Second {
		t.Errorf("пауза после сброса = %v, ожидалась 1s", got)
	}
}

func TestBackoffGiveUpJumpsToLastDelay(t *testing.T) {
	b := backoff{delays: []time.Duration{time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond}}
	b.giveUp()
	want := []time.Duration{3 * time.Millisecond, 3 * time.Millisecond}
	if got := pauses(&b, 2); !slices.Equal(got, want) {
		t.Fatalf("паузы после отказа = %v, ожидались %v", got, want)
	}
}

func TestSleepStopsOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	if !sleep(ctx, time.Millisecond) {
		t.Fatal("короткая пауза прервана без отмены")
	}
	cancel()
	started := time.Now()
	if sleep(ctx, time.Hour) || time.Since(started) > time.Second {
		t.Fatal("пауза не прервана отменой")
	}
}

func TestValidDelays(t *testing.T) {
	for _, delays := range [][]time.Duration{nil, {}, {time.Second, 0}, {-time.Second}} {
		if validDelays(delays) {
			t.Errorf("приняты паузы %v", delays)
		}
	}
	if !validDelays([]time.Duration{time.Millisecond}) {
		t.Error("отвергнута положительная пауза")
	}
}
