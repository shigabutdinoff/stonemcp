package engine

import (
	"fmt"
	"math"
	"time"
)

const (
	// CondFrames означает ожидание числа кадров.
	CondFrames = "frames"
	// CondSeconds означает ожидание реальных секунд.
	CondSeconds = "seconds"
	// CondFoePresent означает наличие врага.
	CondFoePresent = "foe_present"
	// CondFoeChanged означает изменение типа или наличия врага.
	CondFoeChanged = "foe_changed"
	// CondFoeElementChanged означает смену стихии текущего врага.
	CondFoeElementChanged = "foe_element_changed"
	// CondHPBelow означает здоровье строго ниже доли максимума.
	CondHPBelow = "hp_below"
	// CondLocChanged означает изменение идентификатора локации.
	CondLocChanged = "loc_changed"
	// ReasonTimeout означает истечение срока ожидания без события.
	ReasonTimeout = "timeout"
)

// Condition задаёт условие ожидания и его числовой параметр.
type Condition struct {
	// Kind содержит вид условия.
	Kind string `json:"kind"`
	// Value содержит число кадров, секунд или долю здоровья.
	Value float64 `json:"value,omitempty"`
}

func (c Condition) validate() error {
	if math.IsNaN(c.Value) || math.IsInf(c.Value, 0) || c.Value < 0 {
		return fmt.Errorf("недопустимое значение условия %q", c.Kind)
	}
	switch c.Kind {
	case CondFrames:
		if c.Value != math.Trunc(c.Value) || c.Value >= float64(math.MaxInt64) {
			return fmt.Errorf("число кадров должно быть целым и помещаться в int64")
		}
	case CondSeconds:
		if c.Value >= float64(math.MaxInt64)/float64(time.Second) {
			return fmt.Errorf("слишком большой срок ожидания")
		}
	case CondHPBelow:
		if c.Value > 1 {
			return fmt.Errorf("доля здоровья должна быть от 0 до 1")
		}
	case CondFoePresent, CondFoeChanged, CondFoeElementChanged, CondLocChanged:
	default:
		return fmt.Errorf("неизвестное условие %q", c.Kind)
	}
	return nil
}

func (c Condition) met(base, cur *Snapshot, elapsed time.Duration) bool {
	switch c.Kind {
	case CondFrames:
		return cur.Frame-base.Frame >= int64(c.Value)
	case CondSeconds:
		return elapsed >= time.Duration(c.Value*float64(time.Second))
	case CondFoePresent:
		return cur.Foe != nil
	case CondFoeChanged:
		if base.Foe == nil || cur.Foe == nil {
			return base.Foe != cur.Foe
		}
		return base.Foe.ID != cur.Foe.ID
	case CondFoeElementChanged:
		element := foeElement(cur)
		return element != "" && element != foeElement(base)
	case CondHPBelow:
		return cur.Player.MaxHP > 0 && float64(cur.Player.HP) < c.Value*float64(cur.Player.MaxHP)
	case CondLocChanged:
		return cur.Loc.ID != base.Loc.ID
	}
	return false
}
