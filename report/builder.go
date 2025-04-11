package report

import (
	"codeleft-cli/filter" // Assuming this path is correct
	"path/filepath"
	"sort"
	"strings"
)

// TreeBuilder is responsible for constructing the ReportNode tree.
// SRP: Focused on tree building logic.
type TreeBuilder struct{}

func NewTreeBuilder() *TreeBuilder {
	return &TreeBuilder{}
}

// GroupGradeDetailsByPath groups the flat list of details into a map
// where the key is the file path (FileName) and the value is a slice
// of all GradeDetails for that path.
// SRP: Focused on grouping input data.
func (tb *TreeBuilder) GroupGradeDetailsByPath(details []filter.GradeDetails) map[string][]filter.GradeDetails {
	grouped := make(map[string][]filter.GradeDetails)
	for _, d := range details {
		// Normalize path separators for consistency
		normalizedPath := filepath.ToSlash(d.FileName)
		grouped[normalizedPath] = append(grouped[normalizedPath], d)
	}
	return grouped
}

// BuildReportTree constructs the basic tree hierarchy from file paths.
// It does not calculate coverage here.
func (tb *TreeBuilder) BuildReportTree(groupedDetails map[string][]filter.GradeDetails) []*ReportNode {
    roots := []*ReportNode{}
    // Use a map to keep track of created directory nodes by their full path
    // Ensures we don't create duplicate nodes for the same directory
    dirs := make(map[string]*ReportNode)

    // Sort paths for potentially more structured processing (optional but can help)
    paths := make([]string, 0, len(groupedDetails))
    for p := range groupedDetails {
        paths = append(paths, p)
    }
    sort.Strings(paths)

    for _, fullPath := range paths {
        details := groupedDetails[fullPath] // Get the details for this file
        parts := strings.Split(fullPath, "/")
        if len(parts) == 0 {
            continue // Skip empty paths
        }

        var parent *ReportNode
        currentPath := ""

        for i, part := range parts {
            isLastPart := (i == len(parts)-1)
            if currentPath == "" {
                currentPath = part
            } else {
                currentPath = currentPath + "/" + part
            }

            // Check if node already exists (could be a dir created by a previous path)
            existingNode, found := dirs[currentPath]

            if isLastPart { // This is the file part
                fileNode := &ReportNode{
                    Name:           part,
                    Path:           fullPath, // Store the full original path
                    IsDir:          false,
                    Details:        details, // Store associated details
                    ToolCoverages:  make(map[string]float64),
                    ToolCoverageOk: make(map[string]bool),
                }
                if parent == nil { // File in root
                    roots = append(roots, fileNode)
                } else {
                    parent.Children = append(parent.Children, fileNode)
                }
                // Don't add files to the 'dirs' map
            } else { // This is a directory part
                if found {
                    // Directory node already exists, just update parent pointer
                    parent = existingNode
                } else {
                    // Create new directory node
                    dirNode := &ReportNode{
                        Name:           part,
                        Path:           currentPath, // Path up to this directory
                        IsDir:          true,
                        Children:       []*ReportNode{},
                        ToolCoverages:  make(map[string]float64),
                        ToolCoverageOk: make(map[string]bool),
                    }
                    dirs[currentPath] = dirNode // Add to map for lookup

                    if parent == nil { // Directory in root
                        roots = append(roots, dirNode)
                    } else {
                        // Check if child already exists in parent (can happen with sorting/processing order)
                        childExists := false
                        for _, child := range parent.Children {
                            if child.Path == dirNode.Path {
                                childExists = true
                                break
                            }
                        }
                        if !childExists {
                            parent.Children = append(parent.Children, dirNode)
                        }
                    }
                    parent = dirNode // This new dir becomes the parent for the next part
                }
            }
        }
    }
    return roots
}