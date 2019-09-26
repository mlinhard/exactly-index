package server

import "testing"

type TestFileWalk struct {
	t        *testing.T
	fileWalk *FileWalk
}

func (this *TestFileWalk) assertOK() {
	if this.fileWalk.HasErrors() {
		this.t.Errorf("File walk has errors %v", this.fileWalk.Errors)
	}
}

func (this *TestFileWalk) assertContains(path string) {
	if this.fileWalk.Contains(path) {
		this.t.Errorf("File walk should contain path %v", path)
	}
}

func (this *TestFileWalk) assertSize(expectedSize int) {
	if this.fileWalk.Size() != expectedSize {
		this.t.Errorf("unexpected walk size: %v", this.fileWalk.Size())
	}
}
