package filter

import (
	"time"
)

type GradeDetails struct {
	Grade    string
	Score    int
	Coverage int
	FileName string
}

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

type HistoriesWrapper struct {
	Histories Histories `json:"histories"`
}

type History struct {
	AssessingTool  string         `json:"assessingTool"`
	FilePath       string         `json:"filePath"`
	Grade          string         `json:"grade"`
	Username       string         `json:"username"`
	TimeStamp      time.Time      `json:"timeStamp"`
	CodeReview     map[string]any `json:"codeReview"`
	GradingDetails map[string]any `json:"gradingDetails"`
	Hash           string         `json:"hash"`
}

type Histories []History

// Implementing sort.Interface for Histories based on the TimeStamp field.
func (h Histories) Len() int {
	return len(h)
}

func (h Histories) Less(i, j int) bool {
	return h[i].TimeStamp.Before(h[j].TimeStamp)
}

func (h Histories) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
