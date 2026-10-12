package engine

import (
	"fmt"
	"slices"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

// MaxQueries ограничивает число запросов Evaluate вместе с проверками.
const MaxQueries = 256

// Query описывает чтение переменной или вызов функции игры.
type Query struct {
	// Name содержит имя переменной или функции.
	Name string
	// Args содержит аргументы функции.
	Args []mc.Arg
	// Guard задаёт необязательную проверку перед чтением.
	Guard *Guard
}

// Guard разрешает чтение только после успешной проверки в том же кадре.
type Guard struct {
	// Check задаёт запрос проверки без вложенного Guard.
	Check Query
	// OK сообщает, разрешено ли чтение по результату проверки.
	// Функция должна завершаться сразу и не вызывать API движка.
	OK func(mc.Value) bool
	// Else содержит результат при невыполненной проверке.
	Else mc.Value
}

func prepareQueries(queries []Query, maxPacket int) ([]Query, int, error) {
	if len(queries) == 0 || len(queries) > MaxQueries {
		return nil, 0, fmt.Errorf("Evaluate ожидает от 1 до %d запросов", MaxQueries)
	}
	copy := slices.Clone(queries)
	var reqs []mc.Request
	for i := range copy {
		q := &copy[i]
		q.Args = slices.Clone(q.Args)
		reqs = append(reqs, q.request())
		if q.Guard != nil {
			guard := *q.Guard
			if guard.OK == nil || guard.Check.Guard != nil {
				return nil, 0, fmt.Errorf("недопустимая проверка запроса")
			}
			guard.Check.Args = slices.Clone(guard.Check.Args)
			q.Guard = &guard
			reqs = append(reqs, guard.Check.request())
		}
	}
	if len(reqs) > MaxQueries {
		return nil, 0, fmt.Errorf("число запросов с проверками превышает %d", MaxQueries)
	}
	return copy, len(reqs), fitAll(reqs, maxPacket)
}

func (q Query) request() mc.Request { return mc.Var(q.Name, q.Args...) }

func fitAll(reqs []mc.Request, maxPacket int) error {
	for _, r := range reqs {
		if err := mc.Fits(r, maxPacket); err != nil {
			return err
		}
	}
	return nil
}
