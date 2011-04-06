package roar

import (
	"fmt"
	"gomatrix.googlecode.com/hg/matrix"
	"gostat.googlecode.com/hg/stat"
)

/*
G ~ CRP(alpha_G) // clusters to groups
C ~ CRP(alpha_C) // data to clusteres
mu_i = mean of X_{t|C_t=i}
Sigma_i = V_{G_i}
V_j ~ IW(m, Psi)
X_{t|C_t=i} ~ N(mu_i, Sigma_i)
*/

type PosteriorCFG struct {
	Partition      int
	Calpha, Galpha float64
	M              int
	Psi            *matrix.DenseMatrix
}

func PosteriorCFGDefault() (cfg PosteriorCFG) {
	cfg.Partition = 1
	cfg.Calpha, cfg.Galpha = 1, 1
	cfg.M = 1
	cfg.Psi = matrix.Eye(2)

	return
}

type Posterior struct {
	Cfg PosteriorCFG

	beforePsi *matrix.DenseMatrix
	afterPsi  *matrix.DenseMatrix

	emptyV *IWPosterior // for likelihoods of new clusters/groups
	p      int

	X []*matrix.DenseMatrix // 1...T

	C               *HList
	ClusterIW       []*IWPosterior // 1...I
	SmallClusterIW  []*IWPosterior // 1...I
	ClusterScatters []*Scatter     // 1...I

	G *HList
	V []*IWPosterior // 1...J

	block chan bool
}

func (this *Posterior) Copy() (next *Posterior) {
	this.block <- true
	defer func() { <-this.block }()
	next = new(Posterior)
	*next = *this

	next.X = append([]*matrix.DenseMatrix{}, this.X...)
	next.C = this.C.Copy()
	next.ClusterScatters = make([]*Scatter, len(this.ClusterScatters))
	next.ClusterIW = make([]*IWPosterior, len(this.ClusterScatters))
	next.SmallClusterIW = make([]*IWPosterior, len(this.ClusterScatters))
	for i, s := range this.ClusterScatters {
		next.ClusterScatters[i] = s.Copy()
		next.ClusterIW[i] = this.ClusterIW[i].Copy()
		next.SmallClusterIW[i] = this.SmallClusterIW[i].Copy()
	}
	next.G = this.G.Copy()
	next.V = make([]*IWPosterior, len(this.V))
	for i, v := range this.V {
		next.V[i] = v.Copy()
	}
	next.block = make(chan bool, 1)

	return
}

func New(Cfg PosteriorCFG) (this *Posterior) {
	this = new(Posterior)

	this.Cfg = Cfg
	this.p = this.Cfg.Psi.Rows()

	this.emptyV = NewIWPosterior(this.Cfg.M, this.Cfg.Psi)
	this.G, this.V = new(HList), []*IWPosterior{}
	this.C, this.ClusterScatters = new(HList), []*Scatter{}
	this.ClusterIW = []*IWPosterior{}
	this.SmallClusterIW = []*IWPosterior{}

	this.beforePsi = this.Cfg.Psi.GetMatrix(0, 0, this.Cfg.Partition, this.Cfg.Partition)
	this.afterPsi = this.Cfg.Psi.GetMatrix(this.Cfg.Partition, this.Cfg.Partition, this.p-this.Cfg.Partition, this.p-this.Cfg.Partition)

	this.block = make(chan bool, 1)

	return
}

func (this *Posterior) Insert(x *matrix.DenseMatrix) {
	/*
		fmt.Printf("+Insert\n")
		defer fmt.Printf("-Insert\n")
		fmt.Printf("GHist=%v\nG=%v\n\n", this.GHist, this.G)
		fmt.Printf("CHist=%v\nC=%v\n\n", this.CHist, this.C)
		CheckHistogram(this.C, this.CHist)
		CheckHistogram(this.G, this.GHist)
	*/
	this.block <- true
	defer func() { <-this.block }()

	this.X = append(this.X, x)
}

func (this *Posterior) DropAllAssignments() {
	for t := 0; t < len(this.X); t++ {
		this.DropClusterAssignment(t)
	}
}

