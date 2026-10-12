package mcpserver

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func connect(t *testing.T, addr string, opts Options) *mcp.ClientSession {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	eng := engine.New(engine.Config{Addr: addr, ReconnectDelays: []time.Duration{5 * time.Millisecond, 20 * time.Millisecond}})
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = eng.Run(ctx)
	}()
	server := mcp.NewServer(&mcp.Implementation{Name: "stonemcp", Version: opts.Version}, nil)
	Register(server, eng, opts)
	serverT, clientT := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, serverT, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = cs.Close()
		_ = ss.Close()
		cancel()
		<-done
	})
	return cs
}

func online(t *testing.T, g *fakegame.Game) *mcp.ClientSession {
	t.Helper()
	return onlineWith(t, g, Options{Version: "27.0"})
}

func onlineWith(t *testing.T, g *fakegame.Game, opts Options) *mcp.ClientSession {
	t.Helper()
	cs := connect(t, g.Addr(), opts)
	if !eventually(t, 3*time.Second, func() bool { return !call(t, cs, "stone_snapshot", nil).IsError }) {
		t.Fatal("игра так и не подключилась")
	}
	return cs
}

func eventually(t *testing.T, within time.Duration, cond func() bool) bool {
	t.Helper()
	deadline := time.Now().Add(within)
	for !cond() {
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(2 * time.Millisecond)
	}
	return true
}

func call(t *testing.T, cs *mcp.ClientSession, name string, args any) *mcp.CallToolResult {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: ошибка протокола %v", name, err)
	}
	return res
}

func decode[T any](t *testing.T, res *mcp.CallToolResult) T {
	t.Helper()
	if res.IsError {
		t.Fatalf("ошибка инструмента %s", text(res))
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("не разобран ответ %s: %v", raw, err)
	}
	return out
}

func text(res *mcp.CallToolResult) string {
	var parts []string
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "\n")
}

func wantToolError(t *testing.T, res *mcp.CallToolResult, want string) {
	t.Helper()
	if !res.IsError || !strings.Contains(text(res), want) {
		t.Errorf("ожидалась ошибка с %q, isError=%v text=%q", want, res.IsError, text(res))
	}
}

func wantFastError(t *testing.T, cs *mcp.ClientSession, tool string, args any, want string) {
	t.Helper()
	start := time.Now()
	wantToolError(t, call(t, cs, tool, args), want)
	if took := time.Since(start); took > time.Second {
		t.Errorf("%s ждал %v вместо быстрой ошибки", tool, took)
	}
}
