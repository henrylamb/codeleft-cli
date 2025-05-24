package filter

type CollectGrades interface {
	CollectGrades(histories Histories, threshold string) []GradeDetails
}

type GradeCollection struct {
	GradeCalculator GradeCalculator
}

func NewGradeCollection(calculator GradeCalculator) CollectGrades {
	return &GradeCollection{
		GradeCalculator: calculator,
	}
}

func (g *GradeCollection) CollectGrades(histories Histories, threshold string) []GradeDetails {
	gradeDetails := []GradeDetails{}
	for _, history := range histories {
		newDetails := NewGradeDetails(history.Grade, g.GradeCalculator.GradeNumericalValue(history.Grade), history.FilePath, history.AssessingTool, history.TimeStamp)
		newDetails.GetCoverage(g.GradeCalculator.GradeNumericalValue(threshold))
		gradeDetails = append(gradeDetails, newDetails)

	}
	return gradeDetails
}

type GradeCalculator interface {
	GradeNumericalValue(grade string) int
}

type GradeStringCalculator struct{}

func NewGradeStringCalculator() GradeCalculator {
	return &GradeStringCalculator{}
}

func (g *GradeStringCalculator) GradeNumericalValue(grade string) int {
	switch grade {
	case "A*":
		return 90
	case "A+":
		return 90 // Minimum score for A+ in the new system
	case "A":
		return 85 // Minimum score for A
	case "A-":
		return 80 // Minimum score for A-
	case "B+":
		return 75 // Minimum score for B+
	case "B":
		return 70 // Minimum score for B
	case "B-":
		return 65 // Minimum score for B-
	case "C+":
		return 60 // Minimum score for C+
	case "C":
		return 55 // Minimum score for C
	case "C-":
		return 50 // Minimum score for C-
	case "D+":
		return 45 // Minimum score for D+
	case "D":
		return 40 // Minimum score for D
	case "D-":
		return 30 // Minimum score for D- (Lowest passing grade)
	case "F":
		// Although F is for scores below 30, the request is to return 30
		// to maintain similarity with the previous system's representative F score.
		return 20
	default:
		// For any unrecognized grade, return 30, aligning with the F case
		// and the previous system's representative F score location.
		return 20
	}
}