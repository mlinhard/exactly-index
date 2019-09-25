// Enhanced Suffix Array routines
package esa

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/golang-collections/collections/stack"
	"github.com/mlinhard/sais-go/sais"
)

const (
	UNDEF  = int32(-1)
	CUNDEF = int16(-1)
)

type int32stack []int32

func (s int32stack) Peek() int32 {
	return s[len(s)-1]
}

func (s int32stack) Push(v int32) int32stack {
	return append(s, v)
}

func (s int32stack) Pop() (int32stack, int32) {
	l := len(s)
	return s[:l-1], s[l-1]
}

type Interval struct {
	Length int32
	Start  int32
	End    int32
}

type EnhancedSuffixArray struct {
	Data         []byte
	SA           []int32
	LCP          []int32
	Rank         []int32
	Up           []int32
	Down         []int32
	Next         []int32
	rootInterval Interval
}

func (this *Interval) String() string {
	return fmt.Sprintf("%v-[%v, %v]", this.Length, this.Start, this.End)
}

func New(data []byte) (*EnhancedSuffixArray, error) {
	esa := newESA(data)
	err := esa.computeSA()
	if err != nil {
		return nil, err
	}
	esa.computeLCPKeepRank(false)
	esa.computeUpDown()
	esa.computeNext()
	esa.rootInterval = Interval{0, 0, int32(len(esa.SA) - 1)}
	return esa, nil
}

func (esa *EnhancedSuffixArray) findNonExistentChar(parent *Interval, sepLen int32, occurence []bool) int16 {
	for i := range occurence {
		occurence[i] = false
	}
	esa.forEachChild(parent, func(child *Interval) {
		if esa.SA[child.Start]+sepLen < int32(len(esa.Data)) {
			edgeStart := esa.Data[esa.SA[child.Start]+sepLen]
			occurence[edgeStart] = true
		}
	})
	for i := int16(0); i < int16(len(occurence)); i++ {
		if !occurence[i] {
			return i
		}
	}
	return CUNDEF
}

func (esa *EnhancedSuffixArray) buildSeparator(saIdx int32, sepLen int32, tail byte) []byte {
	separator := make([]byte, sepLen+1)
	dataStart := esa.SA[saIdx]
	for i := int32(0); i < sepLen; i++ {
		separator[i] = esa.Data[dataStart+i]
	}
	separator[sepLen] = tail
	return separator
}

type sepLenInterval struct {
	sepLen int32
	Interval
}

func (esa *EnhancedSuffixArray) findSeparator() []byte {
	var intervalStack stack.Stack
	occurenceBuf := make([]bool, 256)
	intervalStack.Push(sepLenInterval{0, esa.rootInterval})

	for intervalStack.Len() != 0 {
		t := intervalStack.Pop().(sepLenInterval)
		nonExistentChar := esa.findNonExistentChar(&t.Interval, t.sepLen, occurenceBuf)
		if nonExistentChar != CUNDEF {
			return esa.buildSeparator(t.Interval.Start, t.sepLen, byte(nonExistentChar))
		} else {
			esa.forEachChild(&t.Interval, func(child *Interval) {
				intervalStack.Push(sepLenInterval{t.sepLen + 1, *child})
			})
		}
	}

	return nil
}

func NewMulti(combinedContent []byte, offsets []int32) (*EnhancedSuffixArray, []byte, error) {
	separatorEsa := newESA(combinedContent)
	err := separatorEsa.computeSA()
	if err != nil {
		return nil, nil, err
	}
	separatorEsa.computeLCPKeepRank(true)
	separatorEsa.computeUpDown()
	separatorEsa.computeNext()
	separatorEsa.rootInterval = Interval{0, 0, int32(len(separatorEsa.SA) - 1)}
	separator := separatorEsa.findSeparator()
	separatorEsa.introduceSeparators(offsets, separator)
	esa, err := New(separatorEsa.Data)
	if err != nil {
		return nil, nil, err
	}
	return esa, separator, nil
}

func newESA(data []byte) *EnhancedSuffixArray {
	esa := new(EnhancedSuffixArray)
	esa.Data = data
	return esa
}

func (esa *EnhancedSuffixArray) computeSA() error {
	n := len(esa.Data)
	esa.SA = make([]int32, n+1)
	esa.SA[n] = UNDEF
	return sais.Sais32(esa.Data, esa.SA[:n])
}

func (esa *EnhancedSuffixArray) Print() string {
	s := " i: SA[i] lcp[i] up[i] down[i] next[i]  suffix[SA[i]]\n"
	for i := range esa.SA {
		suffixStart := esa.SA[i]
		suffix := "$"
		if suffixStart != UNDEF {
			suffix = string(esa.Data[suffixStart:])
		}
		s += fmt.Sprintf("%2v: %4v %6v %5v %7v %7v %v\n", i, suffixStart, esa.LCP[i], esa.Up[i], esa.Down[i], esa.Next[i], suffix)
	}
	return s
}

