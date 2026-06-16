package handlers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/compiler"
	"github.com/pauvalls/grimorio/internal/domain"
)

// ExportHandlers handles campaign export MCP tools.
type ExportHandlers struct {
	exporters map[string]compiler.Exporter
	baseDir   string
}

// NewExportHandlers creates new export handlers with the given exporters.
func NewExportHandlers(baseDir, pdfEngine string) *ExportHandlers {
	exporters := map[string]compiler.Exporter{
		"pdf":      compiler.NewPDFExporter(pdfEngine),
		"markdown": compiler.NewMarkdownExporter(),
		"epub":     compiler.NewEPUBExporter(),
	}
	return &ExportHandlers{exporters: exporters, baseDir: baseDir}
}

// HandleExportCampaign handles the export_campaign tool.
func (h *ExportHandlers) HandleExportCampaign() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		format := getStringArg(args, "format")
		if format == "" {
			format = "pdf"
		}

		if !domain.IsValidExportFormat(format) {
			return mcp.NewToolResultError(fmt.Sprintf("unsupported export format: %s", format)), nil
		}

		exporter, ok := h.exporters[format]
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("exporter not available for format: %s", format)), nil
		}

		campaignDir := filepath.Join(h.baseDir, campaign)
		title := getStringArg(args, "title")

		path, err := exporter.Export(ctx, campaignDir, title)
		if err != nil {
			return ToToolResult(err), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Campaign exported to %s: %s", format, path)), nil
	}
}
