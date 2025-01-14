package filter

import (
	"time"
)

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

func (h Histories) Len() int {
	return len(h)
}

func (h Histories) Less(i, j int) bool {
	return h[i].TimeStamp.Before(h[j].TimeStamp)
}

func (h Histories) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
