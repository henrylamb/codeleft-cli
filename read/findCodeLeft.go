package read

import (
	"fmt"
	"os"
	"path/filepath"
)

// findCodeleftRecursive searches recursively for a directory named ".codeleft" under "root".
func findCodeleftRecursive(root string) (string, error) {
	var codeleftPath string

	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		// Check if current path is a directory named ".codeLeft"
		if info.IsDir() && filepath.Base(path) == ".codeLeft" {
			codeleftPath = path
			// Skip descending further once we've found a match
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if codeleftPath == "" {
		return "", fmt.Errorf(".codeLeft directory does not exist anywhere under: %s", root)
	}

	return codeleftPath, nil
}
