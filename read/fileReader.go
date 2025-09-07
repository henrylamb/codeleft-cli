package read

import (
	"bufio"
	"bytes"
	"codeleft-cli/filter"
	"encoding/json"
	"fmt"
	"io"
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
	
	// Use a more robust line-by-line reader that can handle very large lines
	reader := bufio.NewReader(file)
	lineNumber := 0
	
	for {
		lineNumber++
		
		// Read a complete line, handling very large lines gracefully
		var lineBuffer bytes.Buffer
		for {
			chunk, isPrefix, err := reader.ReadLine()
			if err != nil {
				if err == io.EOF {
					// Process any remaining content in the buffer
					if lineBuffer.Len() > 0 {
						line := bytes.TrimSpace(lineBuffer.Bytes())
						if len(line) > 0 {
							var item filter.History
							if err := json.Unmarshal(line, &item); err != nil {
								return nil, fmt.Errorf("failed to decode history.ndjson at line %d: %w", lineNumber, err)
							}
							histories = append(histories, item)
						}
					}
					goto done
				}
				return nil, fmt.Errorf("error reading history.ndjson at line %d: %w", lineNumber, err)
			}
			
			lineBuffer.Write(chunk)
			
			// If isPrefix is false, we've read the complete line
			if !isPrefix {
				break
			}
		}
		
		// Process the complete line
		line := bytes.TrimSpace(lineBuffer.Bytes())
		
		// Skip empty lines
		if len(line) == 0 {
			continue
		}
		
		// Parse the JSON line
		var item filter.History
		if err := json.Unmarshal(line, &item); err != nil {
			return nil, fmt.Errorf("failed to decode history.ndjson at line %d: %w", lineNumber, err)
		}
		
		histories = append(histories, item)
	}
	
done:
	history = histories

	return history, nil
}
