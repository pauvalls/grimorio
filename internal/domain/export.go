package domain

// ExportFormat represents supported campaign export formats.
type ExportFormat string

const (
	ExportFormatPDF      ExportFormat = "pdf"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatEPUB     ExportFormat = "epub"
)

// IsValidExportFormat checks if the given string is a valid export format.
func IsValidExportFormat(s string) bool {
	switch ExportFormat(s) {
	case ExportFormatPDF, ExportFormatMarkdown, ExportFormatEPUB:
		return true
	}
	return false
}
