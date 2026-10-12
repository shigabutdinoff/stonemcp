package engine

import (
	"fmt"
	"strings"
	"unicode/utf8"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func (c Config) validateBanner() error {
	if c.Banner == "" {
		return nil
	}
	if strings.Contains(c.Banner, "@") {
		return fmt.Errorf("строка на экране содержит @, игра примет её за выражение")
	}
	_, err := Print(c.Banner, &Point{}, "", 0)
	return err
}

func (e *Engine) bannerRequests() []mc.Request {
	if e.cfg.Banner == "" {
		return nil
	}
	e.mu.Lock()
	snapshot := e.snapshot
	e.mu.Unlock()
	if snapshot == nil {
		return nil
	}
	x := max(snapshot.Screen.W-utf8.RuneCountInString(e.cfg.Banner), 0)
	cmd, _ := Print(e.cfg.Banner, &Point{X: x}, "", 0)
	return []mc.Request{cmd.request()}
}
