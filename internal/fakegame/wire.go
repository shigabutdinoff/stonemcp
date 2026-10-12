//go:build !release

package fakegame

// Command описывает команду, полученную от клиента.
type Command struct {
	// Frame содержит номер кадра, в котором пришла команда.
	Frame int
	// Phase содержит шаг кадра, "pre" или "post".
	Phase string
	// Packet содержит номер пакета соединения, в котором пришла команда.
	Packet int
	// Name содержит имя команды.
	Name string
	// Arg содержит строковый аргумент команды.
	Arg string
}

// Call описывает вызов функции игры, заданной тестом.
type Call struct {
	// Frame содержит номер кадра, в котором пришёл вызов.
	Frame int
	// Args содержит строковые аргументы вызова.
	Args []string
}

type pair struct{ Name, Value string }

type foe struct {
	ID string
	HP int
}

type toggle struct {
	Name string
	On   bool
}

type switchAt struct {
	Name          string
	At            int
	Before, After string
}

type funcArgs struct {
	Name  string
	Args  []string
	Value string
}

type report struct{ Violations, Failures []string }
