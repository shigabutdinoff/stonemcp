package engine

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

var endStep = []byte(mc.EndStep)

type session struct {
	conn        net.Conn
	buf         []byte
	id          int
	cfg         Config
	screenFrame int64
}

func (e *Engine) session(ctx context.Context, conn net.Conn) error {
	s := &session{conn: conn, buf: make([]byte, mc.MaxPacket), cfg: e.cfg}
	expected := mc.PhasePre
	for {
		phase, version, err := s.signal()
		if err != nil {
			return err
		}
		if phase != expected {
			return fmt.Errorf("нарушен порядок шагов игры, получен %s", phase)
		}
		e.connected(version)
		if phase == mc.PhasePre {
			if err := e.pre(s); err != nil {
				return err
			}
			expected = mc.PhasePost
		} else {
			if err := e.post(s); err != nil {
				return err
			}
			if err := e.freezeStep(ctx, s); err != nil {
				return err
			}
			expected = mc.PhasePre
		}
		conn.SetWriteDeadline(time.Now().Add(e.cfg.ReadTimeout))
		if err := s.write(endStep); err != nil {
			return fmt.Errorf("%w: отправка конца шага %s", err, phase)
		}
	}
}

func (e *Engine) connected(version string) {
	e.mu.Lock()
	changed := e.status.ProtocolVersion != version
	e.status.Connected, e.status.ProtocolVersion = true, version
	e.status.LastError, e.failure = "", nil
	e.mu.Unlock()
	if changed && version != mc.Version && e.cfg.Logger != nil {
		e.cfg.Logger.Warn("незнакомая версия протокола игры", "версия", version)
	}
}

func (s *session) signal() (mc.Phase, string, error) {
	s.conn.SetReadDeadline(time.Time{})
	var signal []byte
	for len(signal) < 128 {
		n, err := s.conn.Read(s.buf[:128-len(signal)])
		if n > 0 && len(signal) == 0 {
			s.conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))
		}
		signal = append(signal, s.buf[:n]...)
		if errors.Is(err, os.ErrDeadlineExceeded) {
			return mc.ParseStalledSignal(signal)
		}
		phase, version, parseErr := mc.ParseSignal(signal)
		if errors.Is(parseErr, mc.ErrUnsupportedSignal) {
			return 0, "", ErrUnsupportedMode
		}
		if !errors.Is(parseErr, mc.ErrIncompleteSignal) {
			return phase, version, parseErr
		}
		if err != nil {
			return 0, "", fmt.Errorf("%w: чтение сигнала игры", err)
		}
	}
	return 0, "", fmt.Errorf("слишком длинный сигнал игры")
}
