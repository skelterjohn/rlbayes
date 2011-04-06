package roar

import (
	"gomatrix.googlecode.com/hg/matrix"
)

type Scatter struct {
	S     *matrix.DenseMatrix
	Mean  *matrix.DenseMatrix
	Count int
}

func NewScatter(p int) (this *Scatter) {
	this = new(Scatter)
	this.S = matrix.Zeros(p, p)
	this.Mean = matrix.Zeros(p, 1)
	this.Count = 0
	return
}
func (this *Scatter) Copy() (other *Scatter) {
	other = new(Scatter)
	other.S = this.S.Copy()
	other.Mean = this.Mean.Copy()
	other.Count = this.Count
	return
}
func (this *Scatter) Insert(x *matrix.DenseMatrix) {
	/*
		count++;
		Matrix delta = x.minus(mean);
		if (useSampleMean)
			mean = mean.plus(delta.times(1.0/count));
		scatter = scatter.plus(delta.times(x.minus(mean).transpose()));
	*/
	this.Count++
	delta, _ := x.MinusDense(this.Mean)
	deltaOverC := delta.Copy()
	deltaOverC.Scale(1 / float64(this.Count))
	this.Mean.Add(deltaOverC)
	xMinusMean, _ := x.MinusDense(this.Mean)
	xMinusMeanT := xMinusMean.Transpose()
	deltaTimesXMinusMeanT, _ := delta.TimesDense(xMinusMeanT)
	this.S.Add(deltaTimesXMinusMeanT)
}

func (this *Scatter) Remove(x *matrix.DenseMatrix) {
	/*
	   int new_count = count-1;
	   Matrix new_mu = mean;
	   if (useSampleMean)
	   	new_mu = mean.plus(mean.minus(x).times(1.0/new_count));
	   scatter = scatter.minus((x.minus(new_mu)).times(x.minus(mean).transpose()));
	   count = new_count;
	   mean = new_mu;
	   if (count == 0) {
	   	if (useSampleMean)
	   		mean = new Matrix(p, 1, 0);
	       scatter = new Matrix(p, p, 0);
	   }
	*/
	newCount := this.Count - 1
	if newCount == 0 {
		p := this.S.Rows()
		this.S = matrix.Zeros(p, p)
		this.Mean = matrix.Zeros(p, 1)
		this.Count = newCount
		return
	}

	newMean, _ := this.Mean.MinusDense(x)
	newMean.Scale(1 / float64(newCount))
	newMean.Add(this.Mean)

	xMinusNewMean, _ := x.MinusDense(newMean)
	xMinusMean, _ := x.MinusDense(this.Mean)
	xMinusMeanT := xMinusMean.Transpose()
	xCross, _ := xMinusNewMean.TimesDense(xMinusMeanT)

	this.S.Subtract(xCross)

	this.Count = newCount
	this.Mean = newMean
}

func (this *Scatter) InsertScatterOffset(other *Scatter) {
	this.S.Add(other.S)
	this.Count += other.Count
}
func (this *Scatter) InsertScatter(other *Scatter) {
	/*
		int newCount = count+s.getCount();

		Matrix newMean = mean.times(count).plus(s.getMean().times(s.getCount())).times(1.0/newCount);

		Matrix newScatter = scatter.plus(s.getScatter());
		newScatter = newScatter.plus(mean.times(mean.transpose()).times(count));
		newScatter = newScatter.plus(s.getMean().times(s.getMean().transpose()).times(s.getCount()));
		newScatter = newScatter.minus(newMean.times(newMean.transpose()).times(newCount));

		count = newCount;
		mean = newMean;
		scatter = newScatter;
	*/
	if other.Count == 0 {
		return
	}

	newCount := this.Count + other.Count

	newMean := this.Mean.Copy()
	newMean.Scale(float64(this.Count))
	otherMean := other.Mean.Copy()
	otherMean.Scale(float64(other.Count))
	newMean.Add(otherMean)
	newMean.Scale(1 / float64(newCount))

	this.S.Add(other.S)
	thisCross, _ := this.Mean.Times(this.Mean.Transpose())
	thisCross.Scale(float64(this.Count))
	otherCross, _ := other.Mean.Times(other.Mean.Transpose())
	otherCross.Scale(float64(other.Count))
	newCross, _ := newMean.Times(newMean.Transpose())
	newCross.Scale(float64(newCount))

	this.S.Add(thisCross)
	this.S.Add(otherCross)
	this.S.Subtract(newCross)

	this.Count = newCount
	this.Mean = newMean
}

func (this *Scatter) RemoveScatter(other *Scatter) {
	/*
		int newCount = count-s.getCount();

		Matrix newMean = mean.times(count).minus(s.getMean().times(s.getCount())).times(1.0/newCount);

		Matrix newScatter = scatter.minus(s.getScatter());
		newScatter = newScatter.minus(s.getMean().times(s.getMean().transpose()).times(s.getCount()));
		newScatter = newScatter.minus(newMean.times(newMean.transpose()).times(newCount));
		newScatter = newScatter.plus(mean.times(mean.transpose()).times(count));

		count = newCount;
		mean = newMean;
		scatter = newScatter;
	*/
	if other.Count == 0 {
		return
	}

	newCount := this.Count - other.Count

	if newCount == 0 {
		p := this.S.Rows()
		this.S = matrix.Zeros(p, p)
		this.Mean = matrix.Zeros(p, 1)
		this.Count = newCount
		return
	}

	newMean := this.Mean.Copy()
	newMean.Scale(float64(this.Count))
	otherMean := other.Mean.Copy()
	otherMean.Scale(float64(other.Count))
	newMean.Subtract(otherMean)
	newMean.Scale(1 / float64(newCount))

	this.S.Subtract(other.S)
	thisCross, _ := this.Mean.Times(this.Mean.Transpose())
	thisCross.Scale(float64(this.Count))
	otherCross, _ := other.Mean.Times(other.Mean.Transpose())
	otherCross.Scale(float64(other.Count))
	newCross, _ := newMean.Times(newMean.Transpose())
	newCross.Scale(float64(newCount))

	this.S.Add(thisCross)
	this.S.Subtract(otherCross)
	this.S.Subtract(newCross)

	this.Count = newCount
	this.Mean = newMean
}
