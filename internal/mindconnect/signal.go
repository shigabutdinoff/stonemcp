package mindconnect

import (
	"errors"
	"fmt"
	"strings"
)

// Phase обозначает половину игрового кадра, которую открывает сигнал.
type Phase int

const (
	// PhasePre обозначает шаг до исполнения скрипта Mind Stone.
	PhasePre Phase = iota + 1
	// PhasePost обозначает шаг после исполнения скрипта Mind Stone.
	PhasePost
)

var phaseWire = [...]string{PhasePre: "pre", PhasePost: "post"}

// String возвращает "pre" или "post", а для прочих значений "неизвестная".
func (p Phase) String() string {
	if p == PhasePre || p == PhasePost {
		return phaseWire[p]
	}
	return "неизвестная"
}

// Version задаёт версию протокола; сверять её с сигналом решает вызывающий.
const Version = "0.3"

// ErrUnsupportedSignal означает сигнал режима MindConnect, отличного от sync.
var ErrUnsupportedSignal = errors.New("mindconnect: сигнал не поддержан")

// ErrIncompleteSignal означает, что сигнал прочитан не целиком, нужно дочитать.
var ErrIncompleteSignal = errors.New("mindconnect: сигнал не дочитан")

// ParseSignal разбирает сигнал кадра вида "\x06pre0.3" в шаг и версию.
// Оборванный сигнал даёт ErrIncompleteSignal, соединение не рвут, а дочитывают.
// Версия без точки тоже считается оборванной, её принимает ParseStalledSignal.
func ParseSignal(b []byte) (Phase, string, error) {
	return parseSignal(b, false)
}

// ParseStalledSignal разбирает сигнал, после которого байтов больше не будет.
// Версию без точки она принимает. Недописанный сигнал даёт ошибку обрыва,
// а не ErrIncompleteSignal.
func ParseStalledSignal(b []byte) (Phase, string, error) {
	return parseSignal(b, true)
}

func parseSignal(b []byte, stalled bool) (Phase, string, error) {
	s, ok := strings.CutPrefix(string(b), SignalPrefix)
	if !ok {
		if len(b) == 0 {
			return 0, "", incomplete(stalled, "ещё ничего не прочитано")
		}
		return 0, "", fmt.Errorf("mindconnect: не сигнал: %q", string(b))
	}
	for _, p := range []Phase{PhasePre, PhasePost} {
		name := phaseWire[p]
		if strings.HasPrefix(name, s) {
			return 0, "", incomplete(stalled, fmt.Sprintf("%q", s))
		}
		version, ok := strings.CutPrefix(s, name)
		if !ok {
			continue
		}
		switch versionOf(version) {
		case versionOK:
			return p, version, nil
		case versionNoDot:
			if stalled {
				return p, version, nil
			}
			return 0, "", incomplete(stalled, fmt.Sprintf("%q", s))
		case versionPartial:
			return 0, "", incomplete(stalled, fmt.Sprintf("%q", s))
		}
		return 0, "", fmt.Errorf("mindconnect: у сигнала %q неверная версия протокола", s)
	}
	return 0, "", fmt.Errorf("%w: %q (поддержан только режим sync)", ErrUnsupportedSignal, s)
}

func incomplete(stalled bool, what string) error {
	if stalled {
		return fmt.Errorf("mindconnect: сигнал оборван: %s", what)
	}
	return fmt.Errorf("%w: %s", ErrIncompleteSignal, what)
}

type versionForm int

const (
	versionOK versionForm = iota
	versionNoDot
	versionPartial
	versionBad
)

func versionOf(v string) versionForm {
	dots := 0
	prevDot := true
	for _, c := range v {
		switch {
		case c >= '0' && c <= '9':
			prevDot = false
		case c == '.':
			if prevDot {
				return versionBad
			}
			dots++
			prevDot = true
		default:
			return versionBad
		}
	}
	if prevDot {
		return versionPartial
	}
	if dots == 0 {
		return versionNoDot
	}
	return versionOK
}
