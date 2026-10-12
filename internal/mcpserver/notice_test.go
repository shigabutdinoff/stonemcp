package mcpserver

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestCanceledCallLeavesNotice(t *testing.T) {
	g := fakegame.Start(t)
	cs := online(t, g)
	first := decode[engine.Snapshot](t, call(t, cs, "stone_freeze", map[string]any{"limit_seconds": 60}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		args := map[string]any{"conditions": []map[string]any{{"kind": "loc_changed"}}, "timeout_seconds": 60}
		_, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "stone_wait_for", Arguments: args})
		done <- err
	}()
	if !eventually(t, 2*time.Second, func() bool { return g.Frame() > first.Time }) {
		t.Fatal("висящее ожидание не отпустило игру")
	}
	thawByLimit(t, cs, g)
	cancel()
	if err := <-done; err == nil {
		t.Fatal("отменённый вызов завершился без ошибки")
	}
	out := decode[WaitOut](t, call(t, cs, "stone_wait_for", map[string]any{"conditions": []map[string]any{{"kind": "frames", "value": 3}}}))
	if !strings.Contains(out.Notice, "снята по лимиту") {
		t.Fatalf("сообщение о лимите после отмены = %q", out.Notice)
	}
}
