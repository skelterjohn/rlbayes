package roar

import (
	"math"
	"gomatrix.googlecode.com/hg/matrix"
	"gostat.googlecode.com/hg/stat"
)

type IWPosterior struct {
	M   int
	Psi *matrix.DenseMatrix
	S   *Scatter

	//distribution quickies
	psiHat *matrix.DenseMatrix

	covarSampler func() [][]float64
	precSampler  func() [][]float64

	logP float64
}

func (this *IWPosterior) Copy() (next *IWPosterior) {
	next = new(IWPosterior)
	*next = *this
	next.S = this.S.Copy()
	return
}

func NewIWPosterior(M int, Psi *matrix.DenseMatrix) (this *IWPosterior) {
	this = new(IWPosterior)
	this.M, this.Psi = M, Psi.Copy()
	this.S = NewScatter(Psi.Rows())
	return
}
func (this *IWPosterior) Insert(x *matrix.DenseMatrix) {
	this.S.Insert(x)
	this.reset()
}
func (this *IWPosterior) InsertScatter(other *Scatter) {
	this.S.InsertScatter(other)
	this.reset()
}
func (this *IWPosterior) Remove(x *matrix.DenseMatrix) {
	this.S.Remove(x)
	this.reset()
}
func (this *IWPosterior) RemoveScatter(other *Scatter) {
	this.S.RemoveScatter(other)
	this.reset()
}
func (this *IWPosterior) reset() {
	this.psiHat = nil
	this.covarSampler = nil
	this.precSampler = nil
	this.logP = 0
}
func (this *IWPosterior) getPsiHat() (psiHat *matrix.DenseMatrix) {
	if this.psiHat == nil {
		this.psiHat, _ = this.Psi.PlusDense(this.S.S)
	}
	return this.psiHat
}
func (this *IWPosterior) NextCovar() (Sigma *matrix.DenseMatrix) {
	if this.covarSampler == nil {
		this.covarSampler = stat.InverseWishart(this.M+this.S.Count, this.getPsiHat().Arrays())
	}
	Sigma = matrix.MakeDenseMatrixStacked(this.covarSampler())
	return
}
func (this *IWPosterior) NextPrecision() (Sigma *matrix.DenseMatrix) {
	if this.precSampler == nil {
		psiHatInv, _ := this.getPsiHat().Inverse()
		this.precSampler = stat.Wishart(this.M+this.S.Count, psiHatInv.Arrays())
	}
	Sigma = matrix.MakeDenseMatrixStacked(this.precSampler())
	return
}
func (this *IWPosterior) InsertLogRatio(x *matrix.DenseMatrix) (lr float64) {
	/*
				Matrix PsiHatPrime = Psi.plus(scatter.getPretendScatter(x));

		        int p = Psi.getColumnDimension();

		        double ll = 0;

		        ll -= 0.5*p*Math.log(Math.PI);

		        ll += Math.log(PsiHat.det())*(m+n)/2;
		        ll -= Math.log(PsiHatPrime.det())*(m+n+1)/2;

		        ll += GammaF.logGammaPRatio(p, m+n+1, m+n);
	*/
	//fmt.Printf("+IWPosterior.InsertLogRatio\n")
	//defer fmt.Printf("-IWPosterior.InsertLogRatio\n")
	insertScatter := this.S.Copy()
	insertScatter.Insert(x)
	psiHatPrime, _ := this.Psi.PlusDense(insertScatter.S)

	p := this.Psi.Rows()

	lr -= 0.5 * float64(p) * math.Log(math.Pi)
	//println(lr)

	n := this.S.Count
	psiHatDet := this.getPsiHat().Det()
	//fmt.Printf("psiHat = \n%v\n", this.getPsiHat())
	//fmt.Printf("psiHatDet = %f\n", psiHatDet)
	lr += math.Log(psiHatDet) * float64(this.M+n) / 2
	//println(lr)
	psiHatPrimeDet := psiHatPrime.Det()
	//fmt.Printf("psiHatPrime = \n%v\n", psiHatPrime)
	//fmt.Printf("psiHatPrimeDet = %f\n", psiHatPrimeDet)
	lr -= math.Log(psiHatPrimeDet) * float64(this.M+n+1) / 2
	//println(lr)

	lr += stat.LnΓpRatio(p, float64(this.M+n+1), float64(this.M+n))
	//println(lr)

	return
}
func (this *IWPosterior) InsertScatterLogRatio(other *Scatter) (lr float64) {
	/*
				Matrix PsiHatPrime = Psi.plus(scatter.getPretendScatter(x));

		        int p = Psi.getColumnDimension();

		        double ll = 0;

		        ll -= 0.5*p*Math.log(Math.PI);

		        ll += Math.log(PsiHat.det())*(m+n)/2;
		        ll -= Math.log(PsiHatPrime.det())*(m+n+1)/2;

		        ll += GammaF.logGammaPRatio(p, m+n+1, m+n);
	*/
	//fmt.Printf("+IWPosterior.InsertScatterLogRatio\n")
	//defer fmt.Printf("-IWPosterior.InsertScatterLogRatio\n")

	insertScatter := this.S.Copy()
	insertScatter.InsertScatter(other)

	//	fmt.Printf("my scatter =\n%v\n", this.S.S)
	//	fmt.Printf("other =\n%v\n", other.S)
	//	fmt.Printf("insert scatter =\n%v\n", insertScatter.S)

	psiHatPrime, _ := this.Psi.PlusDense(insertScatter.S)
	//	fmt.Printf("PsiHat' =\n%v\n", psiHatPrime)
	p := this.Psi.Rows()

	lr -= 0.5 * float64(p) * math.Log(math.Pi)
	//	println(lr)

	n := this.S.Count
	lr += math.Log(this.getPsiHat().Det()) * float64(this.M+n) / 2
	//	println(lr)
	lr -= math.Log(psiHatPrime.Det()) * float64(this.M+n+other.Count) / 2
	//	println(lr)

	lr += stat.LnΓpRatio(p, float64(this.M+n+other.Count), float64(this.M+n))
	//	println(lr)
	return
}

/*
   logp += Math.log(Psi.det())*m*0.5;
   logp -= Math.log(PsiHat.det())*(m+n)*0.5;
   logp += GammaF.logGammaPRatio(p, m+n, m);
   logp -= 0.5*n*p*Math.log(Math.PI);
*/
func (this *IWPosterior) LogP() (ll float64) {
	if this.logP == 0 && this.S.Count != 0 {
		n := this.S.Count
		this.logP += math.Log(this.Psi.Det()) * float64(this.M) * 0.5
		this.logP -= math.Log(this.getPsiHat().Det()) * float64(this.M+n) * 0.5
		this.logP += stat.LnΓpRatio(this.Psi.Rows(), float64(this.M+n), float64(this.M))
		this.logP -= 0.5 * float64(n) * float64(this.Psi.Rows()) * math.Log(math.Pi)
	}

	return this.logP
}
