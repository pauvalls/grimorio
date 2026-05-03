package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// QuestHandlers handles quest-related MCP tools
type QuestHandlers struct {
	service *services.QuestService
}

// NewQuestHandlers creates new quest handlers
func NewQuestHandlers(service *services.QuestService) *QuestHandlers {
	return &QuestHandlers{service: service}
}

// HandleCreateQuest handles quest creation
func (h *QuestHandlers) HandleCreateQuest() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		characterName := getStringArg(args, "character_name")
		title := getStringArg(args, "quest_title")
		questType := getStringArg(args, "quest_type")
		hook := getStringArg(args, "hook")
		stakes := getStringArg(args, "stakes")
		reward := getStringArg(args, "reward")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		if title == "" {
			return mcp.NewToolResultError("quest_title is required"), nil
		}

		var qType domain.QuestType
		if characterName != "" {
			qType = domain.QuestTypePersonal
		} else {
			qType = domain.QuestTypeGroup
		}
		if questType != "" {
			qType = domain.QuestType(questType)
		}

		quest, err := h.service.CreateQuest(campaign, title, qType, hook, "", stakes, nil)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := fmt.Sprintf("Quest '%s' created in campaign '%s' (ID: %s, Type: %s)",
			quest.Title, campaign, quest.ID, quest.Type)
		if reward != "" {
			result += fmt.Sprintf("\nReward: %s", reward)
		}

		return mcp.NewToolResultText(result), nil
	}
}

// HandleUpdateQuestStatus handles quest status updates
func (h *QuestHandlers) HandleUpdateQuestStatus() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		questID := getStringArg(args, "quest_id")
		status := getStringArg(args, "status")
		notes := getStringArg(args, "notes")

		if campaign == "" || questID == "" || status == "" {
			return mcp.NewToolResultError("campaign, quest_id, and status are required"), nil
		}

		questStatus := domain.QuestStatus(status)
		if err := h.service.UpdateQuestStatus(campaign, questID, questStatus, notes); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Quest %s updated to '%s'", questID, status)), nil
	}
}

// HandleListQuests handles listing quests
func (h *QuestHandlers) HandleListQuests() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		quests, err := h.service.ListQuests(campaign)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(quests) == 0 {
			return mcp.NewToolResultText("No quests found in campaign '" + campaign + "'"), nil
		}

		var result string
		for _, q := range quests {
			charInfo := ""
			if q.CharacterID != nil {
				charInfo = fmt.Sprintf(" (Character: %s)", *q.CharacterID)
			}
			result += fmt.Sprintf("- [%s] %s%s - %s\n", q.Status, q.Title, charInfo, q.Type)
		}
		return mcp.NewToolResultText(result), nil
	}
}
