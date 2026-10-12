package mcpserver

import (
	"context"
	"errors"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

const defaultPrintHold = 3.0

// PrintIn задаёт текст, его позицию, цвет и длительность.
type PrintIn struct {
	// Text содержит выводимый текст.
	Text string `json:"text" jsonschema:"текст на экране"`
	// X содержит столбец текста.
	X *int `json:"x,omitempty" jsonschema:"столбец; задаётся вместе с y"`
	// Y содержит строку текста.
	Y *int `json:"y,omitempty" jsonschema:"строка; задаётся вместе с x"`
	// Color содержит цвет текста.
	Color string `json:"color,omitempty" jsonschema:"имя цвета или #rrggbb; требует x и y"`
	// HoldSeconds задаёт длительность показа в секундах.
	HoldSeconds *float64 `json:"hold_seconds,omitempty" jsonschema:"сколько секунд держать текст, по умолчанию 3, максимум 600"`
}

// PrintOut содержит ответ игры на вывод текста.
type PrintOut struct {
	// Result содержит ответ игры.
	Result string `json:"result" jsonschema:"cmd_done, если игра приняла текст"`
	noticed
}

func printHold(in PrintIn) (time.Duration, error) {
	if in.HoldSeconds == nil {
		return holdFor(defaultPrintHold)
	}
	return holdFor(*in.HoldSeconds)
}

func (h *handler) print(ctx context.Context, in PrintIn) (PrintOut, error) {
	if (in.X == nil) != (in.Y == nil) {
		return PrintOut{}, errors.New("координаты x и y задаются только вместе")
	}
	var at *engine.Point
	if in.X != nil {
		at = &engine.Point{X: *in.X, Y: *in.Y}
	}
	hold, err := printHold(in)
	if err != nil {
		return PrintOut{}, err
	}
	cmd, err := engine.Print(in.Text, at, in.Color, hold)
	if err != nil {
		return PrintOut{}, err
	}
	out, err := h.do(ctx, cmd, 0)
	return PrintOut{Result: out.Result}, err
}
