package filter

func CalculateCoverageScore(grade, thresholdGrade string) float64 {
    // These calls will now use the modified getGradeIndex function
    gradeIndex := GetGradeIndex(grade)
    thresholdIndex := GetGradeIndex(thresholdGrade)

    // Logic must precisely match the Javascript implementation using the new indices
    if gradeIndex > thresholdIndex {
        return 120.0
    } else if gradeIndex == thresholdIndex {
        return 100.0
    } else if gradeIndex == thresholdIndex-1 { // Check for difference of 1
        return 90.0
    } else if gradeIndex == thresholdIndex-2 { // Check for difference of 2
        return 80.0
    } else if gradeIndex == thresholdIndex-3 { // Check for difference of 3
        return 70.0
    } else if gradeIndex == thresholdIndex-4 { // Check for difference of 4
        return 50.0
    } else if gradeIndex == thresholdIndex-5 { // Check for difference of 5
        return 30.0
    } else { // Covers gradeIndex < thresholdIndex - 5 and any other lower cases
        return 10.0
    }
}