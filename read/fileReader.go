package read

import (
	"codeleft-cli/filter"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CodeLeftReader interface {
	ReadHistory() (filter.Histories, error)
}

// HistoryReader is responsible for reading the history.json file.
type HistoryReader struct {
	RepoRoot     string
	CodeleftPath string
}

// NewHistoryReader creates a new instance of HistoryReader.
// It searches for the .codeleft folder directly within the RepoRoot.
// If repoRoot is empty, it defaults to the current working directory.
// Returns an error if .codeleft is not found.
func NewHistoryReader() (CodeLeftReader, error) {
	repoRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	hr := &HistoryReader{RepoRoot: repoRoot}

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

	hr.CodeleftPath = expectedCodeleftPath

	return hr, nil
}

// ReadHistory reads the history.json file from the .codeleft directory.
// Returns an error if the history.json file is not found or cannot be read.
func (hr *HistoryReader) ReadHistory() (filter.Histories, error) {
	// If .codeleft was not found, return an error
	if hr.CodeleftPath == "" {
		return nil, fmt.Errorf(".codeleft folder not found in the repository root: %s", hr.RepoRoot)
	}

	historyPath := filepath.Join(hr.CodeleftPath, "history.json")

	// Check if history.json exists
	info, err := os.Stat(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("history.json does not exist at path: %s", historyPath)
		}
		return nil, fmt.Errorf("error accessing history.json: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("history.json exists but is a directory: %s", historyPath)
	}

	// Open the history.json file
	file, err := os.Open(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history.json: %w", err)
	}
	defer file.Close()

	// Decode the JSON into a slice of History
	var history filter.Histories
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to decode history.json: %w", err)
	}

	return history, nil
}
