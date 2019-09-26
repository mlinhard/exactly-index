package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	config     *ServerConfig
	httpServer *http.Server
	indexer    *Indexer
	listeners  *listeners
}

func NewServer(config *ServerConfig) *Server {
	server := new(Server)
	server.config = config
	http.Handle("/", http.NotFoundHandler())
	http.HandleFunc("/search", server.handleSearch)
	http.HandleFunc("/stats", server.handleStats)
	http.HandleFunc("/document", server.handleDocument)
	server.httpServer = &http.Server{
		Addr:    config.ListenAddress,
		Handler: nil,
	}
	server.indexer = NewIndexer(config)
	server.listeners = new(listeners)
	return server
}

func (this *Server) OnDoneCrawling(listener func(int, int)) {
	this.listeners.OnDoneCrawling(listener)
}

func (this *Server) OnDoneLoading(listener func(int, int)) {
	this.listeners.OnDoneLoading(listener)
}

func (this *Server) OnDoneIndexing(listener func(int, int)) {
	this.listeners.OnDoneIndexing(listener)
}

func (this *Server) OnError(listener func(error)) {
	this.listeners.OnError(listener)
}

func (this *Server) Start() {
	go this.indexer.start(this.listeners)
	go this.listenAndServe()
}

func (this *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()
	err := this.httpServer.Shutdown(ctx)
	if err != nil {
		fmt.Printf("HTTP Server stop: %v", err)
	}
}

func (this *Server) listenAndServe() {
	err := this.httpServer.ListenAndServe()
	if err != nil {
		fmt.Printf("HTTP Server listen: %v", err)
	}
}

func (this *Server) handleDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		docRequest := new(DocumentRequest)
		err := json.NewDecoder(r.Body).Decode(docRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Decoding request: %v", err), http.StatusBadRequest)
		} else {
			response, err, status := this.handleDocumentRequest(docRequest)
			if err != nil {
				http.Error(w, fmt.Sprintf("Processing request: %v", err), status)
			}
			json.NewEncoder(w).Encode(response)
		}
	} else {
		http.Error(w, fmt.Sprintf("Method %v not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func (this *Server) handleDocumentRequest(request *DocumentRequest) (*DocumentResponse, error, int) {
	if this.indexer == nil || !this.indexer.getStats().DoneIndexing {
		return nil, fmt.Errorf("Indexer not ready"), http.StatusBadRequest
	}
	doc, err := this.indexer.Document(request)
	if err != nil {
		return nil, err, http.StatusBadRequest
	}
	return doc, nil, http.StatusOK
}

func (this *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		searchRequest := new(SearchRequest)
		err := json.NewDecoder(r.Body).Decode(searchRequest)
		// bytes, err := ioutil.ReadAll(r.Body)
		// fmt.Printf("RECEIVED SEARCH: %v\n", string(bytes))
		// json.Unmarshal(bytes, searchRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Decoding request: %v", err), http.StatusBadRequest)
		} else {
			response, err, status := this.handleSearchRequest(searchRequest)
			if err != nil {
				http.Error(w, fmt.Sprintf("Processing request: %v", err), status)
			}
			json.NewEncoder(w).Encode(response)
		}
	} else {
		http.Error(w, fmt.Sprintf("Method %v not supported", r.Method), http.StatusMethodNotAllowed)
	}
}

func (this *Server) handleSearchRequest(request *SearchRequest) (*SearchResponse, error, int) {
	if this.indexer == nil || !this.indexer.getStats().DoneIndexing {
		return nil, fmt.Errorf("Indexer not ready"), http.StatusBadRequest
	}
	if len(request.Pattern) == 0 {
		return nil, fmt.Errorf("You have to specify non-empty pattern"), http.StatusBadRequest
	}
	return this.indexer.Search(request), nil, http.StatusOK
}

func (this *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(this.getStats())
}

func (this *Server) getStats() *SearchServerStats {
	return this.indexer.getStats()
}