func (esa *EnhancedSuffixArray) computeLCPKeepRank(keepRank bool) {
	start := (int32)(0)
	length := (int32)(len(esa.Data))
	esa.Rank = make([]int32, length)
	for i := (int32)(0); i < length; i++ {
		esa.Rank[esa.SA[i]] = i
	}
	h := (int32)(0)
	esa.LCP = make([]int32, length+1)
	for i := (int32)(0); i < length; i++ {
		k := esa.Rank[i]
		if k == 0 {
			esa.LCP[k] = -1
		} else {
			j := esa.SA[k-1]
			for i+h < length && j+h < length && esa.Data[start+i+h] == esa.Data[start+j+h] {
				h++
			}
			esa.LCP[k] = h
		}
		if h > 0 {
			h--
		}
	}
	esa.LCP[0] = 0
	esa.LCP[length] = 0
	if !keepRank {
		esa.Rank = nil
	}
}

func (esa *EnhancedSuffixArray) computeUpDown() {
	esa.Up = make([]int32, len(esa.LCP))
	esa.Down = make([]int32, len(esa.LCP))
	for i := range esa.Up {
		esa.Up[i] = UNDEF
		esa.Down[i] = UNDEF
	}
	lastIndex := UNDEF
	var stack int32stack
	stack = stack.Push(0)
	for i := (int32)(1); i < (int32)(len(esa.LCP)); i++ {
		for esa.LCP[i] < esa.LCP[stack.Peek()] {
			stack, lastIndex = stack.Pop()
			if esa.LCP[i] <= esa.LCP[stack.Peek()] && esa.LCP[stack.Peek()] != esa.LCP[lastIndex] {
				esa.Down[stack.Peek()] = lastIndex
			}
		}
		if lastIndex != UNDEF {
			esa.Up[i] = lastIndex
			lastIndex = UNDEF
		}
		stack = stack.Push(i)
	}
}

func (esa *EnhancedSuffixArray) computeNext() {
	esa.Next = make([]int32, len(esa.LCP))
	for i := range esa.Up {
		esa.Next[i] = UNDEF
	}
	var stack int32stack
	var lastIndex int32
	stack = stack.Push(0)
	for i := (int32)(0); i < (int32)(len(esa.LCP)); i++ {
		for esa.LCP[i] < esa.LCP[stack.Peek()] {
			stack, _ = stack.Pop()
		}
		if esa.LCP[i] == esa.LCP[stack.Peek()] {
			stack, lastIndex = stack.Pop()
			esa.Next[lastIndex] = i
		}
		stack = stack.Push(i)
	}
}

func (esa *EnhancedSuffixArray) introduceSeparators(offsets []int32, separator []byte) {
	separatorExtraSpace := (int32)((len(offsets) - 1) * len(separator))
	newData := make([]byte, (int32)(len(esa.Data))+separatorExtraSpace)
	lastIdx := (int32)(len(offsets) - 1)
	for i := (int32)(0); i < lastIdx; i++ {
		oldOffset := offsets[i]
		separatorExtraSpace = i * (int32)(len(separator))
		esa.MoveSegment(oldOffset, offsets[i+1], separatorExtraSpace, newData)
		offsets[i] = oldOffset + separatorExtraSpace
	}
	oldOffset := offsets[lastIdx]
	separatorExtraSpace = lastIdx * (int32)(len(separator))
	esa.MoveSegment(oldOffset, (int32)(len(esa.Data)), separatorExtraSpace, newData)
	offsets[lastIdx] = oldOffset + separatorExtraSpace

	for i := (int32)(0); i < (int32)(len(separator)); i++ {
		sepChar := separator[i]
		for j := (int32)(1); j < (int32)(len(offsets)); j++ {
			newData[offsets[j]-(int32)(len(separator))+i] = sepChar
		}
	}

	esa.Data = newData
}

func (esa *EnhancedSuffixArray) MoveSegment(start, end, separatorExtraSpace int32, newData []byte) {
	for i := start; i < end; i++ {
		newData[i+separatorExtraSpace] = esa.Data[i]
	}
	for j := start; j < end; j++ {
		esa.SA[esa.Rank[j]] += separatorExtraSpace
	}
}

func (esa *EnhancedSuffixArray) interval(i, j int32) *Interval {
	cup := esa.Up[j]
	if cup < j && i < cup {
		return &Interval{esa.LCP[cup], i, j}
	}
	return &Interval{esa.LCP[esa.Down[i]], i, j}
}

func (esa *EnhancedSuffixArray) createInterval(parent *Interval, childStart, childEnd int32) *Interval {
	if childEnd == UNDEF {
		childEnd = parent.End
	}
	if childStart+1 < childEnd {
		return esa.interval(childStart, childEnd)
	} else if childStart != childEnd {
		return &Interval{parent.Length, childStart, childEnd}
	} else {
		return nil
	}
}

