package filter

type GradeDetail interface {
	GetCoverage(int)
}

type GradeDetails struct {
	Grade    string
	Score    int
	Coverage int
	FileName string
	Tool     string
}

func NewGradeDetails(grade string, score int, fileName string, tool string) GradeDetails {
	return GradeDetails{
		Grade:    grade,
		Score:    score,
		FileName: fileName,
		Tool:     tool,
	}
}

// GetCoverage calculates the coverage percentage based on the score and threshold
func (g *GradeDetails) GetCoverage(thresholdAsNum int) {
	if g.Score > thresholdAsNum {
		g.Coverage = 120
	} else if g.Score == thresholdAsNum {
		g.Coverage = 100
	} else if g.Score >= thresholdAsNum-1 {
		g.Coverage = 70
	} else if g.Score >= thresholdAsNum-2 {
		g.Coverage = 50
	} else if g.Score >= thresholdAsNum-3 {
		g.Coverage = 30
	} else {
		g.Coverage = 10
	}
}
