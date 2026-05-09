package handlers

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services"
)

// CharacterHandlers handles character-related MCP tools
type CharacterHandlers struct {
	service *services.CharacterService
}

// NewCharacterHandlers creates new character handlers
func NewCharacterHandlers(service *services.CharacterService) *CharacterHandlers {
	return &CharacterHandlers{service: service}
}

// HandleGenerateCharacter handles character creation
func (h *CharacterHandlers) HandleGenerateCharacter() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		name := getStringArg(args, "name")
		race := getStringArg(args, "race")
		class := getStringArg(args, "class")
		level := getIntArg(args, "level")
		background := getStringArg(args, "background")
		alignment := getStringArg(args, "alignment")
		withBackstory := getBoolArg(args, "with_backstory")

		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}
		if name == "" {
			return mcp.NewToolResultError("name is required"), nil
		}
		if level <= 0 {
			level = 1
		}

		var character *domain.Character
		var err error

		if withBackstory {
			character, err = h.service.GenerateWithBackstory(campaign, name, race, class, level, background, alignment)
		} else {
			character, err = h.service.CreateCharacter(campaign, name, race, class, level, background, alignment)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		msg := fmt.Sprintf("Character '%s' created in campaign '%s' (Level %d %s %s)",
			character.Name, campaign, character.Level, character.Race, character.Class)
		if withBackstory {
			msg += " with backstory hooks, secrets, and goals"
		}

		return mcp.NewToolResultText(msg), nil
	}
}

// HandleGetCharacter handles retrieving a character
func (h *CharacterHandlers) HandleGetCharacter() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		name := getStringArg(args, "name")

		if campaign == "" || name == "" {
			return mcp.NewToolResultError("campaign and name are required"), nil
		}

		character, err := h.service.GetCharacter(campaign, name)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format character sheet as text
		result := formatCharacterSheet(character)
		return mcp.NewToolResultText(result), nil
	}
}

// HandleListCharacters handles listing characters
func (h *CharacterHandlers) HandleListCharacters() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		characters, err := h.service.ListCharacters(campaign)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		if len(characters) == 0 {
			return mcp.NewToolResultText("No characters found in campaign '" + campaign + "'"), nil
		}

		var result string
		for _, c := range characters {
			result += fmt.Sprintf("- %s (Level %d %s %s) - %s\n",
				c.Name, c.Level, c.Race, c.Class, c.Status)
		}
		return mcp.NewToolResultText(result), nil
	}
}

// HandleSaveCharacters handles saving multiple characters
func (h *CharacterHandlers) HandleSaveCharacters() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaign := getStringArg(args, "campaign")
		if campaign == "" {
			return mcp.NewToolResultError("campaign is required"), nil
		}

		characters := getArrayArg(args, "characters")
		if len(characters) == 0 {
			return mcp.NewToolResultError("at least one character is required"), nil
		}

		var savedCount int
		var lastErr error
		for _, charData := range characters {
			charMap, ok := charData.(map[string]any)
			if !ok {
				continue
			}

			name := getStringArg(charMap, "name")
			if name == "" {
				lastErr = fmt.Errorf("character name is required")
				continue
			}

			level := getIntArg(charMap, "level")
			if level <= 0 {
				level = 1
			}

			character := &domain.Character{
				CampaignID: campaign,
				Name:       name,
				Race:       getStringArg(charMap, "race"),
				Class:      getStringArg(charMap, "class"),
				Level:      level,
				Background: getStringArg(charMap, "background"),
				Alignment:  getStringArg(charMap, "alignment"),
				Status:     "alive",
			}

			if err := h.service.SaveCharacter(character); err != nil {
				lastErr = err
				continue
			}
			savedCount++
		}

		if lastErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("saved %d/%d characters, last error: %v", savedCount, len(characters), lastErr)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Saved %d characters to campaign '%s'", savedCount, campaign)), nil
	}
}

func formatCharacterSheet(c *domain.Character) string {
	// This is a simplified formatter - in production would be more detailed
	return fmt.Sprintf(`
# %s
**Level %d %s %s**

**Stats:** STR %d | DEX %d | CON %d | INT %d | WIS %d | CHA %d
**HP:** %d/%d | **AC:** %d
**Status:** %s
`, c.Name, c.Level, c.Race, c.Class,
		c.Stats.STR, c.Stats.DEX, c.Stats.CON, c.Stats.INT, c.Stats.WIS, c.Stats.CHA,
		c.HP.Current, c.HP.Maximum, c.AC, c.Status)
}
