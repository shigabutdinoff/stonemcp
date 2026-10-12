package mcpserver

import (
	"context"
	"fmt"
	"math"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

// QueryIn задаёт чтение переменной или вызов функции игры.
type QueryIn struct {
	// Name содержит имя переменной или функции Stonescript.
	Name string `json:"name" jsonschema:"имя переменной или функции Stonescript, например foe.armor или item.GetCooldown"`
	// Args содержит аргументы функции.
	Args []any `json:"args,omitempty" jsonschema:"аргументы функции, только строки и целые числа"`
}

// EvaluateIn задаёт набор чтений одного кадра.
type EvaluateIn struct {
	// Queries содержит чтения в порядке ответа.
	Queries []QueryIn `json:"queries" jsonschema:"от 1 до 256 чтений, ответ из одного кадра"`
}

// EvaluateOut содержит значения в порядке запросов.
type EvaluateOut struct {
	// Values содержит значения строками.
	Values []string `json:"values" jsonschema:"по одному значению на запрос, в том же порядке"`
	noticed
}

func (h *handler) evaluate(ctx context.Context, in EvaluateIn) (EvaluateOut, error) {
	queries, err := buildQueries(in.Queries)
	if err != nil {
		return EvaluateOut{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, h.Timeout)
	defer cancel()
	values, err := h.eng.Evaluate(ctx, queries)
	if err != nil {
		return EvaluateOut{}, err
	}
	out := EvaluateOut{Values: make([]string, len(values))}
	for i, v := range values {
		if v.IsError() {
			req := mc.Var(queries[i].Name, queries[i].Args...)
			return EvaluateOut{}, fmt.Errorf("игра вернула %s на запрос %s", v, req)
		}
		out.Values[i] = v.String()
	}
	return out, nil
}

func buildQueries(in []QueryIn) ([]engine.Query, error) {
	queries := make([]engine.Query, len(in))
	for i, q := range in {
		args, err := toArgs(q.Args)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", q.Name, err)
		}
		guard, err := checkQuery(q.Name, q.Args, args)
		if err != nil {
			return nil, err
		}
		queries[i] = engine.Query{Name: q.Name, Args: args, Guard: guard}
	}
	return queries, nil
}

func toArgs(raw []any) ([]mc.Arg, error) {
	var args []mc.Arg
	for _, a := range raw {
		arg, err := toArg(a)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func toArg(a any) (mc.Arg, error) {
	switch v := a.(type) {
	case string:
		return mc.Str(v), nil
	case float64:
		if v == math.Trunc(v) && v >= math.MinInt32 && v <= math.MaxInt32 {
			return mc.Int(int(v)), nil
		}
	}
	return mc.Arg{}, fmt.Errorf("аргумент %v недопустим, нужны строки или целые числа int32", a)
}
