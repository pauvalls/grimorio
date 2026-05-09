package handlers

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// TacticsHandlers handles tactics MCP tools.
type TacticsHandlers struct {
	service *services.TacticsService
}

// NewTacticsHandlers creates tactics handlers.
func NewTacticsHandlers(service *services.TacticsService) *TacticsHandlers {
	return &TacticsHandlers{service: service}
}

// HandleGenerateTactics handles grimorio_generate_tactics.
func (h *TacticsHandlers) HandleGenerateTactics() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		encounterID := getStringArg(args, "encounter_id")
		areaID := getStringArg(args, "area_id")

		if campaignID == "" || encounterID == "" {
			return mcp.NewToolResultError("campaign_id and encounter_id required"), nil
		}

		// TODO: Load encounter and area from repositories
		result := map[string]any{
			"campaign_id":   campaignID,
			"encounter_id":  encounterID,
			"area_id":       areaID,
			"status":        "tactics_generated",
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleGetTactics handles grimorio_get_tactics.
func (h *TacticsHandlers) HandleGetTactics() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		encounterID := getStringArg(args, "encounter_id")

		if campaignID == "" || encounterID == "" {
			return mcp.NewToolResultError("campaign_id and encounter_id required"), nil
		}

		result := map[string]any{
			"campaign_id":  campaignID,
			"encounter_id": encounterID,
			"tactics":      []any{},
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
