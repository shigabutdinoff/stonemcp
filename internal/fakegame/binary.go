//go:build !release

package fakegame

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var binaryPath = sync.OnceValues(func() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return binaryIn(dir, runtime.GOOS, runtime.GOARCH)
})

func binaryIn(dir, goos, goarch string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("fakegame: корень модуля с go.mod не найден")
		}
		dir = parent
	}
	name := "fakegame-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, "internal", "fakegame", "bin", name), nil
}
