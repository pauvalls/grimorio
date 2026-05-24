package domain

import "time"

// NarrativeState tracks the mutable session state of a campaign
type NarrativeState struct {
	// Core metadata
	SchemaVersion  string    `json:"schema_version"`
	CampaignID     string    `json:"campaign_id"`
	CurrentSession int       `json:"current_session"`
	LastUpdated    time.Time `json:"last_updated"`

	// Mutable state (serialized in logical order)
	RevealedClues     []RevealedClue   `json:"revealed_clues"`
	ActiveQuests      []QuestState     `json:"-"` // Internal objects
	QuestNames        []string         `json:"active_quests,omitempty"`
	CompletedQuests   []QuestState     `json:"-"` // Internal objects
	CompletedQuestIDs []string         `json:"completed_quests"`
	FailedQuests      []QuestState     `json:"-"` // Internal objects
	FailedQuestIDs    []string         `json:"failed_quests"`
	DeadNPCs          []NPCDeathRecord `json:"dead_npcs"`
	KeyItems          []KeyItem        `json:"-"` // Internal objects
	ItemNames         []string         `json:"key_items,omitempty"`
	LootAcquired      []string         `json:"loot_acquired,omitempty"`
	DMNotes           string           `json:"dm_notes,omitempty"`
	CurrentLocation   string           `json:"current_location,omitempty"`
	PCStatuses        []PCStatus       `json:"pc_status,omitempty"`
	SessionLog        []SessionRecord  `json:"session_log"`
	DMOverrides       []DMOverride     `json:"dm_overrides"`
}

// Validate checks if the narrative state is valid
func (n *NarrativeState) Validate() error {
	if n.SchemaVersion == "" {
		return NewValidationError("schema_version", "schema version is required")
	}
	if n.SchemaVersion != SchemaVersionV2 {
		return NewValidationError("schema_version", "unsupported schema version: "+n.SchemaVersion)
	}
	if n.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if n.CurrentSession < 0 {
		return NewValidationError("current_session", "current session cannot be negative")
	}
	return nil
}

// RevealedClue tracks a clue that has been discovered by the party
type RevealedClue struct {
	ID              string   `json:"id"`
	Description     string   `json:"description"`
	SourceAct       string   `json:"source_act"`
	SourceArea      string   `json:"source_area,omitempty"`
	SessionRevealed int      `json:"session_revealed"`
	IsCritical      bool     `json:"is_critical"`
	Prerequisites   []string `json:"prerequisites,omitempty"`
}

// Validate checks if the revealed clue is valid
func (r *RevealedClue) Validate() error {
	if r.ID == "" {
		return NewValidationError("id", "clue ID is required")
	}
	if r.Description == "" {
		return NewValidationError("description", "clue description is required")
	}
	if r.SourceAct == "" {
		return NewValidationError("source_act", "source act is required")
	}
	if r.SessionRevealed < 0 {
		return NewValidationError("session_revealed", "session revealed cannot be negative")
	}
	return nil
}

// QuestState tracks the state of a quest in the narrative
type QuestState struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	SourceAct     string   `json:"source_act"`
	GiverNPC      string   `json:"giver_npc"`
	RewardClaimed bool     `json:"reward_claimed"`
	Consequences  []string `json:"consequences"`
}

// Validate checks if the quest state is valid
func (q *QuestState) Validate() error {
	if q.ID == "" {
		return NewValidationError("id", "quest ID is required")
	}
	if q.Name == "" {
		return NewValidationError("name", "quest name is required")
	}
	if q.Status == "" {
		return NewValidationError("status", "quest status is required")
	}
	switch q.Status {
	case "active", "completed", "failed", "abandoned":
		// valid
	default:
		return NewValidationError("status", "invalid quest status: "+q.Status)
	}
	return nil
}

// NPCDeathRecord tracks the death of an NPC
type NPCDeathRecord struct {
	NPCID    string `json:"npc_id"`
	Name     string `json:"name"`
	Session  int    `json:"session"`
	Cause    string `json:"cause,omitempty"`
	KilledBy string `json:"killed_by,omitempty"`
	Location string `json:"location,omitempty"`
}

// Validate checks if the death record is valid
func (d *NPCDeathRecord) Validate() error {
	if d.NPCID == "" {
		return NewValidationError("npc_id", "NPC ID is required")
	}
	if d.Name == "" {
		return NewValidationError("name", "NPC name is required")
	}
	if d.Session < 0 {
		return NewValidationError("session", "session cannot be negative")
	}
	return nil
}

// KeyItem tracks an important item in the campaign
type KeyItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Holder       string `json:"holder"`
	SessionFound int    `json:"session_found"`
	IsMcGuffin   bool   `json:"is_mcguffin"`
}

// Validate checks if the key item is valid
func (k *KeyItem) Validate() error {
	if k.ID == "" {
		return NewValidationError("id", "item ID is required")
	}
	if k.Name == "" {
		return NewValidationError("name", "item name is required")
	}
	if k.Holder == "" {
		return NewValidationError("holder", "item holder is required")
	}
	if k.SessionFound < 0 {
		return NewValidationError("session_found", "session found cannot be negative")
	}
	return nil
}

