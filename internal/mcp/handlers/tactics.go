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

// TacticsHandlers handles tactics MCP tools.
type TacticsHandlers struct {
	service        *services.TacticsService
	encounterRepo  services.EncounterRepository
	areaRepo       services.AreaRepository
}

// NewTacticsHandlers creates tactics handlers.
func NewTacticsHandlers(
	service *services.TacticsService,
	encounterRepo services.EncounterRepository,
	areaRepo services.AreaRepository,
) *TacticsHandlers {
	return &TacticsHandlers{
		service:       service,
		encounterRepo: encounterRepo,
		areaRepo:      areaRepo,
	}
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

		// Load encounter and area from repositories
		encounter, err := h.encounterRepo.Read(ctx, campaignID, encounterID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to load encounter: %v", err)), nil
		}

		var area *domain.Area
		if areaID != "" {
			area, err = h.areaRepo.Read(ctx, campaignID, areaID)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("failed to load area: %v", err)), nil
			}
		}

		tacticsList, err := h.service.GenerateEncounterTactics(ctx, encounter, area)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to generate tactics: %v", err)), nil
		}

		result := map[string]any{
			"campaign_id":   campaignID,
			"encounter_id":  encounterID,
			"area_id":       areaID,
			"tactics_count": len(tacticsList),
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

		tacticsList, err := h.service.GetTacticsByEncounter(ctx, campaignID, encounterID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to get tactics: %v", err)), nil
		}

		result := map[string]any{
			"campaign_id":  campaignID,
			"encounter_id": encounterID,
			"tactics":      tacticsList,
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
