package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// FactionHandlers handles faction-related MCP tools
type FactionHandlers struct {
	factionService *services.FactionService
}

// NewFactionHandlers creates new faction handlers
func NewFactionHandlers(factionService *services.FactionService) *FactionHandlers {
	return &FactionHandlers{factionService: factionService}
}

// HandleUpdateFactionReputation handles the update_faction_reputation tool
func (h *FactionHandlers) HandleUpdateFactionReputation() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		factionID := getStringArg(args, "faction_id")
		partyID := getStringArg(args, "party_id")
		reason := getStringArg(args, "reason")
		delta := int8(getIntArg(args, "delta"))

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}
		if factionID == "" {
			return mcp.NewToolResultError("faction_id is required"), nil
		}
		if partyID == "" {
			return mcp.NewToolResultError("party_id is required"), nil
		}

		result, err := h.factionService.UpdateReputation(ctx, campaignID, factionID, partyID, delta, reason, "manual")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