// SessionRecord records what happened in a session
type SessionRecord struct {
	SessionNum   int        `json:"session_num"`
	Date         time.Time  `json:"date"`
	Summary      string     `json:"summary"`
	KeyDecisions []Decision `json:"key_decisions"`
	XPAwarded    int        `json:"xp_awarded"`
	LootAcquired []string   `json:"loot_acquired"`
	DMNotes      string     `json:"dm_notes"`
}

// Validate checks if the session record is valid
func (s *SessionRecord) Validate() error {
	if s.SessionNum < 0 {
		return NewValidationError("session_num", "session number cannot be negative")
	}
	return nil
}

// Decision tracks a key decision made during a session
type Decision struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	ChoiceMade  string `json:"choice_made"`
	ImpactScope string `json:"impact_scope"`
}

// Validate checks if the decision is valid
func (d *Decision) Validate() error {
	if d.ID == "" {
		return NewValidationError("id", "decision ID is required")
	}
	if d.Description == "" {
		return NewValidationError("description", "decision description is required")
	}
	return nil
}

// DMOverride allows the DM to override canonical facts
type DMOverride struct {
	ID          string    `json:"id"`
	TargetType  string    `json:"target_type"`
	TargetID    string    `json:"target_id"`
	Field       string    `json:"field"`
	NewValue    string    `json:"new_value"`
	Reason      string    `json:"reason"`
	SessionNum  int       `json:"session_num"`
	CreatedAt   time.Time `json:"created_at"`
}

// Validate checks if the DM override is valid
func (d *DMOverride) Validate() error {
	if d.ID == "" {
		return NewValidationError("id", "override ID is required")
	}
	if d.TargetID == "" {
		return NewValidationError("target_id", "target ID is required")
	}
	if d.Field == "" {
		return NewValidationError("field", "field is required")
	}
	return nil
}

// StateUpdate represents a batch update to narrative state
type StateUpdate struct {
	SessionNum       int              `json:"session_num"`
	RevealedClues    []RevealedClue   `json:"revealed_clues,omitempty"`
	DeadNPCs         []NPCDeathRecord `json:"dead_npcs,omitempty"`
	CompletedQuests  []string         `json:"completed_quests,omitempty"`
	NewQuests        []QuestState     `json:"new_quests,omitempty"`
	KeyItems         []KeyItem        `json:"key_items,omitempty"`
	KeyDecisions     []Decision       `json:"key_decisions,omitempty"`
	XPAwarded        int              `json:"xp_awarded,omitempty"`
	LootAcquired     []string         `json:"loot_acquired,omitempty"`
	SessionSummary   string           `json:"session_summary,omitempty"`
	DMNotes          string           `json:"dm_notes,omitempty"`
	CurrentLocation  string           `json:"current_location,omitempty"`
	PCStatuses       []PCStatus       `json:"pc_status,omitempty"`
	ReplaceSession   bool             `json:"replace_session,omitempty"` // If true, replace existing session log entry
}

// SessionPrepContext provides context for preparing the next session
type SessionPrepContext struct {
	PreviouslyOn  string       `json:"previously_on"`
	ActiveQuests  []QuestState `json:"active_quests"`
	PendingHooks  []string     `json:"pending_hooks"`
	RelevantNPCs  []CanonEntity `json:"relevant_npcs"`
	WorldChanges  []string     `json:"world_changes"`
	CriticalPaths []string     `json:"critical_paths"`
	DMWarnings    []string     `json:"dm_warnings"`
}

// FactionReputationMatrix tracks reputation with factions
type FactionReputationMatrix struct {
	CampaignID string              `json:"campaign_id"`
	Entries    []ReputationEntry   `json:"entries"`
}

// ReputationEntry tracks party reputation with a faction
type ReputationEntry struct {
	FactionID     string            `json:"faction_id"`
	PartyID       string            `json:"party_id"`
	Score         int8              `json:"score"`
	Status        string            `json:"status"`
	History       []ReputationEvent `json:"history"`
	UnlockedPerks []string          `json:"unlocked_perks"`
}

// ReputationEvent records a change in reputation
type ReputationEvent struct {
	Session    int       `json:"session"`
	Delta      int8      `json:"delta"`
	Reason     string    `json:"reason"`
	ActionType string    `json:"action_type"`
}

// WorldEvent represents a world-level event triggered by player actions
type WorldEvent struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Scope       string `json:"scope"`
	SessionNum  int    `json:"session_num"`
	TriggerType string `json:"trigger_type"`
}

// PCStatus tracks the health and conditions of a player character
type PCStatus struct {
	Name       string   `json:"name"`
	HPCurrent  int      `json:"hp_current"`
	HPMax      int      `json:"hp_max"`
	Conditions []string `json:"conditions,omitempty"`
}
