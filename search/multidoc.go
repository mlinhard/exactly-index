package search

import (
	"fmt"
	"sort"

	"github.com/mlinhard/exactly-index/esa"
)

type MultiDocumentSearch struct {
	esa                *esa.EnhancedSuffixArray
	offsets            []int32
	ids                []string
	separator          []byte
	newLineInSeparator int32
}

type MultiDocumentSearchResult struct {
	MultiDocumentSearch
	interval      esa.Interval
	docIndexCache []int32
}

func toint32(a []int) []int32 {
	r := make([]int32, len(a))
	for i := range a {
		r[i] = int32(a[i])
	}
	return r
}

func NewMulti(combinedContent []byte, offsets []int, docIds []string) (*MultiDocumentSearch, error) {
	search := new(MultiDocumentSearch)
	search.ids = docIds
	search.offsets = toint32(offsets)
	esa, separator, err := esa.NewMulti(combinedContent, search.offsets)
	if err != nil {
		return nil, err
	}
	search.separator = separator
	search.esa = esa
	search.newLineInSeparator = newLineInSeparator(separator)
	return search, nil
}

func newLineInSeparator(separator []byte) int32 {
	for i := int32(0); i < int32(len(separator)); i++ {
		if isNewLine(separator, i) > 0 {
			return i
		}
	}
	return -1
}

func (this *MultiDocumentSearch) separatorAt(pos int32) bool {
	data := this.esa.Data
	separator := this.separator
	lSeparator := int32(len(separator))
	if pos+lSeparator <= int32(len(data)) && pos >= int32(0) {
		for i := int32(0); i < lSeparator; i++ {
			if separator[i] != data[pos+i] {
				return false
			}
		}
		return true
	} else {
		return false
	}
}

func (this *MultiDocumentSearch) separatorAwareMatch(pattern []byte, dataOff int32, patternOff int32, mlen int32) bool {
	data := this.esa.Data
	for i := int32(0); i < mlen; i++ {
		pIdx := patternOff + i
		dIdx := dataOff + i
		if pIdx >= int32(len(pattern)) || dIdx >= int32(len(data)) || pattern[pIdx] != data[dIdx] || this.separatorAt(dIdx) {
			return false
		}
	}
	return true
}

func (search *MultiDocumentSearch) Find(pattern []byte) SearchResult {
	interval := search.esa.Find(pattern, search.separatorAwareMatch)
	if interval == nil {
		return EmptySearchResult(pattern)
	}
	sr := new(MultiDocumentSearchResult)
	sr.interval = *interval
	sr.MultiDocumentSearch = *search
	sr.docIndexCache = make([]int32, sr.interval.End-sr.interval.Start)
	for i := range sr.docIndexCache {
		sr.docIndexCache[i] = esa.UNDEF
	}
	return sr
}

func (search *MultiDocumentSearch) DocumentCount() int {
	return len(search.ids)
}

func (search *MultiDocumentSearch) Document(idx int) *Document {
	start := search.offsets[idx]
	end := int32(len(search.esa.Data))
	if idx < len(search.offsets)-1 {
		end = search.offsets[idx+1] - int32(len(search.separator))
	}
	r := new(Document)
	r.Content = search.esa.Data[start:end]
	r.Id = search.ids[idx]
	r.Index = idx
	return r
}

func (this *MultiDocumentSearchResult) IsEmpty() bool {
	return false
}

func (this *MultiDocumentSearchResult) Size() int {
	return int(this.interval.End - this.interval.Start)
}

func (this *MultiDocumentSearchResult) globalPosition(hitIdx int) int {
	if hitIdx < 0 || int32(hitIdx) >= int32(this.Size()) {
		panic(fmt.Sprintf("Hit index %v exceeds the search result size %v", hitIdx, this.Size()))
	}
	return int(this.esa.SA[this.interval.Start+int32(hitIdx)])
}

func (this *MultiDocumentSearchResult) position(hitIdx int) int {
	return this.globalPosition(hitIdx) - int(this.offsets[this.documentIndex(hitIdx)])
}

func Search32(a []int32, n int32) int32 {
	return int32(sort.Search(len(a), func(i int) bool { return a[i] > n }))
}

func (this *MultiDocumentSearchResult) document(hitIdx int) *Document {
	return this.Document(this.documentIndex(hitIdx))
}

func (this *MultiDocumentSearchResult) documentIndex(hitIdx int) int {
	if this.docIndexCache[hitIdx] == esa.UNDEF {
		pos := int32(this.globalPosition(hitIdx))
		r := Search32(this.offsets, pos)
		this.docIndexCache[hitIdx] = r - 1
	}
	return int(this.docIndexCache[hitIdx])
}

