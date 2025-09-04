package filter

func CalculateCoverageScore(grade, thresholdGrade string) float64 {
    // These calls will now use the modified getGradeIndex function
    gradeIndex := GetGradeIndex(grade)
    thresholdIndex := GetGradeIndex(thresholdGrade)

    // Logic must precisely match the Javascript implementation using the new indices
    switch thresholdIndex - gradeIndex {
    case 0:
        return 100.0
    case 1:
        return 90.0
    case 2:
        return 80.0
    case 3:
        return 70.0
    case 4:
        return 50.0
    case 5:
        return 30.0
    default:
        if gradeIndex > thresholdIndex {
            return 120.0
        }
        return 10.0
    }
}