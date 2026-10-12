package mcpserver

import (
	"fmt"
	"slices"
	"strings"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

var guardedVars = map[string]engine.Guard{
	"buffs.oldest":   {Check: engine.Query{Name: "buffs.count"}, OK: positive},
	"debuffs.oldest": {Check: engine.Query{Name: "debuffs.count"}, OK: positive},
}

func checkQuery(name string, raw []any, args []mc.Arg) (*engine.Guard, error) {
	if g, ok := guardedVars[name]; ok {
		return &g, nil
	}
	if slices.Contains(changing, name) {
		return nil, fmt.Errorf("%s меняет игру, stone_evaluate только читает", name)
	}
	if sig, ok := functions[name]; ok {
		return sig.verify(name, raw, args)
	}
	return nil, nil
}

func (s signature) verify(name string, raw []any, args []mc.Arg) (*engine.Guard, error) {
	if !s.fits(raw) {
		return nil, fmt.Errorf("%s: ожидается %s", name, s.usage)
	}
	if s.check != nil {
		if err := s.check(raw[0]); err != nil {
			return nil, fmt.Errorf("%s: %w; ожидается %s", name, err, s.usage)
		}
	}
	if s.guard == nil {
		return nil, nil
	}
	return s.guard(args), nil
}

func (s signature) fits(raw []any) bool {
	want := append(slices.Clip(s.args), s.optional...)
	if len(raw) != len(s.args) && len(raw) != len(want) {
		return false
	}
	for i, v := range raw {
		if !want[i].accepts(v) {
			return false
		}
	}
	return true
}

func (k kind) accepts(v any) bool {
	switch k {
	case intArg:
		_, ok := v.(float64)
		return ok
	case strArg:
		_, ok := v.(string)
		return ok
	}
	return true
}

func stateName(v any) error {
	if s, _ := v.(string); !slices.Contains(gameStates, s) {
		return fmt.Errorf("неизвестное состояние %q", s)
	}
	return nil
}

func notNegative(v any) error {
	if n, _ := v.(float64); n < 0 {
		return fmt.Errorf("отрицательный номер %d", int(n))
	}
	return nil
}

func summonVar(v any) error {
	if s, _ := v.(string); s != "ignition" {
		return fmt.Errorf("неизвестная переменная призыва %q", s)
	}
	return nil
}

func plainName(v any) error {
	if s := fmt.Sprint(v); strings.Contains(s, ".") {
		return fmt.Errorf("составные имена переменных скрипта не поддерживаются, получено %q", s)
	}
	return nil
}

func positive(v mc.Value) bool {
	n, err := v.Int()
	return err == nil && n > 0
}

func onCinderwisp(args []mc.Arg) *engine.Guard {
	check := engine.Query{Name: "summon.GetId", Args: args[1:]}
	return &engine.Guard{Check: check, OK: func(v mc.Value) bool { return v == "cinderwisp" }}
}
