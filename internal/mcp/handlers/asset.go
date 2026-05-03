package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// AssetHandlers handles asset generation MCP tools
type AssetHandlers struct {
	service *services.AssetService
}

// NewAssetHandlers creates new asset handlers
func NewAssetHandlers(service *services.AssetService) *AssetHandlers {
	return &AssetHandlers{service: service}
}

// HandleGenerateMap handles map generation
func (h *AssetHandlers) HandleGenerateMap() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		filename := getStringArg(args, "filename")
		style := getStringArg(args, "style")
		title := getStringArg(args, "title")
		labels := getStringArg(args, "labels")
		rooms := getIntArg(args, "rooms")
		if rooms <= 0 {
			rooms = 6
		}

		if campaign == "" || filename == "" {
			return mcp.NewToolResultError("campaign and filename are required"), nil
		}

		var labelList []string
		if labels != "" {
			labelList = splitLabels(labels)
		}

		svgContent, err := h.service.GenerateMap(campaign, filename, style, title, rooms, labelList)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Write to filesystem (this maintains backward compatibility)
		assetsDir := filepath.Join(filepath.Join(".", campaign), "assets") // This needs baseDir from config
		// For now, we need to handle this differently - let's pass baseDir through
		_ = assetsDir
		_ = svgContent

		return mcp.NewToolResultText(fmt.Sprintf("Map '%s' generated (%d rooms, %s style)", filename, rooms, style)), nil
	}
}

// HandleGenerateDivider handles divider generation
func (h *AssetHandlers) HandleGenerateDivider() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		filename := getStringArg(args, "filename")
		style := getStringArg(args, "style")
		width := getIntArg(args, "width")
		if width <= 0 {
			width = 600
		}

		if campaign == "" || filename == "" {
			return mcp.NewToolResultError("campaign and filename are required"), nil
		}

		_, err := h.service.GenerateDivider(campaign, filename, style, width)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Divider '%s' generated (%s style, %dpx)", filename, style, width)), nil
	}
}

// HandleGenerateImage handles single image generation
func (h *AssetHandlers) HandleGenerateImage() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		filename := getStringArg(args, "filename")
		prompt := getStringArg(args, "prompt")
		imgType := getStringArg(args, "type")

		if campaign == "" || filename == "" || prompt == "" {
			return mcp.NewToolResultError("campaign, filename, and prompt are required"), nil
		}

		path, err := h.service.GenerateImage(campaign, filename, prompt, imgType)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Image generated: %s (type: %s)", path, imgType)), nil
	}
}

func splitLabels(labels string) []string {
	var result []string
	for _, l := range strings.Split(labels, ",") {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result
}


