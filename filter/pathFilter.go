package filter

import (
	"codeleft-cli/types"
	"path/filepath"
	"strings"
)

// PathFilter is responsible for filtering file paths based on ignored files and folders.
type PathFilter struct {
	ignoredFiles   []types.File
	ignoredFolders []string
}

// NewPathFilter creates a new instance of PathFilter with the provided ignored files and folders.
func NewPathFilter(ignoredFiles []types.File, ignoredFolders []string) *PathFilter {
	return &PathFilter{
		ignoredFiles:   ignoredFiles,
		ignoredFolders: ignoredFolders,
	}
}

// Filter filters out the file paths that match any of the ignored files or reside within ignored folders.
// It returns a new slice containing only the file paths that are not ignored.
func (pf *PathFilter) Filter(histories Histories) Histories {
	var newHistories Histories

	for _, history := range histories {
		if !pf.isIgnored(history.FilePath) {
			newHistories = append(newHistories, history)
		}
	}

	return newHistories
}

// isIgnored checks whether a given file path matches any ignored file or is within any ignored folder.
func (pf *PathFilter) isIgnored(path string) bool {
	// Normalize the path for consistent comparison
	normalizedPath := filepath.ToSlash(path)

	// Check against ignored files
	for _, file := range pf.ignoredFiles {
		// Join the file path and name, then normalize
		ignoredFilePath := filepath.ToSlash(filepath.Join(file.Path, file.Name))
		if normalizedPath == ignoredFilePath {
			return true
		}
	}

	// Split the path into directories
	dirs := strings.Split(normalizedPath, "/")

	// Exclude the last element if it's a file
	// This assumes that the path ends with a file name. Adjust if directories can also be in histories.
	for _, dir := range dirs[:len(dirs)-1] {
		for _, ignoredFolder := range pf.ignoredFolders {
			if dir == ignoredFolder {
				return true
			}
		}
	}

	return false
}
