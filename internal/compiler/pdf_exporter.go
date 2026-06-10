package compiler

import (
	"context"
	"fmt"
)

// PDFExporter wraps the existing Compiler to produce PDF output.
type PDFExporter struct {
	pdfEngine string
}

// NewPDFExporter creates a new PDF exporter with the given engine.
func NewPDFExporter(pdfEngine string) *PDFExporter {
	return &PDFExporter{pdfEngine: pdfEngine}
}

// Format returns the exporter format identifier.
func (e *PDFExporter) Format() string {
	return "pdf"
}

// Export compiles the campaign directory into a PDF.
func (e *PDFExporter) Export(ctx context.Context, campaignDir, title string) (string, error) {
	comp := New(campaignDir, e.pdfEngine)
	path, err := comp.Compile(ctx, title)
	if err != nil {
		return "", fmt.Errorf("pdf compilation failed: %w", err)
	}
	return path, nil
}
