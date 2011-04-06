package bayes

import (
	"fmt"
	"gohash.googlecode.com/hg/hashlessmap"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type RewardBelief interface {
	hashlessmap.HasherLess
	Next(s, a uint64) (r float64)
	Update(s, a uint64, r float64) (next RewardBelief)
}

type TransitionBelief interface {
	hashlessmap.HasherLess
	Next(s, a uint64) (n uint64)
	Update(s, a, n uint64) (next TransitionBelief)
}

type TerminalBelief interface {
	hashlessmap.HasherLess
	Next(s, a uint64) (t bool)
	Update(s, a uint64, t bool) (next TerminalBelief)
}

type KnownBelief interface {
	Update(s, a uint64) (next KnownBelief)
	Known(s, a uint64) (known bool)
}

type BeliefState interface {
	discrete.Oracle
	Update(action uint64, state uint64, reward float64) (next BeliefState)
	UpdateTerminal(action uint64, reward float64) (next BeliefState)
	Teleport(state uint64)
	GetState() uint64
}

type Belief struct {
	State        uint64
	depth        int
	Reward       RewardBelief
	Transition   TransitionBelief
	TerminalB    TerminalBelief
	Known        KnownBelief
	IsTerminal   bool
	ActionFilter func(belief *Belief, action uint64) bool
	hash         uint64
}

func (this *Belief) ActionAvailable(action uint64) bool {
	if this.ActionFilter != nil {
		return this.ActionFilter(this, action)
	}
	return true
}

func (this *Belief) String() (res string) {
	if this.IsTerminal {
		return "{terminal}"
	}
	res = fmt.Sprintf("{s%d %v}", this.State, this.Transition)
	return
}

func NewBelief(state uint64, reward RewardBelief, transition TransitionBelief, terminal TerminalBelief, known KnownBelief) (this *Belief) {
	this = new(Belief)

	this.State = state
	this.Reward = reward
	this.Transition = transition
	this.IsTerminal = false
	this.TerminalB = terminal
	this.Known = known

	this.hash = this.State
	this.hash += this.Reward.Hashcode()
	this.hash += this.Transition.Hashcode()
	this.hash += this.TerminalB.Hashcode()

	return
}

func (this *Belief) Hashcode() (hash uint64) {
	hash = this.hash
	return
}
func (this *Belief) Equals(other interface{}) bool {
	ob := other.(*Belief)
	return !(this.LessThan(ob) || ob.LessThan(this))
}

func (this *Belief) LessThan(other interface{}) bool {
	ob := other.(*Belief)

	if this.State < ob.State {
		return true
	}
	if this.Reward.LessThan(ob.Reward) {
		return true
	}
	if this.Transition.LessThan(ob.Transition) {
		return true
	}
	if this.TerminalB.LessThan(ob.TerminalB) {
		return true
	}
	if !this.IsTerminal && ob.IsTerminal {
		return true
	}

	return false
}
func (this *Belief) Next(action uint64) (o discrete.Oracle, r float64) {
	n := this.Transition.Next(this.State, action)
	r = this.Reward.Next(this.State, action)
	t := this.TerminalB.Next(this.State, action)
	if this.Known != nil && this.Known.Known(this.State, action) {
		next := new(Belief)
		*next = *this
		next.Teleport(n)
		next.IsTerminal = t
		o = next
	} else {
		if !t {
			o = this.Update(action, n, r)
		} else {
			o = this.UpdateTerminal(action, r)
		}
	}

	return
}
func (this *Belief) Terminal() bool {
	return this.IsTerminal
}
func (this *Belief) GetState() (state uint64) {
	state = this.State
	return
}
func (this *Belief) Teleport(state uint64) {
	this.hash -= this.State
	this.State = state
	this.IsTerminal = false
	this.hash += this.State
}
func (this *Belief) UpdateTerminal(action uint64, r float64) BeliefState {
	next := new(Belief)
	*next = *this //shallow copy
	next.IsTerminal = true
	if this.Known == nil || !this.Known.Known(this.State, action) {
		next.hash -= next.TerminalB.Hashcode()
		next.TerminalB = this.TerminalB.Update(next.State, action, true)
		next.hash += this.TerminalB.Hashcode()
	}
	return next
}
func (this *Belief) Update(action uint64, n uint64, r float64) BeliefState {
	next := new(Belief)

	if this.Known != nil && this.Known.Known(this.State, action) {
		*next = *this //shallow copy
		next.IsTerminal = false
		next.Teleport(n)
	} else {
		next.State = n
		next.depth = this.depth + 1
		next.Reward = this.Reward.Update(this.State, action, r)
		next.Transition = this.Transition.Update(this.State, action, next.State)
		next.IsTerminal = false
		next.TerminalB = this.TerminalB.Update(next.State, action, false)
		next.ActionFilter = this.ActionFilter
		next.hash = next.State
		next.hash += next.Reward.Hashcode()
		next.hash += next.Transition.Hashcode()
		next.hash += next.TerminalB.Hashcode()
	}

	if this.Known != nil {
		next.Known = this.Known.Update(this.State, action)
	}

	return next
}
