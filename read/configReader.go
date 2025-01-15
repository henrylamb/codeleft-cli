package read

import (
	"codeleft-cli/types"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigReader is responsible for reading the config.json file.
type ConfigReader struct {
	RepoRoot     string
	CodeleftPath string
}

// NewConfigReader creates a new instance of ConfigReader.
// It walks the RepoRoot directory tree to find the first .codeleft folder it encounters.
// If repoRoot is empty, it defaults to the current working directory.
// Returns an error if .codeleft is not found anywhere in the repo.
func NewConfigReader() (*ConfigReader, error) {
	repoRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Recursively find .codeleft
	codeleftPath, err := findCodeleftRecursive(repoRoot)
	if err != nil {
		return nil, err
	}

	cr := &ConfigReader{
		RepoRoot:     repoRoot,
		CodeleftPath: codeleftPath,
	}
	return cr, nil
}

// ReadConfig reads the config.json file from the .codeleft directory.
// Returns a Config instance and an error if the config.json file is not found or cannot be read.
func (cr *ConfigReader) ReadConfig() (*types.Config, error) {
	// Ensure the .codeleft path is set
	if cr.CodeleftPath == "" {
		return nil, fmt.Errorf(".codeLeft folder not found in the repository root: %s", cr.RepoRoot)
	}

	configPath := filepath.Join(cr.CodeleftPath, "config.json")

	// Check if config.json exists
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config.json does not exist at path: %s", configPath)
		}
		return nil, fmt.Errorf("error accessing config.json: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("config.json exists but is a directory: %s", configPath)
	}

	// Open the config.json file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config.json: %w", err)
	}
	defer file.Close()

	// Decode the JSON into a Config struct
	var config types.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config.json: %w", err)
	}

	return &config, nil
}
