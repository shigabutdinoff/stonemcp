package cli

import (
	"slices"
	"strings"
	"testing"
)

var englishLeftovers = []string{
	"Usage:", "Available Commands", "Flags:", "Global Flags", "help for",
	"Generate the autocompletion", "for more information", "Additional help topics",
	"[flags]", "[command]", "--no-descriptions", "version for",
}

func TestHelp(t *testing.T) {
	for _, path := range paths() {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			out := mustExecute(t, append(slices.Clone(path), "--help")...)
			if short := mustExecute(t, append(slices.Clone(path), "-h")...); short != out {
				t.Errorf("справка по -h отличается от --help:\n%s", short)
			}
			if topic := mustExecute(t, append([]string{"help"}, path...)...); topic != out {
				t.Errorf("справка через help отличается от --help:\n%s", topic)
			}
			if !strings.Contains(out, "Использование:") {
				t.Errorf("в справке нет раздела использования:\n%s", out)
			}
			for _, phrase := range englishLeftovers {
				if strings.Contains(out, phrase) {
					t.Errorf("в справке осталось %q:\n%s", phrase, out)
				}
			}
		})
	}
}

func TestRootHelp(t *testing.T) {
	out := mustExecute(t)
	if out != mustExecute(t, "--help") {
		t.Errorf("без команды ожидалась справка, получено:\n%s", out)
	}
	for _, want := range []string{
		"  stonemcp [флаги]\n  stonemcp [команда]\n",
		"  completion  Вывести скрипт автодополнения для оболочки\n",
		"  help        Показать справку по команде\n",
		"  -h, --help      показать справку по команде\n",
		"  -v, --version   показать версию, номер, дату и коммит сборки\n",
		`Подробнее о команде: "stonemcp [команда] --help".`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("в справке нет %q:\n%s", want, out)
		}
	}
}

func TestShellHelp(t *testing.T) {
	for shell, load := range map[string]string{
		"bash":       "source <(stonemcp completion bash)",
		"zsh":        "source <(stonemcp completion zsh)",
		"fish":       "stonemcp completion fish | source",
		"powershell": "stonemcp completion powershell | Out-String | Invoke-Expression",
	} {
		if out := mustExecute(t, "completion", shell, "--help"); !strings.Contains(out, load) {
			t.Errorf("в справке %s нет строки подключения %q:\n%s", shell, load, out)
		}
	}
}
