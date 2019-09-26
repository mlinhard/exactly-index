package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	CONFIG_FOLDER = "exactly-index"      // inside user's config dir
	CONFIG_FILE   = "server-config.json" // inside config folder
)

type ServerConfig struct {
	ListenAddress  string   `json:"listen_address"`
	NumFileLoaders int      `json:"num_file_loaders"`
	NumFileStaters int      `json:"num_file_staters"`
	Roots          []string `json:"roots"`
	IgnoredDirs    []string `json:"ignored_directories"`
}

func saveConfigTo(configFile string, config *ServerConfig) error {
	bytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("Config marshalling: %v", err)
	}
	configDir := filepath.Dir(configFile)
	err = os.MkdirAll(configDir, 0770)
	if err != nil {
		return fmt.Errorf("Couldn't create config dir: %v: %v", configDir, err)
	}
	err = ioutil.WriteFile(configFile, bytes, 0664)
	if err != nil {
		return fmt.Errorf("Config writing: %v", err)
	}
	return nil
}

func loadConfigFrom(configFile string) (*ServerConfig, error) {
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	config := new(ServerConfig)
	err = json.Unmarshal(bytes, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func getUserConfigFile() (string, error) {
	// TODO: redo, once https://golang.org/pkg/os/#UserConfigDir is complete
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homedir, ".config", CONFIG_FOLDER, CONFIG_FILE), nil
}

func LoadConfig() (*ServerConfig, error) {
	configPath, err := getUserConfigFile()
	if err != nil {
		return nil, fmt.Errorf("Determining home directory: %v", err)
	}
	_, err = os.Stat(configPath)
	var config *ServerConfig
	if os.IsNotExist(err) {
		config = defaultConfig()
		err = saveConfigTo(configPath, config)
		if err != nil {
			return nil, fmt.Errorf("Saving newly created default config: %v", err)
		}
	} else {
		config, err = loadConfigFrom(configPath)
		if err != nil {
			return nil, fmt.Errorf("Loading config: %v", err)
		}
	}
	return config, nil
}

func defaultConfig() *ServerConfig {
	return &ServerConfig{"localhost:8080", 4, 4, []string{"."}, []string{}}
}
