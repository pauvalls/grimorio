package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// DMContextHandlers handles DM context MCP tools.
type DMContextHandlers struct {
	dmContextService *services.DMContextService
}

// NewDMContextHandlers creates new DM context handlers.
func NewDMContextHandlers(dmContextService *services.DMContextService) *DMContextHandlers {
	return &DMContextHandlers{dmContextService: dmContextService}
}

// HandleDMContext handles the dm_session_context tool.
func (h *DMContextHandlers) HandleDMContext() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		sessionNum := getIntArg(args, "session_num")
		includePrologue := getBoolArg(args, "include_prologue")
		includePDFText := getBoolArg(args, "include_pdf_text")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		if !domain.IsValidKebabCase(campaignID) {
			return mcp.NewToolResultError("campaign_id must be kebab-case"), nil
		}

		payload, warnings, err := h.dmContextService.GetContext(ctx, campaignID, sessionNum, includePrologue, includePDFText)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := struct {
			Payload   *domain.DMContextPayload `json:"payload"`
			Warnings  []string                 `json:"warnings"`
		}{
			Payload:  payload,
			Warnings: warnings,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
