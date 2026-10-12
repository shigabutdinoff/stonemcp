package mcpserver

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

func TestSchemaMatchesLimits(t *testing.T) {
	maxHold := fmt.Sprintf("максимум %v", engine.MaxHold.Seconds())
	freezeLimits := []string{fmt.Sprintf("по умолчанию %v", engine.DefaultFreeze.Seconds()), fmt.Sprintf("максимум %v", engine.MaxFreeze.Seconds())}
	for _, tt := range []struct {
		in    any
		field string
		want  []string
	}{
		{EquipIn{}, "HoldSeconds", []string{maxHold}},
		{LoadoutIn{}, "HoldSeconds", []string{maxHold}},
		{PrintIn{}, "HoldSeconds", []string{maxHold, fmt.Sprintf("по умолчанию %v", defaultPrintHold)}},
		{WaitIn{}, "TimeoutSeconds", []string{fmt.Sprintf("по умолчанию %v", defaultWait), fmt.Sprintf("максимум %v", maxWait)}},
		{EvaluateIn{}, "Queries", []string{fmt.Sprintf("до %d", engine.MaxQueries)}},
		{BrewIn{}, "Ingredients", []string{ingredientList}},
		{FreezeIn{}, "LimitSeconds", freezeLimits},
		{WaitIn{}, "FreezeLimitSeconds", freezeLimits},
	} {
		typ := reflect.TypeOf(tt.in)
		f, ok := typ.FieldByName(tt.field)
		if !ok {
			t.Fatalf("%s: нет поля %s", typ.Name(), tt.field)
		}
		tag := f.Tag.Get("jsonschema")
		for _, w := range tt.want {
			if !strings.Contains(tag, w) {
				t.Errorf("%s.%s: описание %q не содержит %q", typ.Name(), tt.field, tag, w)
			}
		}
	}
}
