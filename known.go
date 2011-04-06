package bayes

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
func (this *CountKnown) Update(s, a uint64) (next KnownBelief) {
	nk := new(CountKnown)
	nk.numStates = this.numStates
	nk.visits = make([]int, len(this.visits))
	copy(nk.visits, this.visits)
	nk.threshold = this.threshold

	nk.visits[s+nk.numStates*a]++
	next = nk
	return
}
func (this *CountKnown) Known(s, a uint64) (known bool) {
	return this.visits[s+this.numStates*a] >= this.threshold
}
