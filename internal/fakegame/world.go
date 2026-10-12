//go:build !release

package fakegame

// ElementalistFoe возвращает значение foe для дисангелоса со стихией element.
func ElementalistFoe(element string) string {
	return "dysangelos_elementalist One With the Elements boss unpushable " + element + " phase2 ranged immune_to_stun"
}

// Set задаёт переменную, перекрывая встроенную; SetFunc сильнее Set.
func (g *Game) Set(name, value string) { g.do("Set", pair{name, value}) }

// SetFoe выставляет врага; пустой id убирает его и сбрасывает все поля foe.
func (g *Game) SetFoe(id string, hp int) { g.do("SetFoe", foe{id, hp}) }

// SetElementalist делает врагом дисангелоса со стихией element.
func (g *Game) SetElementalist(element string) { g.do("SetElementalist", element) }

// SetScreen заменяет экран для screen.w, screen.h и draw.GetSymbol.
func (g *Game) SetScreen(lines []string) { g.do("SetScreen", lines) }

// SwitchAt отдаёт name равным before до кадра at и after начиная с него.
func (g *Game) SwitchAt(name string, at int, before, after string) {
	g.do("SwitchAt", switchAt{name, at, before, after})
}

// SetFunc заставляет функцию name отвечать value на любые аргументы.
func (g *Game) SetFunc(name, value string) { g.do("SetFunc", pair{name, value}) }

// SetFuncArgs отвечает value только на аргументы args, иначе fnc_err3.
func (g *Game) SetFuncArgs(name string, args []string, value string) {
	g.do("SetFuncArgs", funcArgs{name, args, value})
}

// HoldFunc придерживает ответ value функции name до ReleaseFunc.
func (g *Game) HoldFunc(name, value string) { g.do("HoldFunc", pair{name, value}) }

// ReleaseFunc отпускает придержанный ответ; повторный вызов безопасен.
func (g *Game) ReleaseFunc(name string) { g.do("ReleaseFunc", name) }

// Calls возвращает вызовы функции name в порядке получения.
func (g *Game) Calls(name string) []Call {
	var calls []Call
	g.call("Calls", name, &calls)
	return calls
}