func (this *Posterior) DropClusterAssignment(t int) {
	i := this.C.Get(t)
	if i == -1 {
		return
	}
	j := this.G.Get(i)
	if j != -1 {
		this.DropGroupAssignment(i)
	}
	this.C.Drop(t)

	xBefore := this.X[t].GetMatrix(0, 0, this.Cfg.Partition, 1)
	xAfter := this.X[t].GetMatrix(this.Cfg.Partition, 0, this.p-this.Cfg.Partition, 1)

	this.SmallClusterIW[i].Remove(xBefore)
	this.ClusterScatters[i].Remove(xAfter)
	this.ClusterIW[i].Remove(this.X[t])
	if j != -1 && this.C.Count(i) != 0 {
		this.SetGroupAssignment(i, j)
	}
}
func (this *Posterior) SetClusterAssignment(t, i int) {
	oldI := this.C.Get(t)
	if i == oldI {
		return
	}
	this.DropClusterAssignment(t)
	j := this.G.Get(i)
	if j != -1 {
		this.DropGroupAssignment(i)
	}
	this.C.Set(t, i)
	xBefore := this.X[t].GetMatrix(0, 0, this.Cfg.Partition, 1)
	xAfter := this.X[t].GetMatrix(this.Cfg.Partition, 0, this.p-this.Cfg.Partition, 1)
	this.SmallClusterIW[i].Insert(xBefore)
	this.ClusterScatters[i].Insert(xAfter)
	this.ClusterIW[i].Insert(this.X[t])
	if j != -1 {
		this.SetGroupAssignment(i, j)
	}
}

func (this *Posterior) DropGroupAssignment(i int) {
	j := this.G.Get(i)
	if j == -1 {
		return
	}
	this.G.Drop(i)
	this.V[j].RemoveScatter(this.ClusterScatters[i])
}
func (this *Posterior) SetGroupAssignment(i, j int) {
	oldJ := this.G.Get(i)
	if j == oldJ {
		return
	}
	this.DropGroupAssignment(i)
	this.G.Set(i, j)
	this.V[j].InsertScatter(this.ClusterScatters[i])
}

func (this *Posterior) SweepG(temperature float64) {
	//fmt.Printf("+SweepG\n")
	//defer fmt.Printf("-SweepG\n")

	//fmt.Printf("C.h = %v\n", this.C.h)
	//fmt.Printf("G.h = %v\n", this.G.h)
	for i := range this.C.Values() {
		this.ResampleG(i, temperature)
	}
	//fmt.Printf("%v\n", this.G.a)
	//fmt.Printf("%v\n\n", this.C.h)
	//println()
}

func (this *Posterior) SweepC(temperature float64) {
	//fmt.Printf("+SweepC\n")
	//defer fmt.Printf("-SweepC\n")

	//fmt.Printf("C = %v\n", this.C)
	//fmt.Printf("G = %v\n", this.G)
	for t := 0; t < len(this.X); t++ {
		this.ResampleC(t, temperature)
	}

}

func (this *Posterior) ResampleC(t int, temperature float64) {
	//fmt.Printf("+ResampleC(t=%d)\n", t)
	//defer fmt.Printf("-ResampleC\n")

	this.block <- true
	defer func() { <-this.block }()

	this.DropClusterAssignment(t)

	plls := CRPPrior(this.Cfg.Calpha, this.C)
	//fmt.Printf("plls = %v\n", plls)
	for i, pll := range plls {
		if pll == NegInf {
			continue
		}
		plls[i] += this.ClusterLoglihoodRatio(this.X[t], i)
	}

	Anneal(plls, temperature)
	//fmt.Printf("plls' = %v\n", plls)

	newCluster := LogChoice(plls)

	this.SetClusterAssignment(t, newCluster)
}
func (this *Posterior) ResampleG(i int, temperature float64) {
	//	fmt.Printf("\nC = %v\nG = %v\n\n", this.C, this.G)
	//	fmt.Printf("+ResampleG(i=%d)\n", i)
	//	defer fmt.Printf("-ResampleG\n")

	if this.C.Count(i) == 0 {
		return
	}
	this.block <- true
	defer func() { <-this.block }()

	this.DropGroupAssignment(i)

	plls := CRPPrior(this.Cfg.Galpha, this.G)
	//fmt.Printf("%v", plls)
	//println(i)
	//fmt.Printf("%v\n",plls)
	for j, pll := range plls {
		if pll == NegInf {
			continue
		}
		plls[j] += this.GroupLoglihoodRatio(i, j)
	}
	//fmt.Printf("%v\n", plls)
	Anneal(plls, temperature)

	newGroup := LogChoice(plls)
	//newGroup = i

	this.SetGroupAssignment(i, newGroup)
}

