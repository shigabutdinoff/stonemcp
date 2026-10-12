package mcpserver

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestToolList(t *testing.T) {
	cs := online(t, fakegame.Start(t))
	list, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"stone_status", "stone_snapshot", "stone_screenshot", "stone_equip", "stone_equip_left", "stone_equip_right",
		"stone_activate", "stone_loadout", "stone_brew", "stone_print", "stone_wait_for", "stone_evaluate", "stone_freeze", "stone_release"}
	var got []string
	for _, tool := range list.Tools {
		got = append(got, tool.Name)
		if tool.Description == "" || tool.InputSchema == nil {
			t.Errorf("%s без описания или схемы входа", tool.Name)
		}
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("инструменты %v, ожидались %v", got, want)
	}
}

func TestToolOutputSchemaHasNotice(t *testing.T) {
	cs := online(t, fakegame.Start(t))
	list, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, tool := range list.Tools {
		raw, err := json.Marshal(tool.OutputSchema)
		if err != nil {
			t.Fatalf("%s: схема выхода не превращается в JSON: %v", tool.Name, err)
		}
		var schema struct {
			Properties map[string]any `json:"properties"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("%s: схема выхода не разобрана: %v", tool.Name, err)
		}
		if _, ok := schema.Properties["notice"]; !ok {
			t.Errorf("%s: схема выхода без поля notice", tool.Name)
		}
	}
}
