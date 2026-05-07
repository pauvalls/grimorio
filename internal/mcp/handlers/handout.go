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

// HandoutHandlers handles handout MCP tools
type HandoutHandlers struct {
	handoutService *services.HandoutService
}

// NewHandoutHandlers creates new handout handlers
func NewHandoutHandlers(handoutService *services.HandoutService) *HandoutHandlers {
	return &HandoutHandlers{handoutService: handoutService}
}

// HandleGenerateHandouts handles the generate_handouts tool
func (h *HandoutHandlers) HandleGenerateHandouts() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		handoutTypeStr := getStringArg(args, "handout_type")
		versionStr := getStringArg(args, "version")
		if versionStr == "" {
			versionStr = "both"
		}

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}
		if handoutTypeStr == "" {
			return mcp.NewToolResultError("handout_type is required"), nil
		}

		var contentRefs []string
		if refsVal, ok := args["content_refs"]; ok {
			if refs, ok := refsVal.([]any); ok {
				for _, r := range refs {
					if s, ok := r.(string); ok {
						contentRefs = append(contentRefs, s)
					}
				}
			}
		}

		handout, err := h.handoutService.GenerateHandout(ctx, campaignID, domain.HandoutType(handoutTypeStr), contentRefs, domain.HandoutVersion(versionStr))
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(handout, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
