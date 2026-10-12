package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
)

const modeFlag = "--mode"

type lookup func(string) (string, bool)

type settings struct {
	Addr    string `json:"addr"`
	Mode    string `json:"mode"`
	LogFile string `json:"log_file"`
}

func configure(flags settings, config string, changed func(string) bool, env lookup) (settings, error) {
	s := settings{Addr: engine.DefaultAddr, Mode: engine.ModeSync}
	from := modeFlag
	override(env, "STONEMCP_CONFIG", &config)
	if config != "" {
		if err := loadSettings(config, &s); err != nil {
			return s, fmt.Errorf("файл конфигурации %s: %w", config, err)
		}
		from = "файл конфигурации " + config + ": mode"
	}
	if changed("addr") {
		s.Addr = flags.Addr
	}
	if changed("mode") {
		s.Mode, from = flags.Mode, modeFlag
	}
	if changed("log-file") {
		s.LogFile = flags.LogFile
	}
	override(env, "STONEMCP_ADDR", &s.Addr)
	if override(env, "STONEMCP_MODE", &s.Mode) {
		from = "STONEMCP_MODE"
	}
	override(env, "STONEMCP_LOG_FILE", &s.LogFile)
	return s, checkMode(s.Mode, from)
}

func override(env lookup, name string, field *string) bool {
	v, ok := env(name)
	if ok {
		*field = v
	}
	return ok
}

func checkMode(mode, from string) error {
	switch mode {
	case engine.ModeSync:
		return nil
	case "async", "exclusive":
		return fmt.Errorf("%s %s: %w", from, mode, engine.ErrUnsupportedMode)
	}
	return fmt.Errorf("%s %s: неизвестный режим", from, mode)
}

func loadSettings(path string, s *settings) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	return dec.Decode(s)
}