func (esa *EnhancedSuffixArray) firstIndex(parent *Interval) int32 {
	if *parent == esa.rootInterval {
		return 0
	}
	cup := esa.Up[parent.End]
	if cup < parent.End && parent.Start < cup {
		return cup
	} else {
		return esa.Down[parent.Start]
	}
}

func (esa *EnhancedSuffixArray) edgeChar(parent *Interval, child *Interval) int16 {
	pos := esa.SA[child.Start] + parent.Length
	if pos >= int32(len(esa.Data)) {
		return -1
	}
	return int16(esa.Data[pos])
}

type intervalIterator struct {
	esa        *EnhancedSuffixArray
	parent     *Interval
	start, end int32
	_next      *Interval
}

func (iter *intervalIterator) hasNext() bool {
	return iter._next != nil
}

func (iter *intervalIterator) next() *Interval {
	r := iter._next
	if iter.end != UNDEF {
		iter.start = iter.end
		iter.end = iter.esa.Next[iter.start]
		iter._next = iter.esa.createInterval(iter.parent, iter.start, iter.end)
	} else {
		iter._next = nil
	}
	return r
}

func (esa *EnhancedSuffixArray) firstLIndex(parent *Interval) int32 {
	if *parent == esa.rootInterval {
		return 0
	} else {
		cup := esa.Up[parent.End]
		if cup < parent.End && parent.Start < cup {
			return cup
		} else {
			return esa.Down[parent.Start]
		}
	}
}

func (esa *EnhancedSuffixArray) getChildren(parent *Interval) *intervalIterator {
	iter := new(intervalIterator)
	iter.esa = esa
	iter.parent = parent
	iter.start = parent.Start
	iter.end = esa.firstLIndex(parent)
	if iter.end == iter.start {
		iter.end = esa.Next[iter.start]
	}
	iter._next = esa.createInterval(parent, iter.start, iter.end)
	return iter
}

func (esa *EnhancedSuffixArray) getInterval(parent *Interval, c int16) *Interval {
	iter := esa.getChildren(parent)
	for iter.hasNext() {
		child := iter.next()
		if c == esa.edgeChar(parent, child) {
			return child
		}
	}
	return nil
}

func (esa *EnhancedSuffixArray) acceptInterval(parent *Interval, childStart, childEnd int32, consumer func(*Interval)) {
	if childEnd == UNDEF {
		childEnd = parent.End
	}
	if childStart+1 < childEnd {
		consumer(esa.interval(childStart, childEnd))
	} else if childStart != childEnd {
		consumer(&Interval{parent.Length, childStart, childEnd})
	}
}

func (esa *EnhancedSuffixArray) forEachChild(parent *Interval, consumer func(*Interval)) {
	i := parent.Start
	nexti := esa.firstLIndex(parent)
	if nexti == i {
		nexti = esa.Next[i]
	}
	esa.acceptInterval(parent, i, nexti, consumer)
	for nexti != UNDEF {
		i = nexti
		nexti = esa.Next[i]
		esa.acceptInterval(parent, i, nexti, consumer)
	}
}

func (esa *EnhancedSuffixArray) Match(pattern []byte, dataOff int32, patternOff int32, mlen int32) bool {
	for i := int32(0); i < mlen; i++ {
		pIdx := patternOff + i
		dIdx := dataOff + i
		if pIdx >= int32(len(pattern)) || dIdx >= int32(len(esa.Data)) || pattern[pIdx] != esa.Data[dIdx] {
			return false
		}
	}
	return true
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func (esa *EnhancedSuffixArray) Find(pattern []byte, match func([]byte, int32, int32, int32) bool) *Interval {
	plen := int32(len(pattern))
	if pattern == nil || plen == 0 {
		panic("You must specify non-empty pattern")
	}
	c := int32(0)
	queryFound := true
	intv := esa.getInterval(&esa.rootInterval, int16(pattern[c]))
	intvLen := int32(0)
	for intv != nil && c < plen && queryFound {
		intvLen = intv.End - intv.Start
		if intvLen > 1 {
			min := min32(intv.Length, plen)
			queryFound = match(pattern, esa.SA[intv.Start]+c, c, min-c)
			c = min
			if c < plen {
				intv = esa.getInterval(intv, int16(pattern[c]))
			}
		} else {
			queryFound = match(pattern, esa.SA[intv.Start]+c, c, plen-c)
			break
		}
	}
	if intv != nil && queryFound {
		return &Interval{plen, intv.Start, intv.End}
	}
	return nil
}

type sortableBA [][]byte

func (b sortableBA) Len() int {
	return len(b)
}

func (b sortableBA) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i], b[j]) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		log.Panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
		return false
	}
}

func (b sortableBA) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// Public
func SortBAs(src [][]byte) [][]byte {
	sorted := sortableBA(src)
	sort.Sort(sorted)
	return sorted
}
