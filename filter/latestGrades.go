package filter

type FindLatestGrades interface {
	FilterLatestGrades(histories Histories) Histories
}

type LatestGrades struct{}

func NewLatestGrades() FindLatestGrades {
	return &LatestGrades{}
}

func (lg *LatestGrades) FilterLatestGrades(histories Histories) Histories {
	// Use a composite key: "FilePath|AssessingTool"
	latestHistory := make(map[string]History)

	for _, history := range histories {
		key := generateCompositeKey(history.FilePath, history.AssessingTool)

		if storedHistory, exists := latestHistory[key]; exists {
			if storedHistory.TimeStamp.Before(history.TimeStamp) {
				latestHistory[key] = history
			}
		} else {
			latestHistory[key] = history
		}
	}

	return ConvertMapToSlice(latestHistory)
}

// generateCompositeKey creates a unique key based on FilePath and AssessingTool
func generateCompositeKey(filePath, assessingTool string) string {
	return filePath + "|" + assessingTool
}

// ConvertMapToSlice converts a map of History to a slice of History.
func ConvertMapToSlice(historyMap map[string]History) Histories {
	historySlice := make(Histories, 0, len(historyMap))
	for _, history := range historyMap {
		historySlice = append(historySlice, history)
	}
	return historySlice
}
