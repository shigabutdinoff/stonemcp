package mcpserver

import (
	"context"
	"math"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

// StatusOut содержит ответ stone_status.
type StatusOut struct {
	// ServerVersion содержит версию сервера.
	ServerVersion string `json:"server_version"`
	// Status содержит состояние соединения с игрой.
	engine.Status
	// Hint объясняет, как подключить игру, и пуст при связи.
	Hint string `json:"hint" jsonschema:"как подключить игру; пусто, если игра подключена"`
	noticed
}

func (h *handler) status(context.Context, struct{}) (StatusOut, error) {
	out := StatusOut{ServerVersion: h.Version, Status: h.eng.Status()}
	out.FPS = math.Round(out.FPS*10) / 10
	if !out.Connected {
		out.Hint = Hint
	}
	return out, nil
}

// SnapshotOut содержит состояние игры и сообщение о снятии заморозки.
type SnapshotOut struct {
	// Snapshot содержит состояние игры.
	engine.Snapshot
	noticed
}

func (h *handler) snapshot(context.Context, struct{}) (SnapshotOut, error) {
	return snapshotOut(h.eng.Snapshot())
}

func snapshotOut(s *engine.Snapshot, err error) (SnapshotOut, error) {
	if err != nil {
		return SnapshotOut{}, err
	}
	return SnapshotOut{Snapshot: *s}, nil
}

// ScreenshotOut содержит ответ stone_screenshot.
type ScreenshotOut struct {
	// Text содержит строки экрана через перевод строки.
	Text string `json:"text" jsonschema:"строки экрана через перевод строки"`
	// Width содержит ширину экрана в клетках.
	Width int `json:"width"`
	// Height содержит высоту экрана в клетках.
	Height int `json:"height"`
	// FramesSpanned содержит число кадров, за которое прочитан экран.
	FramesSpanned int `json:"frames_spanned" jsonschema:"за сколько кадров прочитан экран; больше 1 значит, что картинка смешивает моменты"`
	noticed
}

func (h *handler) screenshot(ctx context.Context, _ struct{}) (ScreenshotOut, error) {
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()
	shot, err := h.eng.Screenshot(ctx)
	if err != nil {
		return ScreenshotOut{}, err
	}
	return ScreenshotOut{Text: shot.Text, Width: shot.Width, Height: shot.Height, FramesSpanned: shot.FramesSpanned}, nil
}
