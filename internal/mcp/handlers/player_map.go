package handlers

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// PlayerMapHandlers handles player map MCP tools.
type PlayerMapHandlers struct {
	service *services.PlayerMapService
}

// NewPlayerMapHandlers creates player map handlers.
func NewPlayerMapHandlers(service *services.PlayerMapService) *PlayerMapHandlers {
	return &PlayerMapHandlers{service: service}
}

// HandleGeneratePlayerMap handles grimorio_generate_player_map.
func (h *PlayerMapHandlers) HandleGeneratePlayerMap() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		dmMapID := getStringArg(args, "dm_map_id")
		areaID := getStringArg(args, "area_id")

		if campaignID == "" || dmMapID == "" {
			return mcp.NewToolResultError("campaign_id and dm_map_id required"), nil
		}

		playerMap, err := h.service.GeneratePlayerVariant(ctx, campaignID, dmMapID, areaID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, _ := json.MarshalIndent(playerMap, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
