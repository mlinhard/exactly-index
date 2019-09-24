package search

import "github.com/mlinhard/exactly-index/esa"

type Document interface {
	Index() int
	Id() string
	Content() []byte
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
	Document(i int) Document
	Find(pattern []byte) SearchResult
}

type interval struct {
	len   int32
	start int32
	end   int32
}

type SingleDocumentSearch struct {
	esa.EnhancedSuffixArray
}

func NewSingle(docId string, docContent []byte) (*SingleDocumentSearch, error) {
	search := new(SingleDocumentSearch)
	esa, err := esa.New(docContent)
	if err != nil {
		return nil, err
	}
	search.SA = esa.SA
	return search, nil
}
