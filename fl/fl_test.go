package fl

import (
	"testing"
	"gonicetrace.googlecode.com/hg/nicetrace"
	"go-glue.googlecode.com/hg/rlglue"
)

func TestChildLearner(t *testing.T) {
	defer nicetrace.Print()
	cfg := ConfigDefault()
	tstr := "VERSION RL-Glue-3.0 PROBLEMTYPE episodic DISCOUNTFACTOR 1 OBSERVATIONS INTS (0 1) (0 1) ACTIONS INTS (0 1) REWARDS (0 1.0)"
	task, _ := rlglue.ParseTaskSpec(tstr)
	dbn := NewDBN(task)
	dbn = dbn.Update(0, 1, true)
	var bg Baggage
	bg.task = task
	bg.cfg = cfg
	bg.numActions = int(task.Act.Ints.Count())
	bg.numStates = int(task.Obs.Ints.Count())
	bg.numFeatures = int(len(task.Obs.Ints))
	bf := NewFBelief(&bg, 0, 0)
	history := make([]OutcomeHist, 8)
	history[0+0+0] = OutcomeHist{0, 1, 3, 2}
	history[0+0+1] = OutcomeHist{3, 1, 3, 2}
	history[0+2+0] = OutcomeHist{0, 1, 5, 2}
	history[0+2+1] = OutcomeHist{1, 1, 1, 2}
	history[4+0+0] = OutcomeHist{0, 1, 3, 2}
	history[4+0+1] = OutcomeHist{3, 1, 3, 2}
	history[4+2+0] = OutcomeHist{0, 1, 5, 2}
	history[4+2+1] = OutcomeHist{1, 1, 1, 2}
	bf.AcquireMappedHistory(history, dbn)
}
func TestOracle(t *testing.T) {
	defer nicetrace.Print()
	cfg := ConfigDefault()
	tstr := "VERSION RL-Glue-3.0 PROBLEMTYPE episodic DISCOUNTFACTOR 1 OBSERVATIONS INTS (0 1) (0 1) ACTIONS INTS (0 0) REWARDS (0 1.0)"
	task, _ := rlglue.ParseTaskSpec(tstr)
	belief := NewBelief(cfg, task)
	var s uint64
	for i := 0; i < 500; i++ {
		n := belief.Next(s, 0)
		belief = belief.Update(s, 0, n).(*Belief)
		s = n
	}
}
