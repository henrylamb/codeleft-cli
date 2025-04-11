package report

import (
	"codeleft-cli/filter" // Assuming this path is correct
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReportNode represents a node (file or directory) in the report tree.
type ReportNode struct {
	Name           string
	Path           string                // Full path relative to root
	IsDir          bool
	Details        []filter.GradeDetails // Stores ALL GradeDetails for this file (if IsDir is false)
	Children       []*ReportNode         // Populated for directories (using pointers)
	Coverage       float64               // Calculated OVERALL coverage for this node
	CoverageOk     bool                  // Flag if overall coverage was calculable
	ToolCoverages  map[string]float64    // Coverage per tool (file's tool coverage OR directory's average coverage per tool)
	ToolCoverageOk map[string]bool       // Flag if coverage for a specific tool was calculable/present
}

// ReportViewData holds all data needed by the HTML template.
type ReportViewData struct {
	RootNodes       []*ReportNode      // Top-level files/dirs (using pointers)
	AllTools        []string           // Sorted list of unique tools found
	OverallAverages map[string]float64 // Average coverage per tool across ALL files
	TotalAverage    float64            // Overall average coverage across ALL files/tools
	ThresholdGrade  string             // The threshold grade used for calculations
}

// GenerateRepoHTMLReport generates the HTML report.
// Takes a slice of GradeDetail structs as input.
func GenerateRepoHTMLReport(gradeDetails []filter.GradeDetails, outputPath string, thresholdGrade string) error {
	if len(gradeDetails) == 0 {
		log.Println("Warning: No grade details provided to generate report.")
		// Optionally create an empty/minimal report or return an error
		// For now, let's proceed and it will likely generate an empty table
	}

	// 1. Group GradeDetails by FileName (path)
	groupedDetails := groupGradeDetailsByPath(gradeDetails)

	// 2. Build the ReportNode tree structure from the grouped paths.
	//    This step only creates the hierarchy, not coverages yet.
	rootNodes := buildReportTree(groupedDetails)

	// 3. Calculate coverages (file, directory averages) recursively,
	//    and collect global stats (tool names, sums for overall averages).
	toolSet := make(map[string]struct{})
	globalToolCoverageSums := make(map[string]float64) // Sum of coverage per tool across ALL FILES
	globalToolFileCounts := make(map[string]int)     // Files assessed per tool across ALL FILES

	for _, node := range rootNodes {
		calculateNodeCoverages(node, groupedDetails, thresholdGrade, toolSet, globalToolCoverageSums, globalToolFileCounts)
	}

	// Sort the tree alphabetically (dirs first) after calculations if needed
	sortReportNodes(rootNodes) // Apply sorting recursively

	// Convert tool set to a sorted slice.
	allTools := make([]string, 0, len(toolSet))
	for tool := range toolSet {
		allTools = append(allTools, tool)
	}
	sort.Strings(allTools)

	// 4. Calculate OVERALL report averages.
	overallAverages := make(map[string]float64)
	var totalCoverageSum float64
	var totalUniqueFilesWithCoverage int // Count unique files with valid coverage

	// Iterate through the *original* grouped data to get file-level data accurately
	processedFilesForTotalAvg := make(map[string]struct{}) // Track files counted

	for filePath, detailsList := range groupedDetails {
		if _, alreadyProcessed := processedFilesForTotalAvg[filePath]; alreadyProcessed {
			continue
		}

		var fileCoverageSum float64
		var fileToolCount int
		fileHasValidCoverage := false

		// Recalculate file's average coverage based *only* on its own tools
		processedToolsThisFile := make(map[string]struct{}) // Handle multiple entries for same tool if needed
		for _, detail := range detailsList {
			if _, toolDone := processedToolsThisFile[detail.Tool]; toolDone {
				continue // Skip if we already processed this tool for this file
			}
			if detail.Tool != "" && detail.Grade != "" {
				cov := calculateCoverage(detail.Grade, thresholdGrade)
				fileCoverageSum += cov
				fileToolCount++
				processedToolsThisFile[detail.Tool] = struct{}{}
				fileHasValidCoverage = true // Mark that this file contributes
			}
		}

		if fileHasValidCoverage && fileToolCount > 0 {
			fileAvg := fileCoverageSum / float64(fileToolCount)
			totalCoverageSum += fileAvg // Add the file's *average* coverage to the total sum
			totalUniqueFilesWithCoverage++
			processedFilesForTotalAvg[filePath] = struct{}{}
		}
	}

	// Calculate average per tool using globally collected sums/counts
	for _, tool := range allTools {
		sum := globalToolCoverageSums[tool]
		count := globalToolFileCounts[tool]
		if count > 0 {
			overallAverages[tool] = sum / float64(count)
		} else {
			overallAverages[tool] = 0 // Or potentially math.NaN()
		}
	}

	// Calculate final total average across all unique files with coverage
	var totalAverage float64
	if totalUniqueFilesWithCoverage > 0 {
		totalAverage = totalCoverageSum / float64(totalUniqueFilesWithCoverage)
	}

	// 5. Prepare data for the template.
	viewData := ReportViewData{
		RootNodes:       rootNodes,
		AllTools:        allTools,
		OverallAverages: overallAverages,
		TotalAverage:    totalAverage,
		ThresholdGrade:  thresholdGrade,
	}

	// 6. Parse and execute the template.
	tmpl, err := template.New("repoReport").Funcs(templateFuncs).Parse(repoReportTemplateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML output file '%s': %w", outputPath, err)
	}
	defer outputFile.Close()

	if err := tmpl.Execute(outputFile, viewData); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	fmt.Printf("Successfully generated repository report: %s\n", outputPath)
	return nil
}

// groupGradeDetailsByPath groups the flat list of details into a map
// where the key is the file path (FileName) and the value is a slice
// of all GradeDetails for that path.
func groupGradeDetailsByPath(details []filter.GradeDetails) map[string][]filter.GradeDetails {
	grouped := make(map[string][]filter.GradeDetails)
	for _, d := range details {
		// Normalize path separators for consistency
		normalizedPath := filepath.ToSlash(d.FileName)
		grouped[normalizedPath] = append(grouped[normalizedPath], d)
	}
	return grouped
}

// buildReportTree constructs the basic tree hierarchy from file paths.
// It does not calculate coverage here.
func buildReportTree(groupedDetails map[string][]filter.GradeDetails) []*ReportNode {
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

// calculateNodeCoverages recursively calculates coverage for nodes and collects global stats.
// It modifies the node directly.
func calculateNodeCoverages(
	node *ReportNode,
	groupedDetails map[string][]filter.GradeDetails, // Pass this down if needed, or use node.Details
	thresholdGrade string,
	toolSet map[string]struct{},
	globalToolCoverageSums map[string]float64,
	globalToolFileCounts map[string]int,
) {
	if node == nil {
		return
	}

	if !node.IsDir {
		// --- Process File Node ---
		var fileOverallCoverageSum float64
		var fileToolCount int
		processedToolsThisFile := make(map[string]struct{}) // Ensure each tool contributes once per file

		// Use the details stored directly on the node now
		for _, detail := range node.Details {
			if detail.Tool == "" || detail.Grade == "" {
				continue // Skip if tool or grade is missing
			}
			if _, toolDone := processedToolsThisFile[detail.Tool]; toolDone {
				continue // Only count first entry for a tool for this specific file node calculation
			}

			coverage := calculateCoverage(detail.Grade, thresholdGrade)
			node.ToolCoverages[detail.Tool] = coverage
			node.ToolCoverageOk[detail.Tool] = true
			toolSet[detail.Tool] = struct{}{} // Add tool to global set

			// Add to global sums/counts *only once* per file/tool combo
			globalToolCoverageSums[detail.Tool] += coverage
			globalToolFileCounts[detail.Tool]++

			fileOverallCoverageSum += coverage
			fileToolCount++
			processedToolsThisFile[detail.Tool] = struct{}{}
		}

		// Calculate the file's overall average coverage
		if fileToolCount > 0 {
			node.Coverage = fileOverallCoverageSum / float64(fileToolCount)
			node.CoverageOk = true
		} else {
			node.CoverageOk = false // No tools/grades found for this file
		}

	} else {
		// --- Process Directory Node ---
		// Recurse first to calculate children coverages
		for _, child := range node.Children {
			calculateNodeCoverages(child, groupedDetails, thresholdGrade, toolSet, globalToolCoverageSums, globalToolFileCounts)
		}

		// Now calculate this directory's averages based on its children
		var dirOverallCoverageSum float64
		dirNodesWithOverallCoverage := 0
		dirToolCoverageSums := make(map[string]float64)
		dirToolCoverageCounts := make(map[string]int)

		for _, child := range node.Children {
			// Aggregate overall coverage for the directory average
			if child.CoverageOk {
				dirOverallCoverageSum += child.Coverage
				dirNodesWithOverallCoverage++
			}

			// Aggregate per-tool coverage for the directory average
			for tool, coverage := range child.ToolCoverages {
				if child.ToolCoverageOk[tool] { // Check if the child had valid coverage for this tool
					dirToolCoverageSums[tool] += coverage
					dirToolCoverageCounts[tool]++
					// toolSet is already populated by file processing or deeper recursion
				}
			}
		}

		// Calculate and set the directory's overall average coverage
		if dirNodesWithOverallCoverage > 0 {
			node.Coverage = dirOverallCoverageSum / float64(dirNodesWithOverallCoverage)
			node.CoverageOk = true
		} else {
			node.CoverageOk = false // No children with valid coverage
		}

		// Calculate and set the directory's per-tool average coverage
		for tool, sum := range dirToolCoverageSums {
			count := dirToolCoverageCounts[tool]
			if count > 0 {
				node.ToolCoverages[tool] = sum / float64(count)
				node.ToolCoverageOk[tool] = true
			}
			// No need for else, ToolCoverageOk map default is false
		}
	}
}

// sortReportNodes recursively sorts children nodes: directories first, then alphabetically.
func sortReportNodes(nodes []*ReportNode) {
	// Sort the current level
	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir // true (directory) comes before false (file)
		}
		return nodes[i].Name < nodes[j].Name
	})

	// Recursively sort children of directories
	for _, node := range nodes {
		if node.IsDir && len(node.Children) > 0 {
			sortReportNodes(node.Children)
		}
	}
}

// --- Utility functions (getGradeIndex, calculateCoverage) remain the same ---
func getGradeIndex(grade string) int {
	gradeIndices := map[string]int{
		"A*": 5, "A": 4, "B": 3, "C": 2, "D": 1, "F": 0,
	}
	// Ensure comparison is case-insensitive
	index, ok := gradeIndices[strings.ToUpper(grade)]
	if !ok {
		log.Printf("Warning: Unrecognized grade '%s', treating as F (0)", grade)
		return 0 // Default to lowest index for unrecognized grades
	}
	return index
}
func calculateCoverage(grade, thresholdGrade string) float64 {
	gradeIndex := getGradeIndex(grade)
	thresholdIndex := getGradeIndex(thresholdGrade)

	// Logic matches the JS example and previous Go version
	if gradeIndex > thresholdIndex {
		return 120.0
	} else if gradeIndex == thresholdIndex {
		return 100.0
	} else if gradeIndex >= thresholdIndex-1 {
		return 70.0
	} else if gradeIndex >= thresholdIndex-2 {
		return 50.0
	} else if gradeIndex >= thresholdIndex-3 {
		return 30.0
	} else {
		return 10.0
	}
}
