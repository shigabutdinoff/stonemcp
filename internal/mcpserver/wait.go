package mcpserver

import (
	"context"
	"fmt"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

const (
	defaultWait = 30.0
	maxWait     = 120.0
)

// WaitCondition задаёт одно условие ожидания.
type WaitCondition struct {
	// Kind содержит вид условия.
	Kind string `json:"kind" jsonschema:"frames, seconds, foe_present, foe_changed, foe_element_changed, hp_below или loc_changed"`
	// Value содержит числовой параметр условия.
	Value float64 `json:"value,omitempty" jsonschema:"frames число кадров; seconds число секунд; hp_below доля максимума здоровья от 0 до 1; для остальных не нужен"`
}

// WaitIn задаёт условия и таймаут ожидания.
type WaitIn struct {
	// Conditions содержит условия, первое выполненное завершает ожидание.
	Conditions []WaitCondition `json:"conditions" jsonschema:"ожидание заканчивает первое выполненное условие"`
	// TimeoutSeconds задаёт таймаут в секундах.
	TimeoutSeconds float64 `json:"timeout_seconds,omitempty" jsonschema:"по умолчанию 30, максимум 120"`
	// Freeze просит заморозить игру на кадре события.
	Freeze bool `json:"freeze,omitempty" jsonschema:"остановить игру на кадре события до stone_release, stone_wait_for или лимита"`
	// FreezeLimitSeconds задаёт лимит заморозки в секундах.
	FreezeLimitSeconds float64 `json:"freeze_limit_seconds,omitempty" jsonschema:"только с freeze; через сколько секунд игра продолжит сама, по умолчанию 30, максимум 120"`
}

// WaitOut содержит причину завершения ожидания и состояние игры.
type WaitOut struct {
	// Reason содержит вид сработавшего условия или timeout.
	Reason string `json:"reason" jsonschema:"вид сработавшего условия или timeout"`
	// Snapshot содержит состояние на момент завершения.
	Snapshot *engine.Snapshot `json:"snapshot"`
	noticed
}

func (h *handler) waitFor(ctx context.Context, in WaitIn) (WaitOut, error) {
	timeout := in.TimeoutSeconds
	if timeout == 0 {
		timeout = defaultWait
	}
	if err := secondsIn(timeout, maxWait, "timeout_seconds"); err != nil {
		return WaitOut{}, err
	}
	if in.FreezeLimitSeconds != 0 && !in.Freeze {
		return WaitOut{}, fmt.Errorf("freeze_limit_seconds задаётся только вместе с freeze")
	}
	conds := make([]engine.Condition, len(in.Conditions))
	for i, c := range in.Conditions {
		conds[i] = engine.Condition(c)
	}
	if !in.Freeze {
		res, err := h.eng.WaitFor(ctx, conds, seconds(timeout))
		return WaitOut{Reason: res.Reason, Snapshot: res.Snapshot}, err
	}
	ctx, cancel := context.WithTimeout(ctx, seconds(timeout)+h.Timeout)
	defer cancel()
	res, err := h.eng.WaitFreeze(ctx, conds, seconds(timeout), seconds(in.FreezeLimitSeconds))
	return WaitOut{Reason: res.Reason, Snapshot: res.Snapshot}, err
}
