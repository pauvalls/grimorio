package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// FlowchartHandlers handles flowchart MCP tools
type FlowchartHandlers struct {
	flowchartService *services.FlowchartService
}

// NewFlowchartHandlers creates new flowchart handlers
func NewFlowchartHandlers(flowchartService *services.FlowchartService) *FlowchartHandlers {
	return &FlowchartHandlers{flowchartService: flowchartService}
}

// HandleGenerateFlowchart handles the generate_flowchart tool
func (h *FlowchartHandlers) HandleGenerateFlowchart() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		detailLevel := getStringArg(args, "detail_level")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		if detailLevel == "" {
			detailLevel = "overview"
		}

		mermaid, err := h.flowchartService.GenerateMermaid(ctx, campaignID, detailLevel)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Also generate SVG
		svg, err := h.flowchartService.GenerateSVG(ctx, campaignID, detailLevel)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := struct {
			Mermaid  string `json:"mermaid"`
			SVG      string `json:"svg"`
			SVGPath  string `json:"svg_path,omitempty"`
		}{
			Mermaid: mermaid,
			SVG:     svg,
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
