package esa

import (
	"testing"
)

func TestSuffixArray(t *testing.T) {
	bytes := ([]byte)("ABRACADABRA")
	esa, err := New(bytes)
	if err != nil {
		t.Errorf("Error creating enhanced suffix array: %v", err)
		return
	}
	var suffixes [][]byte
	for i := range bytes {
		suffixes = append(suffixes, bytes[i:])
	}
	suffixes = SortBAs(suffixes)
	for i := range suffixes {
		expected := string(suffixes[i])
		computed := string(bytes[esa.SA[i]:])
		if expected != computed {
			t.Errorf("for i=%v SA[i]=%v suffix expected: %v computed %v", i, esa.SA[i], expected, computed)
			return
		}
	}

}
