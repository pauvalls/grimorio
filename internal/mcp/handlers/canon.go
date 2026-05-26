package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
		defaultSourceAct := getStringArg(args, "default_source_act")
		if defaultSourceAct == "" {
			defaultSourceAct = "unknown"
		}
		defaultChoiceMade := getStringArg(args, "default_choice_made")
		defaultImpactScope := getStringArg(args, "default_impact_scope")

		update := domain.StateUpdate{
			SessionNum: sessionNum,
		}

		// Parse revealed clues (accepts both strings and objects)
		if clues := getStringArray(args, "revealed_clues"); clues != nil {
			for _, c := range clues {
				switch v := c.(type) {
				case string:
					update.RevealedClues = append(update.RevealedClues, domain.RevealedClue{
						ID:              stableID("clue", campaignID+"|"+v),
						Description:     v,
						SourceAct:       defaultSourceAct,
						SessionRevealed: sessionNum,
					})
				case map[string]any:
					clueID := getStringArg(v, "id")
					// Empty IDs are preserved for validation to reject; stable IDs are only auto-generated for string inputs
					sourceAct := getStringArg(v, "source_act")
					if sourceAct == "" {
						sourceAct = defaultSourceAct
					}
					update.RevealedClues = append(update.RevealedClues, domain.RevealedClue{
						ID:              clueID,
						Description:     getStringArg(v, "description"),
						SourceAct:       sourceAct,
						SourceArea:      getStringArg(v, "source_area"),
						SessionRevealed: sessionNum,
						IsCritical:      getBoolArg(v, "is_critical"),
					})
				}
			}
		}

		// Mark critical clues by index (0-based)
		if criticalIndices := getIntArray(args, "critical_clue_indices"); criticalIndices != nil {
			for _, idx := range criticalIndices {
				if idx >= 0 && idx < len(update.RevealedClues) {
					update.RevealedClues[idx].IsCritical = true
				}
			}
		}

		// Parse completed quests
		if completed := getStringArray(args, "completed_quests"); completed != nil {
			for _, c := range completed {
				if s, ok := c.(string); ok {
					update.CompletedQuests = append(update.CompletedQuests, s)
				}
			}
		}

		// Parse dead NPCs (accepts both strings and objects)
		if deadNPCs := getStringArray(args, "dead_npcs"); deadNPCs != nil {
			for _, n := range deadNPCs {
				switch v := n.(type) {
				case string:
					update.DeadNPCs = append(update.DeadNPCs, domain.NPCDeathRecord{
						NPCID:   stableID("npc", campaignID+"|"+v),
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

		// Parse key decisions (accepts both strings and objects)
		if decisions := getStringArray(args, "key_decisions"); decisions != nil {
			for _, d := range decisions {
				switch v := d.(type) {
				case string:
					update.KeyDecisions = append(update.KeyDecisions, domain.Decision{
						ID:          fmt.Sprintf("decision-%d", len(update.KeyDecisions)+1),
						Description: v,
						ChoiceMade:  defaultChoiceMade,
						ImpactScope: defaultImpactScope,
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

		// Parse active quests (accepts both strings and objects)
		if active := getStringArray(args, "active_quests"); active != nil {
			for _, q := range active {
				switch v := q.(type) {
				case string:
					update.NewQuests = append(update.NewQuests, domain.QuestState{
						ID:     stableID("quest", campaignID+"|"+v),
						Name:   v,
						Status: "active",
					})
				case map[string]any:
					update.NewQuests = append(update.NewQuests, domain.QuestState{
						ID:       getStringArg(v, "id"),
						Name:     getStringArg(v, "name"),
						Status:   getStringArg(v, "status"),
						SourceAct: getStringArg(v, "source_act"),
						GiverNPC: getStringArg(v, "giver_npc"),
					})
				}
			}
		}

		// Parse key items (accepts both strings and objects)
		if items := getStringArray(args, "key_items"); items != nil {
			for _, item := range items {
				switch v := item.(type) {
				case string:
					update.KeyItems = append(update.KeyItems, domain.KeyItem{
						ID:           stableID("item", campaignID+"|"+v),
						Name:         v,
						Holder:       "party",
						SessionFound: sessionNum,
					})
				case map[string]any:
					update.KeyItems = append(update.KeyItems, domain.KeyItem{
						ID:           getStringArg(v, "id"),
						Name:         getStringArg(v, "name"),
						Holder:       getStringArg(v, "holder"),
						SessionFound: sessionNum,
						IsMcGuffin:   getBoolArg(v, "is_mcguffin"),
					})
				}
			}
		}

		// Parse session metadata
		update.SessionSummary = getStringArg(args, "session_summary")
		update.XPAwarded = getIntArg(args, "xp_awarded")
		update.DMNotes = getStringArg(args, "dm_notes")

		// Parse loot acquired
		if loot := getStringArray(args, "loot_acquired"); loot != nil {
			for _, l := range loot {
				if s, ok := l.(string); ok {
					update.LootAcquired = append(update.LootAcquired, s)
				}
			}
		}

		// Parse current location
		update.CurrentLocation = getStringArg(args, "current_location")

		// Parse PC statuses
		if pcStatusVal, ok := args["pc_status"]; ok {
			if pcStatuses, ok := pcStatusVal.([]any); ok {
				for _, ps := range pcStatuses {
					if pcMap, ok := ps.(map[string]any); ok {
						status := domain.PCStatus{
							Name:      getStringArg(pcMap, "name"),
							HPCurrent: getIntArg(pcMap, "hp_current"),
							HPMax:     getIntArg(pcMap, "hp_max"),
						}
						if conditionsVal, ok := pcMap["conditions"]; ok {
							if conditions, ok := conditionsVal.([]any); ok {
								for _, c := range conditions {
									if s, ok := c.(string); ok {
										status.Conditions = append(status.Conditions, s)
									}
								}
							}
						}
						update.PCStatuses = append(update.PCStatuses, status)
					}
				}
			}
		}

		// Parse replace_session flag
		update.ReplaceSession = getBoolArg(args, "replace_session")

		// Parse sync_to_canon flag
		update.SyncToCanon = getBoolArg(args, "sync_to_canon")

		// Pre-mutation validation: reject empty IDs
		for _, clue := range update.RevealedClues {
			if clue.ID == "" {
				return mcp.NewToolResultError("revealed clue with empty ID: all clues must have an explicit or generated ID"), nil
			}
		}
		for _, npc := range update.DeadNPCs {
			if npc.NPCID == "" {
				return mcp.NewToolResultError("dead NPC with empty npc_id: all dead NPCs must have an explicit or generated ID"), nil
			}
		}
		for _, quest := range update.NewQuests {
			if quest.ID == "" {
				return mcp.NewToolResultError("active quest with empty ID: all quests must have an explicit or generated ID"), nil
			}
		}
		for _, item := range update.KeyItems {
			if item.ID == "" {
				return mcp.NewToolResultError("key item with empty ID: all key items must have an explicit or generated ID"), nil
			}
		}

		state, err := h.stateService.Update(ctx, campaignID, update)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Sync to canon if requested
		warnings := []string{}
		if update.SyncToCanon {
			syncWarnings, syncErr := h.stateService.SyncStateToCanon(ctx, campaignID, update)
			if syncErr != nil {
				warnings = append(warnings, fmt.Sprintf("canon sync failed: %v", syncErr))
			} else {
				warnings = append(warnings, syncWarnings...)
			}
		}

		// Check for duplicate clue descriptions
		descSet := make(map[string]int)
		for _, clue := range state.RevealedClues {
			descSet[clue.Description]++
		}
		duplicateCount := 0
		for _, count := range descSet {
			if count > 1 {
				duplicateCount += count - 1
			}
		}
		if duplicateCount > 0 {
			warnings = append(warnings, fmt.Sprintf("%d duplicate clue descriptions detected", duplicateCount))
		}

		// Check for clues without source_act
		noSourceActCount := 0
		for _, clue := range state.RevealedClues {
			if clue.SourceAct == "" || clue.SourceAct == "unknown" {
				noSourceActCount++
			}
		}
		if noSourceActCount > 0 {
			warnings = append(warnings, fmt.Sprintf("%d clues have no source_act", noSourceActCount))
		}

		result := map[string]any{
			"updated":           true,
			"campaign_id":       campaignID,
			"current_session":   state.CurrentSession,
			"active_quests":     len(state.ActiveQuests),
			"revealed_clues":    len(state.RevealedClues),
			"dead_npcs":         len(state.DeadNPCs),
			"warnings":          warnings,
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

// stableID generates a deterministic 8-hex hash ID from a prefix and seed.
func stableID(prefix, seed string) string {
	h := sha256.Sum256([]byte(seed))
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(h[:])[:8])
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

// getStringArray extracts a string array from the args map, handling both []any and []string
func getStringArray(args map[string]any, key string) []any {
	val, ok := args[key]
	if !ok {
		return nil
	}

	// Try []any first
	if arr, ok := val.([]any); ok {
		return arr
	}

	// Try []string
	if arr, ok := val.([]string); ok {
		result := make([]any, len(arr))
		for i, s := range arr {
			result[i] = s
		}
		return result
	}

	return nil
}

// getIntArray extracts an int array from the args map
func getIntArray(args map[string]any, key string) []int {
	val, ok := args[key]
	if !ok {
		return nil
	}

	// Try []any (float64 from JSON)
	if arr, ok := val.([]any); ok {
		result := make([]int, 0, len(arr))
		for _, v := range arr {
			if f, ok := v.(float64); ok {
				result = append(result, int(f))
			}
		}
		return result
	}

	// Try []int
	if arr, ok := val.([]int); ok {
		return arr
	}

	return nil
}
