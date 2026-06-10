package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// HealthHandlers handles campaign health dashboard MCP tools.
type HealthHandlers struct {
	healthScore *services.CampaignHealthScore
}

// NewHealthHandlers creates new health handlers.
func NewHealthHandlers(healthScore *services.CampaignHealthScore) *HealthHandlers {
	return &HealthHandlers{healthScore: healthScore}
}

// HandleCampaignHealthDashboard handles the campaign_health_dashboard tool.
func (h *HealthHandlers) HandleCampaignHealthDashboard() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		if h.healthScore == nil {
			return mcp.NewToolResultError("health score service not initialized"), nil
		}

		report, err := h.healthScore.Compute(ctx, campaignID)
		if err != nil {
			return ToToolResult(err), nil
		}

		jsonBytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal report: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
