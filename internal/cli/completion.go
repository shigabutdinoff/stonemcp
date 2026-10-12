package cli

import "github.com/spf13/cobra"

func localizeCompletionCmd(completion *cobra.Command) {
	name := completion.Root().Name()
	load := map[string]string{
		"bash":       "source <(" + name + " completion bash)",
		"zsh":        "source <(" + name + " completion zsh)",
		"fish":       name + " completion fish | source",
		"powershell": name + " completion powershell | Out-String | Invoke-Expression",
	}
	completion.Short = "Вывести скрипт автодополнения для оболочки"
	completion.Long = "Выводит скрипт автодополнения команд " + name + " для bash, zsh, fish или powershell.\n" +
		"Как подключить скрипт, сказано в справке каждой оболочки."
	completion.Args = noSubcommand
	completion.RunE = showHelp
	for _, shell := range completion.Commands() {
		shell.Short = "Скрипт автодополнения для " + shell.Name()
		shell.Long = "Выводит скрипт автодополнения для " + shell.Name() + ". Подключить в текущем сеансе:\n\n  " + load[shell.Name()]
		shell.Args = NoArgs
	}
}
