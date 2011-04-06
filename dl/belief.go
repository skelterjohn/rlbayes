package dl

import (
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rlglue"
	"github.com/skelterjohn/rlbayes"
)

type BBaggage struct {
	cfg			Config
	task			*rlglue.TaskSpec
	numStates, numActions	uint64
	stateValues		[][]int32
}
type Belief struct {
	bg		*BBaggage
	learners	[]*DepLearner
	totals		[]uint64
	hash		uint64
}

func NewBelief(cfg Config, task *rlglue.TaskSpec) (this *Belief) {
	this = new(Belief)
	this.bg = new(BBaggage)
	this.bg.cfg = cfg
	this.bg.task = task
	this.bg.numStates = task.Obs.Ints.Count()
	this.bg.numActions = task.Act.Ints.Count()
	this.bg.stateValues = make([][]int32, this.bg.numStates)
	for s := range this.bg.stateValues {
		this.bg.stateValues[s] = task.Obs.Ints.Values(uint64(s))
	}
	this.learners = make([]*DepLearner, len(task.Obs.Ints))
	for child := range this.learners {
		this.learners[child] = NewDepLearner(child, cfg, task.Obs.Ints, task.Act.Ints)
	}
	this.totals = make([]uint64, this.bg.numStates*this.bg.numActions)
	return
}
func (this *Belief) Hashcode() uint64 {
	return this.hash
}
func (this *Belief) LessThan(oi interface{}) bool {
	return this.Compare(oi.(*Belief)) < 0
}
func (this *Belief) Compare(o *Belief) (c int) {
	for i, l := range this.learners {
		c = l.Compare(o.learners[i])
		if c != 0 {
			return
		}
	}
	return
}
func (this *Belief) Next(s, a uint64) (n uint64) {
	nv := make([]int32, len(this.learners))
	for child, learner := range this.learners {
		nv[child] = learner.Next(s, a)
	}
	n = this.bg.task.Obs.Ints.Index(nv)
	return
}
func (this *Belief) Update(s, a, n uint64) (nextBelief bayes.TransitionBelief) {
	k := a + s*this.bg.numActions
	if this.totals[k] >= this.bg.cfg.M {
		nextBelief = this
		return
	}
	nv := this.bg.stateValues[n]
	next := new(Belief)
	*next = *this
	next.hash = 0
	next.learners = append([]*DepLearner{}, this.learners...)
	for child := range this.learners {
		next.learners[child] = next.learners[child].Update(s, a, nv[child])
		next.hash += next.learners[child].Hashcode() << uint(child)
	}
	next.totals = append([]uint64{}, this.totals...)
	next.totals[k]++
	nextBelief = next
	return
}
func (this *Belief) ConsiderRandomFlip() {
	child := stat.NextRange(int64(len(this.learners)))
	this.learners[child].ConsiderRandomFlip()
}
func (this *Belief) ConsiderRandomFlipAll() {
	for _, learner := range this.learners {
		learner.ConsiderRandomFlip()
	}
}
func (this *Belief) ConsiderFlipAll() {
	for _, learner := range this.learners {
		learner.ConsiderFlipAll()
	}
}
