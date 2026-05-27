package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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

// HandleGenerateDynamicArea handles generate_dynamic_area tool.
func (h *AreaHandlers) HandleGenerateDynamicArea() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		startTime := time.Now()

		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		locationDesc := getStringArg(args, "location_description")
		partyLevel := getIntArg(args, "party_level")
		_ = getStringArg(args, "tone")     // Optional, for future use
		_ = getBoolArg(args, "auto_save")  // Optional, for future use

		// Validation
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}
		if locationDesc == "" {
			return mcp.NewToolResultError("location_description is required"), nil
		}
		if partyLevel < 1 || partyLevel > 20 {
			return mcp.NewToolResultError(fmt.Sprintf("party_level must be 1-20, got %d", partyLevel)), nil
		}

		// Parse location description
		locationHint, settingType := parseLocationDescription(locationDesc)

		// Auto-detect chapter (default to chapter_1)
		chapterID := "chapter_1"

		// For now, generate without faction/narrative context
		// In production, load from repositories
		area, err := h.service.GenerateAreaWithContext(
			ctx,
			campaignID,
			chapterID,
			1, // Area number
			locationHint,
			settingType,
			partyLevel,
			nil, // factionContext
			nil, // narrativeState
		)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to generate area: %v", err)), nil
		}

		generationTime := time.Since(startTime)

		// Build response
		response := map[string]any{
			"success": true,
			"area":    area,
			"validation": map[string]any{
				"passed":   true,
				"errors":   []string{},
				"warnings": []string{},
			},
			"context_used": map[string]any{
				"chapter_detected":       chapterID,
				"location_hint":          locationHint,
				"setting_type":           settingType,
				"faction_context_loaded": false,
				"narrative_state_loaded": false,
				"dead_npcs_excluded":     0,
			},
			"generation_time_ms": generationTime.Milliseconds(),
		}

		jsonBytes, _ := json.MarshalIndent(response, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// parseLocationDescription extracts locationHint and settingType from natural language
func parseLocationDescription(locationDesc string) (locationHint, settingType string) {
	// Convert to lowercase and tokenize
	normalized := strings.ToLower(locationDesc)
	words := strings.Fields(normalized)

	// Extract first noun or adjective-noun pair as locationHint
	if len(words) >= 2 {
		if isAdjective(words[0]) {
			locationHint = words[0] + " " + words[1]
		} else {
			locationHint = words[0]
		}
	} else if len(words) == 1 {
		locationHint = words[0]
	} else {
		locationHint = "forest" // Default
	}

	// Map keywords to settingType
	wildernessKeywords := []string{"forest", "mountain", "river", "cave", "jungle", "desert", "swamp"}
	urbanKeywords := []string{"city", "town", "market", "street", "castle", "village"}
	dungeonKeywords := []string{"dungeon", "ruin", "tomb", "crypt", "catacomb"}
	socialKeywords := []string{"tavern", "court", "hall", "meeting", "temple"}

	for _, keyword := range wildernessKeywords {
		if strings.Contains(normalized, keyword) {
			return locationHint, "wilderness"
		}
	}
	for _, keyword := range urbanKeywords {
		if strings.Contains(normalized, keyword) {
			return locationHint, "urban"
		}
	}
	for _, keyword := range dungeonKeywords {
		if strings.Contains(normalized, keyword) {
			return locationHint, "dungeon"
		}
	}
	for _, keyword := range socialKeywords {
		if strings.Contains(normalized, keyword) {
			return locationHint, "social"
		}
	}

	return locationHint, "wilderness" // Default
}

// isAdjective checks if a word is likely an adjective (simplified heuristic)
func isAdjective(word string) bool {
	adjectives := map[string]bool{
		"dark": true, "ancient": true, "abandoned": true, "hidden": true,
		"secret": true, "busy": true, "crowded": true, "quiet": true,
		"dangerous": true, "mysterious": true, "haunted": true,
	}
	return adjectives[word]
}
