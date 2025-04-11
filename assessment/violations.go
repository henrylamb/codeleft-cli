package assessment

import (
	"codeleft-cli/filter"
	"fmt"
)

// ViolationReporter interface for reporting violations
type ViolationReporter interface {
	Report(violations []filter.GradeDetails)
}

// ConsoleViolationReporter implements ViolationReporter and prints violations to the console
type ConsoleViolationReporter struct{}

func NewConsoleViolationReporter() ViolationReporter {
	return &ConsoleViolationReporter{}
}

func (c *ConsoleViolationReporter) Report(violations []filter.GradeDetails) {
	for _, v := range violations {
		fmt.Printf("Violation: File: %s, Grade: %s, Coverage: %d\n", v.FileName, v.Grade, v.Coverage)
	}
}
