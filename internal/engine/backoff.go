package engine

import (
	"context"
	"slices"
	"time"
)

type backoff struct {
	delays []time.Duration
	next   int
}

func (b *backoff) wait() time.Duration {
	d := b.delays[b.next]
	b.next = min(b.next+1, len(b.delays)-1)
	return d
}

func (b *backoff) reset() { b.next = 0 }

func (b *backoff) giveUp() { b.next = len(b.delays) - 1 }

func validDelays(delays []time.Duration) bool {
	return len(delays) > 0 && !slices.ContainsFunc(delays, func(d time.Duration) bool { return d <= 0 })
}

func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
