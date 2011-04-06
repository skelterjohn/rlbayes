package bayes

import (
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type MDPTransition struct {
	MDP discrete.MDP
}

func (this *MDPTransition) Hashcode() uint64 {
	return 0
}
func (this *MDPTransition) LessThan(oi interface{}) bool {
	return false
}
func (this *MDPTransition) Next(s, a uint64) (n uint64) {
	weights := make([]float64, this.MDP.S())
	for n := uint64(0); n < this.MDP.S(); n++ {
		weights[n] = this.MDP.T(s, a, n)
	}
	n = uint64(stat.NextChoice(weights))
	return
}
func (this *MDPTransition) Update(s, a, n uint64) (next TransitionBelief) {
	return this
}

type MDPReward struct {
	MDP discrete.MDP
}

func (this *MDPReward) Hashcode() uint64 {
	return 0
}
func (this *MDPReward) LessThan(oi interface{}) bool {
	return false
}
func (this *MDPReward) Next(s, a uint64) (r float64) {
	r = this.MDP.R(s, a)
	return
}
func (this *MDPReward) Update(s, a uint64, r float64) (next RewardBelief) {
	next = this
	return
}

type MDPTerminal struct {
	*MDPTransition
}

func (this *MDPTerminal) Hashcode() uint64 {
	return 0
}
func (this *MDPTerminal) LessThan(oi interface{}) bool {
	return false
}
func (this *MDPTerminal) Next(s, a uint64) (t bool) {
	return this.MDPTransition.Next(s, a) == this.MDP.S()
}
func (this *MDPTerminal) Update(s, a uint64, t bool) (next TerminalBelief) {
	return this
}
