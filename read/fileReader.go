package read

import (
	"bufio"
	"codeleft-cli/filter"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CodeLeftReader interface remains the same
type CodeLeftReader interface {
	ReadHistory() (filter.Histories, error)
}

// HistoryReader is responsible for reading the history.ndjson file.
type HistoryReader struct {
	RepoRoot     string
	CodeleftPath string
}

// NewHistoryReader creates a new instance of HistoryReader.
// It walks the RepoRoot directory tree to find the first .codeleft folder it encounters.
// If repoRoot is empty, it defaults to the current working directory.
// Returns an error if .codeleft is not found anywhere in the repo.
func NewHistoryReader() (CodeLeftReader, error) {
	repoRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Recursively find .codeleft
	codeleftPath, err := findCodeleftRecursive(repoRoot)
	if err != nil {
		return nil, err
	}

	hr := &HistoryReader{
		RepoRoot:     repoRoot,
		CodeleftPath: codeleftPath,
	}
	return hr, nil
}

// ReadHistory reads the history.ndjson file from the discovered .codeleft directory.
// Returns an error if the history.ndjson file is not found or cannot be read.
func (hr *HistoryReader) ReadHistory() (filter.Histories, error) {
	// If .codeleft was not found, return an error
	if hr.CodeleftPath == "" {
		return nil, fmt.Errorf(".codeLeft folder not found in the repository root: %s", hr.RepoRoot)
	}

	historyPath := filepath.Join(hr.CodeleftPath, "history.ndjson")

	// Check if history.ndjson exists
	info, err := os.Stat(historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("history.ndjson does not exist at path: %s", historyPath)
		}
		return nil, fmt.Errorf("error accessing history.ndjson: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("history.ndjson exists but is a directory: %s", historyPath)
	}

	// Open the history.ndjson file
	file, err := os.Open(historyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open history.ndjson: %w", err)
	}
	defer file.Close()

	// Decode the NDJSON into a slice of History
	var history filter.Histories
	histories := filter.Histories{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var item filter.History
		if err := json.Unmarshal(scanner.Bytes(), &item); err != nil {
			return nil, fmt.Errorf("failed to decode history.ndjson: %w", err)
		}
		histories = append(histories, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading history.ndjson: %w", err)
	}

	history = histories

	return history, nil
}
