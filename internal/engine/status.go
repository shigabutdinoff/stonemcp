package engine

import "time"

// Status содержит состояние соединения и последнюю причину сбоя.
type Status struct {
	// Connected сообщает, принят ли сигнал поддержанного режима.
	Connected bool `json:"connected"`
	// Mode содержит поддерживаемый режим.
	Mode string `json:"mode"`
	// ProtocolVersion содержит последнюю объявленную версию протокола.
	ProtocolVersion string `json:"protocol_version"`
	// Frame содержит номер последнего опубликованного кадра.
	Frame int64 `json:"frame"`
	// FPS содержит среднюю частоту опубликованных кадров с начала текущего Run.
	FPS float64 `json:"fps"`
	// LocID содержит идентификатор последней локации.
	LocID string `json:"loc_id"`
	// LastError содержит последнюю причину отсутствия связи.
	LastError string `json:"last_error"`
}

// Status возвращает независимую копию состояния соединения.
func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.status
}

// Snapshot возвращает копию последнего состояния с текущим возрастом.
// До первого кадра возвращается ErrNotTicking, без связи ErrDisconnected.
func (e *Engine) Snapshot() (*Snapshot, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.snapshotLocked()
}

// Ticking возвращает ErrNotTicking, если последний кадр старше maxAge
// без учёта заморозки и игра не заморожена. Без связи ErrDisconnected.
func (e *Engine) Ticking(maxAge time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if err := e.apiErrorLocked(); err != nil {
		return err
	}
	if e.snapshot == nil || e.freezer.age(e.snapshot.taken, time.Now()) > maxAge {
		return ErrNotTicking
	}
	return nil
}

func (e *Engine) snapshotLocked() (*Snapshot, error) {
	if err := e.apiErrorLocked(); err != nil {
		return nil, err
	}
	if e.snapshot == nil {
		return nil, ErrNotTicking
	}
	return e.snapshot.aged(time.Now()), nil
}
