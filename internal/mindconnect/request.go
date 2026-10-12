package mindconnect

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var commandNames = map[string]bool{
	">": true, "play": true, "equip": true, "equipL": true, "equipR": true, "equipF": true,
	"loadout": true, "activate": true, "enable": true, "disable": true, "profile": true, "brew": true,
}

func isCommand(name string) bool { return commandNames[name] }

// Request описывает элемент пакета, это переменная, вызов функции или команда.
type Request struct {
	name    string
	args    []Arg
	command bool
}

// Var строит запрос переменной или функции; аргументы переменной не читаются.
func Var(name string, args ...Arg) Request { return Request{name: name, args: args} }

// Cmd строит команду; команда несёт ровно один строковый аргумент.
func Cmd(name, arg string) Request {
	return Request{name: name, args: []Arg{Str(arg)}, command: true}
}

// String возвращает запрос как name, name(a, b) или "name arg".
func (r Request) String() string {
	if len(r.args) == 0 {
		return r.name
	}
	parts := make([]string, len(r.args))
	for i, a := range r.args {
		parts[i] = a.String()
	}
	if r.command {
		return r.name + " " + parts[0]
	}
	return r.name + "(" + strings.Join(parts, ", ") + ")"
}

// ErrInvalidRequest означает негодный запрос или пакет.
var ErrInvalidRequest = errors.New("mindconnect: недопустимый запрос")

// Validate отвергает запросы, которые игра разобрала бы не так, как задумано.
func (r Request) Validate() error {
	if r.name == "" {
		return fmt.Errorf("%w: пустое имя", ErrInvalidRequest)
	}
	if badText(r.name) {
		return fmt.Errorf("%w: управляющий символ или битый UTF-8 в имени %q", ErrInvalidRequest, r.name)
	}
	switch {
	case r.command:
		if err := r.validateCommand(); err != nil {
			return err
		}
	case isCommand(r.name):
		return fmt.Errorf("%w: %q является командой, а не переменной", ErrInvalidRequest, r.name)
	}
	for _, a := range r.args {
		if err := a.validate(r.name); err != nil {
			return err
		}
	}
	return nil
}

func (r Request) validateCommand() error {
	if !isCommand(r.name) {
		return fmt.Errorf("%w: %q не является командой", ErrInvalidRequest, r.name)
	}
	if len(r.args) != 1 || r.args[0].isInt {
		return fmt.Errorf("%w: команде %q нужен один строковый аргумент", ErrInvalidRequest, r.name)
	}
	if strings.Contains(r.args[0].s, "@") {
		return fmt.Errorf("%w: @ в аргументе команды %q недопустим", ErrInvalidRequest, r.name)
	}
	return nil
}

func badText(s string) bool {
	return !utf8.ValidString(s) || strings.ContainsFunc(s, isControl)
}

func isControl(r rune) bool { return r < 0x20 || r == 0x7f }
