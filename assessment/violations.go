package assessment

import (
	"codeleft-cli/filter"
	"fmt"
	"os"
)

type Violations interface {
	Print()
	AddViolation(detail filter.GradeDetails)
}

type Violation struct {
	ListViolations []filter.GradeDetails
}

func NewViolation() Violations {
	return &Violation{}
}

func (v *Violation) AddViolation(detail filter.GradeDetails) {
	v.ListViolations = append(v.ListViolations, detail)
}

func (v *Violation) Print() {
	for _, detail := range v.ListViolations {
		_, err := fmt.Fprintf(os.Stderr, "File: %s Grade: %s Coverage (Percent): %d \n", detail.FileName, detail.Grade, detail.Coverage)
		if err != nil {
			fmt.Println("Error printing violation")
			return
		}
	}
}
