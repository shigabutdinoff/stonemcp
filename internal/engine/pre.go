package engine

import "time"

func (e *Engine) pre(s *session) error {
	values, err := s.exchange(baseRequests)
	if err != nil {
		return err
	}
	e.frame++
	snapshot := buildSnapshot(values, e.frame, time.Now())
	budget := e.cfg.FrameBudget - len(baseVars)
	shot := e.nextShot()
	jobs, dedicated := e.takeEvaluations(budget, shot, s.screenFrame > 0 && s.screenFrame == e.frame-1)
	spent, err := e.runEvaluations(s, jobs)
	if err != nil {
		return err
	}
	budget -= spent
	if !dedicated {
		if err := e.readScreen(s, snapshot, shot, budget); err != nil {
			return err
		}
	}
	e.publish(snapshot)
	return nil
}
