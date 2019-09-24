package search

import (
	"testing"

	"github.com/golang-collections/collections/set"
)

type TestSearch struct {
	search Search
	t      *testing.T
}

func testSearchIn(text string, t *testing.T) *TestSearch {
	search, err := NewSingle("testDoc", []byte(text))
	if err != nil {
		t.Error(err)
		return nil
	}
	return &TestSearch{search, t}
}

func toSet(array []int) *set.Set {
	r := set.New()
	for i := range array {
		r.Insert(array[i])
	}
	return r
}

func (ts *TestSearch) assertPositions(pattern string, positions ...int) {
	result := ts.search.Find([]byte(pattern))
	if result == nil {
		return
	}
	expectedSet := toSet(positions)
	computedSet := toSet(result.Positions())
	if !(expectedSet.SubsetOf(computedSet) && computedSet.SubsetOf(expectedSet)) {
		ts.t.Errorf("Expected positions for pattern %v: %v, but got %v", pattern, positions, result.Positions())
	}
}

func TestAbracadabra(t *testing.T) {
	search := testSearchIn("abracadabra", t)
	if search == nil {
		return
	}
	search.assertPositions("abracadabra", 0)
	search.assertPositions("bracadabra", 1)
	search.assertPositions("racadabra", 2)
	search.assertPositions("acadabra", 3)
	search.assertPositions("cadabra", 4)
	search.assertPositions("adabra", 5)
	search.assertPositions("dabra", 6)
	search.assertPositions("abra", 7, 0)
	search.assertPositions("bra", 8, 1)
	search.assertPositions("ra", 9, 2)
	search.assertPositions("a", 10, 7, 0, 3, 5)
	search.assertPositions("b", 8, 1)
	search.assertPositions("c", 4)
	search.assertPositions("d", 6)
	search.assertPositions("r", 9, 2)
}
