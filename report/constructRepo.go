package report

import (
	"codeleft-cli/filter"
	"strings"
)

//the aim of this is to construct the repo structure from all the file names
type IRepoConstructor interface {
	ConstructRepoStructure(gradeDetails []filter.GradeDetails) map[string]any
}

type RepoConstructor struct {}

func NewRepoConstructor() IRepoConstructor {
	return &RepoConstructor{}
}

func (r *RepoConstructor) ConstructRepoStructure(gradeDetails []filter.GradeDetails) map[string]any {
	repoStructure := make(map[string]any)

	for _, detail := range gradeDetails {
		filePath := detail.FileName
		fileName := detail.FileName
		directory := strings.Split(filePath, "/")

		if len(directory) > 1 {
			fileName = directory[len(directory)-1]
			directory = directory[:len(directory)-1]
		} else {
			directory = []string{}
		}

		currentDirectory := repoStructure

		for _, dir := range directory {
			if _, exists := currentDirectory[dir]; !exists {
				currentDirectory[dir] = make(map[string]any)
			}
			currentDirectory = currentDirectory[dir].(map[string]any)
		}

		currentDirectory[fileName] = detail
	}

	return repoStructure
}