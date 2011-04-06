package fl

import (
	"fmt"
	"math"
	"github.com/skelterjohn/rlbayes/roar"
)

type FBelief struct {
	bg		*Baggage
	child, action	int
	dbn		*DBN
	fhistory	[]OutcomeHist
}

func NewFBelief(bg *Baggage, child, action int) (this *FBelief) {
	this = new(FBelief)
	this.bg = bg
	this.child = child
	this.action = action
	return
}
func (this *FBelief) String() (res string) {
	res = fmt.Sprintf("{%v, %v, %v}", this.child, this.action, this.fhistory)
	return
}
func (this *FBelief) AcquireMappedHistory(history []OutcomeHist, dbn *DBN) {
	if this.dbn == nil || dbn.Hashcode() != this.dbn.Hashcode() {
		this.dbn = dbn
		this.fhistory = this.dbn.MapDownChild(history, this.child, this.action)
	} else {
		this.dbn = dbn
	}
}
func (this *FBelief) LoglihoodRatio(history []OutcomeHist, dbn *DBN) (ll float64) {
	var phistory []OutcomeHist
	if this.dbn == nil || dbn.HashMask(this.child) != this.dbn.HashMask(this.child) {
		phistory = dbn.MapDownChild(history, this.child, this.action)
	} else {
		phistory = this.fhistory
	}
	for _, phistogram := range phistory {
		ll += phistogram.LoglihoodRatio(this.bg.cfg.Alpha)
	}
	return
}
func (this *FBelief) Next(s uint64) (n int32) {
	pindex := this.Index(s)
	histogram := this.fhistory[pindex]
	cll := make([]float64, len(histogram))
	for i := range cll {
		cll[i] = math.Log(float64(histogram[i]) + this.bg.cfg.Alpha)
	}
	o := int32(roar.LogChoice(cll))
	n = o + this.bg.task.Obs.Ints[this.child].Min
	return
}
func (this *FBelief) Index(s uint64) (parentIndex int) {
	parents := this.dbn.stateValues[s]
	parentIndex = int(this.dbn.Index(parents, this.child))
	return
}
func (this *FBelief) Update(s uint64, n int32) (next *FBelief) {
	o := int32(n) - this.bg.task.Obs.Ints[this.child].Min
	next = new(FBelief)
	*next = *this
	next.fhistory = append([]OutcomeHist{}, this.fhistory...)
	pindex := this.Index(s)
	next.fhistory[pindex] = next.fhistory[pindex].Incr(uint64(o))
	return
}
