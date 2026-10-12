package main

import (
	"github.com/spf13/cobra"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

type options struct {
	settings
}

func bindFlags(root *cobra.Command, env lookup) func() (options, error) {
	var flags settings
	var config string
	f := root.Flags()
	f.StringVar(&flags.Addr, "addr", "", "адрес игры, по умолчанию "+engine.DefaultAddr+"``")
	f.StringVar(&flags.Mode, "mode", "", "режим MindConnect, поддержан только "+engine.ModeSync+"``")
	f.StringVar(&flags.LogFile, "log-file", "", "файл журнала, без флага журнал пишется в stderr``")
	f.StringVarP(&config, "config", "c", "", "файл конфигурации JSON с ключами addr, mode, log_file``")
	return func() (options, error) {
		s, err := configure(flags, config, f.Changed, env)
		return options{s}, err
	}
}
