package report

import (
	"sort"
)

// CoverageCalculator encapsulates the logic for calculating various coverage metrics.
// SRP: Focused on coverage calculation logic.
type CoverageCalculator struct {
	ThresholdGrade string
}

func NewCoverageCalculator(thresholdGrade string) *CoverageCalculator {
	return &CoverageCalculator{ThresholdGrade: thresholdGrade}
}

// GlobalStats holds aggregated statistics across the entire report.
type GlobalStats struct {
	ToolSet              map[string]struct{}
	ToolCoverageSums     map[string]float64
	ToolFileCounts       map[string]int
	TotalCoverageSum     float64 // Sum of *average* coverage for each unique file
	UniqueFilesProcessed map[string]struct{}
}

func NewGlobalStats() *GlobalStats {
	return &GlobalStats{
		ToolSet:              make(map[string]struct{}),
		ToolCoverageSums:     make(map[string]float64),
		ToolFileCounts:       make(map[string]int),
		UniqueFilesProcessed: make(map[string]struct{}),
	}
}

// CalculateNodeCoverages recursively calculates coverage for nodes and collects global stats.
// It modifies the node directly and updates global stats.
func (cc *CoverageCalculator) CalculateNodeCoverages(node *ReportNode, stats *GlobalStats) {
	if node == nil {
		return
	}

	if !node.IsDir {
		cc.calculateFileNodeCoverage(node, stats)
	} else {
		cc.calculateDirectoryNodeCoverage(node, stats)
	}
}

// calculateFileNodeCoverage calculates coverage for a single file node.
func (cc *CoverageCalculator) calculateFileNodeCoverage(node *ReportNode, stats *GlobalStats) {
	var fileOverallCoverageSum float64
	var fileToolCount int
	processedToolsThisFile := make(map[string]struct{}) // Ensure each tool contributes once per file

	// Use the details stored directly on the node
	for _, detail := range node.Details {
		if detail.Tool == "" || detail.Grade == "" {
			continue // Skip if tool or grade is missing
		}
		tool := detail.Tool
		if _, toolDone := processedToolsThisFile[tool]; toolDone {
			continue // Only count first entry for a tool for this specific file node calculation
		}

		coverage := calculateCoverageScore(detail.Grade, cc.ThresholdGrade)
		node.ToolCoverages[tool] = coverage
		node.ToolCoverageOk[tool] = true
		stats.ToolSet[tool] = struct{}{} // Add tool to global set

		// Add to global sums/counts *only once* per file/tool combo
		stats.ToolCoverageSums[tool] += coverage
		stats.ToolFileCounts[tool]++

		fileOverallCoverageSum += coverage
		fileToolCount++
		processedToolsThisFile[tool] = struct{}{}
	}

	// Calculate the file's overall average coverage
	if fileToolCount > 0 {
		node.Coverage = fileOverallCoverageSum / float64(fileToolCount)
		node.CoverageOk = true
		// Add to global *total* average calculation *if* not already processed
		if _, processed := stats.UniqueFilesProcessed[node.Path]; !processed {
             stats.TotalCoverageSum += node.Coverage // Add file's average coverage
             stats.UniqueFilesProcessed[node.Path] = struct{}{}
        }
	} else {
		node.CoverageOk = false // No tools/grades found for this file
	}
}

// calculateDirectoryNodeCoverage calculates coverage for a directory node based on its children.
func (cc *CoverageCalculator) calculateDirectoryNodeCoverage(node *ReportNode, stats *GlobalStats) {
	// Recurse first to calculate children coverages
	for _, child := range node.Children {
		cc.CalculateNodeCoverages(child, stats) // Pass stats down
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


// CalculateOverallAverages computes the final report-wide averages from global stats.
// SRP: Focused on calculating final aggregate averages.
func (cc *CoverageCalculator) CalculateOverallAverages(stats *GlobalStats) (overallAvg map[string]float64, totalAvg float64, allTools []string) {
	overallAvg = make(map[string]float64)
	allTools = make([]string, 0, len(stats.ToolSet))
	for tool := range stats.ToolSet {
		allTools = append(allTools, tool)
	}
	sort.Strings(allTools)

	// Calculate average per tool using globally collected sums/counts
	for _, tool := range allTools {
		sum := stats.ToolCoverageSums[tool]
		count := stats.ToolFileCounts[tool]
		if count > 0 {
			overallAvg[tool] = sum / float64(count)
		} else {
			overallAvg[tool] = 0 // Or potentially math.NaN()
		}
	}

	// Calculate final total average across all unique files with coverage
    totalUniqueFilesWithCoverage := len(stats.UniqueFilesProcessed)
	if totalUniqueFilesWithCoverage > 0 {
		totalAvg = stats.TotalCoverageSum / float64(totalUniqueFilesWithCoverage)
	}

	return overallAvg, totalAvg, allTools
}

// calculateCoverage calculates coverage score based on grade and threshold.
// SRP: Focused responsibility for grade-to-coverage conversion.
func calculateCoverageScore(grade, thresholdGrade string) float64 {
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