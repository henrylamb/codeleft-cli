package report

import (
	"log"
	"strings"
)

func getGradeIndex(grade string) int {
    // Use the same index values as the Javascript implementation
    gradeIndices := map[string]int{
        "A*": 11, "A+": 12, "A": 11, "A-": 10,
        "B+": 9,  "B": 8,  "B-": 7,
        "C+": 6,  "C": 5,  "C-": 4,
        "D+": 3,  "D": 2,  "D-": 1,
        "F":  0, // F is 0
    }
    // Ensure comparison is case-insensitive
    index, ok := gradeIndices[strings.ToUpper(grade)]
    if !ok {
        log.Printf("Warning: Unrecognized grade '%s', treating as F (0)", grade)
        return 0 // Default to 0 for unrecognized grades, matching JS behavior
    }
    return index
}