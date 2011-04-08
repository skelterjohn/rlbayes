package bayes

import (
	"fmt"
	"gohash.googlecode.com/hg/hashlessmap"
	"go-glue.googlecode.com/hg/rltools/discrete"
)

type RewardBelief interface {
	hashlessmap.HasherLess
	Next(s discrete.State, a discrete.Action) (r float64)
	Update(s discrete.State, a discrete.Action, r float64) (next RewardBelief)
}

type TransitionBelief interface {
	hashlessmap.HasherLess
	Next(s discrete.State, a discrete.Action) (n discrete.State)
	Update(s discrete.State, a discrete.Action, n discrete.State) (next TransitionBelief)
}

type TerminalBelief interface {
	hashlessmap.HasherLess
	Next(s discrete.State, a discrete.Action) (t bool)
	Update(s discrete.State, a discrete.Action, t bool) (next TerminalBelief)
}

type KnownBelief interface {
	Update(s discrete.State, a discrete.Action) (next KnownBelief)
	Known(s discrete.State, a discrete.Action) (known bool)
}

type BeliefState interface {
	discrete.Oracle
	Update(action discrete.Action, state discrete.State, reward float64) (next BeliefState)
	UpdateTerminal(action discrete.Action, reward float64) (next BeliefState)
	Teleport(state discrete.State)
	GetState() discrete.State
}

type Belief struct {
	State        discrete.State
	depth        int
	Reward       RewardBelief
	Transition   TransitionBelief
	TerminalB    TerminalBelief
	Known        KnownBelief
	IsTerminal   bool
	ActionFilter func(belief *Belief, action discrete.Action) bool
	hash         uint64
}

func (this *Belief) ActionAvailable(action discrete.Action) bool {
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

func NewBelief(state discrete.State, reward RewardBelief, transition TransitionBelief, terminal TerminalBelief, known KnownBelief) (this *Belief) {
	this = new(Belief)

	this.State = state
	this.Reward = reward
	this.Transition = transition
	this.IsTerminal = false
	this.TerminalB = terminal
	this.Known = known

	this.hash = this.State.Hashcode()
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

	if this.State.LessThan(ob.State) {
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
func (this *Belief) Next(action discrete.Action) (o discrete.Oracle, r float64) {
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
func (this *Belief) GetState() (state discrete.State) {
	state = this.State
	return
}
func (this *Belief) Teleport(state discrete.State) {
	this.hash -= this.State.Hashcode()
	this.State = state
	this.IsTerminal = false
	this.hash += this.State.Hashcode()
}
func (this *Belief) UpdateTerminal(action discrete.Action, r float64) BeliefState {
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
func (this *Belief) Update(action discrete.Action, n discrete.State, r float64) BeliefState {
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
		next.hash = next.State.Hashcode()
		next.hash += next.Reward.Hashcode()
		next.hash += next.Transition.Hashcode()
		next.hash += next.TerminalB.Hashcode()
	}

	if this.Known != nil {
		next.Known = this.Known.Update(this.State, action)
	}

	return next
}
