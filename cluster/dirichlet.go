package cluster

import (
	"gostat.googlecode.com/hg/stat"
)

func InsertLoglihood(numActions, numOutcomes uint64, alpha []float64, Oc, Os []SAHist) (ll float64) {
	for a := uint64(0); a < numActions; a++ {
		totalDenom := 0.0
		totalNum := 0.0
		for o := uint64(0); o < numOutcomes; o++ {
			coTerm := alpha[o] + float64(Oc[a][o]) + float64(Os[a][o])
			soTerm := float64(Os[a][o] + 1)
			ll += stat.LnΓ(coTerm)
			ll -= stat.LnΓ(soTerm)
			totalDenom += coTerm
			totalNum += soTerm
		}
		ll += stat.LnΓ(totalNum + 1)
		ll -= stat.LnΓ(totalDenom)
	}
	return
}
