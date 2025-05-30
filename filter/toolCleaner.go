package filter

import "strings"

type IToolCleaner interface {
	Clean(value string) string
}

type ToolCleaner struct{}

func NewToolCleaner() IToolCleaner {
	return &ToolCleaner{}
}

// Clean removes leading and trailing spaces from the input string.
// It is used to ensure that tool names are consistently formatted without extra spaces.
func (t *ToolCleaner) Clean(value string) string {
	value = strings.TrimPrefix(value, " ")
	value = strings.TrimSuffix(value, " ")
	return value
}