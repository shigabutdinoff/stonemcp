package engine

import (
	"fmt"
	"strings"

	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func (j *shotJob) initialize(screen Screen, frame int64) error {
	if j.first != 0 {
		return nil
	}
	w, h := screen.W, screen.H
	if w <= 0 || h <= 0 || w > (1<<20)/h {
		return fmt.Errorf("недопустимые размеры экрана %d на %d", w, h)
	}
	j.width, j.height, j.first = w, h, frame
	j.cells = make([]string, w*h)
	return nil
}

func (j *shotJob) appendValues(values []mc.Value, frame int64) (Screenshot, bool) {
	for i, value := range values {
		cell := value.String()
		if value.IsError() || cell == "" {
			cell = " "
		}
		j.cells[j.next+i] = cell
	}
	j.next += len(values)
	if j.next != len(j.cells) {
		return Screenshot{}, false
	}
	lines := make([]string, j.height)
	for y := range lines {
		lines[y] = strings.Join(j.cells[y*j.width:(y+1)*j.width], "")
	}
	return Screenshot{
		Text: strings.Join(lines, "\n"), Width: j.width, Height: j.height,
		FramesSpanned: int(frame - j.first + 1),
	}, true
}
