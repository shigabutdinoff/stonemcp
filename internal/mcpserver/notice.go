package mcpserver

import (
	"context"
	"fmt"
)

type noticed struct {
	// Notice сообщает о снятии заморозки по лимиту, пусто в остальных случаях.
	Notice string `json:"notice,omitempty" jsonschema:"сообщение, что игра продолжила сама по лимиту заморозки; пусто в остальных случаях"`
}

func (n *noticed) setNotice(s string) { n.Notice = s }

type noticer interface{ setNotice(string) }

func (h *handler) withNotice(ctx context.Context, out noticer, err error) error {
	if ctx.Err() != nil {
		return err
	}
	notice := h.eng.TakeNotice()
	if notice == "" {
		return err
	}
	if err != nil {
		return fmt.Errorf("%w: %s", err, notice)
	}
	out.setNotice(notice)
	return nil
}
