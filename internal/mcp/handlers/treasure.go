package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// TreasureHandlers handles treasure generation MCP tools.
type TreasureHandlers struct {
	service *services.TreasureService
}

// NewTreasureHandlers creates new treasure handlers.
func NewTreasureHandlers(service *services.TreasureService) *TreasureHandlers {
	return &TreasureHandlers{service: service}
}

// HandleGenerateTreasure handles the generate_treasure tool.
func (h *TreasureHandlers) HandleGenerateTreasure() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		genType := getStringArg(args, "type")
		if genType == "" {
			genType = "individual"
		}

		if h.service == nil {
			return mcp.NewToolResultError("treasure service not initialized"), nil
		}

		var result any
		var err error

		switch genType {
		case "individual":
			cr := getIntArg(args, "cr")
			if cr < 0 {
				cr = 0
			}
			result, err = h.service.GenerateIndividualTreasure(ctx, cr)
		case "hoard":
			tier := getIntArg(args, "tier")
			if tier < 1 {
				tier = 1
			}
			result, err = h.service.GenerateTreasureHoard(ctx, tier)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("unknown treasure type: %s", genType)), nil
		}

		if err != nil {
			return ToToolResult(err), nil
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
