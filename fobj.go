package bayes

import (
	"go-glue.googlecode.com/hg/rlglue"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type FObjBaggage struct {
	Task            *rlglue.TaskSpec
	ObjRanges       rlglue.IntRanges
	ObjCount        uint64
	NumObjs         uint64
	Dimensionality  int
	Alpha           float64
	ForgetThreshold uint64
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
func (this *FObjBaggage) GetObjs(s discrete.State) (objs []discrete.State) {
	objs = make([]discrete.State, this.NumObjs)
	for i := range objs {
		objs[i] = s % discrete.State(this.ObjCount)
		s /= discrete.State(this.ObjCount)
	}
	return
}
func (this *FObjBaggage) GetState(objs []discrete.State) (s discrete.State) {
	values := make([]int32, len(this.Task.Obs.Ints))
	for i, obj := range objs {
		objValues := this.ObjRanges.Values(obj.Hashcode())
		copy(values[i*this.Dimensionality:(i+1)*this.Dimensionality], objValues)
	}
	s = discrete.State(this.Task.Obs.Ints.Index(values))
	return
}

type FObjTransition struct {
	bg     *FObjBaggage
	ObjFDM *FDMTransition
	hash   uint64
}

func NewFObjTransition(bg *FObjBaggage) (this *FObjTransition) {
	this = new(FObjTransition)
	this.bg = bg
	var fdmBaggage FDMTransitionBaggage
	fdmBaggage.NumStates = bg.ObjRanges.Count()
	fdmBaggage.NumActions = bg.Task.Act.Ints[1].Count()
	fdmBaggage.NextToOutcome = func(s, n discrete.State) (o discrete.State) {
		return n
	}
	fdmBaggage.OutcomeToNext = func(s, o discrete.State) (n discrete.State) {
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
func (this *FObjTransition) Next(s discrete.State, a discrete.Action) (n discrete.State) {
	avalues := this.bg.Task.Act.Ints.Values(a.Hashcode())
	which, act := avalues[0], avalues[1]
	sobjs := this.bg.GetObjs(s)
	nobjs := append([]discrete.State{}, sobjs...)
	nobjs[which] = this.ObjFDM.Next(sobjs[which], discrete.Action(act))
	n = this.bg.GetState(nobjs)
	return
}
func (this *FObjTransition) Update(s discrete.State, a discrete.Action, n discrete.State) (next TransitionBelief) {
	nt := new(FObjTransition)
	*nt = *this
	avalues := this.bg.Task.Act.Ints.Values(a.Hashcode())
	which, act := avalues[0], avalues[1]
	sobjs := this.bg.GetObjs(s)
	nobjs := this.bg.GetObjs(n)
	nt.ObjFDM = this.ObjFDM.Update(sobjs[which], discrete.Action(act), nobjs[which]).(*FDMTransition)
	next = nt
	return
}
