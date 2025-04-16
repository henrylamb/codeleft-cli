package report

import (
	"codeleft-cli/filter"
	"path/filepath"
	"sort"
	"strings"
)

// CoverageData interface for abstracting coverage data.
type CoverageData interface {
	GetCoverage(tool string) (float64, bool)
	SetCoverage(tool string, coverage float64, ok bool)
}

// PathSplitter interface for splitting file paths.
type PathSplitter interface {
	Split(path string) []string
}

// SeparatorPathSplitter splits paths by the OS-specific separator.
type SeparatorPathSplitter struct{}

func NewSeparatorPathSplitter() PathSplitter {
	return &SeparatorPathSplitter{}
}

func (s *SeparatorPathSplitter) Split(path string) []string {
	return strings.Split(path, string(filepath.Separator))
}

// NodeCreator interface for creating ReportNode instances.
type NodeCreator interface {
	CreateFileNode(name string, path string, details []filter.GradeDetails) *ReportNode
	CreateDirectoryNode(name string, path string) *ReportNode
}

// DefaultNodeCreator is a concrete implementation of NodeCreator.
type DefaultNodeCreator struct{}

func NewDefaultNodeCreator() NodeCreator {
	return &DefaultNodeCreator{}
}

func (c *DefaultNodeCreator) CreateFileNode(name string, path string, details []filter.GradeDetails) *ReportNode {
	return &ReportNode{
		Name:    name,
		Path:    path,
		IsDir:   false,
		Details: details,
	}
}

func (c *DefaultNodeCreator) CreateDirectoryNode(name string, path string) *ReportNode {
	return &ReportNode{
		Name:     name,
		Path:     path,
		IsDir:    true,
		Children: []*ReportNode{},
	}
}

// TreeBuilder is responsible for constructing the ReportNode tree.
type TreeBuilder struct {
	pathSplitter PathSplitter
	nodeCreator  NodeCreator
}

// NewTreeBuilder creates a new TreeBuilder with provided dependencies.
func NewTreeBuilder(pathSplitter PathSplitter, nodeCreator NodeCreator) *TreeBuilder {
	return &TreeBuilder{
		pathSplitter: pathSplitter,
		nodeCreator:  nodeCreator,
	}
}

// GroupGradeDetailsByPath groups grade details by file path.
func (tb *TreeBuilder) GroupGradeDetailsByPath(details []filter.GradeDetails) map[string][]filter.GradeDetails {
	grouped := make(map[string][]filter.GradeDetails)
	for _, d := range details {
		// Normalize path separators for consistency
		normalizedPath := filepath.ToSlash(d.FileName)
		grouped[normalizedPath] = append(grouped[normalizedPath], d)
	}
	return grouped
}

// buildTree recursively builds the report tree from path parts.
func (tb *TreeBuilder) buildTree(roots []*ReportNode, dirs map[string]*ReportNode, parts []string, fullPath string, details []filter.GradeDetails) []*ReportNode {
	if len(parts) == 0 {
		return roots
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

		existingNode, found := dirs[currentPath]

		if isLastPart {
			fileNode := tb.nodeCreator.CreateFileNode(part, fullPath, details)
			if parent == nil {
				roots = append(roots, fileNode)
			} else {
				parent.Children = append(parent.Children, fileNode)
			}
		} else {
			if found {
				parent = existingNode
			} else {
				dirNode := tb.nodeCreator.CreateDirectoryNode(part, currentPath)
				dirs[currentPath] = dirNode

				if parent == nil {
					roots = append(roots, dirNode)
				} else {
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
				parent = dirNode
			}
		}
	}
	return roots
}

// BuildReportTree constructs the basic tree hierarchy from file paths.
// It does not calculate coverage here.
func (tb *TreeBuilder) BuildReportTree(groupedDetails map[string][]filter.GradeDetails) []*ReportNode {
	roots := []*ReportNode{}
	dirs := make(map[string]*ReportNode)

	paths := make([]string, 0, len(groupedDetails))
	for p := range groupedDetails {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, fullPath := range paths {
		details := groupedDetails[fullPath]
		parts := tb.pathSplitter.Split(fullPath)
		if len(parts) == 0 {
			continue
		}
		roots = tb.buildTree(roots, dirs, parts, fullPath, details)
	}

	return roots
}