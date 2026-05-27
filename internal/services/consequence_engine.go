package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// ConsequenceEngine evaluates consequence rules against narrative state
type ConsequenceEngine struct {
	canonRepo repository.CanonRepository
	reactors  map[string]domain.WorldReactorFunc
}

// NewConsequenceEngine creates a new consequence engine
func NewConsequenceEngine(canonRepo repository.CanonRepository) *ConsequenceEngine {
	return &ConsequenceEngine{
		canonRepo: canonRepo,
		reactors:  make(map[string]domain.WorldReactorFunc),
	}
}

// RegisterReactor registers a reactor for a trigger type
func (e *ConsequenceEngine) RegisterReactor(triggerType string, reactor domain.WorldReactorFunc) {
	e.reactors[triggerType] = reactor
}

// Evaluate checks all consequence rules against the current narrative state
func (e *ConsequenceEngine) Evaluate(ctx context.Context, campaignID string, state *domain.NarrativeState) (*domain.ConsequenceEvaluation, error) {
	doc, err := e.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	eval := &domain.ConsequenceEvaluation{
		CampaignID:       campaignID,
		SessionNum:       state.CurrentSession,
		TriggeredRules:   []domain.ConsequenceRule{},
		ImmediateEffects: []domain.Effect{},
		DelayedEffects:   []domain.DelayedEffect{},
	}

	var rules []domain.ConsequenceRule
	for _, canonRule := range doc.Rules {
		if canonRule.Domain != "consequence" {
			continue
		}
		var rule domain.ConsequenceRule
		if err := json.Unmarshal([]byte(canonRule.Statement), &rule); err != nil {
			continue // skip malformed rules
		}
		rule.ID = canonRule.ID
		rules = append(rules, rule)
	}

	for _, rule := range rules {
		if e.ruleMatches(rule, state) {
			eval.TriggeredRules = append(eval.TriggeredRules, rule)
		}
	}

	// Sort by priority descending
	sort.Slice(eval.TriggeredRules, func(i, j int) bool {
		return eval.TriggeredRules[i].Priority > eval.TriggeredRules[j].Priority
	})

	for _, rule := range eval.TriggeredRules {
		for i, effect := range rule.Effects {
			if effect.Delay != "" {
				applySession := state.CurrentSession + parseDelaySessions(effect.Delay)
				eval.DelayedEffects = append(eval.DelayedEffects, domain.DelayedEffect{
					ID:             makeDelayedEffectID(rule.ID, i),
					Description:    effect.Description,
					EffectType:     effect.Type,
					Target:         effect.Target,
					Effect:         effect,
					TriggerSession: state.CurrentSession,
					ApplySession:   applySession,
					Applied:        false,
				})
			} else {
				eval.ImmediateEffects = append(eval.ImmediateEffects, effect)
			}
		}
	}

	return eval, nil
}

func (e *ConsequenceEngine) ruleMatches(rule domain.ConsequenceRule, state *domain.NarrativeState) bool {
	if rule.DMOverride {
		return true
	}

	trigger := rule.Trigger

	// Evaluate trigger first
	triggerMatched := false

	// Built-in trigger: npc_death
	if trigger.Type == "npc_death" {
		for _, death := range state.DeadNPCs {
			if death.NPCID == trigger.EntityID || trigger.EntityID == "" {
				triggerMatched = true
				break
			}
		}
	}

	// Built-in trigger: any
	if trigger.Type == "any" {
		triggerMatched = true
	}

	// Delegate to registered reactors
	if !triggerMatched {
		if reactor, ok := e.reactors[trigger.Type]; ok {
			effects, err := reactor(trigger, state)
			if err != nil {
				return false
			}
			triggerMatched = len(effects) > 0
		}
	}

	if !triggerMatched {
		return false
	}

	// Evaluate conditions — ALL conditions must pass
	for _, cond := range rule.Conditions {
		if !e.conditionMatches(cond, state) {
			return false
		}
	}

	return true
}

