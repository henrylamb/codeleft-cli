package filter

import (
	"strings"
)

type FilterTools interface {
	Filter(values []string, histories Histories) Histories
}

type ToolFilter struct{
	toolCleaner IToolCleaner
}

func NewToolFilter(toolCleaner IToolCleaner) FilterTools {
	return &ToolFilter{
		toolCleaner: toolCleaner,
	}
}

func (t *ToolFilter) Filter(values []string, histories Histories) Histories {
	filteredHistories := Histories{}
	for _, value := range values {
		value = t.toolCleaner.Clean(value)

		toolFilteredHistories := t.filterByTool(value, histories)
		filteredHistories = append(filteredHistories, toolFilteredHistories...)
	}
	return filteredHistories
}

func (t *ToolFilter) filterByTool(tool string, histories Histories) Histories {
	filteredHistories := Histories{}

	for _, history := range histories {
		if strings.ToUpper(history.AssessingTool) == strings.ToUpper(tool) {
			history.CodeReview = map[string]any{}
			history.GradingDetails = map[string]any{}

			filteredHistories = append(filteredHistories, history)
		}
	}

	return filteredHistories
}