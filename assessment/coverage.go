package assessment

import (
	"codeleft-cli/filter"
	"fmt"
	"os"
)

// CoverageAssessable interface for assessing code coverage
type CoverageAssessable interface {
	AssessCoverage(thresholdPercent int, details []filter.GradeDetails) bool
}

// CoverageAssessment handles code coverage assessment
type CoverageAssessment struct {
	Reporter         ViolationReporter
	ViolationDetails []filter.GradeDetails
}

// NewCoverageAssessment creates a new CoverageAssessment instance
func NewCoverageAssessment(reporter ViolationReporter) CoverageAssessable {
	return &CoverageAssessment{
		Reporter: reporter,
	}
}

// AssessCoverage assesses code coverage against a threshold
func (ca *CoverageAssessment) AssessCoverage(thresholdPercent int, details []filter.GradeDetails) bool {
	total := 0
	ca.ViolationDetails = []filter.GradeDetails{} // Reset violations
	for _, detail := range details {
		total += detail.Coverage
		if detail.Coverage < thresholdPercent {
			ca.ViolationDetails = append(ca.ViolationDetails, detail)
		}
	}
	if len(details) == 0 {
		fmt.Println("No files to assess")
		return false
	}

	average := float32(total) / float32(len(details))
	pass := average >= float32(thresholdPercent)

	if !pass {
		ca.Reporter.Report(ca.ViolationDetails)
	}
	fmt.Fprintf(os.Stderr, "Average coverage: %.2f%%\n", average)
	return pass
}