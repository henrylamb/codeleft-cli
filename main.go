package main


import (
	"codeleft-cli/assessment"
	"codeleft-cli/filter"
	"codeleft-cli/read"
	"codeleft-cli/report"
	"flag"
	"fmt"
	"os"
	"strings"
)

// Version of the CLI tool
const Version = "1.0.18"

// main is the entry point for your CLI tool.
func main() {
	thresholdGrade := flag.String("threshold-grade", "", "Sets the grade threshold.")
	thresholdPercent := flag.Int("threshold-percent", 0, "Sets the percentage threshold.")
	toolsFlag := flag.String("tools", "", "Comma-separated list of tooling (e.g., SOLID,OWASP-Top-10,Clean-Code,...)")
	versionFlag := flag.Bool("version", false, "Displays the current version of the CLI tool.")
	assessGrade := flag.Bool("asses-grade", false, "Assess the grade threshold.")
	assessCoverage := flag.Bool("asses-coverage", false, "Assess the coverage threshold.")
	createReport := flag.Bool("create-report", false, "Create a report of the assessment.")

	// Customize the usage message to include version information
	flag.Usage = func() {
		usageText := `codeleft-cli Version ` + Version + `

Usage:
  codeleft-cli [options]

Options:
`
		fmt.Fprintln(flag.CommandLine.Output(), usageText)
		flag.PrintDefaults()
	}

	// Parse command-line flags
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Fprintf(os.Stderr, "codeleft-cli Version %s\n", Version)
		os.Exit(0)
	}

	// Convert tools into a string slice
	if toolsFlag == nil {
		fmt.Fprintf(os.Stderr, "tools flag is nil")
		os.Exit(1)
	}
	toolsList := parseTools(*toolsFlag)

	// Initialize HistoryReader
	historyReader, err := read.NewHistoryReader()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing history reader: %v\n", err)
		os.Exit(1)
	}

	// Read history
	history, err := historyReader.ReadHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading history: %v\n", err)
		os.Exit(1)
	}

	// Apply filters and assessments
	latestGradeFilter := filter.NewLatestGrades()
	history = latestGradeFilter.FilterLatestGrades(history)

	toolFilter := filter.NewToolFilter(filter.NewToolCleaner())
	history = toolFilter.Filter(toolsList, history)

	//config filtering
	configReader, err := read.NewConfigReader(read.NewOSFileSystem())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config reader: %v\n", err)
		os.Exit(1)
	}
	config, err := configReader.ReadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	if config.Ignore.Folders != nil || config.Ignore.Files != nil {
		ignorefileRule := filter.NewIgnoreFileRule(config.Ignore.Files)
		ignoreFolderRule := filter.NewIgnoreFolderRule(config.Ignore.Folders)

		pathFilter := filter.NewPathFilter(ignorefileRule, ignoreFolderRule)
		history = pathFilter.Filter(history)
	}

	// Collect grades and assess
	violationCounter := assessment.NewConsoleViolationReporter()

	calculator := filter.NewGradeStringCalculator()
	coverageCalculator := filter.NewDefaultCoverageCalculator()
	gradeCollector := filter.NewGradeCollection(calculator, coverageCalculator)
	gradeDetails := gradeCollector.CollectGrades(history, *thresholdGrade)

	accessorGrade := assessment.NewCoverageAssessment(violationCounter)
	if *assessGrade && !accessorGrade.AssessCoverage(*thresholdPercent, gradeDetails) {
		fmt.Fprintf(os.Stderr, "Grade threshold failed :( \n")
		os.Exit(1)
	}

	accessorCoverage := assessment.NewCoverageAssessment(violationCounter)
	if *assessCoverage && !accessorCoverage.AssessCoverage(*thresholdPercent, gradeDetails) {
		fmt.Fprintf(os.Stderr, "Coverage threshold failed :( \n")
		os.Exit(1)
	}

	if *createReport {
		reporter := report.NewHtmlReport()
		if err := reporter.GenerateReport(gradeDetails, *thresholdGrade); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
			os.Exit(1)
		}		
		fmt.Fprintf(os.Stderr, "Report generated successfully!\n")
	}

	fmt.Fprintf(os.Stderr, "All checks passed!\n")
	os.Exit(0)
}

// parseTools splits the comma-separated tools flag into a slice of strings.
func parseTools(toolsFlag string) []string {
	if toolsFlag == "" {
		return []string{}
	}
	// Split on comma and trim spaces
	tools := strings.Split(toolsFlag, ",")
	for i := range tools {
		tools[i] = strings.TrimSpace(tools[i])
	}

	return tools
}
