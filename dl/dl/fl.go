package main

import (
	"gonicetrace.googlecode.com/hg/nicetrace"
	"goargcfg.googlecode.com/hg/argcfg"
	"go-glue.googlecode.com/hg/rlglue"
	"go-glue.googlecode.com/hg/rltools/discrete"
	"github.com/skelterjohn/rlbayes"
	"github.com/skelterjohn/rlalg/bfs3"
	"github.com/skelterjohn/rlbayes/dl"
	"github.com/skelterjohn/rlenv/coffee"
)

func GetCoffeePrior(cfg Config) bfs3.Prior {
	return func(task *rlglue.TaskSpec) (prior bayes.BeliefState) {
		mdp := coffee.NewMDP(cfg.Coffee)
		var transition bayes.TransitionBelief = dl.NewBelief(cfg.DL, task)
		if cfg.FDM {
			bg := new(bayes.FDMTransitionBaggage)
			bg.NumStates = task.Obs.Ints.Count()
			bg.NumActions = task.Act.Ints.Count()
			bg.NextToOutcome = func(s discrete.State, n discrete.State) discrete.State {
				return n
			}
			bg.OutcomeToNext = bg.NextToOutcome
			bg.Alpha = make([]float64, bg.NumStates)
			for i := range bg.Alpha {
				bg.Alpha[i] = .1
			}
			bg.ForgetThreshold = cfg.N
			transition = bayes.NewFDMTransition(bg)
		}
		reward := &bayes.MDPReward{mdp}
		terminal := &bayes.MDPTerminal{&bayes.MDPTransition{mdp}}
		prior = bayes.NewBelief(0, reward, transition, terminal, nil)
		return
	}
}

type Config struct {
	Coffee	coffee.Config
	BFS3	bfs3.Config
	DL	dl.Config
	N	uint64
	FDM	bool
}

var cfg Config

func ConfigDefault() (cfg Config) {
	cfg.Coffee = coffee.ConfigDefault()
	cfg.BFS3 = bfs3.ConfigDefault()
	cfg.DL = dl.ConfigDefault()
	cfg.DL.Alpha = .1
	cfg.DL.M = 10
	cfg.N = 10
	cfg.FDM = false
	return
}
func NewAgent(cfg Config) (agent *FLAgent) {
	prior := GetCoffeePrior(cfg)
	agent = new(FLAgent)
	agent.BFS3Agent = bfs3.New(prior)
	agent.Cfg = cfg.BFS3
	return
}

type FLAgent struct{ *bfs3.BFS3Agent }

func (this *FLAgent) AgentStep(reward float64, state rlglue.Observation) (act rlglue.Action) {
	if !cfg.FDM {
		bayesbelief := this.GetBelief().(*bayes.Belief)
		dltrans := bayesbelief.Transition.(*dl.Belief)
		dltrans.ConsiderFlipAll()
	}
	return this.BFS3Agent.AgentStep(reward, state)
}
func main() {
	defer nicetrace.Print()
	cfg = ConfigDefault()
	argcfg.LoadArgs(&cfg)
	agent := NewAgent(cfg)
	rlglue.LoadAgent(agent)
}
