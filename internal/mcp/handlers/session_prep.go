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

// SessionPrepHandlers handles session prep MCP tools
type SessionPrepHandlers struct {
	sessionPrepService *services.SessionPrepService
}

// NewSessionPrepHandlers creates new session prep handlers
func NewSessionPrepHandlers(sessionPrepService *services.SessionPrepService) *SessionPrepHandlers {
	return &SessionPrepHandlers{sessionPrepService: sessionPrepService}
}

// HandleGenerateSessionPrep handles the generate_session_prep tool
func (h *SessionPrepHandlers) HandleGenerateSessionPrep() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		sessionNum := getIntArg(args, "session_num")
		withScenarios := getBoolArg(args, "with_scenarios")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		var prep *domain.SessionPrep
		var warnings []string
		var err error

		if withScenarios {
			prep, warnings, err = h.sessionPrepService.GetPrepWithScenarios(ctx, campaignID, sessionNum)
		} else {
			prep, warnings, err = h.sessionPrepService.GetPrep(ctx, campaignID, sessionNum)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := struct {
			Prep      any      `json:"prep"`
			Warnings  []string `json:"warnings"`
		}{
			Prep:     prep,
			Warnings: warnings,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
