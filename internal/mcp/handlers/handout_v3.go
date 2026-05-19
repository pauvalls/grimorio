package handlers

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// HandoutV3Handlers handles handout V3 MCP tools.
type HandoutV3Handlers struct {
	service *services.HandoutServiceV3
}

// NewHandoutV3Handlers creates handout V3 handlers.
func NewHandoutV3Handlers(service *services.HandoutServiceV3) *HandoutV3Handlers {
	return &HandoutV3Handlers{service: service}
}

// HandleExportHandout handles grimorio_export_handout.
func (h *HandoutV3Handlers) HandleExportHandout() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		handoutID := getStringArg(args, "handout_id")
		format := getStringArg(args, "format")
		if format == "" {
			format = "text"
		}

		if campaignID == "" || handoutID == "" {
			return mcp.NewToolResultError("campaign_id and handout_id required"), nil
		}

		content, err := h.service.ExportHandout(ctx, campaignID, handoutID, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(content)), nil
	}
}
