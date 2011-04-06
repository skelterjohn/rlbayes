package dl

import (
	"testing"
)

func TestParents(t *testing.T) {
	var ps ParentSet
	ps = ps.Insert(4)
	if !ps.Contains(4) {
		t.Error("Doesn't contain 4")
	}
	c := ps.Iter()
	if 4 != <-c {
		t.Error("bad iter")
	}
	ps = ps.Remove(4)
	if ps.Contains(4) {
		t.Error("Contains 4")
	}
}
