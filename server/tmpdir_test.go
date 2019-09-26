package server

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type TempDir struct {
	t   *testing.T
	dir string
	OK  bool
}

func NewTempDir(t *testing.T, prefix string) *TempDir {
	td := new(TempDir)
	td.t = t
	dir, err := ioutil.TempDir(os.TempDir(), "exactly-index-test")
	if err != nil {
		t.Error(err)
		td.OK = false
	} else {
		td.dir = dir
		td.OK = true
	}
	return td
}
func (this *TempDir) WriteFile(relpath string, content string) {
	dir := filepath.Dir(this.Path(relpath))
	err := os.MkdirAll(dir, 0770)
	if err != nil {
		this.t.Errorf("Creating dir: %v", err)
		this.OK = false
		return
	}
	err = ioutil.WriteFile(this.Path(relpath), []byte(content), 0664)
	if err != nil {
		this.t.Errorf("Writing file: %v", err)
		this.OK = false
	}
}

func (this *TempDir) Remove() {
	err := os.RemoveAll(this.dir)
	if err != nil {
		this.t.Error(err)
	}
}

func (this *TempDir) Paths(relpaths ...string) []string {
	r := make([]string, len(relpaths))
	for i, relpath := range relpaths {
		r[i] = this.Path(relpath)
	}
	return r
}

func (this *TempDir) Path(relpath string) string {
	parts := filepath.SplitList(relpath)
	path := append([]string{string(this.dir)}, parts...)
	return filepath.Join(path...)
}
