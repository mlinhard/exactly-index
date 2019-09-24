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
	// TODO: to implement
	return nil
}

func (ts *TestSearch) assertPositions(pattern string, positions ...int) {
	result := ts.search.Find([]byte(pattern))
	expectedSet := set.New(positions)
	computedSet := set.New(result.Positions())
	if expectedSet != computedSet {
		ts.t.Errorf("Fail")
	}
}

func TestAbracadabra(t *testing.T) {
	search := testSearchIn("abracadabra", t)
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
