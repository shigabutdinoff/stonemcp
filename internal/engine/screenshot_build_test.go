package engine

import (
	"math"
	"testing"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestScreenshotBuildInChunks(t *testing.T) {
	job := &shotJob{}
	if err := job.initialize(Screen{W: 3, H: 2}, 7); err != nil {
		t.Fatal(err)
	}
	if _, complete := job.appendValues([]mc.Value{"A", ""}, 7); complete {
		t.Fatal("неполный экран завершён")
	}
	if err := job.initialize(Screen{W: 1, H: 1}, 8); err != nil {
		t.Fatal(err)
	}
	if _, complete := job.appendValues([]mc.Value{"fnc_err3", " "}, 8); complete {
		t.Fatal("неполный экран завершён")
	}
	got, complete := job.appendValues([]mc.Value{"Ж", " "}, 10)
	want := Screenshot{Text: "A  \n Ж ", Width: 3, Height: 2, FramesSpanned: 4}
	if !complete || got != want {
		t.Fatalf("сборка экрана: получено %+v, ожидалось %+v", got, want)
	}
}

func TestScreenshotBuildRejectsInvalidDimensions(t *testing.T) {
	for _, screen := range []Screen{
		{W: 0, H: 1}, {W: 1, H: 0}, {W: -1, H: 1}, {W: 1, H: -1},
		{W: 1025, H: 1024}, {W: math.MaxInt, H: math.MaxInt},
	} {
		job := &shotJob{}
		if err := job.initialize(screen, 1); err == nil {
			t.Fatalf("недопустимые размеры приняты: %+v", screen)
		}
	}
}
