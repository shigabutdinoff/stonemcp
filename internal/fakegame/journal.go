//go:build !release

package fakegame

import "slices"

// AllowViolations отменяет проверку нарушений в конце теста.
func (g *Game) AllowViolations() { g.allow.Store(true) }

// Commands возвращает все полученные команды.
func (g *Game) Commands() []Command {
	var cmds []Command
	g.call("Commands", struct{}{}, &cmds)
	return cmds
}

// CommandsNamed возвращает полученные команды с именем name.
func (g *Game) CommandsNamed(name string) []Command {
	return slices.DeleteFunc(g.Commands(), func(c Command) bool { return c.Name != name })
}

// Violations перечисляет нарушения протокола, допущенные клиентом.
func (g *Game) Violations() []string {
	var r report
	g.call("Report", struct{}{}, &r)
	return r.Violations
}
