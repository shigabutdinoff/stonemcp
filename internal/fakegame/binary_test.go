package fakegame

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBinaryInFindsModuleRoot(t *testing.T) {
	if _, err := binaryIn(t.TempDir(), "linux", "amd64"); err == nil {
		t.Error("каталог без go.mod не дал ошибку")
	}
	root := t.TempDir()
	deep := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for goos, name := range map[string]string{"linux": "fakegame-linux-amd64", "windows": "fakegame-windows-amd64.exe"} {
		want := filepath.Join(root, "internal", "fakegame", "bin", name)
		if got, err := binaryIn(deep, goos, "amd64"); err != nil || got != want {
			t.Errorf("%s: путь %q, %v, ожидалось %q", goos, got, err, want)
		}
	}
}

func TestLaunchReportsPlatform(t *testing.T) {
	_, err := launch(t, filepath.Join(t.TempDir(), "нет"))
	if err == nil || !strings.Contains(err.Error(), runtime.GOOS+"/"+runtime.GOARCH) {
		t.Errorf("запуск без бинарника: %v, ожидалась ошибка с платформой", err)
	}
}
