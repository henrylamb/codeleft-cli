package filter

import (
	"codeleft-cli/types"
	"path/filepath"
	"strings"
)

// FilterRule defines the interface for any filtering rule.
type FilterRule interface {
	Match(path string) bool
}

// IgnoreFileRule implements FilterRule for ignoring specific files.
// It encapsulates the logic for file path comparison, adhering to SRP.
type IgnoreFileRule struct {
	ignoredFiles []types.File
}

// NewIgnoreFileRule creates a new IgnoreFileRule.
func NewIgnoreFileRule(files []types.File) *IgnoreFileRule {
	return &IgnoreFileRule{ignoredFiles: files}
}

// Match checks if the given path matches any of the ignored files.
func (r *IgnoreFileRule) Match(path string) bool {
	normalizedPath := filepath.ToSlash(path)
	for _, file := range r.ignoredFiles {
		ignoredFilePath := filepath.ToSlash(filepath.Join(file.Path, file.Name))
		if normalizedPath == ignoredFilePath {
			return true
		}
	}
	return false
}

// IgnoreFolderRule implements FilterRule for ignoring files within specific folders.
// It encapsulates the logic for folder path comparison, adhering to SRP.
type IgnoreFolderRule struct {
	ignoredFolders []string
}

// NewIgnoreFolderRule creates a new IgnoreFolderRule.
func NewIgnoreFolderRule(folders []string) *IgnoreFolderRule {
	return &IgnoreFolderRule{ignoredFolders: folders}
}

// Match checks if the given path resides within any of the ignored folders.
func (r *IgnoreFolderRule) Match(path string) bool {
	normalizedPath := filepath.ToSlash(path)
	dirs := strings.Split(normalizedPath, "/")

	// Exclude the last element if it's a file, as we are checking for folder containment.
	// This assumes that the path ends with a file name.
	for _, dir := range dirs[:len(dirs)-1] {
		for _, ignoredFolder := range r.ignoredFolders {
			if dir == ignoredFolder {
				return true
			}
		}
	}
	return false
}

// PathFilter is responsible for orchestrating the filtering of file paths.
// It depends on abstractions (FilterRule interface), adhering to DIP.
type PathFilter struct {
	rules []FilterRule
}

// NewPathFilter creates a new instance of PathFilter with the provided filtering rules.
// This constructor allows injecting different filtering strategies.
func NewPathFilter(rules ...FilterRule) *PathFilter {
	return &PathFilter{
		rules: rules,
	}
}

// Filter filters out the file paths that match any of the configured rules.
// It returns a new slice containing only the file histories that are not ignored.
func (pf *PathFilter) Filter(histories Histories) Histories {
	var newHistories Histories

	for _, history := range histories {
		if !pf.isIgnored(history.FilePath) {
			newHistories = append(newHistories, history)
		}
	}

	return newHistories
}

// isIgnored checks whether a given file path matches any of the configured filter rules.
// This method's responsibility is solely to check against the rules, adhering to SRP.
func (pf *PathFilter) isIgnored(path string) bool {
	for _, rule := range pf.rules {
		if rule.Match(path) {
			return true
		}
	}
	return false
}