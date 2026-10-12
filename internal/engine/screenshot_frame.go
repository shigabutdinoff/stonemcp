package engine

import mc "github.com/shigabutdinoff/stonemcp/internal/mindconnect"

func (e *Engine) readScreen(s *session, snapshot *Snapshot, job *shotJob, budget int) error {
	if budget <= 0 {
		return nil
	}
	if job != nil && stopped(job.stop) {
		job = e.nextShot()
	}
	if job == nil {
		return nil
	}
	batches, ok := e.shotBatches(job, snapshot, budget)
	if !ok {
		return nil
	}
	values, err := s.exchangeBatches(batches)
	if err != nil {
		return err
	}
	s.screenFrame = snapshot.Frame
	if shot, complete := job.appendValues(values, snapshot.Frame); complete {
		e.finishShot(job, shotOutcome{shot: shot})
	}
	return nil
}

func (e *Engine) shotBatches(job *shotJob, snapshot *Snapshot, budget int) ([][]mc.Request, bool) {
	batches, err := job.batches(snapshot, budget, e.cfg.MaxPacket)
	if err != nil {
		e.finishShot(job, shotOutcome{err: err})
	}
	return batches, err == nil
}

func (j *shotJob) batches(snapshot *Snapshot, budget, maxPacket int) ([][]mc.Request, error) {
	if err := j.initialize(snapshot.Screen, snapshot.Frame); err != nil {
		return nil, err
	}
	reqs := make([]mc.Request, min(budget, len(j.cells)-j.next))
	for i := range reqs {
		pos := j.next + i
		reqs[i] = mc.Var("draw.GetSymbol", mc.Int(pos%j.width), mc.Int(pos/j.width))
	}
	return mc.SplitBatches(reqs, maxPacket)
}
