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

type SingleDocumentSearch struct {
	esa   *esa.EnhancedSuffixArray
	docId string
}

type SingleDocumentSearchResult struct {
	SingleDocumentSearch
	interval esa.Interval
}

type HitContext interface {
	Before() []byte
	Pattern() []byte
	After() []byte
	HighlightStart() int // Length of string returned by Before() method
	HighlightEnd() int   // Length of before string + length of pattern
}

// Represents one occurrence of the pattern in the text composed of one or more documents
type Hit interface {
	GlobalPosition() int                                // global position in concatenated string of all documents including separators (will never return position inside of the separator)
	Position() int                                      // position inside of the document, i.e. number of bytes from the document start.
	Document() *Document                                // The document this hit was found in
	CharContext(charsBefore, charsAfter int) HitContext // Context of the found pattern inside of the document given as number of characters
	LineContext(linesBefore, linesAfter int) HitContext
}

// Result of the search for pattern in the text indexed by Search
type SearchResult interface {
	Size() int // Number of occurrences of the pattern found
	IsEmpty() bool
	Hit(i int) Hit
	PatternLength() int                  // Length of the original pattern that we searched for.
	Pattern() []byte                     // Pattern that we searched for.
	HasGlobalPosition(position int) bool // True iff pattern was found on given position
	HitWithGlobalPosition(position int) Hit
	HasPosition(document, position int) bool
	HitWithPosition(document, position int) Hit
	Positions() []int

	document(hitIndex int) *Document
	globalPosition(hitIndex int) int
	position(hitIndex int) int
	charContext(hitIndex int, charsBefore, charsAfter int) HitContext
	lineContext(hitIndex int, linesBefore, linesAfter int) HitContext
}

type Search interface {
	DocumentCount() int
	Document(i int) *Document
	Find(pattern []byte) SearchResult
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

func (this *SingleDocumentSearchResult) IsEmpty() bool {
	return false
}

func (this *SingleDocumentSearchResult) Size() int {
	return int(this.interval.End - this.interval.Start)
}

func (this *SingleDocumentSearchResult) globalPosition(hitIdx int) int {
	if hitIdx < 0 || int32(hitIdx) >= int32(this.Size()) {
		panic(fmt.Sprintf("Hit index %v exceeds the search result size %v", hitIdx, this.Size()))
	}
	return int(this.esa.SA[this.interval.Start+int32(hitIdx)])
}

func (search *SingleDocumentSearchResult) document(hitIdx int) *Document {
	return search.Document(0)
}

func (this *SingleDocumentSearchResult) position(hitIdx int) int {
	return this.globalPosition(hitIdx)
}

func (this *SingleDocumentSearchResult) Hit(hitIdx int) Hit {
	return &HitStruct{this, hitIdx}
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
	r := make([]int, this.Size())
	for i := range r {
		r[i] = int(this.esa.SA[this.interval.Start+int32(i)])
	}
	return r
}

func ifelse(expr bool, onTrue int32, onFalse int32) int32 {
	if expr {
		return onTrue
	} else {
		return onFalse
	}
}

func checkBeforeSingle(pos int32, maxSize int32) int32 {
	r := pos - maxSize
	return ifelse(r < 0, 0, r)
}

func checkAfterSingle(dataLen int32, pos int32, maxSize int32) int32 {
	r := pos + maxSize
	return ifelse(r > dataLen, dataLen, r)
}

func (this *SingleDocumentSearchResult) charContext(hitIndex int, charsBefore, charsAfter int) HitContext {
	if charsBefore < 0 || charsAfter < 0 {
		panic("Negative context length")
	}
	pos := int32(this.globalPosition(hitIndex))
	beforeStart := checkBeforeSingle(pos, int32(charsBefore))
	afterEnd := checkAfterSingle(int32(len(this.esa.Data)), pos+this.interval.Length, int32(charsAfter))
	return &HitContextStruct{
		this.esa.Data,
		beforeStart,
		pos - beforeStart,
		this.interval.Length,
		afterEnd - pos - this.interval.Length}
}

func isNewLine(data []byte, i int32) int32 {
	ldata := int32(len(data))
	if i >= 0 && i < ldata {
		c0 := data[i]
		if c0 == 13 {
			return ifelse(i == ldata-1 || data[i+1] != 10, 1, 2)
		} else if c0 == 10 {
			return ifelse(i == 0 || data[i-1] != 13, 1, 0)
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func (this *SingleDocumentSearchResult) isNewLine(i int32) int32 {
	return isNewLine(this.esa.Data, i)
}

func (this *SingleDocumentSearchResult) linesBeforeStart(hitIndex int, maxLines int) int32 {
	j := int32(this.globalPosition(hitIndex))
	newLine := int32(0)
	lineCount := int32(0)
	for j >= 0 && lineCount <= int32(maxLines) {
		newLine = this.isNewLine(j)
		if newLine > 0 {
			lineCount++
		}
		j--
	}
	return j + 1 + newLine
}

func (this *SingleDocumentSearchResult) linesAfterStart(hitIndex int, maxLines int) int32 {
	j := int32(this.globalPosition(hitIndex)) + this.interval.Length
	lineCount := int32(0)
	dataLength := int32(len(this.esa.Data))
	for j < dataLength && lineCount <= int32(maxLines) {
		if this.isNewLine(j) > 0 {
			lineCount++
		}
		j++
	}
	return ifelse(j == dataLength, j, j-1)
}

func (this *SingleDocumentSearchResult) lineContext(hitIndex int, linesBefore, linesAfter int) HitContext {
	if linesBefore < 0 || linesAfter < 0 {
		panic("Negative context length")
	}
	patternStart := int32(this.globalPosition(hitIndex))
	beforeStart := this.linesBeforeStart(hitIndex, linesBefore)
	afterEnd := this.linesAfterStart(hitIndex, linesAfter)
	return &HitContextStruct{
		this.esa.Data,
		beforeStart,
		patternStart - beforeStart,
		this.interval.Length,
		afterEnd - patternStart - this.interval.Length}
}
