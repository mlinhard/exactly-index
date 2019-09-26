package search

import (
	"testing"
)

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
	search := NewSearchIn(t, "abracadabra")
	search.Find("abracadabra").AssertPositions(0)
	search.Find("bracadabra").AssertPositions(1)
	search.Find("racadabra").AssertPositions(2)
	search.Find("acadabra").AssertPositions(3)
	search.Find("cadabra").AssertPositions(4)
	search.Find("adabra").AssertPositions(5)
	search.Find("dabra").AssertPositions(6)
	search.Find("abra").AssertPositions(7, 0)
	search.Find("bra").AssertPositions(8, 1)
	search.Find("ra").AssertPositions(9, 2)
	search.Find("a").AssertPositions(10, 7, 0, 3, 5)
	search.Find("b").AssertPositions(8, 1)
	search.Find("c").AssertPositions(4)
	search.Find("d").AssertPositions(6)
	search.Find("r").AssertPositions(9, 2)
}

func TestAcaaacatat(t *testing.T) {
	search := NewSearchIn(t, "acaaacatat")
	search.Find("acaaacatat").AssertPositions(0)
	search.Find("caaacatat").AssertPositions(1)
	search.Find("aaacatat").AssertPositions(2)
	search.Find("aacatat").AssertPositions(3)
	search.Find("acatat").AssertPositions(4)
	search.Find("catat").AssertPositions(5)
	search.Find("atat").AssertPositions(6)
	search.Find("tat").AssertPositions(7)
	search.Find("at").AssertPositions(8, 6)
	search.Find("t").AssertPositions(9, 7)

	search.Find("acaaacatat").AssertPositions(0)
	search.Find("acaaacata").AssertPositions(0)
	search.Find("acaaacat").AssertPositions(0)
	search.Find("acaaaca").AssertPositions(0)
	search.Find("acaaac").AssertPositions(0)
	search.Find("acaaa").AssertPositions(0)
	search.Find("acaa").AssertPositions(0)
	search.Find("aca").AssertPositions(0, 4)
	search.Find("ac").AssertPositions(0, 4)
	search.Find("a").AssertPositions(2, 3, 0, 4, 8, 6)

	search.Find("caaacatat").AssertPositions(1)
	search.Find("caaacata").AssertPositions(1)
	search.Find("caaacat").AssertPositions(1)
	search.Find("caaaca").AssertPositions(1)
	search.Find("caaac").AssertPositions(1)
	search.Find("caaa").AssertPositions(1)
	search.Find("caa").AssertPositions(1)
	search.Find("ca").AssertPositions(1, 5)
	search.Find("c").AssertPositions(1, 5)

	search.Find("aaacatat").AssertPositions(2)
	search.Find("aaacata").AssertPositions(2)
	search.Find("aaacat").AssertPositions(2)
	search.Find("aaaca").AssertPositions(2)
	search.Find("aaac").AssertPositions(2)
	search.Find("aaa").AssertPositions(2)
	search.Find("aa").AssertPositions(2, 3)

	search.Find("aacatat").AssertPositions(3)
	search.Find("aacata").AssertPositions(3)
	search.Find("aacat").AssertPositions(3)
	search.Find("aaca").AssertPositions(3)
	search.Find("aac").AssertPositions(3)

	search.Find("acatat").AssertPositions(4)
	search.Find("acata").AssertPositions(4)
	search.Find("acat").AssertPositions(4)

	search.Find("catat").AssertPositions(5)
	search.Find("cata").AssertPositions(5)
	search.Find("cat").AssertPositions(5)

	search.Find("atat").AssertPositions(6)
	search.Find("ata").AssertPositions(6)
}

