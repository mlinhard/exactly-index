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

func testSearchIn(t *testing.T, text ...string) *TestSearch {
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

func toSet(array []int) *set.Set {
	r := set.New()
	for i := range array {
		r.Insert(array[i])
	}
	return r
}

func (ts *TestSearch) find(text string) *TestSearchResult {
	return &TestSearchResult{*ts, ts.search.Find([]byte(text))}
}

func (tsr *TestSearchResult) assertPositions(positions ...int) {
	computedPositions := tsr.result.Positions()
	expectedSet := toSet(positions)
	computedSet := toSet(computedPositions)
	if !(expectedSet.SubsetOf(computedSet) && computedSet.SubsetOf(expectedSet)) {
		tsr.t.Errorf("Expected positions for pattern %v: %v, but got %v", tsr.result.Pattern(), positions, computedPositions)
	}
}

func (tsr *TestSearchResult) assertSingleHit() *TestHit {
	if tsr.result.Size() != 1 {
		tsr.t.Errorf("Expected single hit but got %v", tsr.result.Size())
	}
	return &TestHit{*tsr, tsr.result.Hit(0)}
}

func (th *TestHit) assertPosition(pos int) *TestHit {
	if th.hit.Position() != pos {
		th.t.Errorf("Expected position %v but got %v", pos, th.hit.Position())
	}
	return th
}

func (th *TestHit) assertDocument(doc int) *TestHit {
	docIndex := th.hit.Document().Index
	if docIndex != doc {
		th.t.Errorf("Expected document %v but got %v", doc, docIndex)
	}
	return th
}

func (th *TestHit) assertCtx(maxCtx int, leftCtx string, rightCtx string) *TestHit {
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

func (ts *TestSearch) assertSingleHitCtx(pattern string, doc int, pos int, maxCtx int, leftCtx string, rightCtx string) {
	ts.find(pattern).assertSingleHit().assertDocument(doc).assertPosition(pos).assertCtx(maxCtx, leftCtx, rightCtx)
}

func assertEqual(t *testing.T, a []int32, n int32, expectedR int32) {
	r := Search32(a, n)
	if r != expectedR {
		t.Errorf("Search32(%v, %v) is %v not %v", a, n, r, expectedR)
	}
}

func (th *TestHit) assertLinesAbove(linesAbove int, lines string) *TestHit {
	ctx := th.hit.LineContext(linesAbove, 0)
	actualLines := string(ctx.Before())
	if actualLines != lines {
		th.t.Errorf("Expected lines above %v but got %v", lines, actualLines)
	}
	return th
}

func (th *TestHit) assertLinesBelow(linesBelow int, lines string) *TestHit {
	ctx := th.hit.LineContext(0, linesBelow)
	actualLines := string(ctx.After())
	if actualLines != lines {
		th.t.Errorf("Expected lines below %v but got %v", lines, actualLines)
	}
	return th
}

func TestSearch32(t *testing.T) {
	a := []int32{5, 10, 15}
	assertEqual(t, a, 20, 3)
	assertEqual(t, a, 16, 3)
	assertEqual(t, a, 15, 3)
	assertEqual(t, a, 14, 2)
	assertEqual(t, a, 11, 2)
	assertEqual(t, a, 10, 2)
	assertEqual(t, a, 9, 1)
	assertEqual(t, a, 6, 1)
	assertEqual(t, a, 5, 1)
	assertEqual(t, a, 4, 0)
	assertEqual(t, a, 1, 0)
	assertEqual(t, a, 0, 0)
	assertEqual(t, a, -1, 0)
}
func TestSearch32_2(t *testing.T) {
	a := []int32{0, 5, 10}
	assertEqual(t, a, 10, 3)
	assertEqual(t, a, 6, 2)
	assertEqual(t, a, 5, 2)
	assertEqual(t, a, 4, 1)
	assertEqual(t, a, 1, 1)
	assertEqual(t, a, 0, 1)
}

func TestAbracadabra(t *testing.T) {
	search := testSearchIn(t, "abracadabra")
	search.find("abracadabra").assertPositions(0)
	search.find("bracadabra").assertPositions(1)
	search.find("racadabra").assertPositions(2)
	search.find("acadabra").assertPositions(3)
	search.find("cadabra").assertPositions(4)
	search.find("adabra").assertPositions(5)
	search.find("dabra").assertPositions(6)
	search.find("abra").assertPositions(7, 0)
	search.find("bra").assertPositions(8, 1)
	search.find("ra").assertPositions(9, 2)
	search.find("a").assertPositions(10, 7, 0, 3, 5)
	search.find("b").assertPositions(8, 1)
	search.find("c").assertPositions(4)
	search.find("d").assertPositions(6)
	search.find("r").assertPositions(9, 2)
}

func TestAcaaacatat(t *testing.T) {
	search := testSearchIn(t, "acaaacatat")
	search.find("acaaacatat").assertPositions(0)
	search.find("caaacatat").assertPositions(1)
	search.find("aaacatat").assertPositions(2)
	search.find("aacatat").assertPositions(3)
	search.find("acatat").assertPositions(4)
	search.find("catat").assertPositions(5)
	search.find("atat").assertPositions(6)
	search.find("tat").assertPositions(7)
	search.find("at").assertPositions(8, 6)
	search.find("t").assertPositions(9, 7)

	search.find("acaaacatat").assertPositions(0)
	search.find("acaaacata").assertPositions(0)
	search.find("acaaacat").assertPositions(0)
	search.find("acaaaca").assertPositions(0)
	search.find("acaaac").assertPositions(0)
	search.find("acaaa").assertPositions(0)
	search.find("acaa").assertPositions(0)
	search.find("aca").assertPositions(0, 4)
	search.find("ac").assertPositions(0, 4)
	search.find("a").assertPositions(2, 3, 0, 4, 8, 6)

	search.find("caaacatat").assertPositions(1)
	search.find("caaacata").assertPositions(1)
	search.find("caaacat").assertPositions(1)
	search.find("caaaca").assertPositions(1)
	search.find("caaac").assertPositions(1)
	search.find("caaa").assertPositions(1)
	search.find("caa").assertPositions(1)
	search.find("ca").assertPositions(1, 5)
	search.find("c").assertPositions(1, 5)

	search.find("aaacatat").assertPositions(2)
	search.find("aaacata").assertPositions(2)
	search.find("aaacat").assertPositions(2)
	search.find("aaaca").assertPositions(2)
	search.find("aaac").assertPositions(2)
	search.find("aaa").assertPositions(2)
	search.find("aa").assertPositions(2, 3)

	search.find("aacatat").assertPositions(3)
	search.find("aacata").assertPositions(3)
	search.find("aacat").assertPositions(3)
	search.find("aaca").assertPositions(3)
	search.find("aac").assertPositions(3)

	search.find("acatat").assertPositions(4)
	search.find("acata").assertPositions(4)
	search.find("acat").assertPositions(4)

	search.find("catat").assertPositions(5)
	search.find("cata").assertPositions(5)
	search.find("cat").assertPositions(5)

	search.find("atat").assertPositions(6)
	search.find("ata").assertPositions(6)
}

func TestMississippi(t *testing.T) {
	search := testSearchIn(t, "mississippi")
	search.find("mississippi").assertPositions(0)
	search.find("ississippi").assertPositions(1)
	search.find("ssissippi").assertPositions(2)
	search.find("sissippi").assertPositions(3)
	search.find("issippi").assertPositions(4)
	search.find("ssippi").assertPositions(5)
	search.find("sippi").assertPositions(6)
	search.find("ippi").assertPositions(7)
	search.find("ppi").assertPositions(8)
	search.find("pi").assertPositions(9)
	search.find("i").assertPositions(10, 7, 4, 1)

	search.find("mississippi").assertPositions(0)
	search.find("mississipp").assertPositions(0)
	search.find("mississip").assertPositions(0)
	search.find("mississi").assertPositions(0)
	search.find("mississ").assertPositions(0)
	search.find("missis").assertPositions(0)
	search.find("missi").assertPositions(0)
	search.find("miss").assertPositions(0)
	search.find("mis").assertPositions(0)
	search.find("mi").assertPositions(0)
	search.find("m").assertPositions(0)

	search.find("ississippi").assertPositions(1)
	search.find("ississipp").assertPositions(1)
	search.find("ississip").assertPositions(1)
	search.find("ississi").assertPositions(1)
	search.find("ississ").assertPositions(1)
	search.find("issis").assertPositions(1)
	search.find("issi").assertPositions(1, 4)
	search.find("iss").assertPositions(1, 4)
	search.find("is").assertPositions(1, 4)

	search.find("ssissippi").assertPositions(2)
	search.find("ssissipp").assertPositions(2)
	search.find("ssissip").assertPositions(2)
	search.find("ssissi").assertPositions(2)
	search.find("ssiss").assertPositions(2)
	search.find("ssis").assertPositions(2)
	search.find("ssi").assertPositions(2, 5)
	search.find("ss").assertPositions(2, 5)
	search.find("s").assertPositions(2, 3, 5, 6)

	search.find("sissippi").assertPositions(3)
	search.find("sissipp").assertPositions(3)
	search.find("sissip").assertPositions(3)
	search.find("sissi").assertPositions(3)
	search.find("siss").assertPositions(3)
	search.find("sis").assertPositions(3)
	search.find("si").assertPositions(3, 6)

	search.find("issippi").assertPositions(4)
	search.find("issipp").assertPositions(4)
	search.find("issip").assertPositions(4)

	search.find("ssippi").assertPositions(5)
	search.find("ssipp").assertPositions(5)
	search.find("ssip").assertPositions(5)

	search.find("sippi").assertPositions(6)
	search.find("sipp").assertPositions(6)
	search.find("sip").assertPositions(6)
}

func TestJoin(t *testing.T) {
	search := testSearchIn(t, "abcde", "fghij", "klmno", "pqrst")
	search.find("defg").assertPositions()
	search.find("abc").assertSingleHit().assertDocument(0).assertPosition(0)
	search.find("fgh").assertSingleHit().assertDocument(1).assertPosition(0)
	search.find("klm").assertSingleHit().assertDocument(2).assertPosition(0)
	search.find("pqr").assertSingleHit().assertDocument(3).assertPosition(0)

	search.assertSingleHitCtx("bcd", 0, 1, 2, "a", "e")
	search.assertSingleHitCtx("ghi", 1, 1, 1, "f", "j")
	search.assertSingleHitCtx("lmn", 2, 1, 10, "k", "o")
	search.assertSingleHitCtx("qrs", 3, 1, 100, "p", "t")

	search.find("abcde").assertSingleHit().assertDocument(0).assertPosition(0)
	search.find("fghij").assertSingleHit().assertDocument(1).assertPosition(0)
	search.find("klmno").assertSingleHit().assertDocument(2).assertPosition(0)
	search.find("pqrst").assertSingleHit().assertDocument(3).assertPosition(0)

	search.find("abcde").assertPositions(0)
	search.find("fghij").assertPositions(0)
	search.find("klmno").assertPositions(0)
	search.find("pqrst").assertPositions(0)
}

func TestLineContext(t *testing.T) {
	search := testSearchIn(t, "aaa\nbbb\nccc\nddd\neee")
	result := search.find("ccc")
	hit := result.assertSingleHit()
	hit.assertLinesAbove(0, "")
	hit.assertLinesAbove(1, "bbb\n")
	hit.assertLinesAbove(2, "aaa\nbbb\n")
	hit.assertLinesAbove(3, "aaa\nbbb\n")
	hit.assertLinesBelow(0, "")
	hit.assertLinesBelow(1, "\nddd")
	hit.assertLinesBelow(2, "\nddd\neee")
	hit.assertLinesBelow(3, "\nddd\neee")
}

func TestLineContext2(t *testing.T) {
	search := testSearchIn(t, "aaa\nbbb\nccGGcc\nddd\neee")
	result := search.find("GG")
	hit := result.assertSingleHit()
	hit.assertLinesAbove(0, "cc")
	hit.assertLinesAbove(1, "bbb\ncc")
	hit.assertLinesAbove(2, "aaa\nbbb\ncc")
	hit.assertLinesAbove(3, "aaa\nbbb\ncc")
	hit.assertLinesBelow(0, "cc")
	hit.assertLinesBelow(1, "cc\nddd")
	hit.assertLinesBelow(2, "cc\nddd\neee")
	hit.assertLinesBelow(3, "cc\nddd\neee")
}

func TestAAAA(t *testing.T) {
	search := testSearchIn(t, "aaaaaaaaaaaaaaaaaaaa")
	search.find("aaaa").assertPositions(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
}
