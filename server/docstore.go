package server

import (
	"fmt"
	"os"
	"sync"

	"github.com/mlinhard/exactly-index/search"
)

type DocumentStore struct {
	documentIds []string
	offsets     []int32
	data        []byte
}

type loaderEntry struct {
	idx    int
	path   string
	buffer []byte
}

func LoadDocuments(fileWalk *FileWalk) (*DocumentStore, error) {
	docStore := new(DocumentStore)
	docStore.data = make([]byte, int(fileWalk.TotalBytes))
	docStore.offsets = computeOffsets(fileWalk.Entries)
	docStore.documentIds = documentIds(fileWalk.Entries)

	entries := make(chan loaderEntry)
	successes := make(chan int, PATH_BUFFER_SIZE)
	errors := make(chan error, ERROR_BUFFER_SIZE)
	wgCounter := new(sync.WaitGroup)
	wgLoaders := new(sync.WaitGroup)
	loaded := make([]bool, fileWalk.Size())

	startLoaders(wgLoaders, entries, successes, errors, 4)

	wgCounter.Add(1)
	go counter(wgCounter, successes, loaded)

	for i, entry := range fileWalk.Entries {
		dataStart := docStore.offsets[i]
		dataEnd := int32(len(docStore.data))
		if i < len(docStore.offsets)-1 {
			dataEnd = docStore.offsets[i+1]
		}
		entries <- loaderEntry{i, entry.Path, docStore.data[dataStart:dataEnd]}
	}
	close(entries)
	wgLoaders.Wait()
	close(errors)
	close(successes)
	wgCounter.Wait()

	var errorbuf []error
	for err := range errors {
		errorbuf = append(errorbuf, err)
	}

	if len(errorbuf) > 0 {
		return nil, fmt.Errorf("Errors during file loading: %v", errorbuf)
	}

	for i, flag := range loaded {
		if !flag {
			return nil, fmt.Errorf("File %v not loaded", docStore.documentIds[i])
		}
	}

	return docStore, nil
}

func computeOffsets(entries []FileEntry) []int32 {
	offsets := make([]int32, len(entries))
	currentOffset := int32(0)
	for i, entry := range entries {
		offsets[i] = currentOffset
		currentOffset += int32(entry.Size)
	}
	return offsets
}

func documentIds(entries []FileEntry) []string {
	docIds := make([]string, len(entries))
	for i, entry := range entries {
		docIds[i] = entry.Path
	}
	return docIds
}

func counter(wg *sync.WaitGroup, successes chan int, loaded []bool) {
	defer wg.Done()
	for successIdx := range successes {
		loaded[successIdx] = true
	}
}

func startLoaders(wg *sync.WaitGroup, entries chan loaderEntry, successes chan int, errors chan error, numWorkers int) {
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go loaderWorker(wg, entries, successes, errors)
	}
}

func loaderWorker(wg *sync.WaitGroup, entries chan loaderEntry, successes chan int, errors chan error) {
	defer wg.Done()
	for entry := range entries {
		err := loadFile(entry.path, entry.buffer)
		if err != nil {
			errors <- err
		} else {
			successes <- entry.idx
		}
	}
}

func loadFile(path string, buffer []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	bytesRead, err := f.Read(buffer)
	if err != nil {
		return err
	}
	if bytesRead != len(buffer) {
		return fmt.Errorf("Unexpected length %v read from file %v, expected %v", bytesRead, path, len(buffer))
	}
	return nil
}

func (this *DocumentStore) Size() int {
	return len(this.documentIds)
}

func (this *DocumentStore) DocumentById(documentId string) *search.Document {
	docIdx := this.IndexOf(documentId)
	if docIdx == -1 {
		return nil
	}
	return this.Document(docIdx)
}

func (this *DocumentStore) Document(idx int) *search.Document {
	if idx >= len(this.offsets) || idx >= len(this.documentIds) {
		panic("Illegal document index")
	}
	doc := new(search.Document)
	start := this.offsets[idx]
	end := int32(len(this.data))
	if idx < len(this.offsets)-1 {
		end = this.offsets[idx+1]
	}
	doc.Content = this.data[start:end]
	doc.Id = this.documentIds[idx]
	doc.Index = idx
	return doc
}

func (this *DocumentStore) Data() []byte {
	return this.data
}

func (this *DocumentStore) Offsets() []int32 {
	return this.offsets
}

func (this *DocumentStore) DocumentId(idx int) string {
	return this.documentIds[idx]
}

func (this *DocumentStore) TotalBytes() int {
	return len(this.data)
}

func (this *DocumentStore) IndexOf(documentId string) int {
	for i, id := range this.documentIds {
		if id == documentId {
			return i
		}
	}
	return -1
}

func (this *DocumentStore) ContainsId(documentId string) bool {
	return this.IndexOf(documentId) != -1
}
