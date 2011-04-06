package fl

import (
	"go-glue.googlecode.com/hg/rlglue"
)

type DBN struct {
	connections	[]bool
	task		*rlglue.TaskSpec
	ranges		[]rlglue.IntRanges
	numFactors	int
	numActions	int
	numStates	int
	stateValues	[][]int32
	hash		uint64
}

func NewDBN(task *rlglue.TaskSpec) (this *DBN) {
	this = new(DBN)
	this.numFactors = len(task.Obs.Ints)
	this.numActions = int(task.Act.Ints.Count())
	this.numStates = int(task.Obs.Ints.Count())
	this.task = task
	this.ranges = make([]rlglue.IntRanges, this.numFactors)
	this.connections = make([]bool, this.numFactors*this.numFactors)
	this.stateValues = make([][]int32, this.task.Obs.Ints.Count())
	for s := range this.stateValues {
		this.stateValues[s] = this.task.Obs.Ints.Values(uint64(s))
	}
	return
}
func (this *DBN) String() (res string) {
	for _, c := range this.connections {
		if c {
			res += "1"
		} else {
			res += "0"
		}
	}
	return
}
func (this *DBN) Hashcode() (hash uint64) {
	return this.hash
}
func (this *DBN) HashMask(child int) (hash uint64) {
	var mask uint64
	for parent := 0; parent < this.numFactors; parent++ {
		mask += 1 << uint64(this.numFactors*child+parent)
	}
	return this.hash | mask
}
func (this *DBN) LessThan(oi interface{}) bool {
	other := oi.(*DBN)
	return this.hash < other.hash
}
func (this *DBN) Compare(other *DBN) int {
	return int(this.hash - other.hash)
}
func (this *DBN) UpdateInPlace(child, parent int, connected bool) {
	k := this.numFactors*child + parent
	this.connections = append([]bool{}, this.connections...)
	this.connections[k] = connected
	if connected {
		this.hash += 1 << uint64(k)
	} else {
		this.hash -= 1 << uint64(k)
	}
	this.ranges[child] = make(rlglue.IntRanges, 0, this.numFactors)
	for pi := 0; pi < this.numFactors; pi++ {
		if this.Connection(pi, child) {
			this.ranges[child] = append(this.ranges[child], this.task.Obs.Ints[pi])
		}
	}
}
func (this *DBN) Update(child, parent int, connected bool) (next *DBN) {
	k := this.numFactors*child + parent
	if this.connections[k] == connected {
		return this
	}
	next = new(DBN)
	*next = *this
	next.ranges = append([]rlglue.IntRanges{}, this.ranges...)
	next.connections = append([]bool{}, this.connections...)
	next.UpdateInPlace(child, parent, connected)
	return
}
func (this *DBN) Connection(parent, child int) bool {
	k := this.numFactors*child + parent
	return this.connections[k]
}
func (this *DBN) Filter(parents []int32, child int) (fparents []int32) {
	fparents = make([]int32, 0, len(parents))
	for parent, parentValue := range parents {
		if this.Connection(parent, child) {
			fparents = append(fparents, parentValue)
		}
	}
	return
}
func (this *DBN) Index(parents []int32, child int) (index uint64) {
	fparents := this.Filter(parents, child)
	index = this.ranges[child].Index(fparents)
	return
}
func (this *DBN) Indices(parents []int32) (indices []uint64) {
	indices = make([]uint64, this.numFactors)
	for child := range indices {
		indices[child] = this.Index(parents, child)
	}
	return
}
func (this *DBN) Count(child int) (count uint64) {
	return this.ranges[child].Count()
}
func (this *DBN) MapDownChild(history []OutcomeHist, child, action int) (chist []OutcomeHist) {
	chist = make([]OutcomeHist, this.ranges[child].Count())
	for i := range chist {
		chist[i] = make(OutcomeHist, this.task.Obs.Ints[child].Count())
	}
	numActions := int(this.task.Act.Ints.Count())
	for s, parents := range this.stateValues {
		hist := history[s*numActions+action]
		parentIndex := this.Index(parents, child)
		for n, count := range hist {
			if count == 0 {
				continue
			}
			children := this.stateValues[n]
			childValue := children[child]
			childOutcome := childValue - this.task.Obs.Ints[child].Min
			chist[parentIndex][childOutcome] += count
		}
	}
	return
}
