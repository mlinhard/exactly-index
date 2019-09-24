package search

import (
	"fmt"

	"github.com/mlinhard/exactly-index/esa"
)

type Document struct {
	Index   int
	Id      string
	Content []byte
}

type HitContext interface {
	Before() []byte

	Pattern() []byte

	After() []byte

	/**
	 * @return Length of string returned by {@link #before()} method
	 */
	HighlightStart() int

	/**
	 *
	 * @return Length of before string + length of pattern
	 */
	HighlightEnd() int
}

/**
 * Represents one occurrence of the pattern in the text composed of one or more
 * documents
 */
type Hit interface {

	/**
	 * @return global position in concatenated string of all documents including
	 *         separators (will never return position inside of the separator)
	 */
	GlobalPosition() int

	/**
	 * @return position inside of the document, i.e. number of bytes from the
	 *         document start.
	 */
	Position() int

	/**
	 * @return The document this hit was found in
	 */
	Document() Document

	/**
	 * Context of the found pattern inside of the document given as number of
	 * characters.
	 *
	 * @param charsBefore
	 *            Number of characters / bytes to get. If the position -
	 *            charsBefore is before document start will return characters
	 *            from the beginning of the document
	 * @param charsAfter
	 * @return
	 */
	CharContext(charsBefore, charsAfter int) HitContext

	SafeCharContext(charsBefore, charsAfter int) HitContext

	LineContext(linesBefore, linesAfter int) HitContext
}

/**
 * Result of the search for pattern in the text indexed by Search
 */
type SearchResult interface {
	/**
	 * @return Number of occurrences of the pattern found
	 */
	Size() int

	IsEmpty() bool

	/**
	 * @param i
	 * @return i-th hit (occurence of pattern)
	 */
	Hit(i int) Hit

	Hits() []Hit

	/**
	 *
	 * @return Length of the original pattern that we searched for.
	 */
	PatternLength() int

	/**
	 *
	 * @return Pattern that we searched for.
	 */
	Pattern() []byte

	/**
	 * @param position
	 * @return True iff pattern was found on given position
	 */
	HasGlobalPosition(position int) bool

	HitWithGlobalPosition(position int) Hit

	HasPosition(document, position int) bool

	HitWithPosition(document, position int) Hit

	Positions() []int
}

type Search interface {
	DocumentCount() int
	Document(i int) *Document
	Find(pattern []byte) SearchResult
}

type SingleDocumentSearch struct {
	esa   *esa.EnhancedSuffixArray
	docId string
}

type MultiDocumentSearch struct {
	esa.EnhancedSuffixArray
	offsets            []int32
	ids                []string
	separator          []byte
	newLineInSeparator int
}

func NewSingle(docId string, docContent []byte) (*SingleDocumentSearch, error) {
	search := new(SingleDocumentSearch)
	search.docId = docId
	esa, err := esa.New(docContent)
	if err != nil {
		return nil, err
	}
	search.esa = esa
	return search, nil
}

func (*SingleDocumentSearch) DocumentCount() int {
	return 1
}

func (search *SingleDocumentSearch) Document(idx int) *Document {
	if idx != 0 {
		panic("Single document search contains only index 0")
	}
	r := new(Document)
	r.Content = search.esa.Data
	r.Id = search.docId
	r.Index = 0
	return r
}

func (search *SingleDocumentSearch) Find(pattern []byte) SearchResult {
	interval := search.esa.Find(pattern, search.esa.Match)
	if interval == nil {
		return EmptySearchResult(pattern)
	}
	sr := new(SingleDocumentSearchResult)
	sr.interval = *interval
	sr.SingleDocumentSearch = *search
	return sr
}

type EmptySearchResult []byte

func (this EmptySearchResult) IsEmpty() bool {
	return true
}

func (this EmptySearchResult) Size() int {
	return 0
}

func (this EmptySearchResult) Hit(i int) Hit {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) Hits() []Hit {
	return make([]Hit, 0)
}

func (this EmptySearchResult) PatternLength() int {
	return len(this)
}

func (this EmptySearchResult) Pattern() []byte {
	return this
}

func (this EmptySearchResult) HasGlobalPosition(position int) bool {
	return false
}

func (this EmptySearchResult) HitWithGlobalPosition(position int) Hit {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) HasPosition(document, position int) bool {
	return false
}

func (this EmptySearchResult) HitWithPosition(document, position int) Hit {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) Positions() []int {
	return make([]int, 0)
}

type SingleDocumentSearchResult struct {
	SingleDocumentSearch
	interval esa.Interval
}

func (this *SingleDocumentSearchResult) IsEmpty() bool {
	return false
}

func (this *SingleDocumentSearchResult) Size() int {
	return int(this.interval.Length)
}

func (this *SingleDocumentSearchResult) globalPosition(hitIdx int) int {
	if hitIdx < 0 || int32(hitIdx) >= this.interval.Length {
		panic(fmt.Sprintf("Hit index %v exceeds the search result size %v", hitIdx, this.Size()))
	}
	return int(this.esa.SA[this.interval.Start+int32(hitIdx)])
}

func (this *SingleDocumentSearchResult) position(hitIdx int) int {
	return this.globalPosition(hitIdx)
}

func (this *SingleDocumentSearchResult) Hit(i int) Hit {
	return nil
}

func (this *SingleDocumentSearchResult) Hits() []Hit {
	return nil
}

func (this *SingleDocumentSearchResult) PatternLength() int {
	return int(this.interval.Length)
}

func (this *SingleDocumentSearchResult) Pattern() []byte {
	patternStart := this.esa.SA[this.interval.Start]
	return this.esa.Data[patternStart : patternStart+this.interval.Length]
}

func (this *SingleDocumentSearchResult) HasGlobalPosition(position int) bool {
	return false
}

func (this *SingleDocumentSearchResult) HitWithGlobalPosition(position int) Hit {
	return nil
}

func (this *SingleDocumentSearchResult) HasPosition(document, position int) bool {
	return false
}

func (this *SingleDocumentSearchResult) HitWithPosition(document, position int) Hit {
	return nil
}

func (this *SingleDocumentSearchResult) Positions() []int {
	r := make([]int, this.interval.End-this.interval.Start)
	for i := range r {
		r[i] = int(this.esa.SA[this.interval.Start+int32(i)])
	}
	return r
}

func (search *MultiDocumentSearch) DocumentCount() int {
	return len(search.ids)
}

func (search *MultiDocumentSearch) Document(idx int) *Document {
	start := search.offsets[idx]
	end := len(search.Data)
	if idx < len(search.offsets)-1 {
		end = int(search.offsets[idx+1]) - len(search.separator)
	}
	r := new(Document)
	r.Content = search.Data[start:end]
	r.Id = search.ids[idx]
	r.Index = idx
	return r
}
