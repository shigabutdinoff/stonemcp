package mcpserver

import "context"

// FreezeIn задаёт лимит заморозки.
type FreezeIn struct {
	// LimitSeconds задаёт лимит заморозки в секундах.
	LimitSeconds float64 `json:"limit_seconds,omitempty" jsonschema:"через сколько секунд игра продолжит сама, по умолчанию 30, максимум 120"`
}

// ReleaseOut содержит итог отпускания игры.
type ReleaseOut struct {
	// Released сообщает, была ли игра заморожена.
	Released bool `json:"released" jsonschema:"false, если игра не была заморожена"`
	// Frame содержит кадр, на котором стояла игра.
	Frame int64 `json:"frame" jsonschema:"кадр заморозки; игра продолжает со следующего"`
	noticed
}

func (h *handler) freeze(ctx context.Context, in FreezeIn) (SnapshotOut, error) {
	if err := h.eng.Ticking(h.StaleAfter); err != nil {
		return SnapshotOut{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()
	return snapshotOut(h.eng.Freeze(ctx, seconds(in.LimitSeconds)))
}

func (h *handler) release(context.Context, struct{}) (ReleaseOut, error) {
	frame, ok := h.eng.Release()
	return ReleaseOut{Released: ok, Frame: frame}, nil
}
