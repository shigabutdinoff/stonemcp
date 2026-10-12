//go:build !release

package fakegame

import (
	"errors"
	"net/rpc"
)

func (g *Game) call(method string, args, reply any) {
	g.tb.Helper()
	err := g.rpc.Call("Game."+method, args, reply)
	var fromGame rpc.ServerError
	switch {
	case err == nil:
	case errors.As(err, &fromGame):
		g.tb.Error(err)
	default:
		g.tb.Errorf("fakegame: связь с игрой потеряна на команде %s: %v", method, err)
	}
}

func (g *Game) do(method string, args any) {
	g.tb.Helper()
	g.call(method, args, &struct{}{})
}
