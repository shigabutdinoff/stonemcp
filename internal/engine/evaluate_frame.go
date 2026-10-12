package engine

import mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"

func (e *Engine) runEvaluations(s *session, jobs []*evalJob) (int, error) {
	spent := 0
	for _, job := range jobs {
		if stopped(job.stop) {
			e.removeEval(job)
			continue
		}
		got, err := s.evaluate(job)
		if err != nil {
			return spent, err
		}
		e.finishEval(job, evalOutcome{values: got})
		spent += job.cost
	}
	return spent, nil
}

func (s *session) evaluate(job *evalJob) ([]mc.Value, error) {
	reqs := make([]mc.Request, len(job.queries))
	for i, q := range job.queries {
		if q.Guard != nil {
			q = q.Guard.Check
		}
		reqs[i] = q.request()
	}
	values, err := s.exchange(reqs)
	if err != nil {
		return nil, err
	}
	var guarded []mc.Request
	var positions []int
	for i, q := range job.queries {
		if q.Guard == nil {
			continue
		}
		if q.Guard.OK(values[i]) {
			guarded = append(guarded, q.request())
			positions = append(positions, i)
		} else {
			values[i] = q.Guard.Else
		}
	}
	got, err := s.exchange(guarded)
	if err != nil {
		return nil, err
	}
	for i, pos := range positions {
		values[pos] = got[i]
	}
	return values, nil
}
