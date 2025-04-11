package report

import "codeleft-cli/filter"

type IReport interface {
	GenerateReport(gradeDetails []filter.GradeDetails, threshold string) error
}

type HtmlReport struct {
	ReportType string
}

func NewHtmlReport() IReport {
	return &HtmlReport{
		ReportType: "HTML",
	}
}


func (h *HtmlReport) GenerateReport(gradeDetails []filter.GradeDetails, threshold string) error {
	return GenerateRepoHTMLReport(gradeDetails, "CodeLeft-Coverage-Report.html", threshold)
}