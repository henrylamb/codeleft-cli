package report

import "codeleft-cli/filter"

type IReport interface {
	GenerateReport(gradeDetails []filter.GradeDetails, threshold string) error
}

type HtmlReport struct {
	ReportType string
	RepoConstructor IRepoConstructor
}

func NewHtmlReport(repoConstructor IRepoConstructor) IReport {
	return &HtmlReport{
		ReportType: "HTML",
		RepoConstructor: repoConstructor,
	}
}


func (h *HtmlReport) GenerateReport(gradeDetails []filter.GradeDetails, threshold string) error {
	repoStructure := h.RepoConstructor.ConstructRepoStructure(gradeDetails)

	return GenerateRepoHTMLReport(repoStructure, "CodeLeft-Coverage-Report.html", threshold)
}