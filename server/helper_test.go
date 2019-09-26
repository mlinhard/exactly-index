package server

import (
	"fmt"
	"testing"
)

type testServer struct {
	t      *testing.T
	server *Server
}

func (this *testServer) client() *testClient {
	client := new(testClient)
	client.t = this.t
	client.baseUrl = fmt.Sprintf("http://%v", this.server.config.ListenAddress)
	client.maxContext = 10
	client.maxHits = 10
	return client
}

func createTestServerConfig(roots []string, ignore []string) *ServerConfig {
	config := new(ServerConfig)
	config.ListenAddress = "localhost:9876"
	config.Roots = roots
	config.IgnoredDirs = ignore
	config.NumFileLoaders = 4
	config.NumFileStaters = 4
	return config
}

func createTestServer(t *testing.T, config *ServerConfig) *testServer {
	server := new(testServer)
	server.t = t
	server.server = NewServer(config)
	server.server.OnError(func(err error) {
		t.Errorf("ERROR: %v", err)
	})
	return server
}

func createTestFiles(tmpDir *TempDir) {
	tmpDir.WriteFile("docs/ignored1/file01.txt", "ignored1")
	tmpDir.WriteFile("docs/ignored1/file02.txt", "ignored2")
	tmpDir.WriteFile("docs/ignored1/file03.txt", "ignored3")
	tmpDir.WriteFile("docs/file01.txt", "AAAA")
	tmpDir.WriteFile("docs/file02.txt", "BBBB")
	tmpDir.WriteFile("docs/file03.txt", "CCCC")
	tmpDir.WriteFile("docs/bla1/file04.txt", "DDDD")
	tmpDir.WriteFile("docs/bla1/file05.txt", "EEABC")
	tmpDir.WriteFile("docs/bla2/file06.txt", "FFABC")
	tmpDir.WriteFile("docs/bla2/file07.txt", "GGABC")
	tmpDir.WriteFile("texts/file08.txt", "HHHH")
	tmpDir.WriteFile("texts/file09.txt", "IIII")
	tmpDir.WriteFile("texts/bla3/file10.txt", "JJJJ")
	tmpDir.WriteFile("texts/ignored2/file04.txt", "ignored4")
}

func assertIgnored(t *testing.T, ign *ignorer, path string) {
	if !ign.match(path) {
		t.Errorf("Path %v should be ignored", path)
	}
}

func assertNotIgnored(t *testing.T, ign *ignorer, path string) {
	if ign.match(path) {
		t.Errorf("Path %v should not be ignored", path)
	}
}
