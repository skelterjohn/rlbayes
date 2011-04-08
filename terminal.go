package bayes

import (
	"gostat.googlecode.com/hg/stat"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type KnownTerminal struct {
	Foo func(s discrete.State, a discrete.Action) (t bool)
}

func (this *KnownTerminal) Hashcode() uint64 {
	return 0
}
func (this *KnownTerminal) Equals(other interface{}) bool {
	return this.Foo == other.(*KnownTerminal).Foo
}
func (this *KnownTerminal) LessThan(other interface{}) bool {
	return this.Foo != other.(*KnownTerminal).Foo
}
func (this *KnownTerminal) Next(s discrete.State, a discrete.Action) (t bool) {
	t = this.Foo(s, a)
	return
}
func (this *KnownTerminal) Update(s discrete.State, a discrete.Action, t bool) (next TerminalBelief) {
	return this
}

type BetaTerminal struct {
	NumStates   uint64
	Alpha, Beta float64
	Known       []bool
	Term        []bool
}

func NewBetaTerminal(NumStates, NumActions uint64, Alpha, Beta float64) (this *BetaTerminal) {
	this = new(BetaTerminal)
	this.NumStates = NumStates
	this.Alpha, this.Beta = Alpha, Beta
	this.Known = make([]bool, NumStates*NumActions)
	this.Term = make([]bool, NumStates*NumActions)
	return
}
func (this *BetaTerminal) Hashcode() uint64 {
	return 0
}
func (this *BetaTerminal) Equals(other interface{}) bool {
	obt := other.(*BetaTerminal)
	if this.Alpha != obt.Alpha {
		return false
	}
	if this.Beta != obt.Beta {
		return false
	}
	for i, k := range this.Known {
		if k != obt.Known[i] {
			return false
		}
		if this.Term[i] != obt.Term[i] {
			return false
		}
	}
	return true
}
func (this *BetaTerminal) LessThan(other interface{}) bool {
	obt := other.(*BetaTerminal)
	if this.Alpha < obt.Alpha {
		return true
	}
	if this.Beta < obt.Beta {
		return true
	}
	for i, k := range this.Known {
		if !k && obt.Known[i] {
			return true
		}
		if !this.Term[i] && obt.Term[i] {
			return true
		}
	}
	return false
}
func (this *BetaTerminal) Next(s discrete.State, a discrete.Action) (t bool) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.Known[index] {
		t = this.Term[index]
		return
	}
	prob := this.Alpha / (this.Alpha + this.Beta)
	if stat.NextUniform() < prob {
		t = true
	}
	return
}
func (this *BetaTerminal) Update(s discrete.State, a discrete.Action, t bool) (next TerminalBelief) {
	index := s.Hashcode() + this.NumStates*a.Hashcode()
	if this.Known[index] {
		next = this
		return
	}
	nbt := new(BetaTerminal)
	*nbt = *this
	nbt.Known = append([]bool{}, this.Known...)
	nbt.Term = append([]bool{}, this.Term...)
	nbt.Known[index] = true
	nbt.Term[index] = t
	if t {
		nbt.Alpha++
	} else {
		nbt.Beta++
	}

	return nbt
}
