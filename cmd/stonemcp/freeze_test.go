package main

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func callUntilOK(t *testing.T, l *launch, tool string) {
	t.Helper()
	eventually(t, tool+" без ошибки", func() bool {
		res, err := l.CallTool(context.Background(), &mcp.CallToolParams{Name: tool})
		return err == nil && !res.IsError
	})
}

func TestClientClosedDuringFreeze(t *testing.T) {
	g := fakegame.Start(t)
	l := start(t, "--addr", g.Addr())
	callUntilOK(t, l, "stone_snapshot")
	callUntilOK(t, l, "stone_freeze")
	frame := g.Frame()
	closed := time.Now()
	l.stdin.Close()
	l.want(t, 0)
	if took := time.Since(closed); took > time.Second {
		t.Errorf("stonemcp держал игру %v после закрытия клиента", took)
	}
	callUntilOK(t, start(t, "--addr", g.Addr()), "stone_snapshot")
	eventually(t, "игра идёт после закрытия клиента в заморозке", func() bool { return g.Frame() > frame+3 })
}
