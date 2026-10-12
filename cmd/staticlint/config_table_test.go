package main

import (
	"strings"
	"testing"
)

func TestStaticcheckAnalyzersConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     checksConfig
		wantErr bool
		wantMsg string
	}{
		{"опечатка в exclude", checksConfig{Prefixes: []string{"SA"}, Required: []string{"SA"}, Exclude: []string{"QF1O08"}}, true, ""},
		{"опечатка в prefixes", checksConfig{Prefixes: []string{"SA", "ZZ"}, Required: []string{"SA"}}, true, ""},
		{"пустой prefixes", checksConfig{Required: []string{"SA"}}, true, ""},
		{"класс SA покрыт более широким префиксом", checksConfig{Prefixes: []string{"S"}, Required: []string{"S"}}, false, ""},
		{"класс SA сужен в required", checksConfig{Prefixes: []string{"SA"}, Required: []string{"SA1"}}, true, "класс SA"},
		{"префикс совпал только с исключённой проверкой", checksConfig{Prefixes: []string{"SA", "QF1008"}, Required: []string{"SA"}, Exclude: []string{"QF1008"}}, false, ""},
		{"обязательная проверка в exclude", checksConfig{Prefixes: []string{"SA"}, Required: []string{"SA"}, Exclude: []string{"SA1019"}}, true, "нельзя выключить"},
		{"обязательный префикс не включён", checksConfig{Prefixes: []string{"S1"}, Required: []string{"SA"}}, true, "не включена"},
		{"обязательный префикс включён более широким", checksConfig{Prefixes: []string{"S"}, Required: []string{"SA"}}, false, ""},
		{"обязательный префикс сужен в prefixes", checksConfig{Prefixes: []string{"SA1"}, Required: []string{"SA"}}, true, "не включена"},
		{"неизвестный префикс в required", checksConfig{Prefixes: []string{"SA"}, Required: []string{"SA", "ZZ"}}, true, `"ZZ"`},
		{"опечатка в exclude с обязательным префиксом", checksConfig{Prefixes: []string{"SA"}, Required: []string{"SA"}, Exclude: []string{"SA10190"}}, true, "неизвестная проверка"},
		{"пустой префикс в required", checksConfig{Prefixes: []string{"SA"}, Required: []string{""}, Exclude: []string{"QF1008"}}, true, "пустой префикс"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := staticcheckAnalyzers(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, ждали ошибку %v, включено проверок %d", err, tt.wantErr, len(res))
			}
			if tt.wantMsg != "" && !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("ошибка %v не содержит %q", err, tt.wantMsg)
			}
		})
	}
}
