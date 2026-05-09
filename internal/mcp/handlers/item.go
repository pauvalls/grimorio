package handlers

import (
	"context"
	"encoding/json"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// ItemHandlers handles magic item MCP tools.
type ItemHandlers struct {
	service *services.ItemService
}

// NewItemHandlers creates item handlers.
func NewItemHandlers(service *services.ItemService) *ItemHandlers {
	return &ItemHandlers{service: service}
}

// HandleGenerateMagicItem handles grimorio_generate_magic_item.
func (h *ItemHandlers) HandleGenerateMagicItem() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		rarityStr := getStringArg(args, "rarity")
		itemTypeStr := getStringArg(args, "item_type")
		
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id required"), nil
		}

		// Defaults
		if rarityStr == "" {
			rarityStr = "rare"
		}
		if itemTypeStr == "" {
			itemTypeStr = "weapon"
		}

		rarity := domain.MagicItemRarity(rarityStr)
		itemType := domain.MagicItemType(itemTypeStr)

		cursed := false
		if v, ok := args["cursed"].(bool); ok {
			cursed = v
		}

		var item *domain.MagicItem
		var err error
		if cursed {
			item, err = h.service.GenerateCursedItem(ctx, campaignID, rarity)
		} else {
			item, err = h.service.GenerateItem(ctx, campaignID, rarity, itemType)
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, _ := json.MarshalIndent(item, "", "  ")
		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
