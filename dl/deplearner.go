package dl

import (
	"math"
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rlglue"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type DBaggage struct {
	cfg                   Config
	myRange               rlglue.IntRange
	ranges                rlglue.IntRanges
	numStates, numActions uint64
	numOutcomes           uint64
	stateValues           [][]int32
	alphaLogFactor        float64
}
type DepLearner struct {
	bg              *DBaggage
	history         []Histogram
	parents         ParentSet
	cutRanges       rlglue.IntRanges
	mappedHistory   []Histogram
	mappedLoglihood float64
	consistency     int
	hash            uint64
}

func NewDepLearner(child int, cfg Config, stateRanges, actionRanges rlglue.IntRanges) (this *DepLearner) {
	this = new(DepLearner)
	this.bg = new(DBaggage)
	this.bg.cfg = cfg
	this.bg.myRange = stateRanges[child]
	this.bg.ranges = stateRanges
	this.bg.numStates = stateRanges.Count()
	this.bg.numActions = actionRanges.Count()
	this.bg.numOutcomes = stateRanges[child].Count()
	this.bg.alphaLogFactor = stat.LnΓ(this.bg.cfg.Alpha * float64(this.bg.numOutcomes))
	this.bg.alphaLogFactor -= stat.LnΓ(this.bg.cfg.Alpha) * float64(this.bg.numOutcomes)
	this.bg.stateValues = make([][]int32, this.bg.numStates)
	for s := range this.bg.stateValues {
		this.bg.stateValues[s] = this.bg.ranges.Values(uint64(s))
	}
	this.history = make([]Histogram, this.bg.numStates*this.bg.numActions)
	for i := range this.history {
		this.history[i] = make(Histogram, this.bg.numOutcomes)
	}
	this.cutRanges = this.parents.CutRanges(this.bg.ranges)
	this.mappedHistory = this.MakeMappedHistory(this.parents, this.cutRanges)
	this.mappedLoglihood = this.MappedLoglihoodRatio(this.mappedHistory)
	return
}
func (this *DepLearner) Update(s discrete.State, a discrete.Action, o int32) (next *DepLearner) {
	k := a.Hashcode() + this.bg.numActions*s.Hashcode()
	next = new(DepLearner)
	*next = *this
	oi := this.bg.myRange.Index(o)
	next.history = append([]Histogram{}, this.history...)
	next.history[k] = next.history[k].Incr(oi)
	sv := next.bg.stateValues[s]
	mv := next.parents.CutValues(sv)
	ms := next.cutRanges.Index(mv)
	mk := a.Hashcode() + this.bg.numActions*ms
	next.mappedHistory = append([]Histogram{}, this.mappedHistory...)
	next.mappedLoglihood += next.mappedHistory[mk].LogFactorAlpha(this.bg.cfg.Alpha)
	next.mappedHistory[mk] = next.mappedHistory[mk].Incr(oi)
	next.mappedLoglihood -= next.mappedHistory[mk].LogFactorAlpha(this.bg.cfg.Alpha)
	next.hash += k << oi
	return
}
func (this *DepLearner) Next(s discrete.State, a discrete.Action) (o int32) {
	sv := this.bg.stateValues[s]
	mv := this.parents.CutValues(sv)
	ms := this.cutRanges.Index(mv)
	mk := a.Hashcode() + this.bg.numActions*ms
	h := this.mappedHistory[mk]
	lls := make([]float64, len(h))
	usePrior := h.Sum() < this.bg.cfg.M
	for i, c := range h {
		if usePrior {
			lls[i] = math.Log(this.bg.cfg.Alpha + float64(c))
		} else {
			lls[i] = math.Log(float64(c))
		}
	}
	oi := uint64(stat.NextLogChoice(lls))
	o = this.bg.myRange.Value(oi)
	return
}
func (this *DepLearner) SetParents(parents ParentSet) {
	this.parents = parents
	this.cutRanges = this.parents.CutRanges(this.bg.ranges)
	this.mappedHistory = this.MakeMappedHistory(this.parents, this.cutRanges)
	this.mappedLoglihood = this.MappedLoglihoodRatio(this.mappedHistory)
}
func (this *DepLearner) GetCutRanges(parents ParentSet) (cutRanges rlglue.IntRanges) {
	return parents.CutRanges(this.bg.ranges)
}
func (this *DepLearner) MakeMappedHistory(parents ParentSet, cutRanges rlglue.IntRanges) (mappedHistory []Histogram) {
	numMappedStates := cutRanges.Count()
	mappedHistory = make([]Histogram, numMappedStates*this.bg.numActions)
	for i := range mappedHistory {
		mappedHistory[i] = make(Histogram, this.bg.numOutcomes)
	}
	for sk, h := range this.history {
		a := uint64(sk) % this.bg.numActions
		s := uint64(sk) / this.bg.numActions
		sv := this.bg.stateValues[s]
		mv := parents.CutValues(sv)
		ms := cutRanges.Index(mv)
		mk := a + this.bg.numActions*ms
		mappedHistory[mk] = mappedHistory[mk].UpdateHistogram(h)
	}
	return
}
func (this *DepLearner) ConsiderFlips() {
	for parent := range this.bg.ranges {
		this.ConsiderFlip(uint32(parent))
	}
}
func (this *DepLearner) ConsiderRandomFlip() {
	this.ConsiderFlip(uint32(stat.NextRange(int64(len(this.bg.ranges)))))
}
func (this *DepLearner) ConsiderFlipAll() {
	for i := range this.bg.ranges {
		this.ConsiderFlip(uint32(i))
	}
}
func (this *DepLearner) ConsiderFlip(parent uint32) {
	cp := this.parents.Toggle(parent)
	this.Consider(cp)
}
func (this *DepLearner) Consider(np ParentSet) {
	if np == this.parents {
		return
	}
	if stat.NextBernoulli(1/(1+float64(this.consistency))) == 0 {
		return
	}
	alternateRanges := np.CutRanges(this.bg.ranges)
	alternateHistory := this.MakeMappedHistory(np, alternateRanges)
	alternateLoglihood := this.MappedLoglihoodRatio(alternateHistory)
	choiceLL := []float64{this.mappedLoglihood, alternateLoglihood}
	sizeDiff := float64(this.parents.Size(uint32(len(this.bg.ranges))) - np.Size(uint32(len(this.bg.ranges))))
	if sizeDiff > 0 {
		choiceLL[0] += sizeDiff * this.bg.cfg.Kappa
		choiceLL[1] += sizeDiff * (1 - this.bg.cfg.Kappa)
	} else {
		choiceLL[0] -= sizeDiff * (1 - this.bg.cfg.Kappa)
		choiceLL[1] -= sizeDiff * this.bg.cfg.Kappa
	}
	if stat.NextLogChoice(choiceLL) == 1 {
		this.parents = np
		this.cutRanges = alternateRanges
		this.mappedHistory = alternateHistory
		this.mappedLoglihood = alternateLoglihood
		this.consistency = 0
	} else {
		this.consistency++
	}
}
func (this *DepLearner) ParentSetLoglihoodRatio(p ParentSet) (ll float64) {
	cutRanges := p.CutRanges(this.bg.ranges)
	mappedHistory := this.MakeMappedHistory(p, cutRanges)
	ll = this.MappedLoglihoodRatio(mappedHistory)
	return
}
func (this *DepLearner) MappedLoglihoodRatio(mappedHistory []Histogram) (ll float64) {
	ll += float64(len(mappedHistory)) * this.bg.alphaLogFactor
	for _, mh := range mappedHistory {
		ll -= mh.LogFactorAlpha(this.bg.cfg.Alpha)
	}
	return
}
func (this *DepLearner) Hashcode() uint64 {
	return this.hash
}
func (this *DepLearner) LessThan(oi interface{}) bool {
	return this.Compare(oi.(*DepLearner)) < 0
}
func (this *DepLearner) Compare(o *DepLearner) int {
	hd := len(this.history) - len(o.history)
	if hd != 0 {
		return hd
	}
	for i, th := range this.history {
		id := th.Compare(o.history[i])
		if id != 0 {
			return id
		}
	}
	return 0
}
