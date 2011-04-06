package dl

import (
	"fmt"
	"time"
	"testing"
	"gonicetrace.googlecode.com/hg/nicetrace"
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rlglue"
)

func abs(f float64) float64 {
	if f < 0 {
		return -1 * f
	}
	return f
}

const (
	seed	= 6
	steps	= 1e3
	alpha	= 1
	M	= 20
)

func TestDepMatch(t *testing.T) {
	defer nicetrace.Print()
	stat.Seed(seed)
	tstr := "VERSION RL-Glue-3.0 PROBLEMTYPE episodic DISCOUNTFACTOR 1 OBSERVATIONS INTS (0 4) (-1 1) (0 2) ACTIONS INTS (0 1) REWARDS (0 1.0)"
	task, _ := rlglue.ParseTaskSpec(tstr)
	stateRanges := task.Obs.Ints
	actionRanges := task.Act.Ints
	cfg := ConfigDefault()
	cfg.Alpha = alpha
	cfg.M = M
	genDLs := []*DepLearner{NewDepLearner(0, cfg, stateRanges, actionRanges), NewDepLearner(1, cfg, stateRanges, actionRanges), NewDepLearner(2, cfg, stateRanges, actionRanges)}
	genDLs[0].SetParents(ParentSet(0).Insert(0, 1))
	genDLs[1].SetParents(ParentSet(0).Insert(0, 2))
	genDLs[2].SetParents(ParentSet(0).Insert(2))
	patDLs := []*DepLearner{NewDepLearner(0, cfg, stateRanges, actionRanges), NewDepLearner(1, cfg, stateRanges, actionRanges), NewDepLearner(2, cfg, stateRanges, actionRanges)}
	numStates := stateRanges.Count()
	numActions := actionRanges.Count()
	RS := stat.Range(int64(numStates))
	RA := stat.Range(int64(numActions))
	startTime := time.Nanoseconds()
	lastWrongStep := make([]int, len(genDLs))
	for i := 0; i < steps; i++ {
		s := uint64(RS())
		a := uint64(RA())
		nv := make([]int32, len(genDLs))
		for child := 0; child < len(nv); child++ {
			nv[child] = genDLs[child].Next(s, a)
		}
		for child := 0; child < len(nv); child++ {
			genDLs[child] = genDLs[child].Update(s, a, nv[child])
		}
		for child := 0; child < len(nv); child++ {
			patDLs[child] = patDLs[child].Update(s, a, nv[child])
		}
		if i%1 == 0 {
			for child := 0; child < len(nv); child++ {
				patDLs[child].ConsiderRandomFlip()
				if genDLs[child].parents != patDLs[child].parents {
					lastWrongStep[child] = i
				}
			}
		}
	}
	fmt.Println(lastWrongStep)
	endTime := time.Nanoseconds()
	duration := endTime - startTime
	if true {
		fmt.Printf("Ran in %fms\n", float64(duration)/1e6)
	}
	for child := range genDLs {
		if genDLs[child].parents != patDLs[child].parents {
			t.Error(fmt.Sprintf("%d: %v != %v", child, genDLs[child].parents.Slice(), patDLs[child].parents.Slice()))
		}
	}
}
func TestBeliefMatch(t *testing.T) {
	defer nicetrace.Print()
	stat.Seed(seed)
	tstr := "VERSION RL-Glue-3.0 PROBLEMTYPE episodic DISCOUNTFACTOR 1 OBSERVATIONS INTS (0 4) (-1 1) (0 2) ACTIONS INTS (0 1) REWARDS (0 1.0)"
	task, _ := rlglue.ParseTaskSpec(tstr)
	cfg := ConfigDefault()
	cfg.Alpha = alpha
	cfg.M = M
	beliefG := NewBelief(cfg, task)
	beliefG.learners[0].SetParents(ParentSet(0).Insert(0, 1))
	beliefG.learners[1].SetParents(ParentSet(0).Insert(0, 2))
	beliefG.learners[2].SetParents(ParentSet(0).Insert(2))
	beliefP := NewBelief(cfg, task)
	numStates := task.Obs.Ints.Count()
	numActions := task.Act.Ints.Count()
	RS := stat.Range(int64(numStates))
	RA := stat.Range(int64(numActions))
	startTime := time.Nanoseconds()
	lastWrongStep := make([]int, len(task.Obs.Ints))
	for i := 0; i < steps; i++ {
		s := uint64(RS())
		a := uint64(RA())
		n := beliefG.Next(s, a)
		beliefG = beliefG.Update(s, a, n).(*Belief)
		beliefP = beliefP.Update(s, a, n).(*Belief)
		if i%1 == 0 {
			beliefP.ConsiderRandomFlipAll()
			for child, learner := range beliefP.learners {
				if beliefG.learners[child].parents != learner.parents {
					lastWrongStep[child] = i
				}
			}
		}
	}
	fmt.Println(lastWrongStep)
	endTime := time.Nanoseconds()
	duration := endTime - startTime
	if true {
		fmt.Printf("Ran in %fms\n", float64(duration)/1e6)
	}
	for child := range beliefP.learners {
		if beliefG.learners[child].parents != beliefP.learners[child].parents {
			t.Error(fmt.Sprintf("%d: %v != %v", child, beliefG.learners[child].parents.Slice(), beliefP.learners[child].parents.Slice()))
		}
	}
}
func TestDepLL(t *testing.T) {
	stat.Seed(240)
	cfg := ConfigDefault()
	stateRanges := rlglue.IntRanges{rlglue.IntRange{0, 1}, rlglue.IntRange{0, 1}}
	actionRanges := rlglue.IntRanges{rlglue.IntRange{0, 0}}
	dl := NewDepLearner(0, cfg, stateRanges, actionRanges)
	for s := uint64(0); s < dl.bg.numStates; s++ {
		sv := dl.bg.stateValues[s]
		for a := uint64(0); a < dl.bg.numActions; a++ {
			for i := 0; i < 10; i++ {
				if (sv[0] == 0) != (stat.NextUniform() < .9) {
					dl = dl.Update(s, a, 1)
				} else {
					dl = dl.Update(s, a, 0)
				}
			}
		}
	}
	ll0 := dl.ParentSetLoglihoodRatio(ParentSet(0))
	if abs(ll0-dl.mappedLoglihood) > .0001 {
		t.Error("incremental ll off")
	}
	ll1 := dl.ParentSetLoglihoodRatio(ParentSet(0).Insert(0))
	ll2 := dl.ParentSetLoglihoodRatio(ParentSet(0).Insert(1))
	ll3 := dl.ParentSetLoglihoodRatio(ParentSet(0).Insert(0).Insert(1))
	if abs(ll1-ll3-1.931146) > .0001 || abs(ll2-ll3+9.941318) > .0001 {
		t.Error(fmt.Sprintf("got wrong lls: %f, %f", ll1-ll3, ll2-ll3))
	}
	dl.Consider(ParentSet(0))
	dl.Consider(ParentSet(0).Insert(0))
	dl.Consider(ParentSet(0).Insert(1))
	dl.Consider(ParentSet(0).Insert(0).Insert(1))
	if !dl.parents.Contains(0) || dl.parents.Contains(1) {
		t.Error("Got wrong parents")
	}
}
