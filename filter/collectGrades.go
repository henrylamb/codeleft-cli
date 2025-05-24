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
	return GetGradeIndex(grade)
}