package search

import (
	"fmt"
	"testing"

	"github.com/golang-collections/collections/set"
)

type TestSearch struct {
	search Search
	t      *testing.T
}

type TestSearchResult struct {
	TestSearch
	result SearchResult
}

type TestHit struct {
	TestSearchResult
	hit Hit
}

func NewSearchIn(t *testing.T, text ...string) *TestSearch {
	if len(text) == 0 {
		t.Errorf("You have to specify some texts")
		return nil
	}
	if len(text) == 1 {
		search, err := NewSingle("testDoc", []byte(text[0]))
		if err != nil {
			t.Error(err)
			return nil
		}
		return &TestSearch{search, t}
	} else {
		offsets, combinedData := combine(text)
		search, err := NewMulti(combinedData, offsets, testIds(len(text)))
		if err != nil {
			t.Error(err)
			return nil
		}
		return &TestSearch{search, t}
	}
}

func testIds(n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("testDoc%v", i)
	}
	return ids
}

func combine(text []string) ([]int, []byte) {
	totalLength := 0
	offsets := make([]int, len(text))
	for i := range text {
		offsets[i] = totalLength
		totalLength += len([]byte(text[i]))
	}
	r := make([]byte, totalLength)
	for i := range text {
		data := []byte(text[i])
		offset := offsets[i]
		for j := range data {
			r[offset+j] = data[j]
		}
	}
	return offsets, r
}

func toSet(array []int) *set.Set {
	r := set.New()
	for i := range array {
		r.Insert(array[i])
	}
	return r
}

func (ts *TestSearch) Find(text string) *TestSearchResult {
	return &TestSearchResult{*ts, ts.search.Find([]byte(text))}
}

func (tsr *TestSearchResult) AssertHitCount(hitCount int) {
	if tsr.result.Size() != hitCount {
		tsr.t.Errorf("Unexpected hit count %v (expected %v)", tsr.result.Size(), hitCount)
	}
}

func (tsr *TestSearchResult) AssertPositions(positions ...int) {
	computedPositions := tsr.result.Positions()
	expectedSet := toSet(positions)
	computedSet := toSet(computedPositions)
	if !(expectedSet.SubsetOf(computedSet) && computedSet.SubsetOf(expectedSet)) {
		tsr.t.Errorf("Expected positions for pattern %v: %v, but got %v", tsr.result.Pattern(), positions, computedPositions)
	}
}

func (tsr *TestSearchResult) AssertSingleHit() *TestHit {
	if tsr.result.Size() != 1 {
		tsr.t.Errorf("Expected single hit but got %v", tsr.result.Size())
	}
	return &TestHit{*tsr, tsr.result.Hit(0)}
}

func (th *TestHit) AssertPosition(pos int) *TestHit {
	if th.hit.Position() != pos {
		th.t.Errorf("Expected position %v but got %v", pos, th.hit.Position())
	}
	return th
}

func (th *TestHit) AssertDocument(doc int) *TestHit {
	docIndex := th.hit.Document().Index
	if docIndex != doc {
		th.t.Errorf("Expected document %v but got %v", doc, docIndex)
	}
	return th
}

func (th *TestHit) AssertCtx(maxCtx int, leftCtx string, rightCtx string) *TestHit {
	ctx := th.hit.CharContext(maxCtx, maxCtx)
	aLeftCtx := string(ctx.Before())
	aRightCtx := string(ctx.After())
	if leftCtx != aLeftCtx {
		th.t.Errorf("Expected left context %v got %v", leftCtx, aLeftCtx)
	}
	if rightCtx != aRightCtx {
		th.t.Errorf("Expected right context %v got %v", rightCtx, aRightCtx)
	}
	return th
}

func (ts *TestSearch) AssertSingleHitCtx(pattern string, doc int, pos int, maxCtx int, leftCtx string, rightCtx string) {
	ts.Find(pattern).AssertSingleHit().AssertDocument(doc).AssertPosition(pos).AssertCtx(maxCtx, leftCtx, rightCtx)
}

func (th *TestHit) AssertLinesAbove(linesAbove int, lines string) *TestHit {
	ctx := th.hit.LineContext(linesAbove, 0)
	actualLines := string(ctx.Before())
	if actualLines != lines {
		th.t.Errorf("Expected lines above %v but got %v", lines, actualLines)
	}
	return th
}

func (th *TestHit) AssertLinesBelow(linesBelow int, lines string) *TestHit {
	ctx := th.hit.LineContext(0, linesBelow)
	actualLines := string(ctx.After())
	if actualLines != lines {
		th.t.Errorf("Expected lines below %v but got %v", lines, actualLines)
	}
	return th
}
