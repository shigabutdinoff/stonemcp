package engine

import (
	"context"
	"slices"
)

// Screenshot содержит текст экрана и длительность чтения в кадрах.
type Screenshot struct {
	// Text содержит строки экрана, разделённые переводами строк.
	Text string `json:"text"`
	// Width содержит ширину экрана в клетках.
	Width int `json:"width"`
	// Height содержит высоту экрана в клетках.
	Height int `json:"height"`
	// FramesSpanned содержит число кадров от начала чтения до его конца.
	FramesSpanned int `json:"frames_spanned"`
}

type shotOutcome struct {
	shot Screenshot
	err  error
}
type shotJob struct {
	stop                <-chan struct{}
	done                chan shotOutcome
	width, height, next int
	first               int64
	cells               []string
}

// Screenshot читает экран порциями, размеры фиксируются при первой порции.
// Дедлайн ctx при живом соединении возвращает ErrNotTicking.
func (e *Engine) Screenshot(ctx context.Context) (Screenshot, error) {
	job := &shotJob{stop: ctx.Done(), done: make(chan shotOutcome, 1)}
	e.mu.Lock()
	_, err := e.snapshotLocked()
	if err == nil && ctx.Err() == nil {
		e.shots = append(e.shots, job)
		e.kick()
	}
	e.mu.Unlock()
	if err != nil {
		return Screenshot{}, err
	}
	defer e.removeShot(job)
	result, err := awaitJob(ctx, e, job.done)
	if err != nil {
		return Screenshot{}, err
	}
	return result.shot, result.err
}

func (e *Engine) removeShot(job *shotJob) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.shots = removeJob(e.shots, job)
}

func (e *Engine) finishShot(job *shotJob, outcome shotOutcome) {
	e.removeShot(job)
	job.done <- outcome
}

func (e *Engine) nextShot() *shotJob {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.shots = slices.DeleteFunc(e.shots, func(j *shotJob) bool { return stopped(j.stop) })
	if len(e.shots) == 0 {
		return nil
	}
	return e.shots[0]
}
