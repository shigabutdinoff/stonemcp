package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

type options struct {
	settings
	showVersion bool
}

func parseFlags(args []string, env lookup, stderr io.Writer) (options, error) {
	var o options
	var flags settings
	var config string
	fs := flag.NewFlagSet("stonemcp", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Использование: stonemcp [флаги]")
		fs.PrintDefaults()
	}
	fs.StringVar(&flags.Mode, "mode", engine.ModeSync, "режим MindConnect, поддержан только "+engine.ModeSync)
	fs.StringVar(&flags.Addr, "addr", engine.DefaultAddr, "адрес игры")
	fs.StringVar(&flags.LogFile, "log-file", "", "файл журнала, без флага журнал пишется в stderr")
	fs.StringVar(&config, "c", "", "файл конфигурации JSON, как -config")
	fs.StringVar(&config, "config", "", "файл конфигурации JSON с ключами addr, mode, log_file")
	fs.BoolVar(&o.showVersion, "version", false, "показать версию и выйти")
	if err := fs.Parse(args); err != nil {
		return o, err
	}
	var err error
	o.settings, err = configure(flags, config, fs, env)
	if err != nil {
		fmt.Fprintln(stderr, "stonemcp:", err)
		if errors.As(err, new(usageError)) {
			fs.Usage()
		}
	}
	return o, err
}
