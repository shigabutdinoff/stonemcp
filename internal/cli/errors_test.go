package cli

import "testing"

func TestErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"-x"}, `неизвестный флаг "-x"`},
		{[]string{"-hx"}, `неизвестный флаг "-x"`},
		{[]string{"--help=да"}, `неверное значение "да" флага "--help"`},
		{[]string{"---x"}, `неверный синтаксис флага "---x"`},
		{[]string{"nosuch"}, `неизвестная команда "nosuch"`},
		{[]string{"--nosuch"}, `неизвестный флаг "--nosuch"`},
		{[]string{"completion", "nosuch"}, `неизвестная команда "nosuch"`},
		{[]string{"completion", "--nosuch"}, `неизвестный флаг "--nosuch"`},
		{[]string{"completion", "zsh", "x"}, `лишний аргумент "x"`},
		{[]string{"completion", "zsh", "--nosuch"}, `неизвестный флаг "--nosuch"`},
		{[]string{"help", "nosuch"}, `неизвестная команда "nosuch"`},
	} {
		out, errOut, err := execute(tc.args...)
		if err == nil || err.Error() != tc.want {
			t.Errorf("%q: err = %v, ожидалась ошибка %q", tc.args, err, tc.want)
		}
		if out != "" || errOut != "" {
			t.Errorf("%q: при ошибке напечатано stdout %q, stderr %q", tc.args, out, errOut)
		}
	}
}
