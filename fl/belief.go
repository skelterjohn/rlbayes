package fl

import (
	"fmt"
	"math"
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rlglue"
	"github.com/skelterjohn/rlbayes"
	"github.com/skelterjohn/rlbayes/roar"
)

type Config struct {
	Alpha, Kappa	float64
	M, N		int
}

func ConfigDefault() (cfg *Config) {
	cfg = new(Config)
	cfg.Alpha = .1
	cfg.Kappa = 0.5
	cfg.M = 20
	cfg.N = 3
	return
}

type Baggage struct {
	task			*rlglue.TaskSpec
	cfg			*Config
	numActions, numStates	int
	numFeatures		int
}
type Belief struct {
	bg	*Baggage
	dbn	*DBN
	history	[]OutcomeHist
	cells	[]float64
	totals	[]int
	factors	[]*FBelief
	hash	uint64
}

func NewBelief(cfg *Config, task *rlglue.TaskSpec) (this *Belief) {
	this = new(Belief)
	this.bg = new(Baggage)
	this.bg.task = task
	this.bg.cfg = cfg
	this.bg.numActions = int(task.Act.Ints.Count())
	this.bg.numStates = int(task.Obs.Ints.Count())
	this.bg.numFeatures = len(task.Obs.Ints)
	this.dbn = NewDBN(task)
	this.history = make([]OutcomeHist, this.bg.numStates*this.bg.numActions)
	this.cells = make([]float64, this.bg.numStates*this.bg.numActions)
	this.totals = make([]int, this.bg.numStates*this.bg.numActions)
	for s := range this.history {
		this.history[s] = make(OutcomeHist, this.bg.numStates)
	}
	this.factors = make([]*FBelief, this.bg.numFeatures*this.bg.numActions)
	for child := 0; child < this.bg.numFeatures; child++ {
		for a := 0; a < this.bg.numActions; a++ {
			fb := NewFBelief(this.bg, child, a)
			fb.AcquireMappedHistory(this.history, this.dbn)
			this.factors[child*this.bg.numActions+a] = fb
		}
	}
	return
}
func (this *Belief) String() (res string) {
	res = fmt.Sprintf("{%v, %v, %v}", this.dbn, this.history, this.factors)
	return
}
func (this *Belief) Hashcode() (hash uint64) {
	hash = this.hash
	return
}
func (this *Belief) Compare(other *Belief) int {
	dbnc := this.dbn.Compare(other.dbn)
	if dbnc != 0 {
		return dbnc
	}
	for i, hg := range this.history {
		hgc := hg.Compare(other.history[i])
		if hgc != 0 {
			return hgc
		}
	}
	return 0
}
func (this *Belief) LessThan(oi interface{}) bool {
	other := oi.(*Belief)
	return this.Compare(other) < 0
}
func (this *Belief) Next(s, a uint64) (n uint64) {
	nvals := make([]int32, this.bg.numFeatures)
	for child := 0; child < this.bg.numFeatures; child++ {
		cfb := this.factors[child*this.bg.numActions+int(a)]
		nvals[child] = cfb.Next(s)
	}
	n = this.bg.task.Obs.Ints.Index(nvals)
	return
}
func (this *Belief) Update(s, a, n uint64) (ntb bayes.TransitionBelief) {
	sak := s*uint64(this.bg.numActions) + a
	if this.totals[sak] >= this.bg.cfg.M {
		return this
	}
	next := new(Belief)
	*next = *this
	next.history = append([]OutcomeHist{}, this.history...)
	next.history[sak] = this.history[sak].Incr(n)
	next.cells[sak] = this.history[sak].LogFactorCount()
	next.totals = append([]int{}, this.totals...)
	next.totals[sak]++
	nvals := this.dbn.stateValues[n]
	next.factors = append([]*FBelief{}, this.factors...)
	for child := 0; child < this.bg.numFeatures; child++ {
		ci := child*this.bg.numActions + int(a)
		next.factors[ci] = this.factors[ci].Update(s, nvals[child])
	}
	ntb = next
	return
}
func (this *Belief) ResampleNConnections(n int) {
	for i := 0; i < n; i++ {
		child := int(stat.NextRange(int64(this.bg.numFeatures)))
		parent := int(stat.NextRange(int64(this.bg.numFeatures)))
		this.ResampleConnection(parent, child)
	}
}
func (this *Belief) ResampleDBN() {
	for child := 0; child < this.bg.numFeatures; child++ {
		this.ResampleChild(child)
	}
}
func (this *Belief) ResampleChild(child int) {
	for parent := 0; parent < this.bg.numFeatures; parent++ {
		this.ResampleConnection(parent, child)
	}
}
func (this *Belief) ResampleConnection(parent, child int) {
	connected := this.dbn.Connection(parent, child)
	dbns := []*DBN{this.dbn, this.dbn.Update(child, parent, !connected)}
	lls := []float64{math.Log(1 - this.bg.cfg.Kappa), math.Log(this.bg.cfg.Kappa)}
	if connected {
		lls[0], lls[1] = lls[1], lls[0]
	}
	for a := 0; a < this.bg.numActions; a++ {
		cf := this.factors[child*this.bg.numActions+a]
		for i := 0; i < 2; i++ {
			lls[i] += cf.LoglihoodRatio(this.history, dbns[i])
		}
	}
	newDBN := dbns[roar.LogChoice(lls)]
	for a := 0; a < this.bg.numActions; a++ {
		cf := this.factors[child*this.bg.numActions+a]
		cf.AcquireMappedHistory(this.history, newDBN)
	}
	this.dbn = newDBN
}
