package roar

import (
	"testing"
	"fmt"
	"gomatrix.googlecode.com/hg/matrix"
)

func TestIW(t *testing.T) {
	psi := matrix.MakeDenseMatrix([]float64{1, 0, 0, 1}, 2, 2)
	iwp := NewIWPosterior(1, psi)

	iwp.Insert(matrix.MakeDenseMatrix([]float64{1, 1}, 2, 1))
	iwp.Insert(matrix.MakeDenseMatrix([]float64{2, 1}, 2, 1))
	lr := iwp.InsertLogRatio(matrix.MakeDenseMatrix([]float64{1, 2}, 2, 1))
	println(lr)
}

func TestScatterNeg(t *testing.T) {
	points := []*matrix.DenseMatrix{
		matrix.MakeDenseMatrix([]float64{-5}, 1, 1),
		matrix.MakeDenseMatrix([]float64{-3}, 1, 1),
		matrix.MakeDenseMatrix([]float64{-4}, 1, 1),
		matrix.MakeDenseMatrix([]float64{-6}, 1, 1),
	}
	s1 := NewScatter(1)
	s2 := NewScatter(1)
	for _, point := range points {
		s1.Insert(point)
	}
	fmt.Printf("%v\n", s1)
	for _, point := range points {
		offset, _ := point.PlusDense(matrix.MakeDenseMatrix([]float64{10}, 1, 1))
		s2.Insert(offset)
		//s1.Insert(offset)
	}
	s1.InsertScatter(s2)
	s1.RemoveScatter(s2)
	fmt.Printf("%v\n", s1)
}

func TestScatterRatio(t *testing.T) {
	points := []*matrix.DenseMatrix{
		matrix.MakeDenseMatrix([]float64{-1, 0}, 2, 1),
		matrix.MakeDenseMatrix([]float64{1, 0}, 2, 1),
		matrix.MakeDenseMatrix([]float64{0, 1}, 2, 1),
		matrix.MakeDenseMatrix([]float64{0, -1}, 2, 1),
	}

	iw1 := NewIWPosterior(1, matrix.Eye(2))
	iw2 := NewIWPosterior(1, matrix.Eye(2))

	scatter1 := NewScatter(2)
	scatter2 := NewScatter(2)

	for _, point := range points {
		scatter1.Insert(point)
		offset, _ := point.PlusDense(matrix.MakeDenseMatrix([]float64{5, 0}, 2, 1))
		scatter2.Insert(offset)
	}

	iw1.InsertScatter(scatter1)
	iw2.InsertScatter(scatter2)

	fmt.Printf("%f\n", iw1.InsertScatterLogRatio(scatter1))
	fmt.Printf("%f\n", iw1.InsertScatterLogRatio(scatter2))

	fmt.Printf("%f\n", iw2.InsertScatterLogRatio(scatter1))
	fmt.Printf("%f\n", iw2.InsertScatterLogRatio(scatter2))
}
