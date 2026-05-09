package domain

import (
	"errors"
	"fmt"
)

// ConsequenceTable tracks narrative state changes between acts.
type ConsequenceTable struct {
	ID                  string            `json:"id"`
	CampaignID          string            `json:"campaign_id"`
	FromAct             int               `json:"from_act"`
	ToAct               int               `json:"to_act"`
	QuestOutcomes       []QuestOutcome    `json:"quest_outcomes"`
	FactionChanges      []FactionChange   `json:"faction_changes"`
	NPCChanges          []NPCChange       `json:"npc_changes"`
	NewOpportunities    []Opportunity     `json:"new_opportunities"`
	LockedContent       []LockedContent   `json:"locked_content"`
	WorldStateChanges   []WorldStateChange `json:"world_state_changes"`
}

// QuestOutcome represents the outcome of a quest and its consequences.
type QuestOutcome struct {
	QuestID           string             `json:"quest_id"`
	Outcome           string             `json:"outcome"` // success, failure, partial, abandoned
	Consequences      []string           `json:"consequences"`
	ReputationChanges []ReputationChange `json:"reputation_changes"`
}

// ReputationChange represents a faction reputation change.
type ReputationChange struct {
	FactionID string `json:"faction_id"`
	Delta     int    `json:"delta"` // -100 to 100
	Reason    string `json:"reason"`
}

// FactionChange represents a faction's reputation change with status.
type FactionChange struct {
	FactionID     string `json:"faction_id"`
	PreviousRep   int    `json:"previous_rep"`
	NewRep        int    `json:"new_rep"`
	Delta         int    `json:"delta"`
	Reason        string `json:"reason"`
	StatusChange  string `json:"status_change,omitempty"` // neutral -> friendly, etc.
}

// NPCChange represents an NPC status change.
type NPCChange struct {
	NPCID           string   `json:"npc_id"`
	PreviousStatus  string   `json:"previous_status"` // alive, neutral, friendly, hostile, dead
	NewStatus       string   `json:"new_status"`
	Reason          string   `json:"reason"`
	ImpactOnQuests  []string `json:"impact_on_quests"`
}

// Opportunity represents a newly unlocked opportunity.
type Opportunity struct {
	Type            string `json:"type"` // quest, location, npc, item
	ReferenceID     string `json:"reference_id"`
	UnlockCondition string `json:"unlock_condition"`
	Description     string `json:"description"`
}

// LockedContent represents content that is now locked.
type LockedContent struct {
	Type            string `json:"type"` // quest, location, npc, area
	ReferenceID     string `json:"reference_id"`
	LockReason      string `json:"lock_reason"`
	Unlockable      bool   `json:"unlockable"`
	UnlockCondition string `json:"unlock_condition,omitempty"`
}

// WorldStateChange represents a change to the world state.
type WorldStateChange struct {
	Description     string `json:"description"`
	Scope           string `json:"scope"` // local, regional, global
	Permanent       bool   `json:"permanent"`
	VisibleToPlayers bool  `json:"visible_to_players"`
}

// Validate checks consequence table validity.
func (c *ConsequenceTable) Validate() error {
	if c.ID == "" {
		return errors.New("id is required")
	}
	if c.CampaignID == "" {
		return errors.New("campaign_id is required")
	}
	if c.FromAct < 1 {
		return errors.New("from_act must be at least 1")
	}
	if c.ToAct <= c.FromAct {
		return errors.New("to_act must be greater than from_act")
	}
	// Validate faction reputation bounds
	for _, fc := range c.FactionChanges {
		if fc.NewRep < -100 || fc.NewRep > 100 {
			return fmt.Errorf("faction %s: reputation %d out of bounds (-100 to 100)", fc.FactionID, fc.NewRep)
		}
		if fc.Delta < -100 || fc.Delta > 100 {
			return fmt.Errorf("faction %s: delta %d out of bounds (-100 to 100)", fc.FactionID, fc.Delta)
		}
	}
	// Validate quest outcomes
	for _, qo := range c.QuestOutcomes {
		if qo.QuestID == "" {
			return errors.New("quest_id is required for quest outcome")
		}
		if !isValidQuestOutcome(qo.Outcome) {
			return fmt.Errorf("invalid quest outcome: %s", qo.Outcome)
		}
	}
	return nil
}

// isValidQuestOutcome checks if an outcome is valid.
func isValidQuestOutcome(outcome string) bool {
	switch outcome {
	case "success", "failure", "partial", "abandoned":
		return true
	default:
		return false
	}
}

// GetNetFactionChange calculates net reputation change for a faction.
func (c *ConsequenceTable) GetNetFactionChange(factionID string) int {
	total := 0
	for _, fc := range c.FactionChanges {
		if fc.FactionID == factionID {
			total += fc.Delta
		}
	}
	return total
}

// GetQuestOutcome retrieves outcome for a specific quest.
func (c *ConsequenceTable) GetQuestOutcome(questID string) *QuestOutcome {
	for _, qo := range c.QuestOutcomes {
		if qo.QuestID == questID {
			return &qo
		}
	}
	return nil
}

// HasNPCChange checks if an NPC has a status change.
func (c *ConsequenceTable) HasNPCChange(npcID string) bool {
	for _, nc := range c.NPCChanges {
		if nc.NPCID == npcID {
			return true
		}
	}
	return false
}
