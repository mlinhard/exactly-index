// Empty search result logic
package search

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

func (this EmptySearchResult) document(hitIdx int) *Document {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) position(hitIdx int) int {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) globalPosition(hitIdx int) int {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) charContext(hitIndex int, charsBefore, charsAfter int) HitContext {
	panic("Empty search result has no hits")
}

func (this EmptySearchResult) lineContext(hitIndex int, linesBefore, linesAfter int) HitContext {
	panic("Empty search result has no hits")
}
