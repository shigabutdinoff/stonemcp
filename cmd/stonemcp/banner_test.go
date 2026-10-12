package main

import (
	"strings"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestBannerOnGameScreen(t *testing.T) {
	setBuild(t, "27.0", "27A41", "2026-09-28", "3f2a9c1")
	g := fakegame.Start(t)
	l := start(t, "--addr", g.Addr())
	eventually(t, "строка версии на экране игры", func() bool {
		cmds := g.CommandsNamed(">")
		return len(cmds) > 0 && strings.HasSuffix(cmds[0].Arg, ",0,stonemcp 27.0 (27A41)")
	})
	l.stdin.Close()
	l.want(t, 0)
}
