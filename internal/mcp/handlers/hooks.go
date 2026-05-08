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

// HookHandlers handles character hook-related MCP tools
type HookHandlers struct {
	service *services.PlayerHookService
}

// NewHookHandlers creates new hook handlers
func NewHookHandlers(service *services.PlayerHookService) *HookHandlers {
	return &HookHandlers{service: service}
}

// HandleGenerateCharacterHooks handles the generate_character_hooks tool
func (h *HookHandlers) HandleGenerateCharacterHooks() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		hooks, warnings, err := h.service.GenerateHooks(ctx, campaign)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format hooks as markdown table
		result := formatCharacterHooks(hooks, warnings)
		return mcp.NewToolResultText(result), nil
	}
}

// formatCharacterHooks formats hooks as a markdown table with per-area breakdown
func formatCharacterHooks(hooks []domain.CharacterHook, warnings []string) string {
	if len(hooks) == 0 {
		return "No character hooks generated. Add characters to the campaign first."
	}

	result := "# Character Hooks\n\n"
	
	if len(warnings) > 0 {
		result += "## Warnings\n\n"
		for _, w := range warnings {
			result += fmt.Sprintf("- %s\n", w)
		}
		result += "\n---\n\n"
	}

	result += "## Hooks por Personaje\n\n"

	for _, hook := range hooks {
		result += fmt.Sprintf("### %s\n", hook.CharacterName)
		result += fmt.Sprintf("**Background:** %s | **Class:** %s\n\n", hook.Background, hook.Class)
		result += fmt.Sprintf("### Gancho Personal\n\n%s\n\n", hook.Hook)
		result += fmt.Sprintf("**Conexión con la Trama:** %s\n\n", hook.ConnectionToPlot)
		result += "---\n\n"
	}

	result += "## Hooks por Área (para incluir en cada área)\n\n"
	result += "| Área | Personaje | Background | Gancho |\n"
	result += "|------|-----------|------------|--------|\n"

	for i, hook := range hooks {
		areaNum := (i % 5) + 1 // Distribute hooks across areas 1-5
		result += fmt.Sprintf("| Área %d | %s | %s | %s |\n",
			areaNum,
			hook.CharacterName,
			hook.Background,
			truncateString(hook.Hook, 50))
	}

	result += "\n## Instrucciones para el DM\n\n"
	result += "1. **Antes de la Sesión Cero:** Generá estos hooks y compartilos individualmente con cada jugador\n"
	result += "2. **Durante el Juego:** Incluí referencias a estos hooks en las áreas correspondientes\n"
	result += "3. **Evolución:** Actualizá los hooks según las decisiones de los personajes\n"
	result += "4. **Recompensas:** Los hooks bien interpretados pueden dar ventaja en tiradas sociales o encuentros con NPCs aliados\n"

	return result
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// HandleValidateCharacterHooks handles validation of character hooks in area content
func (h *HookHandlers) HandleValidateCharacterHooks() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		content := getStringArg(args, "content")
		if content == "" {
			return mcp.NewToolResultError("content is required"), nil
		}

		// Use the validators package
		// Note: This would need import of validators package
		// For now, return a placeholder
		result := map[string]any{
			"valid":   true,
			"message": "Character hooks validation not yet implemented in MCP handler",
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}
