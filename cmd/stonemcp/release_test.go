package main

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const (
	releaseTag  = "release"
	internalPkg = "github.com/shigabutdinoff/stonemcp/internal/"
)

var testOnly = []string{internalPkg + "fakegame", internalPkg + "linktest"}

func TestReleaseExcludesTestCode(t *testing.T) {
	t.Parallel()
	for _, platform := range platforms(t) {
		goos, goarch, _ := strings.Cut(platform, "/")
		t.Run(goos+"-"+goarch, func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("go", "list", "-tags", releaseTag, "-deps", ".")
			cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("список пакетов не получен: %v\n%s", err, out)
			}
			for _, pkg := range strings.Fields(string(out)) {
				if pkg == "testing" || slices.Contains(testOnly, pkg) {
					t.Errorf("в сборку попал пакет %s", pkg)
				}
			}
		})
	}
	for _, pkg := range testOnly {
		out, err := exec.Command("go", "list", "-tags", releaseTag, "-e", "-f", "{{len .GoFiles}}", pkg).CombinedOutput()
		if got := strings.TrimSpace(string(out)); err != nil || got != "0" {
			t.Errorf("с тегом %q в пакете %s собираются файлы: %q, err = %v", releaseTag, pkg, got, err)
		}
	}
}

func platforms(t *testing.T) []string {
	t.Helper()
	makefile, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatalf("Makefile не прочитан: %v", err)
	}
	for _, line := range strings.Split(string(makefile), "\n") {
		if name, list, ok := strings.Cut(line, ":="); ok && strings.TrimSpace(name) == "PLATFORMS" {
			return strings.Fields(list)
		}
	}
	t.Fatal("в Makefile нет списка PLATFORMS")
	return nil
}