func TestMississippi(t *testing.T) {
	search := NewSearchIn(t, "mississippi")
	search.Find("mississippi").AssertPositions(0)
	search.Find("ississippi").AssertPositions(1)
	search.Find("ssissippi").AssertPositions(2)
	search.Find("sissippi").AssertPositions(3)
	search.Find("issippi").AssertPositions(4)
	search.Find("ssippi").AssertPositions(5)
	search.Find("sippi").AssertPositions(6)
	search.Find("ippi").AssertPositions(7)
	search.Find("ppi").AssertPositions(8)
	search.Find("pi").AssertPositions(9)
	search.Find("i").AssertPositions(10, 7, 4, 1)

	search.Find("mississippi").AssertPositions(0)
	search.Find("mississipp").AssertPositions(0)
	search.Find("mississip").AssertPositions(0)
	search.Find("mississi").AssertPositions(0)
	search.Find("mississ").AssertPositions(0)
	search.Find("missis").AssertPositions(0)
	search.Find("missi").AssertPositions(0)
	search.Find("miss").AssertPositions(0)
	search.Find("mis").AssertPositions(0)
	search.Find("mi").AssertPositions(0)
	search.Find("m").AssertPositions(0)

	search.Find("ississippi").AssertPositions(1)
	search.Find("ississipp").AssertPositions(1)
	search.Find("ississip").AssertPositions(1)
	search.Find("ississi").AssertPositions(1)
	search.Find("ississ").AssertPositions(1)
	search.Find("issis").AssertPositions(1)
	search.Find("issi").AssertPositions(1, 4)
	search.Find("iss").AssertPositions(1, 4)
	search.Find("is").AssertPositions(1, 4)

	search.Find("ssissippi").AssertPositions(2)
	search.Find("ssissipp").AssertPositions(2)
	search.Find("ssissip").AssertPositions(2)
	search.Find("ssissi").AssertPositions(2)
	search.Find("ssiss").AssertPositions(2)
	search.Find("ssis").AssertPositions(2)
	search.Find("ssi").AssertPositions(2, 5)
	search.Find("ss").AssertPositions(2, 5)
	search.Find("s").AssertPositions(2, 3, 5, 6)

	search.Find("sissippi").AssertPositions(3)
	search.Find("sissipp").AssertPositions(3)
	search.Find("sissip").AssertPositions(3)
	search.Find("sissi").AssertPositions(3)
	search.Find("siss").AssertPositions(3)
	search.Find("sis").AssertPositions(3)
	search.Find("si").AssertPositions(3, 6)

	search.Find("issippi").AssertPositions(4)
	search.Find("issipp").AssertPositions(4)
	search.Find("issip").AssertPositions(4)

	search.Find("ssippi").AssertPositions(5)
	search.Find("ssipp").AssertPositions(5)
	search.Find("ssip").AssertPositions(5)

	search.Find("sippi").AssertPositions(6)
	search.Find("sipp").AssertPositions(6)
	search.Find("sip").AssertPositions(6)
}

func TestJoin(t *testing.T) {
	search := NewSearchIn(t, "abcde", "fghij", "klmno", "pqrst")
	search.Find("defg").AssertPositions()
	search.Find("abc").AssertSingleHit().AssertDocument(0).AssertPosition(0)
	search.Find("fgh").AssertSingleHit().AssertDocument(1).AssertPosition(0)
	search.Find("klm").AssertSingleHit().AssertDocument(2).AssertPosition(0)
	search.Find("pqr").AssertSingleHit().AssertDocument(3).AssertPosition(0)

	search.AssertSingleHitCtx("bcd", 0, 1, 2, "a", "e")
	search.AssertSingleHitCtx("ghi", 1, 1, 1, "f", "j")
	search.AssertSingleHitCtx("lmn", 2, 1, 10, "k", "o")
	search.AssertSingleHitCtx("qrs", 3, 1, 100, "p", "t")

	search.Find("abcde").AssertSingleHit().AssertDocument(0).AssertPosition(0)
	search.Find("fghij").AssertSingleHit().AssertDocument(1).AssertPosition(0)
	search.Find("klmno").AssertSingleHit().AssertDocument(2).AssertPosition(0)
	search.Find("pqrst").AssertSingleHit().AssertDocument(3).AssertPosition(0)

	search.Find("abcde").AssertPositions(0)
	search.Find("fghij").AssertPositions(0)
	search.Find("klmno").AssertPositions(0)
	search.Find("pqrst").AssertPositions(0)
}

func TestLineContext(t *testing.T) {
	search := NewSearchIn(t, "aaa\nbbb\nccc\nddd\neee")
	result := search.Find("ccc")
	hit := result.AssertSingleHit()
	hit.AssertLinesAbove(0, "")
	hit.AssertLinesAbove(1, "bbb\n")
	hit.AssertLinesAbove(2, "aaa\nbbb\n")
	hit.AssertLinesAbove(3, "aaa\nbbb\n")
	hit.AssertLinesBelow(0, "")
	hit.AssertLinesBelow(1, "\nddd")
	hit.AssertLinesBelow(2, "\nddd\neee")
	hit.AssertLinesBelow(3, "\nddd\neee")
}

func TestLineContext2(t *testing.T) {
	search := NewSearchIn(t, "aaa\nbbb\nccGGcc\nddd\neee")
	result := search.Find("GG")
	hit := result.AssertSingleHit()
	hit.AssertLinesAbove(0, "cc")
	hit.AssertLinesAbove(1, "bbb\ncc")
	hit.AssertLinesAbove(2, "aaa\nbbb\ncc")
	hit.AssertLinesAbove(3, "aaa\nbbb\ncc")
	hit.AssertLinesBelow(0, "cc")
	hit.AssertLinesBelow(1, "cc\nddd")
	hit.AssertLinesBelow(2, "cc\nddd\neee")
	hit.AssertLinesBelow(3, "cc\nddd\neee")
}

func TestAAAA(t *testing.T) {
	search := NewSearchIn(t, "aaaaaaaaaaaaaaaaaaaa")
	search.Find("aaaa").AssertPositions(0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16)
}

func assertEqual(t *testing.T, a []int32, n int32, expectedR int32) {
	r := Search32(a, n)
	if r != expectedR {
		t.Errorf("Search32(%v, %v) is %v not %v", a, n, r, expectedR)
	}
}
