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

func releaseConfig(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../.goreleaser.yaml")
	if err != nil {
		t.Fatalf("конфиг GoReleaser не прочитан: %v", err)
	}
	return string(data)
}

func platforms(t *testing.T) []string {
	t.Helper()
	for _, line := range strings.Split(releaseConfig(t), "\n") {
		if list, ok := strings.CutPrefix(strings.TrimSpace(line), "targets:"); ok {
			return strings.Fields(strings.NewReplacer("[", " ", "]", " ", ",", " ", "_", "/").Replace(list))
		}
	}
	t.Fatal("в конфиге GoReleaser нет списка targets")
	return nil
}

func TestReleaseConfig(t *testing.T) {
	t.Parallel()
	want := []string{"darwin/amd64", "darwin/arm64", "linux/amd64", "linux/arm64", "windows/amd64", "windows/arm64"}
	if got := platforms(t); !slices.Equal(got, want) {
		t.Errorf("платформы сборки = %v, ожидались %v", got, want)
	}
	cfg := releaseConfig(t)
	for _, part := range []string{
		"tags: [" + releaseTag + "]", "main: ./cmd/stonemcp", "binary: stonemcp-",
		"go run ./tools/buildnumber .buildnumber",
		"-X main.buildVersion=", "-X main.buildNumber=", "-X main.buildDate=", "-X main.buildCommit=",
	} {
		if !strings.Contains(cfg, part) {
			t.Errorf("в конфиге GoReleaser нет %q", part)
		}
	}
}
