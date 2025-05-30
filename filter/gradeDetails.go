package filter

import "time"

// ICoverageCalculator defines the interface for calculating coverage based on score and threshold.
type ICoverageCalculator interface {
	CalculateCoverage(score int, threshold int) int
}

// DefaultCoverageCalculator implements ICoverageCalculator using a rule-based approach.
type DefaultCoverageCalculator struct{}

func NewDefaultCoverageCalculator() ICoverageCalculator {
	return &DefaultCoverageCalculator{}
}

// coverageRules defines the mapping from score difference to coverage percentage.
// The rules are ordered from highest score (smallest or negative offset) to lowest
// to ensure correct application of the first matching rule.
var coverageRules = []struct {
	MinScoreOffset int // score >= threshold + MinScoreOffset
	Coverage       int
}{
	{MinScoreOffset: 1, Coverage: 120},
	{MinScoreOffset: 0, Coverage: 100}, 
	{MinScoreOffset: -1, Coverage: 90},
	{MinScoreOffset: -2, Coverage: 80},
	{MinScoreOffset: -3, Coverage: 70}, 
	{MinScoreOffset: -4, Coverage: 50},
	{MinScoreOffset: -5, Coverage: 30}, 
}

// CalculateCoverage determines the coverage percentage based on the given score and threshold.
// It iterates through predefined rules to find the appropriate coverage.
func (c *DefaultCoverageCalculator) CalculateCoverage(score int, threshold int) int {
	for _, rule := range coverageRules {
		if score >= threshold+rule.MinScoreOffset {
			return rule.Coverage
		}
	}
	return 10 // Default coverage for scores significantly below threshold (score < threshold - 5)
}

// GradeDetails holds information about a grade, including its calculated coverage.
type GradeDetails struct {
	Grade      string `json:"grade"`
	Score      int    `json:"score"`
	Coverage   int    `json:"coverage"`
	FileName   string `json:"fileName"`
	Tool       string `json:"tool"`
	Timestamp  time.Time `json:"timestamp"`
	calculator ICoverageCalculator // Injected dependency for coverage calculation
}

// NewGradeDetails creates a new GradeDetails instance.
func NewGradeDetails(grade string, score int, fileName string, tool string, timeStamp time.Time, calculator ICoverageCalculator) GradeDetails {
	return GradeDetails{
		Grade:      grade,
		Score:      score,
		FileName:   fileName,
		Tool:       tool,
		Timestamp:  timeStamp,
		calculator: calculator,
	}
}

// UpdateCoverage calculates and sets the Coverage field of GradeDetails using the injected calculator.
func (g *GradeDetails) UpdateCoverage(thresholdAsNum int) {
	g.Coverage = g.calculator.CalculateCoverage(g.Score, thresholdAsNum)
}