package server

import (
	"fmt"

	"github.com/mlinhard/exactly-index/search"
)

type Indexer struct {
	config       *ServerConfig
	doneCrawling bool
	doneLoading  bool
	doneIndexing bool
	errors       []string
	fileWalk     *FileWalk
	docStore     *DocumentStore
	search       search.Search
}

func NewIndexer(config *ServerConfig) *Indexer {
	indexer := new(Indexer)
	indexer.config = config
	return indexer
}

func (this *Indexer) start(listeners *listeners) {
	this.fileWalk = NewFileWalk(this.config.Roots, this.config.IgnoredDirs)
	if this.fileWalk.HasErrors() {
		this.errorState(this.fileWalk.Errors...)
		listeners.Error(fmt.Errorf("filewalk: %v", this.fileWalk.Errors))
		return
	}
	this.doneCrawling = true
	listeners.DoneCrawling(this.fileWalk.Size(), int(this.fileWalk.TotalBytes))
	docStore, err := LoadDocuments(this.fileWalk)
	if err != nil {
		this.errorState(err)
		listeners.Error(err)
		return
	}
	this.docStore = docStore
	this.doneLoading = true
	listeners.DoneLoading(this.fileWalk.Size(), int(this.fileWalk.TotalBytes))
	var search search.Search
	search, err = createSearch(this.docStore)
	if err != nil {
		this.errorState(err)
		listeners.Error(err)
		return
	}
	this.search = search
	this.doneIndexing = true
	listeners.DoneIndexing(this.fileWalk.Size(), int(this.fileWalk.TotalBytes))
}

func createSearch(docStore *DocumentStore) (search.Search, error) {
	if docStore.Size() < 1 {
		return nil, fmt.Errorf("No documents to index")
	}
	if docStore.Size() == 1 {
		return search.NewSingle(docStore.DocumentId(0), docStore.Data())
	}
	return search.NewMulti32(docStore.Data(), docStore.offsets, docStore.documentIds)
}

func (this *Indexer) errorState(errors ...error) {
	this.errors = make([]string, len(errors))
	for i, err := range errors {
		this.errors[i] = fmt.Sprintf("%v", err)
	}
}

func (this *Indexer) Search(request *SearchRequest) *SearchResponse {
	response := new(SearchResponse)
	result := this.search.Find(request.Pattern)
	start := 0
	if request.Offset != nil {
		start = *request.Offset
	}
	end := result.Size()
	userLimit := start + request.MaxHits
	if userLimit < end {
		end = userLimit
	}
	response.Hits = make([]Hit, end-start)
	for i := start; i < end; i++ {
		hit := result.Hit(i)
		hitContext := hit.CharContext(request.MaxContext, request.MaxContext)
		var respHit Hit
		respHit.Position = hit.Position()
		respHit.DocumentId = hit.Document().Id
		respHit.ContextBefore = hitContext.Before()
		respHit.ContextAfter = hitContext.After()
		response.Hits[i] = respHit
	}
	return response
}

func (this *Indexer) Document(request *DocumentRequest) (*DocumentResponse, error) {
	if request.DocumentId == nil && request.DocumentIndex == nil {
		return nil, fmt.Errorf("You have to specify document index or document id")
	}
	resp := new(DocumentResponse)
	var doc *search.Document
	if request.DocumentIndex == nil {
		doc = this.docStore.DocumentById(*request.DocumentId)
		if doc == nil {
			return nil, nil
		}
	} else {
		docIdx := *request.DocumentIndex
		if docIdx < 0 || docIdx >= this.docStore.Size() {
			return nil, fmt.Errorf("Document index %v out of range", docIdx)
		}
		doc = this.docStore.Document(docIdx)
	}
	resp.Content = doc.Content
	resp.DocumentId = doc.Id
	resp.DocumentIndex = doc.Index
	return resp, nil
}

func (this *Indexer) getStats() *SearchServerStats {
	stats := new(SearchServerStats)
	stats.DoneCrawling = this.doneCrawling
	stats.DoneLoading = this.doneLoading
	stats.DoneIndexing = this.doneIndexing
	stats.Errors = this.errors
	if this.doneLoading {
		stats.IndexedBytes = this.docStore.TotalBytes()
		stats.IndexedFiles = this.docStore.Size()
	} else {
		stats.IndexedFiles = 0
		stats.IndexedBytes = 0
	}
	return stats
}
