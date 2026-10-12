package engine

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

const probeWait = 10 * time.Millisecond

func (s *session) exchange(reqs []mc.Request) ([]mc.Value, error) {
	batches, err := mc.SplitBatches(reqs, s.cfg.MaxPacket)
	if err != nil {
		return nil, err
	}
	return s.exchangeBatches(batches)
}

func (s *session) exchangeBatches(batches [][]mc.Request) ([]mc.Value, error) {
	n := 0
	for _, batch := range batches {
		n += len(batch)
	}
	values := make([]mc.Value, 0, n)
	for _, batch := range batches {
		s.id = mc.NextID(s.id)
		packet, reader, err := mc.EncodeBatch(s.id, batch, s.cfg.MaxPacket)
		if err != nil {
			return nil, err
		}
		s.conn.SetDeadline(time.Now().Add(s.cfg.ReadTimeout))
		if err := s.write(packet); err != nil {
			return nil, fmt.Errorf("%w: отправка пакета %d", err, s.id)
		}
		if err := s.response(reader); err != nil {
			return nil, err
		}
		if values, err = reader.AppendValues(values); err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (s *session) response(reader *mc.ResponseReader) error {
	for {
		n, err := s.conn.Read(s.buf)
		if n > 0 {
			if done, feedErr := reader.Feed(s.buf[:n]); done || feedErr != nil {
				return feedErr
			}
		}
		if errors.Is(err, os.ErrDeadlineExceeded) {
			err = reader.Stalled()
			return err
		}
		if err != nil {
			return fmt.Errorf("%w: чтение ответа на пакет %d", err, s.id)
		}
	}
}

func (s *session) write(packet []byte) error {
	n, err := s.conn.Write(packet)
	if err == nil && n != len(packet) {
		return io.ErrShortWrite
	}
	return err
}

func (s *session) probe() error {
	s.conn.SetReadDeadline(time.Now().Add(probeWait))
	n, err := s.conn.Read(s.buf[:1])
	if n > 0 {
		return fmt.Errorf("игра прислала данные во время заморозки")
	}
	if err == nil || errors.Is(err, os.ErrDeadlineExceeded) {
		return nil
	}
	return fmt.Errorf("%w: проверка связи в заморозке", err)
}
