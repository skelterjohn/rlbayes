package bayes

import (
	"fmt"
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type DirSA struct {
	counts  []float64
	total   float64
	visits  uint64
	weights []float64
	hash    uint64
}

func NewDirSA(alpha []float64) (next *DirSA) {
	next = new(DirSA)
	next.counts = make([]float64, len(alpha))
	copy(next.counts, alpha)
	for _, a := range alpha {
		next.total += a
	}
	return
}

func (this *DirSA) Hashcode() uint64 {
	return this.hash
}
func (this *DirSA) Equals(other interface{}) bool {
	ot := other.(*DirSA)
	if this.hash != ot.hash {
		return false
	}
	for i, c := range this.counts {
		if c != ot.counts[i] {
			return false
		}
	}
	return true
}
func (this *DirSA) LessThan(other interface{}) bool {
	unv := this == nil || this.visits == 0
	ot := other.(*DirSA)
	ounv := ot == nil || ot.visits == 0

	if unv {
		return !ounv
	}
	if ounv {
		return false
	}

	for i, c := range this.counts {
		if c < ot.counts[i] {
			return true
		}
	}
	return false
}
func (this *DirSA) Next() (n discrete.State) {
	if this.weights == nil {
		this.weights = make([]float64, len(this.counts))
		for i, c := range this.counts {
			this.weights[i] = c / this.total
		}
	}
	n = discrete.State(stat.NextChoice(this.weights))
	return
}
func (this *DirSA) Update(n discrete.State) (next *DirSA) {
	next = new(DirSA)
	next.counts = make([]float64, len(this.counts))
	copy(next.counts, this.counts)
	if n >= discrete.State(len(next.counts)) {
		panic(fmt.Sprintf("%d for %d", n, len(next.counts)))
	}
	next.counts[n] += 1
	next.total = this.total + 1
	next.visits = this.visits + 1
	next.hash = this.hash + n.Hashcode()
	return
}
func (this *DirSA) ForgetPrior(alpha []float64) {
	for i, a := range alpha {
		this.counts[i] -= a
		this.total -= a
	}
	this.weights = nil
}
func (this *DirSA) String() string {
	if this == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", this.counts)
}

type FDMTransitionBaggage struct {
	NumStates, NumActions uint64
	NextToOutcome         func(s, n discrete.State) (o discrete.State)
	OutcomeToNext         func(s, o discrete.State) (n discrete.State)
	Alpha                 []float64
	ForgetThreshold       uint64
}

type FDMTransition struct {
	bg   *FDMTransitionBaggage
	sas  []*DirSA
	hash uint64
}

func NewFDMTransition(bg *FDMTransitionBaggage) (this *FDMTransition) {
	this = new(FDMTransition)
	this.bg = bg
	this.sas = make([]*DirSA, bg.NumStates*bg.NumActions)
	this.hash = 0
	return
}

func (this *FDMTransition) Hashcode() (hash uint64) {
	return hash
}
func (this *FDMTransition) Equals(other interface{}) bool {
	ot := other.(*FDMTransition)
	for i, dirsa := range this.sas {
		if dirsa != nil && ot.sas[i] != nil && !dirsa.Equals(ot.sas[i]) {
			return false
		}
	}
	return true
}
func (this *FDMTransition) LessThan(other interface{}) bool {
	ot := other.(*FDMTransition)
	for i, dirsa := range this.sas {
		if dirsa.LessThan(ot.sas[i]) {
			return true
		}
	}
	return false
}
func (this *FDMTransition) Next(s discrete.State, a discrete.Action) (n discrete.State) {
	k := s.Hashcode() + a.Hashcode()*this.bg.NumStates
	dsa := this.sas[k]
	if dsa == nil {
		dsa = NewDirSA(this.bg.Alpha)
		this.sas[k] = dsa
	}
	n = this.bg.OutcomeToNext(s, dsa.Next())
	return
}
func (this *FDMTransition) Update(s discrete.State, a discrete.Action, n discrete.State) (next TransitionBelief) {
	o := this.bg.NextToOutcome(s, n)
	k := s.Hashcode() + a.Hashcode()*this.bg.NumStates
	dsa := this.sas[k]
	if dsa == nil {
		dsa = NewDirSA(this.bg.Alpha)
		this.sas[k] = dsa
	}

	if this.bg.ForgetThreshold != 0 && dsa.visits >= this.bg.ForgetThreshold {
		next = this
		return
	}

	nextFDM := new(FDMTransition)
	nextFDM.bg = this.bg
	nextFDM.sas = make([]*DirSA, len(this.sas))
	copy(nextFDM.sas, this.sas)
	nextFDM.sas[k] = dsa.Update(o)
	if nextFDM.sas[k].visits == this.bg.ForgetThreshold {
		nextFDM.sas[k].ForgetPrior(this.bg.Alpha)
		//fmt.Printf("%v\n", nextFDM.sas[k])
	}
	nextFDM.hash = this.hash - this.sas[k].Hashcode() + nextFDM.sas[k].Hashcode()
	next = nextFDM

	return
}
func (this *FDMTransition) String() (res string) {
	for s := range discrete.AllStates64(this.bg.NumStates) {
	//for s = 0; s < this.bg.NumStates; s++ {
		res += fmt.Sprintf("\ns%d:", s)
		for a := range discrete.AllActions64(uint64(len(this.sas))/this.bg.NumStates) {
		//for a = 0; a < uint64(len(this.sas))/this.bg.NumStates; a++ {
			k := s.Hashcode() + a.Hashcode()*this.bg.NumStates
			res += this.sas[k].String()
		}
	}
	return
}
