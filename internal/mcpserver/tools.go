package mcpserver

import (
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

const (
	defaultTimeout    = 10 * time.Second
	defaultStaleAfter = 10 * time.Second
)

// Hint объясняет, как подключить игру к серверу.
const Hint = "запустите Stone Story RPG, сделайте первой строкой программы Mind Stone //MindConnect: sync, включите Mind Stone и зайдите в локацию"

// Options задаёт версию сервера и лимиты вызова.
// Нули выбирают значения по умолчанию.
type Options struct {
	// Version содержит версию сервера для stone_status.
	Version string
	// Timeout ограничивает вызов инструмента, по умолчанию 10 с.
	// У stone_wait_for свой таймаут.
	Timeout time.Duration
	// StaleAfter задаёт возраст кадра, после которого действие
	// сразу отвечает, что игра не тикает, по умолчанию 10 с.
	StaleAfter time.Duration
}

type handler struct {
	eng *engine.Engine
	Options
}

// Register добавляет на сервер все инструменты stone_*.
func Register(s *mcp.Server, eng *engine.Engine, opts Options) {
	if opts.Timeout <= 0 {
		opts.Timeout = defaultTimeout
	}
	if opts.StaleAfter <= 0 {
		opts.StaleAfter = defaultStaleAfter
	}
	h := &handler{eng: eng, Options: opts}
	add(s, h, "stone_status", descStatus, h.status)
	add(s, h, "stone_snapshot", descSnapshot, h.snapshot)
	add(s, h, "stone_screenshot", descScreenshot, h.screenshot)
	add(s, h, "stone_equip", descEquip, h.equipWith(engine.Equip))
	add(s, h, "stone_equip_left", descEquipLeft, h.equipWith(engine.EquipLeft))
	add(s, h, "stone_equip_right", descEquipRight, h.equipWith(engine.EquipRight))
	add(s, h, "stone_activate", descActivate, h.activate)
	add(s, h, "stone_loadout", descLoadout, h.loadout)
	add(s, h, "stone_brew", descBrew, h.brew)
	add(s, h, "stone_print", descPrint, h.print)
	add(s, h, "stone_wait_for", descWait, h.waitFor)
	add(s, h, "stone_evaluate", descEvaluate, h.evaluate)
	add(s, h, "stone_freeze", descFreeze, h.freeze)
	add(s, h, "stone_release", descRelease, h.release)
}

func add[In, Out any, P interface {
	*Out
	noticer
}](s *mcp.Server, h *handler, name, desc string, fn func(context.Context, In) (Out, error)) {
	mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		out, err := fn(ctx, in)
		return nil, out, h.withNotice(ctx, P(&out), explain(err))
	})
}
