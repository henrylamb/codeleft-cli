package report

import (
	"codeleft-cli/filter" // Assuming this path is correct
	"fmt"
	"log"
)

// GenerateReport orchestrates the report generation process.
// It depends on abstractions (ReportWriter) and coordinates different components.
func GenerateReport(gradeDetails []filter.GradeDetails, outputPath string, thresholdGrade string, writer ReportWriter) error {
	if len(gradeDetails) == 0 {
		log.Println("Warning: No grade details provided to generate report.")
		// Handle appropriately - maybe write an empty report or return specific error
		// Creating empty data for the writer
		return writer.Write(ReportViewData{ ThresholdGrade: thresholdGrade }, outputPath)
	}

	// 1. Build the tree structure
	builder := NewTreeBuilder()
	groupedDetails := builder.GroupGradeDetailsByPath(gradeDetails)
	rootNodes := builder.BuildReportTree(groupedDetails)

	// 2. Calculate coverages and aggregate stats
	calculator := NewCoverageCalculator(thresholdGrade)
	stats := NewGlobalStats()
	for _, node := range rootNodes {
		calculator.CalculateNodeCoverages(node, stats) // Modifies nodes and stats
	}

	// 3. Sort the tree (optional, could be done after building or before writing)
	sortReportNodes(rootNodes)

    // 4. Calculate final overall averages
    overallAverages, totalAverage, allTools := calculator.CalculateOverallAverages(stats)

	// 5. Prepare data for the view
	viewData := ReportViewData{
		RootNodes:       rootNodes,
		AllTools:        allTools, // Already sorted by calculator
		OverallAverages: overallAverages,
		TotalAverage:    totalAverage,
		ThresholdGrade:  thresholdGrade,
	}

	// 6. Write the report using the injected writer
	err := writer.Write(viewData, outputPath)
	if err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	fmt.Printf("Successfully generated repository report: %s\n", outputPath)
	return nil
}