package roar

import (
	"rand"
	"go-glue.googlecode.com/hg/rlglue"
	"gomatrix.googlecode.com/hg/matrix"
)

type AgentCFG struct {
	Scale     float64
	Intensity int
}

func AgentCFGDefault() (cfg AgentCFG) {
	cfg.Scale = 1
	cfg.Intensity = 1
	return
}

type ROARAgent struct {
	Cfg         AgentCFG
	task        *rlglue.TaskSpec
	LastObs     rlglue.Observation
	LastAct     rlglue.Action
	numFeatures int
	rpost       []*Posterior
}

func NewROARAgent(Cfg AgentCFG) (this *ROARAgent) {
	this = new(ROARAgent)
	this.Cfg = Cfg
	return
}
func (this *ROARAgent) AgentInit(taskString string) {
	this.task, _ = rlglue.ParseTaskSpec(taskString)
	this.numFeatures = len(this.task.Obs.Doubles)
}
func (this *ROARAgent) AgentStart(obs rlglue.Observation) rlglue.Action {
	this.LastObs = obs
	return this.GetAction()
}
func (this *ROARAgent) AgentStep(reward float64, obs rlglue.Observation) rlglue.Action {
	last := matrix.MakeDenseMatrix(this.LastObs.Doubles(), this.numFeatures, 1)
	current := matrix.MakeDenseMatrix(obs.Doubles(), this.numFeatures, 1)
	rm := matrix.MakeDenseMatrix([]float64{reward}, 1, 1)
	outcome, _ := current.MinusDense(last)
	sor, _ := last.Augment(outcome)
	sor, _ = sor.Augment(rm)
	actionIndex := this.task.Act.Ints.Index(this.LastAct.Ints())
	this.rpost[actionIndex].Insert(sor)
	this.LastObs = obs
	return this.GetAction()
}
func (this *ROARAgent) AgentEnd(reward float64) {
}
func (this *ROARAgent) AgentCleanup() {
}
func (this *ROARAgent) AgentMessage(message string) string {
	return ""
}
func (this *ROARAgent) GetAction() (act rlglue.Action) {
	index := uint64(rand.Int63n(int64(this.task.Obs.Ints.Count())))
	act = rlglue.NewAction(this.task.Obs.Ints.Values(index), []float64{}, []byte{})
	this.LastAct = act
	return
}
