package bayes

import (
	"go-glue.googlecode.com/hg/rlglue"
)

type FObjBaggage struct {
	Task		*rlglue.TaskSpec
	ObjRanges	rlglue.IntRanges
	ObjCount	uint64
	NumObjs		uint64
	Dimensionality	int
	Alpha		float64
	ForgetThreshold	uint64
}

func NewFObjBaggage(Task *rlglue.TaskSpec, NumObjs uint64, Alpha float64, ForgetThreshold uint64) (this *FObjBaggage) {
	this = new(FObjBaggage)
	this.Task = Task
	this.NumObjs = NumObjs
	this.Alpha = Alpha
	this.ForgetThreshold = ForgetThreshold
	this.Dimensionality = len(this.Task.Obs.Ints) / int(NumObjs)
	this.ObjRanges = this.Task.Obs.Ints[0:this.Dimensionality]
	this.ObjCount = this.ObjRanges.Count()
	return
}
func (this *FObjBaggage) GetObjs(s uint64) (objs []uint64) {
	objs = make([]uint64, this.NumObjs)
	for i := range objs {
		objs[i] = s % this.ObjCount
		s /= this.ObjCount
	}
	return
}
func (this *FObjBaggage) GetState(objs []uint64) (s uint64) {
	values := make([]int32, len(this.Task.Obs.Ints))
	for i, obj := range objs {
		objValues := this.ObjRanges.Values(obj)
		copy(values[i*this.Dimensionality:(i+1)*this.Dimensionality], objValues)
	}
	s = this.Task.Obs.Ints.Index(values)
	return
}

type FObjTransition struct {
	bg	*FObjBaggage
	ObjFDM	*FDMTransition
	hash	uint64
}

func NewFObjTransition(bg *FObjBaggage) (this *FObjTransition) {
	this = new(FObjTransition)
	this.bg = bg
	var fdmBaggage FDMTransitionBaggage
	fdmBaggage.NumStates = bg.ObjRanges.Count()
	fdmBaggage.NumActions = bg.Task.Act.Ints[1].Count()
	fdmBaggage.NextToOutcome = func(s, n uint64) (o uint64) {
		return n
	}
	fdmBaggage.OutcomeToNext = func(s, o uint64) (n uint64) {
		return o
	}
	fdmBaggage.Alpha = make([]float64, fdmBaggage.NumStates)
	for i := range fdmBaggage.Alpha {
		fdmBaggage.Alpha[i] = bg.Alpha
	}
	fdmBaggage.ForgetThreshold = bg.ForgetThreshold
	this.ObjFDM = NewFDMTransition(&fdmBaggage)
	this.hash += this.ObjFDM.Hashcode()
	return
}
func (this *FObjTransition) Hashcode() (hash uint64) {
	return this.hash
}
func (this *FObjTransition) Equals(other interface{}) bool {
	ot := other.(*FObjTransition)
	if this.hash != ot.hash {
		return false
	}
	if !this.ObjFDM.Equals(ot.ObjFDM) {
		return false
	}
	return true
}
func (this *FObjTransition) LessThan(other interface{}) bool {
	ot := other.(*FObjTransition)
	if this.hash < ot.hash {
		return true
	}
	if this.ObjFDM.LessThan(ot.ObjFDM) {
		return true
	}
	return false
}
func (this *FObjTransition) Next(s, a uint64) (n uint64) {
	avalues := this.bg.Task.Act.Ints.Values(a)
	which, act := avalues[0], avalues[1]
	sobjs := this.bg.GetObjs(s)
	nobjs := append([]uint64{}, sobjs...)
	nobjs[which] = this.ObjFDM.Next(sobjs[which], uint64(act))
	n = this.bg.GetState(nobjs)
	return
}
func (this *FObjTransition) Update(s, a, n uint64) (next TransitionBelief) {
	nt := new(FObjTransition)
	*nt = *this
	avalues := this.bg.Task.Act.Ints.Values(a)
	which, act := avalues[0], avalues[1]
	sobjs := this.bg.GetObjs(s)
	nobjs := this.bg.GetObjs(n)
	nt.ObjFDM = this.ObjFDM.Update(sobjs[which], uint64(act), nobjs[which]).(*FDMTransition)
	next = nt
	return
}
