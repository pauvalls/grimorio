package handlers

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// AreaHandlers handles area MCP tools.
type AreaHandlers struct {
	service *services.AreaService
}

// NewAreaHandlers creates area handlers.
func NewAreaHandlers(service *services.AreaService) *AreaHandlers {
	return &AreaHandlers{service: service}
}

// HandleGenerateArea handles grimorio_generate_area.
func (h *AreaHandlers) HandleGenerateArea() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		chapterNum := getIntArg(args, "chapter_number")
		areaNum := getIntArg(args, "area_number")
		levelMin := getIntArg(args, "level_min")
		levelMax := getIntArg(args, "level_max")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id required"), nil
		}

		// Defaults
		if chapterNum <= 0 {
			chapterNum = 1
		}
		if areaNum <= 0 {
			areaNum = 1
		}
		if levelMin <= 0 {
			levelMin = 1
		}
		if levelMax <= 0 {
			levelMax = 5
		}

		levelRange := domain.LevelRange{Min: levelMin, Max: levelMax}
		area, err := h.service.GenerateArea(ctx, 
			"chapter_"+string(rune(chapterNum+'0')), 
			areaNum, 
			levelRange)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, _ := json.MarshalIndent(area, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleGenerateAreasChapter handles grimorio_generate_areas_chapter.
func (h *AreaHandlers) HandleGenerateAreasChapter() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		chapterNum := getIntArg(args, "chapter_number")
		count := getIntArg(args, "count")
		levelMin := getIntArg(args, "level_min")
		levelMax := getIntArg(args, "level_max")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id required"), nil
		}

		// Defaults
		if chapterNum <= 0 {
			chapterNum = 1
		}
		if count <= 0 {
			count = 12
		}
		if levelMin <= 0 {
			levelMin = 1
		}
		if levelMax <= 0 {
			levelMax = 5
		}

		levelRange := domain.LevelRange{Min: levelMin, Max: levelMax}
		areas, err := h.service.GenerateChapterAreas(ctx, 
			"chapter_"+string(rune(chapterNum+'0')), 
			levelRange, 
			count)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, _ := json.MarshalIndent(areas, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