func (this *MultiDocumentSearchResult) Hit(hitIdx int) Hit {
	return &HitStruct{this, hitIdx}
}

func (this *MultiDocumentSearchResult) PatternLength() int {
	return int(this.interval.Length)
}

func (this *MultiDocumentSearchResult) Pattern() []byte {
	patternStart := this.esa.SA[this.interval.Start]
	return this.esa.Data[patternStart : patternStart+this.interval.Length]
}

func (this *MultiDocumentSearchResult) HasGlobalPosition(position int) bool {
	return false
}

func (this *MultiDocumentSearchResult) HitWithGlobalPosition(position int) Hit {
	return nil
}

func (this *MultiDocumentSearchResult) HasPosition(document, position int) bool {
	return false
}

func (this *MultiDocumentSearchResult) HitWithPosition(document, position int) Hit {
	return nil
}

func (this *MultiDocumentSearchResult) Positions() []int {
	r := make([]int, this.Size())
	for i := range r {
		r[i] = int(this.position(i))
	}
	return r
}

func (this *MultiDocumentSearchResult) charContext(hitIndex int, charsBefore, charsAfter int) HitContext {
	if charsBefore < 0 || charsAfter < 0 {
		panic("Negative context length")
	}
	pos := int32(this.globalPosition(hitIndex))
	beforeStart := this.checkBefore(pos, int32(charsBefore))
	afterEnd := this.checkAfter(pos+this.interval.Length, int32(charsAfter))
	return &HitContextStruct{
		this.esa.Data,
		beforeStart,
		pos - beforeStart,
		this.interval.Length,
		afterEnd - pos - this.interval.Length}
}

func (this *MultiDocumentSearchResult) lineContext(hitIndex int, linesBefore, linesAfter int) HitContext {
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

func (this *MultiDocumentSearchResult) checkBefore(pos int32, maxSize int32) int32 {
	leftLimit := checkBeforeSingle(pos, maxSize)
	sepLen := int32(len(this.separator))
	for i := pos - sepLen; i >= leftLimit; i-- {
		if this.separatorAt(i) {
			return i + sepLen
		}
	}
	return leftLimit
}

func (this *MultiDocumentSearchResult) checkAfter(pos int32, maxSize int32) int32 {
	rightLimit := checkAfterSingle(int32(len(this.esa.Data)), pos, maxSize)
	sepLen := int32(len(this.separator))
	sepRightLimit := rightLimit - sepLen
	for i := pos; i <= sepRightLimit; i++ {
		if this.separatorAt(i) {
			return i
		}
	}
	return rightLimit
}

func (this *MultiDocumentSearchResult) linesBeforeStart(hitIndex int, maxLines int) int32 {
	j := int32(this.globalPosition(hitIndex))
	newLine := int32(0)
	lineCount := int32(0)
	sepLen := int32(len(this.separator))
	sep := this.separatorAt(j)
	for j >= 0 && !sep && lineCount <= int32(maxLines) {
		newLine = isNewLine(this.esa.Data, j)
		if newLine > 0 {
			lineCount++
		}
		j--
		sep = this.separatorAt(j)
	}
	/*
	 * if separator is contained in (newLineInSeparator == -1) or equal
	 * to (newLineInSeparator == 0) newline sequence this means that the
	 * newline sequence never appears in the data. That means that
	 * isNewLine always returns 0, lineCount never increases and
	 * therefore the loop is ended only by the separator. in both cases
	 * we want to return j + 1 + separator.length
	 *
	 * if newLine is contained (but not equal) in the separator
	 * (newLineInSeparator > 0) we want to return
	 *
	 */
	newLineEnd := j + 1 + newLine
	if this.newLineInSeparator == -1 {
		return newLineEnd + ifelse(sep, sepLen-1, 0)
	} else {
		limit := ifelse(j-this.newLineInSeparator < 0, 0, j-this.newLineInSeparator)
		for j >= limit && !sep {
			j--
			sep = this.separatorAt(j)
		}
		return ifelse(sep, j+sepLen, newLineEnd)
	}
}

func (this *MultiDocumentSearchResult) linesAfterStart(hitIndex int, maxLines int) int32 {
	j := int32(this.globalPosition(hitIndex)) + this.interval.Length
	lineCount := int32(0)
	dataLen := int32(len(this.esa.Data))
	sep := this.separatorAt(j)
	for j < dataLen && !sep && lineCount <= int32(maxLines) {
		if isNewLine(this.esa.Data, j) > 0 {
			lineCount++
		}
		j++
		sep = this.separatorAt(j)
	}
	return ifelse(j == dataLen || sep, j, j-1)
}
