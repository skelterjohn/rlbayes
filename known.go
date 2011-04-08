package bayes

import (
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type CountKnown struct {
	numStates uint64
	visits    []int
	threshold int
}

func NewCountKnown(numStates, numActions uint64, threshold int) (this *CountKnown) {
	this = new(CountKnown)
	this.numStates = numStates
	this.visits = make([]int, numStates*numActions)
	this.threshold = threshold
	return
}
func (this *CountKnown) Update(s discrete.State, a discrete.Action) (next KnownBelief) {
	nk := new(CountKnown)
	nk.numStates = this.numStates
	nk.visits = make([]int, len(this.visits))
	copy(nk.visits, this.visits)
	nk.threshold = this.threshold

	k := s.Hashcode()+nk.numStates*a.Hashcode()

	nk.visits[k]++
	next = nk
	return
}
func (this *CountKnown) Known(s discrete.State, a discrete.Action) (known bool) {
	k := s.Hashcode()+this.numStates*a.Hashcode()
	return this.visits[k] >= this.threshold
}
