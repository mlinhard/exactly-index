package server

type SearchRequest struct {
	Pattern    []byte `json:"pattern"`
	MaxHits    int    `json:"max_hits"`
	MaxContext int    `json:"max_context"`
	Offset     *int   `json:"offset,omitempty"`
}

type Hit struct {
	Position      int    `json:"pos"`
	DocumentId    string `json:"doc_id"`
	ContextBefore []byte `json:"ctx_before"`
	ContextAfter  []byte `json:"ctx_after"`
}

type Cursor struct {
	CompleteSize int `json:"complete_size"`
	Offset       int `json:"offset"`
}

type SearchResponse struct {
	Hits   []Hit   `json:"hits"`
	Cursor *Cursor `json:"cursor,omitempty"`
}

type SearchServerStats struct {
	IndexedBytes int      `json:"indexed_bytes"`
	IndexedFiles int      `json:"indexed_files"`
	DoneCrawling bool     `json:"done_crawling"`
	DoneLoading  bool     `json:"done_loading"`
	DoneIndexing bool     `json:"done_indexing"`
	Errors       []string `json:"errors"`
}

type DocumentRequest struct {
	DocumentId    *string `json:"document_id,omitempty"`
	DocumentIndex *int    `json:"document_index,omitempty"`
}

type DocumentResponse struct {
	DocumentId    string `json:"document_id"`
	DocumentIndex int    `json:"document_index"`
	Content       []byte `json:"content"`
}
