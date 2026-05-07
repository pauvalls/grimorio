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

// TableHandlers handles random table MCP tools
type TableHandlers struct {
	tableService *services.RandomTableService
}

// NewTableHandlers creates new table handlers
func NewTableHandlers(tableService *services.RandomTableService) *TableHandlers {
	return &TableHandlers{tableService: tableService}
}

// HandleGenerateRandomTables handles the generate_random_tables tool
func (h *TableHandlers) HandleGenerateRandomTables() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		tableTypeStr := getStringArg(args, "table_type")

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}
		if tableTypeStr == "" {
			return mcp.NewToolResultError("table_type is required"), nil
		}

		ctx2 := domain.TableContext{
			LevelRange:   getStringArg(args, "level_range"),
			SettingType:  getStringArg(args, "setting_type"),
			PartySize:    getIntArg(args, "party_size"),
			LocationHint: getStringArg(args, "location_hint"),
		}

		tbl, err := h.tableService.GenerateTable(ctx, campaignID, domain.TableType(tableTypeStr), ctx2)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(tbl, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
