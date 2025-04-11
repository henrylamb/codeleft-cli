package report

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

type ReportWriter interface {
	Write(data ReportViewData, outputPath string) error
}

type HTMLReportWriter struct {
	template *template.Template
}

func NewHTMLReportWriter() (*HTMLReportWriter, error) {
	tmpl, err := template.New("repoReport").Funcs(templateFuncs).Parse(repoReportTemplateHTML)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}
	return &HTMLReportWriter{template: tmpl}, nil
}

func (w *HTMLReportWriter) Write(data ReportViewData, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}

	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create HTML output file '%s': %w", outputPath, err)
	}
	defer outputFile.Close()

	if err := w.template.Execute(outputFile, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}
	return nil
}