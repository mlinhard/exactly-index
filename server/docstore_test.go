package server

import "testing"

type TestDocStore struct {
	t        *testing.T
	tempDir  *TempDir
	docStore *DocumentStore
}

func loadTestDocStore(tempDir *TempDir, testFileWalk *TestFileWalk) *TestDocStore {
	tds := new(TestDocStore)
	tds.t = testFileWalk.t
	tds.tempDir = tempDir
	docStore, err := LoadDocuments(testFileWalk.fileWalk)
	if err != nil {
		tds.t.Errorf("Error loading document store %v", err)
		return nil
	}
	tds.docStore = docStore
	return tds
}

func (this *TestDocStore) assertTotalBytes(expectedTotalBytes int) {
	if this.docStore.TotalBytes() != expectedTotalBytes {
		this.t.Errorf("Unexpected document store data size: %v (expected %v)", this.docStore.TotalBytes(), expectedTotalBytes)
	}
}

func (this *TestDocStore) assertSize(expectedSize int) {
	if this.docStore.Size() != expectedSize {
		this.t.Errorf("Unexpected document store size: %v (expected %v)", this.docStore.Size(), expectedSize)
	}
}

func (this *TestDocStore) assertContains(relPath string, docContent string) {
	docId := this.tempDir.Path(relPath)
	doc := this.docStore.DocumentById(docId)
	if doc == nil {
		this.t.Errorf("Store doesn't contain document %v", docId)
		return
	}
	actualContent := string(doc.Content)
	if actualContent != docContent {
		this.t.Errorf("Document %v content is %v (expected %v)", docId, actualContent, docContent)
	}
}