func (e *ConsequenceEngine) conditionMatches(cond domain.Condition, state *domain.NarrativeState) bool {
	switch cond.Type {
	case "npc_alive":
		for _, death := range state.DeadNPCs {
			if death.NPCID == cond.Target {
				return false
			}
		}
		return true
	case "quest_active":
		for _, quest := range state.ActiveQuests {
			if quest.ID == cond.Target {
				return true
			}
		}
		return false
	case "clue_revealed":
		for _, clue := range state.RevealedClues {
			if clue.ID == cond.Target {
				return true
			}
		}
		return false
	case "session_min":
		minSession, ok := cond.Value.(float64)
		if !ok {
			// Try int
			if minInt, ok := cond.Value.(int); ok {
				return state.CurrentSession >= minInt
			}
			return false
		}
		return state.CurrentSession >= int(minSession)
	case "session_max":
		maxSession, ok := cond.Value.(float64)
		if !ok {
			if maxInt, ok := cond.Value.(int); ok {
				return state.CurrentSession <= maxInt
			}
			return false
		}
		return state.CurrentSession <= int(maxSession)
	case "has_item":
		for _, item := range state.KeyItems {
			if item.ID == cond.Target {
				return true
			}
		}
		return false
	default:
		// Unknown condition type: pass (permissive)
		return true
	}
}

// ApplyEvaluation appends non-duplicate delayed effects from an evaluation to the narrative state.
// It respects IsRepeatable: if a rule is not repeatable and the state already contains an effect
// from that rule, no new effects from that rule are appended.
func (e *ConsequenceEngine) ApplyEvaluation(eval *domain.ConsequenceEvaluation, state *domain.NarrativeState) error {
	if eval == nil || state == nil {
		return nil
	}

	// Build map of rule IDs to their IsRepeatable status from triggered rules
	ruleRepeatable := make(map[string]bool, len(eval.TriggeredRules))
	for _, rule := range eval.TriggeredRules {
		ruleRepeatable[rule.ID] = rule.IsRepeatable
	}

	existingIDs := make(map[string]bool, len(state.PendingEffects))
	for _, pe := range state.PendingEffects {
		existingIDs[pe.ID] = true
	}

	for _, de := range eval.DelayedEffects {
		if existingIDs[de.ID] {
			continue // dedup by exact ID
		}
		sourceRuleID := extractRuleIDFromEffectID(de.ID)
		if sourceRuleID != "" {
			isRepeatable, known := ruleRepeatable[sourceRuleID]
			if known && !isRepeatable && e.hasExistingEffectFromRule(sourceRuleID, state) {
				continue // non-repeatable rule already has an effect in state
			}
		}
		de.Applied = false
		de.TriggerSession = eval.SessionNum
		state.PendingEffects = append(state.PendingEffects, de)
		existingIDs[de.ID] = true
	}

	return nil
}

// GetPendingEffects returns all pending effects due for the given session number,
// sorted by ApplySession ascending.
func (e *ConsequenceEngine) GetPendingEffects(state *domain.NarrativeState, sessionNum int) []domain.DelayedEffect {
	if state == nil {
		return nil
	}

	var result []domain.DelayedEffect
	for _, de := range state.PendingEffects {
		if !de.Applied && de.ApplySession <= sessionNum {
			result = append(result, de)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ApplySession < result[j].ApplySession
	})

	return result
}

// hasExistingEffectFromRule checks if the state already contains any pending
// effect derived from the given rule ID.
func (e *ConsequenceEngine) hasExistingEffectFromRule(ruleID string, state *domain.NarrativeState) bool {
	for _, pe := range state.PendingEffects {
		if extractRuleIDFromEffectID(pe.ID) == ruleID {
			return true
		}
	}
	return false
}

func extractRuleIDFromEffectID(effectID string) string {
	// Format: {ruleID}-{idx}
	for i := len(effectID) - 1; i >= 0; i-- {
		if effectID[i] == '-' {
			return effectID[:i]
		}
	}
	return ""
}

func makeDelayedEffectID(ruleID string, effectIndex int) string {
	return fmt.Sprintf("%s-%d", ruleID, effectIndex)
}

func parseDelaySessions(delay string) int {
	// Simple parser: "2 sessions" → 2
	var sessions int
	if _, err := fmt.Sscanf(delay, "%d", &sessions); err != nil {
		return 1
	}
	if sessions <= 0 {
		return 1
	}
	return sessions
}
