package engine

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestFPSCountsOnlyCurrentRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := New(Config{})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		for _, frame := range []int64{20, 40} {
			if err := e.Run(ctx); err != nil {
				t.Fatal(err)
			}
			if status := e.Status(); status.FPS != 0 {
				t.Errorf("FPS до первого кадра нового запуска = %v, ожидался 0", status.FPS)
			}
			time.Sleep(time.Second)
			e.frame += 20
			e.publish(&Snapshot{Frame: e.frame})
			if status := e.Status(); status.Frame != frame || status.FPS != 20 {
				t.Errorf("статус запуска = %+v, ожидались кадр %d и FPS 20", status, frame)
			}
		}
	})
}

func TestFPSKeepsRunBaselineAfterReconnect(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		e := New(Config{})
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := e.Run(ctx); err != nil {
			t.Fatal(err)
		}
		e.connected("0.3")
		time.Sleep(time.Second)
		e.frame = 20
		e.publish(&Snapshot{Frame: e.frame})
		e.disconnect(ErrDisconnected)
		time.Sleep(time.Second)
		e.connected("0.3")
		time.Sleep(time.Second)
		e.frame = 30
		e.publish(&Snapshot{Frame: e.frame})
		if status := e.Status(); status.Frame != 30 || status.FPS != 10 {
			t.Fatalf("статус после переподключения = %+v, ожидались кадр 30 и FPS 10", status)
		}
	})
}

func TestRunRestartPreservesFrameSequence(t *testing.T) {
	g := fakegame.Start(t)
	e := New(Config{Addr: g.Addr()})
	var previous int64
	for range 2 {
		stop := runEngine(t, e)
		func() {
			defer stop()
			eventually(t, "кадр после запуска", func() bool {
				s, err := e.Snapshot()
				return err == nil && s.Frame > previous
			})
		}()
		previous = e.Status().Frame
	}
}
