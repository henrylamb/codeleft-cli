package filter

import "time"

type GradeDetail interface {
	GetCoverage(int)
}

type GradeDetails struct {
	Grade    string `json:"grade"`
	Score    int `json:"score"`
	Coverage int `json:"coverage"`
	FileName string `json:"fileName"`
	Tool     string `json:"tool"`
	Timestamp time.Time `json:"timestamp"`
}

func NewGradeDetails(grade string, score int, fileName string, tool string, timeStamp time.Time) GradeDetails {
	return GradeDetails{
		Grade:    grade,
		Score:    score,
		FileName: fileName,
		Tool:     tool,
		Timestamp: timeStamp,
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
