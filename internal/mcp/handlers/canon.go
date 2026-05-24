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
	gateService    *services.ConsistencyGateService
}

// NewCanonHandlers creates new canon handlers
func NewCanonHandlers(
	canonService *services.CanonService,
	stateService *services.NarrativeStateService,
	validationEngine *services.ValidationEngine,
	gateService *services.ConsistencyGateService,
) *CanonHandlers {
	return &CanonHandlers{
		canonService:     canonService,
		stateService:     stateService,
		validationEngine: validationEngine,
		gateService:      gateService,
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
			Name:             getStringArg(args, "name"),
			BriefDescription: getStringArg(args, "brief_description"),
			LevelRange:       getStringArg(args, "level_range"),
			Tone:             getStringArg(args, "tone"),
			SettingType:      getStringArg(args, "setting_type"),
			VillainType:      getStringArg(args, "villain_type"),
			McGuffinType:     getStringArg(args, "mcguffin_type"),
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
		factionContext := getStringArg(args, "faction_context")

		proposal := domain.ContentProposal{
			ID:             proposalID,
			Type:           proposalType,
			Content:        content,
			FactionContext: factionContext,
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

		update := domain.StateUpdate{
			SessionNum: sessionNum,
		}

		// Parse revealed clues (accepts both strings and objects)
		if cluesVal, ok := args["revealed_clues"]; ok {
			if clues, ok := cluesVal.([]any); ok {
				for _, c := range clues {
					switch v := c.(type) {
					case string:
						update.RevealedClues = append(update.RevealedClues, domain.RevealedClue{
							ID:              fmt.Sprintf("clue-%d", len(update.RevealedClues)+1),
							Description:     v,
							SourceAct:       "unknown",
							SessionRevealed: sessionNum,
						})
					case map[string]any:
						update.RevealedClues = append(update.RevealedClues, domain.RevealedClue{
							ID:              getStringArg(v, "id"),
							Description:     getStringArg(v, "description"),
							SourceAct:       getStringArg(v, "source_act"),
							SourceArea:      getStringArg(v, "source_area"),
							SessionRevealed: sessionNum,
							IsCritical:      getBoolArg(v, "is_critical"),
						})
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

		// Parse dead NPCs (accepts both strings and objects)
		if deadVal, ok := args["dead_npcs"]; ok {
			if deadNPCs, ok := deadVal.([]any); ok {
				for _, n := range deadNPCs {
					switch v := n.(type) {
					case string:
						update.DeadNPCs = append(update.DeadNPCs, domain.NPCDeathRecord{
							NPCID:   fmt.Sprintf("npc-%d", len(update.DeadNPCs)+1),
							Name:    v,
							Session: sessionNum,
						})
					case map[string]any:
						update.DeadNPCs = append(update.DeadNPCs, domain.NPCDeathRecord{
							NPCID:    getStringArg(v, "npc_id"),
							Name:     getStringArg(v, "name"),
							Session:  sessionNum,
							Cause:    getStringArg(v, "cause"),
							KilledBy: getStringArg(v, "killed_by"),
							Location: getStringArg(v, "location"),
						})
					}
				}
			}
		}

		// Parse key decisions (accepts both strings and objects)
		if decisionsVal, ok := args["key_decisions"]; ok {
			if decisions, ok := decisionsVal.([]any); ok {
				for _, d := range decisions {
					switch v := d.(type) {
					case string:
						update.KeyDecisions = append(update.KeyDecisions, domain.Decision{
							ID:          fmt.Sprintf("decision-%d", len(update.KeyDecisions)+1),
							Description: v,
						})
					case map[string]any:
						update.KeyDecisions = append(update.KeyDecisions, domain.Decision{
							ID:          getStringArg(v, "id"),
							Description: getStringArg(v, "description"),
							ChoiceMade:  getStringArg(v, "choice_made"),
							ImpactScope: getStringArg(v, "impact_scope"),
						})
					}
				}
			}
		}

		// Parse active quests as strings (creates QuestState objects)
		if activeVal, ok := args["active_quests"]; ok {
			if active, ok := activeVal.([]any); ok {
				for _, q := range active {
					if s, ok := q.(string); ok {
						update.NewQuests = append(update.NewQuests, domain.QuestState{
							ID:     fmt.Sprintf("quest-%d", len(update.NewQuests)+1),
							Name:   s,
							Status: "active",
						})
					}
				}
			}
		}

		// Parse key items as strings (creates KeyItem objects)
		if itemsVal, ok := args["key_items"]; ok {
			if items, ok := itemsVal.([]any); ok {
				for _, item := range items {
					if s, ok := item.(string); ok {
						update.KeyItems = append(update.KeyItems, domain.KeyItem{
							ID:           fmt.Sprintf("item-%d", len(update.KeyItems)+1),
							Name:         s,
							Holder:       "party",
							SessionFound: sessionNum,
						})
					}
				}
			}
		}

		// Parse session metadata
		update.SessionSummary = getStringArg(args, "session_summary")
		update.XPAwarded = getIntArg(args, "xp_awarded")
		update.DMNotes = getStringArg(args, "dm_notes")

		// Parse loot acquired
		if lootVal, ok := args["loot_acquired"]; ok {
			if loot, ok := lootVal.([]any); ok {
				for _, l := range loot {
					if s, ok := l.(string); ok {
						update.LootAcquired = append(update.LootAcquired, s)
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

// HandleProcessConsistencyGate handles the process_consistency_gate tool
func (h *CanonHandlers) HandleProcessConsistencyGate() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		campaignID := getStringArg(args, "campaign_id")
		batchID := getStringArg(args, "batch_id")
		fastMode := getBoolArg(args, "fast_mode")
		attempt := getIntArg(args, "attempt")
		if attempt == 0 {
			attempt = 1
		}

		if campaignID == "" {
			return mcp.NewToolResultError("campaign_id is required"), nil
		}
		if batchID == "" {
			return mcp.NewToolResultError("batch_id is required"), nil
		}

		// Parse proposals from args
		var proposals []domain.ContentProposal
		if proposalsVal, ok := args["proposals"]; ok {
			if proposalsArr, ok := proposalsVal.([]any); ok {
				for _, p := range proposalsArr {
					if pMap, ok := p.(map[string]any); ok {
						proposal := domain.ContentProposal{
							ID:      getStringArg(pMap, "id"),
							Type:    getStringArg(pMap, "type"),
							Content: getStringArg(pMap, "content"),
						}
						if refsVal, ok := pMap["entity_references"]; ok {
							if refsArr, ok := refsVal.([]any); ok {
								for _, r := range refsArr {
									if rMap, ok := r.(map[string]any); ok {
										ref := domain.EntityReference{
											EntityID: getStringArg(rMap, "entity_id"),
										}
										proposal.EntityReferences = append(proposal.EntityReferences, ref)
									}
								}
							}
						}
						proposals = append(proposals, proposal)
					}
				}
			}
		}

		batchProposal := domain.BatchProposal{
			BatchID:    batchID,
			CampaignID: campaignID,
			Artifacts:  proposals,
			Attempt:    attempt,
		}

		result, err := h.gateService.ProcessBatch(ctx, batchProposal, fastMode)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonBytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
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
