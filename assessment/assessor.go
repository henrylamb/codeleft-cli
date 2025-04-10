package assessment

import (
	"codeleft-cli/filter"
	"fmt"
	"os"
)

type Accessor interface {
	Assess(threshold string, details []filter.GradeDetails) bool
}

type AccessorGrade struct {
	Calculator       filter.GradeCalculator
	ViolationCounter Violations
}

func NewAccessorGrade(calculator filter.GradeCalculator, violationCounter Violations) Accessor {
	return &AccessorGrade{
		Calculator:       calculator,
		ViolationCounter: violationCounter,
	}
}

func (ag *AccessorGrade) Assess(threshold string, details []filter.GradeDetails) bool {
	failed := true
	for _, detail := range details {
		if ag.Calculator.GradeNumericalValue(detail.Grade) < ag.Calculator.GradeNumericalValue(threshold) {
			failed = false
			ag.ViolationCounter.AddViolation(detail)
		}
	}
	if !failed {
		ag.ViolationCounter.Print()
	}
	return failed
}

type AccessorCoverage interface {
	Assess(thresholdPercent int, details []filter.GradeDetails) bool
}

type AccessorAverageCoverage struct {
	ViolationCounter Violations
}

func NewAccessorAverageCoverage(violationCounter Violations) AccessorCoverage {
	return &AccessorAverageCoverage{
		ViolationCounter: violationCounter,
	}
}

func (aac *AccessorAverageCoverage) Assess(thresholdPercent int, details []filter.GradeDetails) bool {
	total := 0
	for _, detail := range details {
		total += detail.Coverage
		if detail.Coverage < thresholdPercent {
			aac.ViolationCounter.AddViolation(detail)
		}
	}
	if (len(details)) == 0 {
		fmt.Println("No files to assess")
		return false
	}

	average := float32(total) / float32(len(details))
	pass := average >= float32(thresholdPercent)

	if !pass {
		aac.ViolationCounter.Print()
	}
	fmt.Fprintf(os.Stderr, "Average coverage: %.2f%%\n", average)
	return pass
}
