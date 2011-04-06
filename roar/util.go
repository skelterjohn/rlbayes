package roar

import (
	"math"
	"gostat.googlecode.com/hg/stat"
)

var NegInf = math.Log(0)

func LogChoice(lws []float64) int {
	max := lws[0]
	for _, lw := range lws[1:len(lws)] {
		if lw > max {
			max = lw
		}
	}
	ws := make([]float64, len(lws))
	var sum float64
	for i, lw := range lws {
		ws[i] = math.Exp(lw - max)
		sum += ws[i]
	}
	norm := 1 / sum
	for i := range ws {
		ws[i] *= norm
	}
	return int(stat.NextChoice(ws))
}

func CRPPrior(alpha float64, hl *HList) (plls []float64) {
	plls = make([]float64, len(hl.h)+1)

	foundNew := false
	for i := range plls {
		var count float64
		if i != len(hl.h) {
			count = float64(hl.Count(i))
		}
		if count == 0 && !foundNew {

			count += alpha
			foundNew = true
		}
		plls[i] = math.Log(float64(count))
	}

	return
}

type HList struct {
	a []int //assignments
	h []int //histogram
}

func (this *HList) Hashcode() (hash uint64) {
	return 100
}

func (this *HList) LessThan(other interface{}) bool {
	o := other.(*HList)
	for i := range this.a {
		if this.a[i] < o.a[i] {
			return true
		}
		if this.a[i] > o.a[i] {
			return false
		}
	}
	return false
}

func (this *HList) Copy() (next *HList) {
	next = new(HList)
	next.a = append([]int{}, this.a...)
	next.h = append([]int{}, this.h...)
	return
}
func (this *HList) Set(index, value int) {
	for index >= len(this.a) {
		this.a = append(this.a, -1)
	}
	for value >= len(this.h) {
		this.h = append(this.h, 0)
	}
	if this.a[index] != -1 {
		this.Drop(index)
	}
	this.a[index] = value
	this.h[value]++
}
func (this *HList) Get(index int) int {
	if index < 0 || index >= len(this.a) {
		return -1
	}
	return this.a[index]
}
func (this *HList) Values() (ch <-chan int) {
	ich := make(chan int, len(this.h))
	ch = ich
	for v, c := range this.h {
		if c != 0 {
			ich <- v
		}
	}
	close(ich)

	return
}
func (this *HList) Count(value int) int {
	if value >= len(this.h) || value < 0 {
		return 0
	}
	return this.h[value]
}
func (this *HList) Drop(index int) {
	if this.a[index] == -1 {
		return
	}
	v := this.a[index]
	this.h[v]--
	this.a[index] = -1
}

func Anneal(lls []float64, temp float64) {
	if temp == 0 {
		var maxProb float64 = NegInf
		var maxIndex int
		for i, p := range lls {
			if p > maxProb {
				maxIndex = i
				maxProb = p
			}
		}
		for i, _ := range lls {
			if i != maxIndex {
				lls[i] = NegInf
			}
		}
	} else {
		for i := range lls {
			lls[i] /= temp
		}
	}
}

func IsNaN(x float64) bool {
	return x != x
}
