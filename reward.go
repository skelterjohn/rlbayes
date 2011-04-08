package bayes

import (
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rltools/discrete"
)
// prior for a known reward

type KnownReward struct {
	Foo func(s discrete.State, a discrete.Action) (r float64)
}

func (this *KnownReward) Hashcode() uint64 {
	return 0
}
func (this *KnownReward) Equals(other interface{}) bool {
	return this.Foo == other.(*KnownReward).Foo
}
func (this *KnownReward) LessThan(other interface{}) bool {
	return this.Foo != other.(*KnownReward).Foo
}
func (this *KnownReward) Next(s discrete.State, a discrete.Action) (r float64) {
	return this.Foo(s, a)
}
func (this *KnownReward) Update(s discrete.State, a discrete.Action, r float64) (next RewardBelief) {
	return this
}

/*
 Prior for deterministic yet unknown reward. Prior weight 1-eps for
 it being Rmax, and eps for it being uniform, infinitesmal eps.

 Basically, it will tell you the reward is Rmax until you correct it,
 and from then on it will use that value.
*/

type RmaxReward struct {
	NumStates  uint64
	R          []float64
	countKnown uint64
}

func NewRmaxReward(NumStates, NumActions uint64, Rmax float64) (this *RmaxReward) {
	this = new(RmaxReward)
	this.NumStates = NumStates
	this.R = make([]float64, NumStates*NumActions)
	for i := range this.R {
		this.R[i] = Rmax
	}
	return
}
func (this *RmaxReward) Hashcode() uint64 {
	return this.countKnown
}
func (this *RmaxReward) Equals(other interface{}) bool {
	orr := other.(*RmaxReward)
	if this.countKnown != orr.countKnown {
		return false
	}
	for i := range this.R {
		if this.R[i] != orr.R[i] {
			return false
		}
	}
	return true
}
func (this *RmaxReward) LessThan(other interface{}) bool {
	orr := other.(*RmaxReward)
	if this.countKnown < orr.countKnown {
		return true
	}
	if this.countKnown > orr.countKnown {
		return false
	}
	for i := range this.R {
		if this.R[i] < orr.R[i] {
			return true
		}
		if this.R[i] > orr.R[i] {
			return false
		}
	}
	//equal -> false
	return false
}
func (this *RmaxReward) Next(s discrete.State, a discrete.Action) (r float64) {
	return this.R[s.Hashcode()+this.NumStates*a.Hashcode()]
}
func (this *RmaxReward) Update(s discrete.State, a discrete.Action, r float64) (next RewardBelief) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.R[index] == r {
		return this
	}
	nrr := new(RmaxReward)
	*nrr = *this
	nrr.R = make([]float64, len(this.R))
	copy(nrr.R, this.R)
	nrr.R[index] = r
	nrr.countKnown++
	return nrr
}

type DeterministicReward struct {
	//data structure to keep track of known rewards
	RmaxReward
	//which have been observed
	Known []bool
	//the prior generator
	BaseSampler func() float64
}

func NewDeterministicReward(NumStates, NumActions uint64, BaseSampler func() float64) (this *DeterministicReward) {
	this = new(DeterministicReward)
	this.NumStates = NumStates
	this.R = make([]float64, NumStates*NumActions)
	this.Known = make([]bool, NumStates*NumActions)
	this.BaseSampler = BaseSampler
	return
}

func (this *DeterministicReward) Hashcode() uint64 {
	return this.countKnown
}
func (this *DeterministicReward) Equals(other interface{}) bool {
	odr := other.(*DeterministicReward)
	if !this.RmaxReward.Equals(&odr.RmaxReward) {
		return false
	}

	for i := range this.Known {
		if this.Known[i] != odr.Known[i] {
			return false
		}
	}

	return true
}
func (this *DeterministicReward) LessThan(other interface{}) bool {
	odr := other.(*DeterministicReward)
	if this.RmaxReward.LessThan(&odr.RmaxReward) {
		return true
	}

	for i := range this.Known {
		if !this.Known[i] && odr.Known[i] {
			return true
		}
	}

	return false
}
func (this *DeterministicReward) Next(s discrete.State, a discrete.Action) (r float64) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.Known[index] {
		r = this.R[index]
		return
	}
	r = this.BaseSampler()
	return
}
func (this *DeterministicReward) Update(s discrete.State, a discrete.Action, r float64) (next RewardBelief) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.Known[index] {
		return this
	}
	ndr := new(DeterministicReward)
	*ndr = *this
	ndr.Known = make([]bool, len(this.Known))
	copy(ndr.Known, this.Known)
	ndr.R = make([]float64, len(this.R))
	copy(ndr.R, this.R)
	ndr.Known[index] = true
	ndr.R[index] = r
	ndr.countKnown++
	next = ndr
	return
}

