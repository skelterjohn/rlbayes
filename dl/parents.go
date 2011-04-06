package dl

import (
	"go-glue.googlecode.com/hg/rlglue"
)

type ParentSet uint32

func ParentIter(max uint32) (i <-chan ParentSet) {
	ic := make(chan ParentSet)
	go func(ic chan<- ParentSet) {
		for p := ParentSet(0); p < (1 << max); p++ {
			ic <- p
		}
		close(ic)
	}(ic)
	i = ic
	return
}
func (this ParentSet) Contains(parent uint32) bool {
	return (this & (1 << parent)) != 0
}
func (this ParentSet) Insert(parents ...uint32) (next ParentSet) {
	next = this
	for _, parent := range parents {
		next = next | (1 << uint32(parent))
	}
	return
}
func (this ParentSet) Remove(parents ...uint32) (next ParentSet) {
	next = this
	for _, parent := range parents {
		next = next & (^(1 << parent))
	}
	return
}
func (this ParentSet) Toggle(parents ...uint32) (next ParentSet) {
	next = this
	for _, parent := range parents {
		next = next ^ (1 << parent)
	}
	return
}
func (this ParentSet) Size(max uint32) (size int) {
	for p := uint32(0); p < max; p++ {
		if this.Contains(p) {
			size++
		}
	}
	return
}
func (this ParentSet) Slice() (parents []uint32) {
	for p := uint32(0); p < 32; p++ {
		if this.Contains(p) {
			parents = append(parents, p)
		}
	}
	return
}
func (this ParentSet) Iter() (r <-chan uint32) {
	rc := make(chan uint32)
	r = rc
	go func(c chan<- uint32) {
		for p := uint32(0); p < 32; p++ {
			if this.Contains(p) {
				c <- p
			}
		}
		close(c)
	}(rc)
	return
}
func (this ParentSet) CutRanges(full rlglue.IntRanges) (cut rlglue.IntRanges) {
	for p := range this.Iter() {
		cut = append(cut, full[p])
	}
	return
}
func (this ParentSet) CutValues(full []int32) (cut []int32) {
	for p := range this.Iter() {
		cut = append(cut, full[p])
	}
	return
}
