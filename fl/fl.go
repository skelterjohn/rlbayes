package fl

import (
	"gostat.googlecode.com/hg/stat"
)

type OutcomeHist []int

func (this OutcomeHist) Update(o uint64, c int) (next OutcomeHist) {
	next = append([]int{}, this...)
	next[o] = c
	return
}

func (this OutcomeHist) Incr(o uint64) (next OutcomeHist) {
	next = append([]int{}, this...)
	next[o]++
	return
}

func (this OutcomeHist) LessThan(oi interface{}) bool {
	return this.Compare(oi.(OutcomeHist)) < 0
}

func (this OutcomeHist) Compare(other OutcomeHist) int {
	for i, v := range this {
		c := v - other[i]
		if c == 0 {
			continue
		}
		return c
	}
	return 0
}

func (this OutcomeHist) LogFactorCount() (lf float64) {
	var sum float64
	for _, c := range this {
		sum += float64(c)
		lf -= stat.LnΓ(float64(c + 1))
	}
	lf += stat.LnΓ(sum + 1)
	return
}

func (this OutcomeHist) LogFactorAlpha(alpha float64) (lf float64) {
	var sum float64
	for _, c := range this {
		sum += float64(c)
		lf += stat.LnΓ(float64(c) + alpha)
	}
	lf += stat.LnΓ(sum + alpha*float64(len(this)))
	return
}

func (this OutcomeHist) LoglihoodRatio(alpha float64) (ll float64) {
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
