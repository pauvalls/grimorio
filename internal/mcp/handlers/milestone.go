package handlers

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// MilestoneHandlers handles milestone XP MCP tools.
type MilestoneHandlers struct {
	service *services.MilestoneService
}

// NewMilestoneHandlers creates milestone handlers.
func NewMilestoneHandlers(service *services.MilestoneService) *MilestoneHandlers {
	return &MilestoneHandlers{service: service}
}

// HandleGenerateXPTable handles grimorio_generate_xp_table.
func (h *MilestoneHandlers) HandleGenerateXPTable() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		chapterNum := getIntArg(args, "chapter_number")
		levelMin := getIntArg(args, "level_min")
		levelMax := getIntArg(args, "level_max")

		if campaignID == "" || chapterNum <= 0 {
			return mcp.NewToolResultError("campaign_id and chapter_number required"), nil
		}

		// Defaults
		if levelMin <= 0 {
			levelMin = 1
		}
		if levelMax <= 0 {
			levelMax = 5
		}

		levelRange := domain.LevelRange{Min: levelMin, Max: levelMax}
		chapterID := getStringArg(args, "chapter_id")
		if chapterID == "" {
			chapterID = "chapter_" + string(rune(chapterNum+'0'))
		}
		
		table, err := h.service.GenerateChapterTable(ctx, chapterID, "Chapter "+string(rune(chapterNum+'0')), levelRange)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, _ := json.MarshalIndent(table, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleTrackPartyProgress handles grimorio_track_party_progress.
func (h *MilestoneHandlers) HandleTrackPartyProgress() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		partyID := getStringArg(args, "party_id")

		if campaignID == "" || partyID == "" {
			return mcp.NewToolResultError("campaign_id and party_id required"), nil
		}

		level, err := h.service.CalculatePartyLevel(ctx, campaignID, partyID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]any{
			"campaign_id": campaignID,
			"party_id":    partyID,
			"level":       level,
		}
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