type CRPReward struct {
	DeterministicReward
	Alpha       float64
	Total       uint64
	Counts      []uint64
	SeenRewards []float64
	chooser     func() int64
}

func NewCRPReward(NumStates, NumActions uint64, Alpha float64, BaseSampler func() float64) (this *CRPReward) {
	this = new(CRPReward)
	this.NumStates = NumStates
	this.R = make([]float64, NumStates*NumActions)
	this.Known = make([]bool, NumStates*NumActions)
	this.BaseSampler = BaseSampler
	this.Alpha = Alpha
	this.Counts = make([]uint64, 0)
	this.SeenRewards = make([]float64, 0)
	return
}
func (this *CRPReward) Equals(other interface{}) bool {
	ocr := other.(*CRPReward)
	if !this.DeterministicReward.Equals(&ocr.DeterministicReward) {
		return false
	}
	if len(this.SeenRewards) != len(ocr.SeenRewards) {
		return false
	}
	for i, r := range this.SeenRewards {
		if r != ocr.SeenRewards[i] {
			return false
		}
		if this.Counts[i] != ocr.Counts[i] {
			return false
		}
	}
	return true
}
func (this *CRPReward) LessThan(other interface{}) bool {
	ocr := other.(*CRPReward)
	if this.DeterministicReward.LessThan(&ocr.DeterministicReward) {
		return true
	}
	if len(this.SeenRewards) < len(ocr.SeenRewards) {
		return true
	}
	if len(this.SeenRewards) > len(ocr.SeenRewards) {
		return false
	}
	for i, r := range this.SeenRewards {
		if r < ocr.SeenRewards[i] {
			return true
		}
		if this.Counts[i] < ocr.Counts[i] {
			return true
		}
	}
	return false
}
func (this *CRPReward) Next(s discrete.State, a discrete.Action) (r float64) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.Known[index] {
		r = this.R[index]
		return
	}

	if this.chooser == nil {
		if len(this.Counts) == 0 {
			this.chooser = func() int64 { return 0 }
		} else {
			normalizer := 1.0 / (float64(this.Total) + this.Alpha)
			weights := make([]float64, len(this.Counts))
			for i := range weights {
				weights[i] = float64(this.Counts[i]) * normalizer
			}
			this.chooser = stat.Choice(weights)
		}
	}

	which := int(this.chooser())
	if which == len(this.SeenRewards) {
		r = this.BaseSampler()
	} else {
		r = this.SeenRewards[which]
	}

	return
}
func (this *CRPReward) Update(s discrete.State, a discrete.Action, r float64) (next RewardBelief) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.Known[index] {
		return this
	}
	ndr := new(CRPReward)
	*ndr = *this
	ndr.Known = make([]bool, len(this.Known))
	copy(ndr.Known, this.Known)
	ndr.R = make([]float64, len(this.R))
	copy(ndr.R, this.R)
	ndr.Known[index] = true
	ndr.R[index] = r
	ndr.countKnown++

	ndr.SeenRewards = append([]float64{r}, this.SeenRewards...)
	ndr.Counts = append([]uint64{1}, this.Counts...)
	var seen bool
	for i, sr := range this.SeenRewards {
		if i != 0 && sr == r {
			seen = true
			ndr.Counts[i]++
			break
		}
	}
	if seen {
		ndr.SeenRewards = ndr.SeenRewards[1:len(ndr.SeenRewards)]
		ndr.Counts = ndr.Counts[1:len(ndr.Counts)]
	}

	ndr.Total++

	next = ndr
	return
}
