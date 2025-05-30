package filter

type CollectGrades interface {
	CollectGrades(histories Histories, threshold string) []GradeDetails
}

type GradeCollection struct {
	GradeCalculator GradeCalculator
	CoverageCalculator ICoverageCalculator
}

func NewGradeCollection(calculator GradeCalculator, coverageCalculator ICoverageCalculator) CollectGrades {
	return &GradeCollection{
		GradeCalculator: calculator,
		CoverageCalculator: coverageCalculator,
	}
}

func (g *GradeCollection) CollectGrades(histories Histories, threshold string) []GradeDetails {
	gradeDetails := []GradeDetails{}
	for _, history := range histories {
		newDetails := NewGradeDetails(history.Grade, g.GradeCalculator.GradeNumericalValue(history.Grade), history.FilePath, history.AssessingTool, history.TimeStamp, g.CoverageCalculator)
		newDetails.UpdateCoverage(g.GradeCalculator.GradeNumericalValue(threshold))

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