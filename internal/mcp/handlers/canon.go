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

// CanonHandlers handles canon and narrative coherence MCP tools
type CanonHandlers struct {
	canonService   *services.CanonService
	stateService   *services.NarrativeStateService
	validationEngine *services.ValidationEngine
}

// NewCanonHandlers creates new canon handlers
func NewCanonHandlers(
	canonService *services.CanonService,
	stateService *services.NarrativeStateService,
	validationEngine *services.ValidationEngine,
) *CanonHandlers {
	return &CanonHandlers{
		canonService:     canonService,
		stateService:     stateService,
		validationEngine: validationEngine,
	}
}

// HandleGenerateAdventureBible handles the generate_adventure_bible tool
func (h *CanonHandlers) HandleGenerateAdventureBible() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		brief := domain.CampaignBrief{
			Name:         getStringArg(args, "name"),
			LevelRange:   getStringArg(args, "level_range"),
			Tone:         getStringArg(args, "tone"),
			SettingType:  getStringArg(args, "setting_type"),
			VillainType:  getStringArg(args, "villain_type"),
			McGuffinType: getStringArg(args, "mcguffin_type"),
		}

		// Parse themes array if provided
		if themesVal, ok := args["themes"]; ok {
			if themes, ok := themesVal.([]any); ok {
				for _, t := range themes {
					if s, ok := t.(string); ok {
						brief.Themes = append(brief.Themes, s)
					}
				}
			}
		}

		if brief.Name == "" {
			return mcp.NewToolResultError("campaign name is required"), nil
		}

		doc, err := h.canonService.InitializeCanon(ctx, brief)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]any{
			"canon_id":            doc.CampaignID,
			"campaign_id":         doc.CampaignID,
			"facts_count":         len(doc.Facts),
			"entities_count":      len(doc.Entities),
			"timeline_events_count": len(doc.Timeline),
			"rules_count":         len(doc.Rules),
			"canon_summary":       fmt.Sprintf("Canon initialized for '%s' with %d facts, %d entities, %d timeline events, and %d rules.", doc.CampaignID, len(doc.Facts), len(doc.Entities), len(doc.Timeline), len(doc.Rules)),
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleValidateCanon handles the validate_canon tool
func (h *CanonHandlers) HandleValidateCanon() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		proposalID := getStringArg(args, "proposal_id")
		proposalType := getStringArg(args, "proposal_type")
		content := getStringArg(args, "content")

		proposal := domain.ContentProposal{
			ID:      proposalID,
			Type:    proposalType,
			Content: content,
		}

		report, err := h.canonService.ValidateProposal(ctx, campaignID, proposal)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal report: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleUpdateNarrativeState handles the update_narrative_state tool
func (h *CanonHandlers) HandleUpdateNarrativeState() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		sessionNum := getIntArg(args, "session_num")
		if sessionNum <= 0 {
			return mcp.NewToolResultError("session_num must be positive"), nil
		}

		update := domain.StateUpdate{
			SessionNum: sessionNum,
		}

		// Parse revealed clues
		if cluesVal, ok := args["revealed_clues"]; ok {
			if clues, ok := cluesVal.([]any); ok {
				for _, c := range clues {
					if clueMap, ok := c.(map[string]any); ok {
						clue := domain.RevealedClue{
							ID:          getStringArg(clueMap, "id"),
							Description: getStringArg(clueMap, "description"),
							SourceAct:   getStringArg(clueMap, "source_act"),
							SourceArea:  getStringArg(clueMap, "source_area"),
							SessionRevealed: sessionNum,
							IsCritical:  getBoolArg(clueMap, "is_critical"),
						}
						update.RevealedClues = append(update.RevealedClues, clue)
					}
				}
			}
		}

		// Parse completed quests
		if completedVal, ok := args["completed_quests"]; ok {
			if completed, ok := completedVal.([]any); ok {
				for _, c := range completed {
					if s, ok := c.(string); ok {
						update.CompletedQuests = append(update.CompletedQuests, s)
					}
				}
			}
		}

		// Parse dead NPCs
		if deadVal, ok := args["dead_npcs"]; ok {
			if deadNPCs, ok := deadVal.([]any); ok {
				for _, n := range deadNPCs {
					if npcMap, ok := n.(map[string]any); ok {
						npc := domain.NPCDeathRecord{
							NPCID:    getStringArg(npcMap, "npc_id"),
							Name:     getStringArg(npcMap, "name"),
							Session:  sessionNum,
							Cause:    getStringArg(npcMap, "cause"),
							KilledBy: getStringArg(npcMap, "killed_by"),
							Location: getStringArg(npcMap, "location"),
						}
						update.DeadNPCs = append(update.DeadNPCs, npc)
					}
				}
			}
		}

		// Parse key decisions
		if decisionsVal, ok := args["key_decisions"]; ok {
			if decisions, ok := decisionsVal.([]any); ok {
				for _, d := range decisions {
					if decMap, ok := d.(map[string]any); ok {
						decision := domain.Decision{
							ID:          getStringArg(decMap, "id"),
							Description: getStringArg(decMap, "description"),
							ChoiceMade:  getStringArg(decMap, "choice_made"),
							ImpactScope: getStringArg(decMap, "impact_scope"),
						}
						update.KeyDecisions = append(update.KeyDecisions, decision)
					}
				}
			}
		}

		state, err := h.stateService.Update(ctx, campaignID, update)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]any{
			"updated":           true,
			"campaign_id":       campaignID,
			"current_session":   state.CurrentSession,
			"active_quests":     len(state.ActiveQuests),
			"revealed_clues":    len(state.RevealedClues),
			"dead_npcs":         len(state.DeadNPCs),
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// HandleCheckConsistency handles the check_consistency tool
func (h *CanonHandlers) HandleCheckConsistency() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}

		scopeStr := getStringArg(args, "scope")
		if scopeStr == "" {
			scopeStr = "full"
		}

		scope := domain.ConsistencyScope(scopeStr)

		report, err := h.validationEngine.CheckConsistency(ctx, campaignID, scope)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal report: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonBytes)), nil
	}
}

// getBoolArg extracts a bool argument from the args map
func getBoolArg(args map[string]any, key string) bool {
	if val, ok := args[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
