package engine

import (
	"fmt"
	"testing"

	"github.com/shigabutdinoff/stonemcp/internal/fakegame"
	mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"
)

func TestScreenshotProgressWithQueuedEvaluations(t *testing.T) {
	for _, tc := range []struct {
		budget, cost, frames int
	}{
		{6, 1, 8}, {5, 1, 8}, {6, 4, 7}, {6, 80, 7}, {1, 1, 47},
	} {
		t.Run(fmt.Sprintf("%d/%d", tc.budget, tc.cost), func(t *testing.T) {
			g := fakegame.Start(t)
			g.SetScreen([]string{"12345678", "abcdefgh", "ABCDEFGH"})
			e := New(Config{FrameBudget: len(baseVars) + tc.budget, MaxPacket: 100})
			s := &session{conn: g.Dial(t), buf: make([]byte, mc.MaxPacket), cfg: e.cfg}
			shot := &shotJob{done: make(chan shotOutcome, 1)}
			e.shots = append(e.shots, shot)
			jobs := make([]*evalJob, 128)
			for i := range jobs {
				queries := make([]Query, tc.cost)
				for j := range queries {
					queries[j].Name = "time"
				}
				jobs[i] = &evalJob{queries: queries, cost: tc.cost, done: make(chan evalOutcome, 1)}
			}
			e.evals = append(e.evals, jobs...)
			for range tc.frames {
				advanceFrame(t, e, s)
			}
			select {
			case out := <-shot.done:
				if out.err != nil || out.shot.Text != "12345678\nabcdefgh\nABCDEFGH" || out.shot.FramesSpanned != tc.frames {
					t.Fatalf("экран при очереди чтений: %+v", out)
				}
			default:
				t.Fatal("очередь чтений задержала экран")
			}
			completed := 0
			for _, job := range jobs {
				select {
				case out := <-job.done:
					completed++
					if out.err != nil || len(out.values) != tc.cost {
						t.Fatalf("чтение при очереди экрана: %+v", out)
					}
					for _, value := range out.values {
						if value != out.values[0] {
							t.Fatalf("значения разных кадров: %v", out.values)
						}
					}
				default:
				}
			}
			if completed == 0 || completed == len(jobs) {
				t.Fatalf("очередь должна продвигаться и оставаться непустой: %d", completed)
			}
		})
	}
}

func advanceFrame(t *testing.T, e *Engine, s *session) {
	t.Helper()
	for _, want := range []mc.Phase{mc.PhasePre, mc.PhasePost} {
		phase, _, err := s.signal()
		if err != nil || phase != want {
			t.Fatalf("шаг игры: %v, %v", phase, err)
		}
		if phase == mc.PhasePre {
			if err := e.pre(s); err != nil {
				t.Fatal(err)
			}
		}
		if err := s.write([]byte(mc.EndStep)); err != nil {
			t.Fatal(err)
		}
	}
}
