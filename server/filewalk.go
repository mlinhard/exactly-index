package server

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/karrick/godirwalk"
)

const (
	PATH_BUFFER_SIZE  = 1024 * 1024
	ERROR_BUFFER_SIZE = 1024 * 10
)

type FileEntry struct {
	Path string
	Size int64
}

type FileWalk struct {
	Roots      []string
	Ignore     []string
	TotalBytes int64
	Entries    []FileEntry
	Errors     []error
}

func NewFileWalk(roots []string, ignore []string) *FileWalk {
	fileWalk := new(FileWalk)
	fileWalk.TotalBytes = 0
	fileWalk.Roots = roots
	fileWalk.Ignore = ignore
	paths := make(chan string, PATH_BUFFER_SIZE)
	errors := make(chan error, ERROR_BUFFER_SIZE)
	entries := make(chan FileEntry, PATH_BUFFER_SIZE)
	wgCollectors := new(sync.WaitGroup)
	wgCounter := new(sync.WaitGroup)
	go fileWalk.entryCollector(wgCollectors, entries)
	go fileWalk.errorCollector(wgCollectors, errors)
	go countSizes(wgCounter, paths, entries, errors, 8)
	walk(roots, ignore, paths, errors)
	close(paths)
	wgCounter.Wait()
	close(entries)
	close(errors)
	wgCollectors.Wait()
	return fileWalk
}

func (this *FileWalk) entryCollector(wg *sync.WaitGroup, entries chan FileEntry) {
	wg.Add(1)
	defer wg.Done()
	for entry := range entries {
		this.Entries = append(this.Entries, entry)
		this.TotalBytes += entry.Size
	}
}

func (this *FileWalk) errorCollector(wg *sync.WaitGroup, errors chan error) {
	wg.Add(1)
	defer wg.Done()
	for err := range errors {
		this.Errors = append(this.Errors, err)
	}
}

func countSizes(wg *sync.WaitGroup, paths chan string, entries chan FileEntry, errors chan error, numWorkers int) {
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go countSizeWorker(wg, paths, errors, entries, i)
	}
}

func countSizeWorker(wg *sync.WaitGroup, paths chan string, errors chan error, entries chan FileEntry, i int) {
	defer wg.Done()
	for path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			errors <- err
		} else {
			entries <- FileEntry{path, info.Size()}
		}
	}
}

func walk(roots []string, ignore []string, paths chan string, errors chan error) {
	var wg sync.WaitGroup
	wg.Add(len(roots))
	ignorer := createIgnorer(ignore...)
	for _, root := range roots {
		go walkOne(&wg, root, ignorer, paths, errors)
	}
	wg.Wait()
}

type ignorer struct {
	abs []string
	rel []string
}

func createIgnorer(ignore ...string) *ignorer {
	ignorer := new(ignorer)
	ignorer.rel, ignorer.abs = relAbsSplit(ignore)
	return ignorer
}

func relAbsSplit(ignore []string) ([]string, []string) {
	var rel, abs []string
	for _, dir := range ignore {
		if filepath.IsAbs(dir) {
			abs = append(abs, dir)
		} else {
			rel = append(rel, dir)
		}
	}
	return rel, abs
}

func startsWith(basepath, path string) bool {
	if len(basepath) > len(path) {
		return false
	}
	if basepath != path[:len(basepath)] {
		return false
	}
	list := append(filepath.SplitList(basepath), filepath.SplitList(path[len(basepath):])...)
	if path != filepath.Join(list...) {
		return false
	}
	return true
}

func (this *ignorer) match(path string) bool {
	for _, ignorePath := range this.abs {
		if startsWith(ignorePath, path) {
			return true
		}
	}
	base := filepath.Base(path)
	for _, ignorePath := range this.rel {
		if ignorePath == base {
			return true
		}
	}
	return false
}

func walkOne(wg *sync.WaitGroup, root string, ignorer *ignorer, paths chan string, errors chan error) {
	defer wg.Done()
	err := godirwalk.Walk(root, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsRegular() {
				paths <- osPathname
			} else if de.IsDir() && ignorer.match(osPathname) {
				return filepath.SkipDir
			}
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			errors <- err
			return godirwalk.SkipNode
		},
	})
	if err != nil {
		errors <- err
	}
}

func (this *FileWalk) HasErrors() bool {
	return len(this.Errors) > 0
}

func (this *FileWalk) Size() int {
	return len(this.Entries)
}

func (this *FileWalk) IndexOf(path string) int {
	for i, entry := range this.Entries {
		if entry.Path == path {
			return i
		}
	}
	return -1
}

func (this *FileWalk) Contains(path string) bool {
	return this.IndexOf(path) != -1
}
