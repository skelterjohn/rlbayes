package fl

import (
	"testing"
	"time"
	"fmt"
	"go-glue.googlecode.com/hg/rlglue"
	"github.com/skelterjohn/rlbayes"
	"gostat.googlecode.com/hg/stat"
)

func TestDBNConnections(t *testing.T) {
}
func TestAccuracy(t *testing.T) {
	tstr := "VERSION RL-Glue-3.0 PROBLEMTYPE episodic DISCOUNTFACTOR 1 OBSERVATIONS INTS (0 2) (0 2) ACTIONS INTS (0 0) REWARDS (0 1.0)"
	task, _ := rlglue.ParseTaskSpec(tstr)
	dbn := NewDBN(task)
	dbn = dbn.Update(0, 0, true)
	dbn = dbn.Update(0, 1, true)
	dbn = dbn.Update(1, 1, true)
	fmt.Printf("(%v)\n", dbn)
	ways := []uint64{dbn.Count(0), dbn.Count(1)}
	limits := []uint64{task.Obs.Ints[0].Count(), task.Obs.Ints[1].Count()}
	stat.Seed(4)
	mults := [][][]float64{{}, {}}
	for child := 0; child < len(mults); child++ {
		alpha := make([]float64, limits[child])
		for j := range alpha {
			alpha[j] = 1
		}
		for i := uint64(0); i < ways[child]; i++ {
			m := make([]float64, len(alpha))
			m[int(i)%len(alpha)] = 1
			m = stat.NextDirichlet(alpha)
			mults[child] = append(mults[child], m)
		}
	}
	cfg := ConfigDefault()
	cfg.M = 100
	cfg.Alpha = 1
	cfg.Kappa = 0.5
	var belief bayes.TransitionBelief = NewBelief(cfg, task)
	for i := 0; i < 100000; i++ {
		s := uint64(stat.NextRange(int64(task.Obs.Ints.Count())))
		sv := task.Obs.Ints.Values(s)
		nv := make([]int32, len(task.Obs.Ints))
		for child := 0; child < len(task.Obs.Ints); child++ {
			ci := dbn.Index(sv, child)
			nm := mults[child][ci]
			nv[child] = int32(stat.NextChoice(nm))
		}
		n := uint64(task.Obs.Ints.Index(nv))
		belief = belief.Update(s, 0, n)
	}
	stat.Seed(time.Nanoseconds())
	b := belief.(*Belief)
	fmt.Printf("%v\n", b.dbn)
	for j := 0; j < 10; j++ {
		for i := 0; i < 100; i++ {
			b.ResampleConnection(0, 0)
			b.ResampleConnection(0, 1)
			b.ResampleConnection(1, 0)
			b.ResampleConnection(1, 1)
		}
		fmt.Printf("%v\n", b.dbn)
	}
}
