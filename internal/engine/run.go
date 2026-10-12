package engine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// Run обслуживает игру до отмены ctx, после отмены возвращает nil.
// Повторный одновременный вызов и неверные лимиты возвращают ошибку.
func (e *Engine) Run(ctx context.Context) error {
	if !e.running.CompareAndSwap(false, true) {
		return fmt.Errorf("движок уже запущен")
	}
	defer e.running.Store(false)
	if err := e.cfg.validate(); err != nil {
		return err
	}
	e.mu.Lock()
	e.started, e.startFrame = e.freezer.clock(time.Now()), e.frame
	e.status.FPS = 0
	e.mu.Unlock()
	defer e.disconnect(ErrDisconnected)
	b := backoff{delays: e.cfg.ReconnectDelays}
	for {
		err := e.connect(ctx)
		if e.Status().Connected {
			b.reset()
		}
		if stopped(ctx.Done()) {
			return nil
		}
		if errors.Is(e.disconnect(err), ErrUnsupportedMode) {
			b.giveUp()
		}
		pause := b.wait()
		if e.cfg.Logger != nil {
			e.cfg.Logger.Warn("связь с игрой прервана", "ошибка", err, "пауза", pause)
		}
		if !sleep(ctx, pause) {
			return nil
		}
	}
}

func (e *Engine) connect(ctx context.Context) error {
	dialer := net.Dialer{Timeout: e.cfg.ReadTimeout}
	conn, err := dialer.DialContext(ctx, "tcp4", e.cfg.Addr)
	if err != nil {
		return fmt.Errorf("%w: подключение к игре", err)
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	return e.session(ctx, conn)
}
