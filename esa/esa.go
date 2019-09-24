// Enhanced Suffix Array routines
package esa

import (
	"bytes"
	"fmt"
	"log"
	"sort"

	"github.com/mlinhard/sais-go/sais"
)

const (
	UNDEF = (int32)(-1)
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

type EnhancedSuffixArray struct {
	data []byte
	SA   []int32
	lcp  []int32
	rank []int32
	up   []int32
	down []int32
	next []int32
}

func New(data []byte) (*EnhancedSuffixArray, error) {
	esa := new(EnhancedSuffixArray)
	esa.data = data
	esa.SA = make([]int32, len(data))
	err := sais.Sais32(data, esa.SA)
	if err != nil {
		return nil, err
	}
	return esa, nil
}

func (esa EnhancedSuffixArray) Print() string {
	s := " i: SA[i] suffix\n"
	for i := range esa.SA {
		s += fmt.Sprintf("%2v: %4v %v\n", i, esa.SA[i], string(esa.data[esa.SA[i]:]))
	}
	return s
}

func (esa EnhancedSuffixArray) ComputeLCP() {
	esa.ComputeLCPKeepRank(false)
}

func (esa EnhancedSuffixArray) ComputeLCPKeepRank(keepRank bool) {
	start := (int32)(0)
	length := (int32)(len(esa.data))
	esa.rank = make([]int32, length)
	for i := (int32)(0); i < length; i++ {
		esa.rank[i] = i
	}
	h := (int32)(0)
	esa.lcp = make([]int32, length+1)
	for i := (int32)(0); i < length; i++ {
		k := esa.rank[i]
		if k == 0 {
			esa.lcp[k] = -1
		} else {
			j := esa.SA[k-1]
			for i+h < length && j+h < length && esa.data[start+i+h] == esa.data[start+j+h] {
				h++
			}
			esa.lcp[k] = h
		}
		if h > 0 {
			h--
		}
	}
	esa.lcp[0] = 0
	esa.lcp[length] = 0
	if !keepRank {
		esa.rank = nil
	}
}

func (esa EnhancedSuffixArray) ComputeUpDown() {
	esa.up = make([]int32, len(esa.lcp))
	esa.down = make([]int32, len(esa.lcp))
	for i := range esa.up {
		esa.up[i] = UNDEF
		esa.down[i] = UNDEF
	}
	lastIndex := UNDEF
	var stack int32stack
	stack = stack.Push(0)
	for i := (int32)(0); i < (int32)(len(esa.lcp)); i++ {
		for esa.lcp[i] < esa.lcp[stack.Peek()] {
			stack, lastIndex = stack.Pop()
			if esa.lcp[i] <= esa.lcp[stack.Peek()] && esa.lcp[stack.Peek()] != esa.lcp[lastIndex] {
				esa.down[stack.Peek()] = lastIndex
			}
		}
		if lastIndex != UNDEF {
			esa.up[i] = lastIndex
			lastIndex = UNDEF
		}
		stack = stack.Push(i)
	}
}

func (esa EnhancedSuffixArray) ComputeNext() {
	esa.next = make([]int32, len(esa.lcp))
	for i := range esa.up {
		esa.next[i] = UNDEF
	}
	var stack int32stack
	var lastIndex int32
	stack = stack.Push(0)
	for i := (int32)(0); i < (int32)(len(esa.lcp)); i++ {
		for esa.lcp[i] < esa.lcp[stack.Peek()] {
			stack, _ = stack.Pop()
		}
		if esa.lcp[i] == esa.lcp[stack.Peek()] {
			stack, lastIndex = stack.Pop()
			esa.next[lastIndex] = i
		}
		stack = stack.Push(i)
	}
}

func (esa EnhancedSuffixArray) IntroduceSeparators(offsets []int32, separator []byte) {
	separatorExtraSpace := (int32)((len(offsets) - 1) * len(separator))
	newData := make([]byte, (int32)(len(esa.data))+separatorExtraSpace)
	lastIdx := (int32)(len(offsets) - 1)
	for i := (int32)(0); i < lastIdx; i++ {
		oldOffset := offsets[i]
		separatorExtraSpace = i * (int32)(len(separator))
		esa.MoveSegment(oldOffset, offsets[i+1], separatorExtraSpace, newData)
		offsets[i] = oldOffset + separatorExtraSpace
	}
	oldOffset := offsets[lastIdx]
	separatorExtraSpace = lastIdx * (int32)(len(separator))
	esa.MoveSegment(oldOffset, (int32)(len(esa.data)), separatorExtraSpace, newData)
	offsets[lastIdx] = oldOffset + separatorExtraSpace

	for i := (int32)(0); i < (int32)(len(separator)); i++ {
		sepChar := separator[i]
		for j := (int32)(1); j < (int32)(len(offsets)); j++ {
			newData[offsets[j]-(int32)(len(separator))+i] = sepChar
		}
	}

	esa.data = newData
}

func (esa EnhancedSuffixArray) MoveSegment(start, end, separatorExtraSpace int32, newData []byte) {
	for i := start; i < end; i++ {
		newData[i+separatorExtraSpace] = esa.data[i]
	}
	for j := start; j < end; j++ {
		esa.SA[esa.rank[j]] += separatorExtraSpace
	}
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
