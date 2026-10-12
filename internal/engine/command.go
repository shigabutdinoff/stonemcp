package engine

import (
	"fmt"
	"strings"
	"time"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

// MaxHold ограничивает длительность удержания команды.
const MaxHold = 600 * time.Second

// Command задаёт игровую команду и длительность её повторения.
type Command struct {
	// Name содержит имя разрешённой команды игры.
	Name string
	// Arg содержит строковый аргумент команды.
	Arg string
	// Hold задаёт удержание, ноль означает однократную отправку.
	Hold time.Duration
}

// CommandResult содержит ответ игры и состояние после исполнения.
type CommandResult struct {
	// Result содержит код ответа игры.
	Result string `json:"result"`
	// Snapshot содержит первый следующий кадр или кадр заморозки.
	// Значение равно nil при отказе игры и при обрыве до нового кадра.
	Snapshot *Snapshot `json:"snapshot"`
	// Frozen сообщает, что команда принята в заморозке и подействует
	// с первого кадра после отпускания.
	Frozen bool `json:"frozen,omitempty"`
}

func (c Command) request() mc.Request { return mc.Cmd(c.Name, c.Arg) }

func (c Command) validate(maxPacket int) error {
	spec, ok := commandSpecs[c.Name]
	if !ok {
		return fmt.Errorf("недопустимая команда %q", c.Name)
	}
	empty := c.Arg == ""
	if !spec.blankOK {
		empty = strings.TrimSpace(c.Arg) == ""
	}
	if empty {
		return fmt.Errorf("пустой аргумент команды %q", c.Name)
	}
	if c.Hold < 0 || c.Hold > MaxHold {
		return fmt.Errorf("удержание должно быть от нуля до %v", MaxHold)
	}
	if c.Hold > 0 && spec.slot == nil {
		return fmt.Errorf("команду %q нельзя удерживать", c.Name)
	}
	if spec.arg != nil {
		if err := spec.arg(c); err != nil {
			return err
		}
	}
	return mc.Fits(c.request(), maxPacket)
}
