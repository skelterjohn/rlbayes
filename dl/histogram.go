package dl

import (
	//"fmt"
	"gostat.googlecode.com/hg/stat"
)

type Histogram []int

func (this Histogram) Sum() (total uint64) {
	for _, c := range this {
		total += uint64(c)
	}
	return
}

func (this Histogram) Update(o uint64, c int) (next Histogram) {
	next = append([]int{}, this...)
	next[o] = c
	return
}
func (this Histogram) UpdateHistogram(h Histogram) (next Histogram) {
	next = append([]int{}, this...)
	for i, o := range this {
		next[i] = o + h[i]
	}
	return
}

func (this Histogram) Next() (o uint64) {
	return this.GetChoice()()
}

func (this Histogram) GetChoice() (chooser func() uint64) {
	var sum float64
	for _, v := range this {
		sum += float64(v)
	}
	norm := 1 / sum
	weights := make([]float64, len(this))
	for i, v := range this {
		weights[i] = float64(v) * norm
	}
	chooser = func() uint64 { chooser := stat.Choice(weights); return uint64(chooser()) }
	return
}

func (this Histogram) Incr(o uint64) (next Histogram) {
	next = append([]int{}, this...)
	next[o]++
	return
}

func (this Histogram) LessThan(oi interface{}) bool {
	return this.Compare(oi.(Histogram)) < 0
}

func (this Histogram) Compare(other Histogram) int {
	for i, v := range this {
		c := v - other[i]
		if c == 0 {
			continue
		}
		return c
	}
	return 0
}

func (this Histogram) LogFactorCount() (lf float64) {
	var sum float64
	for _, c := range this {
		sum += float64(c)
		lf -= stat.LnΓ(float64(c + 1))
	}
	lf += stat.LnΓ(sum + 1)
	//fmt.Printf("%v %f\n", this, lf)
	return
}

func (this Histogram) LogFactorAlpha(alpha float64) (lf float64) {
	var sum float64
	for _, c := range this {
		sum += float64(c)
		lf -= stat.LnΓ(float64(c) + alpha)
	}
	lf += stat.LnΓ(sum + alpha*float64(len(this)))
	return
}

func (this Histogram) LoglihoodRatio(alpha float64) (ll float64) {
	//factorials * 1/gammas
	total := 0.0
	for _, c := range this {
		cf := float64(c)
		total += cf
		//ll -= stat.LnΓ(cf + 1)
		ll -= stat.LnΓ(alpha)
		ll += stat.LnΓ(cf + alpha)
	}
	//ll += stat.LnΓ(total + 1)
	ll += stat.LnΓ(alpha * float64(len(this)))
	ll -= stat.LnΓ(total + alpha*float64(len(this)))
	return
}
