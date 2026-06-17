package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// ChapterPartHandlers handles sequential chapter generation MCP tools
type ChapterPartHandlers struct {
	service *services.CampaignService
}

// NewChapterPartHandlers creates new chapter part handlers
func NewChapterPartHandlers(service *services.CampaignService) *ChapterPartHandlers {
	return &ChapterPartHandlers{service: service}
}

// HandleSaveChapterPart handles the save_chapter_part tool
func (h *ChapterPartHandlers) HandleSaveChapterPart() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		chapterNum := getIntArg(args, "chapter_number")
		partName := getStringArg(args, "part_name")
		content := getStringArg(args, "content")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		if chapterNum < 0 {
			return mcp.NewToolResultError("chapter_number must be >= 0 (0 for prologue)"), nil
		}
		if partName == "" {
			return mcp.NewToolResultError("part_name is required"), nil
		}
		if content == "" {
			return mcp.NewToolResultError("content is required"), nil
		}

		result, err := h.service.SaveChapterPart(campaign, chapterNum, partName, content)
		if err != nil {
			return ToToolResult(err), nil
		}

		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(data)), nil
	}
}

// HandleFinalizeChapter handles the finalize_chapter tool
func (h *ChapterPartHandlers) HandleFinalizeChapter() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		chapterNum := getIntArg(args, "chapter_number")
		title := getStringArg(args, "title")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		if chapterNum < 0 {
			return mcp.NewToolResultError("chapter_number must be >= 0 (0 for prologue)"), nil
		}
		if title == "" {
			return mcp.NewToolResultError("title is required"), nil
		}

		result, err := h.service.FinalizeChapter(campaign, chapterNum, title)
		if err != nil {
			return ToToolResult(err), nil
		}

		data, _ := json.Marshal(result)
		return mcp.NewToolResultText(fmt.Sprintf("Chapter %d '%s' finalized: %s", chapterNum, title, string(data))), nil
	}
}
