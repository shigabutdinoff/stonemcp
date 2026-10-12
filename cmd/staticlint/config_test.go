package main

import (
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"
)

func TestLoadConfig(t *testing.T) {
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("config.json не разобран: %v", err)
	}
	if len(cfg.Prefixes) == 0 {
		t.Error("prefixes пуст, ни одна проверка staticcheck не включится")
	}
	if !slices.Contains(cfg.Required, "SA") {
		t.Error("в required нет SA, класс SA перестал быть обязательным")
	}
}

func TestStaticcheckAnalyzers(t *testing.T) {
	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("config.json не разобран: %v", err)
	}
	res, err := staticcheckAnalyzers(cfg)
	if err != nil {
		t.Fatalf("проверки staticcheck не выбраны: %v", err)
	}
	names := analyzerNames(res)
	for _, a := range staticcheck.Analyzers {
		if !slices.Contains(names, a.Analyzer.Name) {
			t.Errorf("проверка %q класса SA не включена", a.Analyzer.Name)
		}
	}
	if slices.Contains(names, "QF1008") {
		t.Error("QF1008 включена, хотя перечислена в exclude")
	}
}

func TestAllAnalyzers(t *testing.T) {
	all, err := allAnalyzers()
	if err != nil {
		t.Fatalf("набор анализаторов не собран: %v", err)
	}
	names := analyzerNames(all)
	for _, want := range []string{"printf", "nilerr", "osexitcheck", "SA1000", "S1000", "ST1005", "QF1001"} {
		if !slices.Contains(names, want) {
			t.Errorf("анализатор %q не попал в набор", want)
		}
	}
}

func analyzerNames(analyzers []*analysis.Analyzer) []string {
	names := make([]string, 0, len(analyzers))
	for _, a := range analyzers {
		names = append(names, a.Name)
	}
	return names
}
