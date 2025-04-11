package assessment

import (
	"codeleft-cli/filter"
)

// GradeAssessable interface for assessing code grades
type GradeAssessable interface {
	AssessGrade(threshold string, details []filter.GradeDetails) bool
}

// GradeAssessment handles grade assessment
type GradeAssessment struct {
	Calculator       filter.GradeCalculator
	Reporter         ViolationReporter
	ViolationDetails []filter.GradeDetails
}

// NewGradeAssessment creates a new GradeAssessment instance
func NewGradeAssessment(calculator filter.GradeCalculator, reporter ViolationReporter) GradeAssessable {
	return &GradeAssessment{
		Calculator: calculator,
		Reporter:   reporter,
	}
}

// AssessGrade assesses code grades against a threshold
func (ga *GradeAssessment) AssessGrade(threshold string, details []filter.GradeDetails) bool {
	passed := true
	ga.ViolationDetails = []filter.GradeDetails{} // Reset violations
	for _, detail := range details {
		if ga.Calculator.GradeNumericalValue(detail.Grade) < ga.Calculator.GradeNumericalValue(threshold) {
			passed = false
			ga.ViolationDetails = append(ga.ViolationDetails, detail)
		}
	}
	if !passed {
		ga.Reporter.Report(ga.ViolationDetails)
	}
	return passed
}