package handlers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/services"
)

// PrologueHandlers handles prologue-related MCP tools
type PrologueHandlers struct {
	prologueService *services.PrologueService
	campaignService *services.CampaignService
}

// NewPrologueHandlers creates new prologue handlers
func NewPrologueHandlers(ps *services.PrologueService, cs *services.CampaignService) *PrologueHandlers {
	return &PrologueHandlers{
		prologueService: ps,
		campaignService: cs,
	}
}

// HandleGeneratePrologue handles the grimorio_generate_prologue tool
func (h *PrologueHandlers) HandleGeneratePrologue() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		tone := getStringArg(args, "tone")
		characterHooks := getArrayArg(args, "character_hooks")
		regenerate := getBoolArg(args, "regenerate")

		// Check if prologue already exists (unless regenerate is true)
		if !regenerate {
			prologuePath := filepath.Join(h.campaignService.GetBaseDir(), campaign, "prologue.md")
			if _, err := os.Stat(prologuePath); err == nil {
				return mcp.NewToolResultText(fmt.Sprintf("Prologue already exists for campaign '%s'. Use regenerate=true to overwrite.", campaign)), nil
			}
		}

		// Convert characterHooks []any to []string
		var hooks []string
		for _, h := range characterHooks {
			if s, ok := h.(string); ok {
				hooks = append(hooks, s)
			}
		}

		// Generate prologue via service
		prologue, warnings, err := h.prologueService.GeneratePrologue(ctx, campaign, tone, hooks)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Save to prologue.md
		if err := h.campaignService.SavePrologue(campaign, prologue); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to save prologue: %v", err)), nil
		}

		savedTo := filepath.Join(campaign, "prologue.md")

		result := fmt.Sprintf("Prologue generated for campaign '%s'.\n\nSaved to: %s\n\n", campaign, savedTo)

		if len(warnings) > 0 {
			result += "**Warnings:**\n"
			for _, w := range warnings {
				result += fmt.Sprintf("- %s\n", w)
			}
			result += "\n"
		}

		// Include the rendered markdown
		if prologue != nil {
			result += fmt.Sprintf("**Prologue Parts:** %d\n\n", len(prologue.Parts))
			for i, part := range prologue.Parts {
				result += fmt.Sprintf("**Part %d: %s**\n%s\n\n", part.Order, part.Title, part.Content)
				if i < len(prologue.Parts)-1 {
					result += "---\n\n"
				}
			}
		}

		return mcp.NewToolResultText(result), nil
	}
}