func (this *Posterior) PrepareCluster(i int) {
	for i >= len(this.ClusterScatters) {
		this.ClusterScatters = append(this.ClusterScatters, NewScatter(this.p-this.Cfg.Partition))
		this.SmallClusterIW = append(this.SmallClusterIW, NewIWPosterior(this.Cfg.M, this.beforePsi))
		this.ClusterIW = append(this.ClusterIW, NewIWPosterior(this.Cfg.M, this.Cfg.Psi))
	}
}
func (this *Posterior) PrepareGroup(j int) {
	for j >= len(this.V) {
		this.V = append(this.V, NewIWPosterior(this.Cfg.M, this.afterPsi))
	}
}

func (this *Posterior) ClusterLoglihoodRatio(x *matrix.DenseMatrix, i int) (ll float64) {
	this.PrepareCluster(i)

	ll = this.ClusterIW[i].InsertLogRatio(x)

	return
}

func (this *Posterior) SmallClusterLoglihoodRatio(x1 *matrix.DenseMatrix, i int) (ll float64) {
	this.PrepareCluster(i)

	ll = this.SmallClusterIW[i].InsertLogRatio(x1)

	return
}

func (this *Posterior) GroupLoglihoodRatio(i, j int) (ll float64) {
	this.PrepareGroup(j)

	ll = this.V[j].InsertScatterLogRatio(this.ClusterScatters[i])

	return
}

func (this *Posterior) BestCluster(y *matrix.DenseMatrix) (chosenCluster int) {

	plls := CRPPrior(this.Cfg.Calpha, this.C)
	//fmt.Printf("plls = %v\n", plls)
	for i, pll := range plls {
		if pll == NegInf {
			continue
		}
		plls[i] += this.ClusterLoglihoodRatio(y, i)
	}

	ll := NegInf

	for i, cll := range plls {
		if cll > ll {
			ll = cll
			chosenCluster = i
		}
	}

	if this.C.Count(chosenCluster) == 0 {
		return -1
	}

	return
}

func (this *Posterior) ConditionalSample(x1 *matrix.DenseMatrix) (x2 *matrix.DenseMatrix) {

	this.block <- true
	defer func() { <-this.block }()
	//first we choose one

	plls := CRPPrior(this.Cfg.Calpha, this.C)
	//fmt.Printf("plls = %v\n", plls)
	for i, pll := range plls {
		if pll == NegInf {
			continue
		}
		plls[i] += this.SmallClusterLoglihoodRatio(x1, i)
	}

	chosenCluster := LogChoice(plls)

	newCluster := this.C.Count(chosenCluster) == 0
	chosenGroup := this.G.Get(chosenCluster)
	if chosenGroup == -1 || newCluster {
		glls := CRPPrior(this.Cfg.Galpha, this.G)
		chosenGroup = LogChoice(glls)
	}

	this.PrepareGroup(chosenGroup)
	theV := this.V[chosenGroup]
	Sigma := theV.NextCovar()
	x2f := stat.NextMVNormal(theV.S.Mean.Array(), Sigma.Arrays())
	x2 = matrix.MakeDenseMatrix(x2f, Sigma.Rows(), 1)

	return
}

func (this *Posterior) ListClusters() {
	for i, iwp := range this.ClusterIW {
		if this.C.Count(i) == 0 {
			continue
		}
		mean := iwp.S.Mean
		scatter := iwp.S.S
		fmt.Printf("%v\n%v\n\n", mean, scatter)
	}
}
