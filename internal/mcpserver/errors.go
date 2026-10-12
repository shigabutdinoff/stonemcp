package mcpserver

import (
	"errors"
	"fmt"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

func explain(err error) error {
	switch {
	case errors.Is(err, engine.ErrDisconnected):
		return fmt.Errorf("%w: %s", err, Hint)
	case errors.Is(err, engine.ErrNotTicking):
		return fmt.Errorf("%w: игра подключена, но не присылает кадры (пауза, меню, анимация или герой вне локации)", err)
	}
	return err
}
