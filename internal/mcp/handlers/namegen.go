package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/namegen"
)

// NamegenHandlers handles name generation MCP tools.
type NamegenHandlers struct{}

// NewNamegenHandlers creates new name generation handlers.
func NewNamegenHandlers() *NamegenHandlers {
	return &NamegenHandlers{}
}

// HandleGenerateNames handles the generate_names tool.
func (h *NamegenHandlers) HandleGenerateNames() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		categoryStr := getStringArg(args, "category")
		styleStr := getStringArg(args, "style")
		count := getIntArg(args, "count")
		seed := getInt64Arg(args, "seed")

		if categoryStr == "" {
			return mcp.NewToolResultError("category is required"), nil
		}
		if styleStr == "" {
			styleStr = string(namegen.StyleGenericFantasy)
		}
		if count < 1 || count > 50 {
			return mcp.NewToolResultError("count must be between 1 and 50"), nil
		}

		var g *namegen.NameGenerator
		if seed != 0 {
			g = namegen.NewWithSeed(seed)
		} else {
			g = namegen.New()
		}

		names, err := g.Generate(namegen.Category(categoryStr), namegen.Style(styleStr), count)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("name generation failed: %v", err)), nil
		}

		jsonBytes, err := json.MarshalIndent(names, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
