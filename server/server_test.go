package server

import (
	"sync"
	"testing"
)

func TestConfigSaveLoad(t *testing.T) {
	tmpDir := NewTempDir(t, "exactly-index-test")
	if !tmpDir.OK {
		return
	}
	defer tmpDir.Remove()
	tmpConfigFile := tmpDir.Path("folder1/folder2/test-config.json")
	config := &ServerConfig{
		"192.168.10.10:9090",
		3,
		6,
		[]string{"/home/mlinhard/Documents1", "/home/mlinhard/Documents2"},
		[]string{".git"}}
	err := saveConfigTo(tmpConfigFile, config)
	if err != nil {
		t.Error(err)
		return
	}
	var loadedConfig *ServerConfig
	loadedConfig, err = loadConfigFrom(tmpConfigFile)
	if err != nil {
		t.Error(err)
		return
	}
	if loadedConfig.ListenAddress != "192.168.10.10:9090" {
		t.Errorf("Loaded wrong listen_address: %v", config.ListenAddress)
	}
	if loadedConfig.NumFileLoaders != 3 {
		t.Errorf("Loaded wrong num_file_loaders: %v", config.ListenAddress)
	}
	if loadedConfig.NumFileStaters != 6 {
		t.Errorf("Loaded wrong nim_file_staters: %v", config.ListenAddress)
	}
	if len(loadedConfig.Roots) != 2 || loadedConfig.Roots[0] != "/home/mlinhard/Documents1" || loadedConfig.Roots[1] != "/home/mlinhard/Documents2" {
		t.Errorf("Loaded wrong roots: %v", config.Roots)
	}
	if len(loadedConfig.IgnoredDirs) != 1 || loadedConfig.IgnoredDirs[0] != ".git" {
		t.Errorf("Loaded wrong ignored_directories: %v", config.IgnoredDirs)
	}
}

func TestFileWalkDocumentLoad(t *testing.T) {
	tmpDir := NewTempDir(t, "exactly-test")
	if !tmpDir.OK {
		return
	}
	defer tmpDir.Remove()
	createTestFiles(tmpDir)
	if !tmpDir.OK {
		return
	}
	roots := tmpDir.Paths("docs", "texts")
	ignored := []string{tmpDir.Path("docs/ignored1"), "ignored2"}
	fileWalk := &TestFileWalk{t, NewFileWalk(roots, ignored)}
	fileWalk.assertOK()
	fileWalk.assertSize(10)
	fileWalk.assertContains("docs/file01.txt")
	fileWalk.assertContains("docs/file02.txt")
	fileWalk.assertContains("docs/file03.txt")
	fileWalk.assertContains("docs/bla1/file04.txt")
	fileWalk.assertContains("docs/bla1/file05.txt")
	fileWalk.assertContains("docs/bla2/file06.txt")
	fileWalk.assertContains("docs/bla2/file07.txt")
	fileWalk.assertContains("texts/file08.txt")
	fileWalk.assertContains("texts/file09.txt")
	fileWalk.assertContains("texts/bla3/file10.txt")

	docStore := loadTestDocStore(tmpDir, fileWalk)
	docStore.assertTotalBytes(43)
	docStore.assertSize(10)
	docStore.assertContains("docs/file01.txt", "AAAA")
	docStore.assertContains("docs/file02.txt", "BBBB")
	docStore.assertContains("docs/file03.txt", "CCCC")
	docStore.assertContains("docs/bla1/file04.txt", "DDDD")
	docStore.assertContains("docs/bla1/file05.txt", "EEABC")
	docStore.assertContains("docs/bla2/file06.txt", "FFABC")
	docStore.assertContains("docs/bla2/file07.txt", "GGABC")
	docStore.assertContains("texts/file08.txt", "HHHH")
	docStore.assertContains("texts/file09.txt", "IIII")
	docStore.assertContains("texts/bla3/file10.txt", "JJJJ")
}

func TestStartsWith(t *testing.T) {
	if startsWith("haha", "/home/mlinhard") {
		t.Errorf("error")
	}
	if startsWith("/home/mli", "/home/mlinhard") {
		t.Errorf("error")
	}
	if !startsWith("/home", "/home/mlinhard") {
		t.Errorf("error")
	}
	if !startsWith("/home/mlinhard", "/home/mlinhard") {
		t.Errorf("error")
	}
}

func TestIgnorer(t *testing.T) {
	ignorer := createIgnorer("ign1", "/home/docs/ign2")
	assertIgnored(t, ignorer, "ign1")
	assertIgnored(t, ignorer, "/home/ign1")
	assertNotIgnored(t, ignorer, "/home/ign1/bla")
	assertIgnored(t, ignorer, "/home/docs/ign2")
	assertIgnored(t, ignorer, "/home/docs/ign2/subdir")
	assertIgnored(t, ignorer, "/home/docs/ign2/subdir/aa")
	assertNotIgnored(t, ignorer, "/home/bla")
	assertNotIgnored(t, ignorer, "/home/ing2/docs")
	assertNotIgnored(t, ignorer, "ign2")
}

func TestSearchServer(t *testing.T) {
	tmpDir := NewTempDir(t, "exactly-test")
	if !tmpDir.OK {
		return
	}
	defer tmpDir.Remove()
	createTestFiles(tmpDir)
	if !tmpDir.OK {
		return
	}
	roots := tmpDir.Paths("docs", "texts")
	ignore := []string{tmpDir.Path("docs/ignored1"), "ignored2"}
	config := createTestServerConfig(roots, ignore)
	testServer := createTestServer(t, config)
	testServer.server.Start()
	defer testServer.server.Stop()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	testServer.server.OnDoneIndexing(func(int, int) {
		wg.Done()
	})
	client := testServer.client()
	wg.Wait()
	resp := client.search("ABC")
	if resp == nil {
		return
	}
	resp.AssertHitCount(3)
	resp.HitByDocumentId(tmpDir.Path("docs/bla1/file05.txt")).AssertBefore("EE").AssertAfter("")
	resp.HitByDocumentId(tmpDir.Path("docs/bla2/file06.txt")).AssertBefore("FF").AssertAfter("")
	resp.HitByDocumentId(tmpDir.Path("docs/bla2/file07.txt")).AssertBefore("GG").AssertAfter("")

	resp = client.searchBounded("ABC", 10, 2)
	if resp == nil {
		return
	}
	resp.AssertHitCount(2)

}
