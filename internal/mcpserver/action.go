package mcpserver

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

const animationWait = 15 * time.Second

var ingredientList = strings.Join(engine.Ingredients, ", ")

// ActionOut содержит ответ игры на действие и состояние после него.
type ActionOut struct {
	// Result содержит ответ игры.
	Result string `json:"result" jsonschema:"cmd_done, если игра приняла команду"`
	// Snapshot содержит первый кадр после команды или кадр заморозки.
	Snapshot *engine.Snapshot `json:"snapshot" jsonschema:"состояние игры с первого кадра после исполнения команды или замороженного кадра"`
	// Frozen сообщает, что команда принята в заморозке.
	Frozen bool `json:"frozen,omitempty" jsonschema:"true, если игра заморожена и команда подействует с первого кадра после отпускания"`
	noticed
}

// EquipIn задаёт предмет и удержание экипировки.
type EquipIn struct {
	// Item содержит критерии предмета в синтаксисе Stonescript.
	Item string `json:"item" jsonschema:"критерии предмета как в Stonescript, например poison sword *8 +5"`
	// HoldSeconds задаёт удержание в секундах.
	HoldSeconds float64 `json:"hold_seconds,omitempty" jsonschema:"повторять команду каждый кадр столько секунд, чтобы скрипт игры её не отменил; 0 по умолчанию отправляет один раз, максимум 600"`
}

// ActivateIn задаёт цель активации.
type ActivateIn struct {
	// Target содержит способность, зелье или предмет.
	Target string `json:"target" jsonschema:"potion, P, L, R или имя предмета"`
}

// LoadoutIn задаёт номер набора и удержание.
type LoadoutIn struct {
	// Number содержит номер набора экипировки.
	Number int `json:"number" jsonschema:"номер набора экипировки"`
	// HoldSeconds задаёт удержание в секундах.
	HoldSeconds float64 `json:"hold_seconds,omitempty" jsonschema:"повторять команду каждый кадр столько секунд; 0 по умолчанию отправляет один раз, максимум 600"`
}

// BrewIn задаёт ингредиенты зелья.
type BrewIn struct {
	// Ingredients содержит ингредиенты зелья.
	Ingredients []string `json:"ingredients" jsonschema:"ингредиенты из stone, wood, tar, bronze"`
}

func (h *handler) do(ctx context.Context, cmd engine.Command, extra time.Duration) (ActionOut, error) {
	if err := h.eng.Ticking(h.StaleAfter); err != nil {
		return ActionOut{}, err
	}
	wait := h.Timeout + extra
	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	res, err := h.eng.Do(ctx, cmd, wait)
	return ActionOut{Result: res.Result, Snapshot: res.Snapshot, Frozen: res.Frozen}, err
}

func (h *handler) held(ctx context.Context, seconds float64, cmd engine.Command) (ActionOut, error) {
	hold, err := holdFor(seconds)
	if err != nil {
		return ActionOut{}, err
	}
	cmd.Hold = hold
	return h.do(ctx, cmd, 0)
}

func (h *handler) equipWith(build func(string, time.Duration) engine.Command) func(context.Context, EquipIn) (ActionOut, error) {
	return func(ctx context.Context, in EquipIn) (ActionOut, error) {
		return h.held(ctx, in.HoldSeconds, build(in.Item, 0))
	}
}

func (h *handler) loadout(ctx context.Context, in LoadoutIn) (ActionOut, error) {
	return h.held(ctx, in.HoldSeconds, engine.Loadout(in.Number, 0))
}

func (h *handler) activate(ctx context.Context, in ActivateIn) (ActionOut, error) {
	return h.do(ctx, engine.Activate(in.Target), animationWait)
}

func (h *handler) brew(ctx context.Context, in BrewIn) (ActionOut, error) {
	if len(in.Ingredients) == 0 {
		return ActionOut{}, fmt.Errorf("нужен хотя бы один ингредиент, допустимы %s", ingredientList)
	}
	for _, i := range in.Ingredients {
		if !slices.Contains(engine.Ingredients, i) {
			return ActionOut{}, fmt.Errorf("недопустимый ингредиент %q, допустимы %s", i, ingredientList)
		}
	}
	return h.do(ctx, engine.Brew(in.Ingredients), 0)
}
