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
// It searches for the .codeleft folder directly within the RepoRoot.
// If repoRoot is empty, it defaults to the current working directory.
// Returns an error if .codeleft is not found.
func NewConfigReader() (*ConfigReader, error) {
	repoRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	cr := &ConfigReader{RepoRoot: repoRoot}

	// Define the expected path for .codeleft directly under RepoRoot
	expectedCodeleftPath := filepath.Join(repoRoot, ".codeleft")

	// Check if .codeleft exists directly under RepoRoot
	info, err := os.Stat(expectedCodeleftPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(".codeleft directory does not exist in the repository root: %s", repoRoot)
		}
		return nil, fmt.Errorf("error accessing .codeleft directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf(".codeleft exists but is not a directory: %s", expectedCodeleftPath)
	}

	cr.CodeleftPath = expectedCodeleftPath

	return cr, nil
}

// ReadConfig reads the config.json file from the .codeleft directory.
// Returns a Config instance and an error if the config.json file is not found or cannot be read.
func (cr *ConfigReader) ReadConfig() (*types.Config, error) {
	// Ensure the .codeleft path is set
	if cr.CodeleftPath == "" {
		return nil, fmt.Errorf(".codeleft folder not found in the repository root: %s", cr.RepoRoot)
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
