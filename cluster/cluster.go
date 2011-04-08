package cluster

import (
	"github.com/skelterjohn/rlbayes"
	"github.com/skelterjohn/rlbayes/roar"
	"go-glue.googlecode.com/hg/rltools/discrete"
	"gostat.googlecode.com/hg/stat"
)

type Baggage struct {
	OutcomeToNext                      func(s, o discrete.State) (n discrete.State)
	NextToOutcome                      func(s, n discrete.State) (o discrete.State)
	Alpha                              float64
	Beta                               []float64
	NumStates, NumActions, NumOutcomes uint64
}
type SAHist []uint

func (this SAHist) Incr(NumOutcomes uint64, o discrete.State) (next SAHist) {
	next = make([]uint, NumOutcomes)
	if this != nil {
		copy(next, this)
	}
	next[o] += 1
	return
}

type Posterior struct {
	bg          *Baggage
	stateData   []SAHist
	clusterData []SAHist
	C           *roar.HList
	hash        uint64
}

func New(bg *Baggage) (this *Posterior) {
	pairs := bg.NumStates * bg.NumActions
	this = new(Posterior)
	this.bg = bg
	this.stateData = make([]SAHist, pairs)
	this.clusterData = make([]SAHist, pairs)
	this.C = new(roar.HList)
	return
}
func (this *Posterior) Hashcode() (hash uint64) {
	return this.hash
}
func (this *Posterior) LessThan(other interface{}) bool {
	op := other.(*Posterior)
	if this.C.LessThan(op.C) {
		return true
	}
	return false
}
func (this *Posterior) Next(s discrete.State, a discrete.Action) (n discrete.State) {
	c := uint64(this.C.Get(int(s)))
	ck := c*this.bg.NumActions + a.Hashcode()
	hist := this.clusterData[ck]
	fhist := append([]float64{}, this.bg.Beta...)
	total := 0.0
	for i, c := range hist {
		fhist[i] += float64(c)
		total += fhist[i]
	}
	for i := range fhist {
		fhist[i] /= total
	}
	o := discrete.State(stat.NextChoice(fhist))
	n = this.bg.OutcomeToNext(s, o)
	return
}
func (this *Posterior) Update(s discrete.State, a discrete.Action, n discrete.State) (next bayes.TransitionBelief) {
	o := this.bg.NextToOutcome(s, n)
	nextPost := this.UpdatePosterior(s, a, o)
	next = nextPost
	return
}
func (this *Posterior) UpdatePosterior(s discrete.State, a discrete.Action, o discrete.State) (next *Posterior) {
	next = new(Posterior)
	*next = *this
	next.stateData = append([]SAHist{}, this.stateData...)
	next.clusterData = append([]SAHist{}, this.clusterData...)
	next.C = this.C.Copy()
	k := s.Hashcode()*this.bg.NumActions + a.Hashcode()
	next.stateData[k] = next.stateData[k].Incr(this.bg.NumOutcomes, o)
	return
}
func (this *Posterior) Sweep() {
	for s := range discrete.AllStates64(this.bg.NumStates) {
	//for uint64(0); s < this.bg.NumStates; s++ {
		this.ResampleState(s)
	}
}
func (this *Posterior) SampleRandomState() {
	s := discrete.State(stat.NextRange(int64(this.bg.NumStates)))
	this.ResampleState(s)
}
func (this *Posterior) ResampleState(s discrete.State) {
	plls := roar.CRPPrior(this.bg.Alpha, this.C)
	for c := range plls {
		ck := uint64(c) * this.bg.NumActions
		Oc := this.clusterData[ck : ck+this.bg.NumActions]
		sk := s.Hashcode() * this.bg.NumActions
		Os := this.clusterData[sk : sk+this.bg.NumActions]
		plls[c] += InsertLoglihood(this.bg.NumActions, this.bg.NumOutcomes, this.bg.Beta, Oc, Os)
	}
	newCluster := uint(roar.LogChoice(plls))
	this.InsertState(s, newCluster)
}
func (this *Posterior) RemoveState(s discrete.State) {
	hist := this.stateData[s]
	if hist == nil {
		return
	}
	c := uint64(this.C.Get(int(s)))
	this.C.Drop(int(s))
	ck := c * this.bg.NumActions
	for a := uint64(0); a < this.bg.NumActions; a++ {
		for o, n := range hist {
			this.clusterData[ck+a][o] -= n
		}
	}
}
func (this *Posterior) InsertState(s discrete.State, c uint) {
	hist := this.stateData[s]
	if hist == nil {
		return
	}
	this.C.Set(int(s), int(c))
	ck := uint64(c) * this.bg.NumActions
	if this.clusterData[c] == nil {
		this.clusterData[c] = make([]uint, this.bg.NumOutcomes)
	}
	for a := uint64(0); a < this.bg.NumActions; a++ {
		for o, n := range hist {
			this.clusterData[ck+a][o] += n
		}
	}
}
