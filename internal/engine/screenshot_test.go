package engine

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
)

func TestScreenshotSpansFourFrames(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen([]string{"12345678", "abcdefgh", "ABCDEFGH"})
	e := startEngine(t, g, Config{FrameBudget: 40})
	shot, err := e.Screenshot(testContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if shot.Text != "12345678\nabcdefgh\nABCDEFGH" || shot.Width != 8 || shot.Height != 3 || shot.FramesSpanned != 4 {
		t.Fatalf("неверный экран: %+v", shot)
	}
}

func TestScreenshotPreservesSpaces(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen([]string{" A  ", "    ", "B   "})
	e := startEngine(t, g, Config{FrameBudget: 40})
	shot, err := e.Screenshot(testContext(t))
	if err != nil || shot.Text != " A  \n    \nB   " || shot.Width != 4 || shot.Height != 3 {
		t.Fatalf("экран потерял клетки с пробелами: %+v, %v", shot, err)
	}
}

func TestScreenshotCancellationDoesNotBlockNext(t *testing.T) {
	g := fakegame.Start(t)
	g.SetScreen([]string{strings.Repeat("x", 80)})
	e := startEngine(t, g, Config{FrameBudget: 35})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := e.Screenshot(ctx); !errors.Is(err, ErrNotTicking) {
		t.Fatalf("истечение ожидания экрана: %v", err)
	}
	g.SetScreen([]string{"new"})
	eventually(t, "размер нового экрана", func() bool { s, _ := e.Snapshot(); return s.Screen.W == 3 })
	shot, err := e.Screenshot(testContext(t))
	if err != nil || shot.Text != "new" || shot.FramesSpanned != 3 {
		t.Fatalf("следующий экран: %+v, %v", shot, err)
	}
}
