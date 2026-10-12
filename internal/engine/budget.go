package engine

import "slices"

func (e *Engine) takeEvaluations(budget int, shot *shotJob, screenProgress bool) ([]*evalJob, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.evals = slices.DeleteFunc(e.evals, func(j *evalJob) bool { return stopped(j.stop) })
	if shot != nil {
		budget /= 2
		if len(e.evals) > 0 && e.evals[0].cost > budget && screenProgress {
			return []*evalJob{e.evals[0]}, true
		}
	}
	var jobs []*evalJob
	for _, job := range e.evals {
		if job.cost > budget && (len(jobs) > 0 || shot != nil) {
			break
		}
		jobs = append(jobs, job)
		budget -= job.cost
		if budget <= 0 {
			break
		}
	}
	return jobs, false
}
