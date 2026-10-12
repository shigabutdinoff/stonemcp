package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

const configName = "config.json"

//go:embed config.json
var configData []byte

type checksConfig struct {
	Prefixes []string `json:"prefixes"`
	Required []string `json:"required"`
	Exclude  []string `json:"exclude"`
}

func loadConfig() (checksConfig, error) {
	var cfg checksConfig
	dec := json.NewDecoder(bytes.NewReader(configData))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("%s не разобран: %w", configName, err)
	}
	return cfg, nil
}

func staticcheckAnalyzers(cfg checksConfig) ([]*analysis.Analyzer, error) {
	all := slices.Concat(staticcheck.Analyzers, simple.Analyzers, stylecheck.Analyzers, quickfix.Analyzers)
	names := make([]string, 0, len(all))
	for _, a := range all {
		names = append(names, a.Analyzer.Name)
	}
	if err := checkConfig(cfg, names); err != nil {
		return nil, err
	}
	res := make([]*analysis.Analyzer, 0, len(all))
	for _, a := range all {
		name := a.Analyzer.Name
		if !slices.Contains(cfg.Exclude, name) && hasPrefix(name, cfg.Prefixes) {
			res = append(res, a.Analyzer)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("%s не включил ни одной проверки staticcheck", configName)
	}
	if err := checkRequired(cfg.Required, names, res); err != nil {
		return nil, err
	}
	return res, nil
}

func checkConfig(cfg checksConfig, names []string) error {
	if slices.Contains(cfg.Required, "") {
		return fmt.Errorf("%s: пустой префикс в required", configName)
	}
	if !slices.ContainsFunc(cfg.Required, func(prefix string) bool { return strings.HasPrefix("SA", prefix) }) {
		return fmt.Errorf("%s: класс SA не объявлен обязательным", configName)
	}
	for _, name := range cfg.Exclude {
		if !slices.Contains(names, name) {
			return fmt.Errorf("%s: неизвестная проверка %q в exclude", configName, name)
		}
		if hasPrefix(name, cfg.Required) {
			return fmt.Errorf("%s: проверка %q обязательна, её нельзя выключить", configName, name)
		}
	}
	for _, prefix := range cfg.Prefixes {
		if !slices.ContainsFunc(names, func(name string) bool { return strings.HasPrefix(name, prefix) }) {
			return fmt.Errorf("%s: префикс %q не совпал ни с одной проверкой", configName, prefix)
		}
	}
	return nil
}

func checkRequired(required, names []string, enabled []*analysis.Analyzer) error {
	for _, prefix := range required {
		matched := false
		for _, name := range names {
			if !strings.HasPrefix(name, prefix) {
				continue
			}
			matched = true
			if !slices.ContainsFunc(enabled, func(a *analysis.Analyzer) bool { return a.Name == name }) {
				return fmt.Errorf("%s: обязательная проверка %q не включена", configName, name)
			}
		}
		if !matched {
			return fmt.Errorf("%s: обязательный префикс %q не совпал ни с одной проверкой", configName, prefix)
		}
	}
	return nil
}

func hasPrefix(name string, prefixes []string) bool {
	return slices.ContainsFunc(prefixes, func(prefix string) bool { return strings.HasPrefix(name, prefix) })
}
