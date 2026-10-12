package cli

import (
	"errors"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/pflag"
)

var flagErrors = []struct{ arg, want string }{
	{"--name", `флагу "--name" нужно значение`},
	{"-n", `флагу "-n" нужно значение`},
	{"--nosuch", `неизвестный флаг "--nosuch"`},
	{"-x", `неизвестный флаг "-x"`},
	{"-hx", `неизвестный флаг "-x"`},
	{"--help=да", `неверное значение "да" флага "--help"`},
	{"---x", `неверный синтаксис флага "---x"`},
	{"--name=x", "неверный флаг"},
}

func parseFlag(arg string) error {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringP("name", "n", "", "")
	flags.BoolP("help", "h", false, "")
	if err := flags.Parse([]string{arg}); err != nil {
		return err
	}
	return errors.New(arg)
}

func TestFlagError(t *testing.T) {
	for _, tc := range flagErrors {
		if got := flagError(nil, parseFlag(tc.arg)).Error(); got != tc.want {
			t.Errorf("ошибка флага %q = %q, ожидалось %q", tc.arg, got, tc.want)
		}
	}
}

func FuzzFlagError(f *testing.F) {
	for _, tc := range flagErrors {
		f.Add(tc.arg)
	}
	russian := regexp.MustCompile(`^(неизвестный флаг|флагу|неверное значение|неверный синтаксис флага|неверный флаг)`)
	f.Fuzz(func(t *testing.T, arg string) {
		if got := flagError(nil, parseFlag(arg)).Error(); !russian.MatchString(got) {
			t.Errorf("ошибка флага %q не на русском: %q", arg, got)
		}
	})
}
