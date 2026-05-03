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

// HandleGenerateImagesBatch handles batch image generation
func (h *AssetHandlers) HandleGenerateImagesBatch() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		imagesRaw, ok := args["images"].([]interface{})
		if !ok || len(imagesRaw) == 0 {
			return mcp.NewToolResultError("images array is required and must not be empty"), nil
		}

		var specs []services.BatchImageSpec
		for i, raw := range imagesRaw {
			img, ok := raw.(map[string]interface{})
			if !ok {
				return mcp.NewToolResultError(fmt.Sprintf("image %d: invalid format", i)), nil
			}

			spec := services.BatchImageSpec{}
			if v, ok := img["filename"].(string); ok {
				spec.Filename = v
			}
			if v, ok := img["prompt"].(string); ok {
				spec.Prompt = v
			}
			if v, ok := img["type"].(string); ok {
				spec.Type = v
			}
			if spec.Filename == "" || spec.Prompt == "" {
				return mcp.NewToolResultError(fmt.Sprintf("image %d: filename and prompt are required", i)), nil
			}
			specs = append(specs, spec)
		}

		results, err := h.service.GenerateImagesBatch(campaign, specs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var successes, failures []string
		for _, r := range results {
			if r.Success {
				successes = append(successes, fmt.Sprintf("  ✅ %s → %s", r.Filename, r.Path))
			} else {
				failures = append(failures, fmt.Sprintf("  ❌ %s: %s", r.Filename, r.Error))
			}
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Batch image generation complete\n"))
		sb.WriteString(fmt.Sprintf("Total: %d | Success: %d | Failed: %d\n\n", len(specs), len(successes), len(failures)))

		if len(successes) > 0 {
			sb.WriteString("Generated:\n")
			for _, s := range successes {
				sb.WriteString(s + "\n")
			}
		}

		if len(failures) > 0 {
			sb.WriteString("\nFailed:\n")
			for _, f := range failures {
				sb.WriteString(f + "\n")
			}
		}

		return mcp.NewToolResultText(sb.String()), nil
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


