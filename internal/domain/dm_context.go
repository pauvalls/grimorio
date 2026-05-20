package domain

import (
	"fmt"
	"time"
)

// DMContextPayload aggregates all campaign data for the AI Dungeon Master.
type DMContextPayload struct {
	CampaignID     string                    `json:"campaign_id"`
	SessionNum     int                       `json:"session_num"`
	GeneratedAt    time.Time                 `json:"generated_at"`
	Canon          *CanonContext             `json:"canon"`
	NarrativeState *NarrativeContext         `json:"narrative_state"`
	SessionPrep    *DMContextSessionPrep     `json:"session_prep"`
	Characters     []CharacterContext        `json:"characters"`
	Areas          map[string]AreaContext    `json:"areas"`
	NPCs           map[string]NPCContext     `json:"npcs"`
	Bestiary       map[string]MonsterContext `json:"bestiary"`
	Prologue       *PrologueContext          `json:"prologue,omitempty"`
	Factions       map[string]FactionContext `json:"factions"`
	Quests         []QuestContext            `json:"quests"`
	PDFAvailable   bool                      `json:"pdf_available"`
	PDFPath        string                    `json:"pdf_path,omitempty"`
	PDFText        string                    `json:"pdf_text,omitempty"`
	DMNotes        DMNotesContext            `json:"dm_notes"`
}

// Validate checks if the DM context payload is valid.
func (d *DMContextPayload) Validate() error {
	if d.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if !IsValidKebabCase(d.CampaignID) {
		return NewValidationError("campaign_id", "campaign ID must be kebab-case")
	}
	if d.SessionNum < 1 {
		return NewValidationError("session_num", "session number must be at least 1")
	}
	return nil
}

// CanonContext provides the canonical source of truth for a campaign.
type CanonContext struct {
	Facts         []CanonFact         `json:"facts"`
	Entities      []CanonEntity       `json:"entities"`
	Timeline      []CanonTimelineEvent `json:"timeline"`
	Rules         []CanonRule         `json:"rules"`
	Relationships []CanonRelationship `json:"relationships"`
}

// Validate checks if the canon context is valid.
func (c *CanonContext) Validate() error {
	for i, f := range c.Facts {
		if err := f.Validate(); err != nil {
			return NewValidationError(fmt.Sprintf("facts[%d]", i), err.Error())
		}
	}
	for i, e := range c.Entities {
		if err := e.Validate(); err != nil {
			return NewValidationError(fmt.Sprintf("entities[%d]", i), err.Error())
		}
	}
	return nil
}

// NarrativeContext tracks the mutable session state of a campaign.
type NarrativeContext struct {
	CurrentSession  int              `json:"current_session"`
	RevealedClues   []RevealedClue   `json:"revealed_clues"`
	ActiveQuests    []QuestState     `json:"active_quests"`
	CompletedQuests []QuestState     `json:"completed_quests"`
	FailedQuests    []QuestState     `json:"failed_quests"`
	DeadNPCs        []NPCDeathRecord `json:"dead_npcs"`
	KeyItems        []KeyItem        `json:"key_items"`
	SessionLog      []SessionRecord  `json:"session_log"`
}

// Validate checks if the narrative context is valid.
func (n *NarrativeContext) Validate() error {
	if n.CurrentSession < 0 {
		return NewValidationError("current_session", "current session cannot be negative")
	}
	return nil
}

// DMContextSessionPrep provides context for the upcoming session within the DM context payload.
type DMContextSessionPrep struct {
	PreviouslyOn    string   `json:"previously_on"`
	ActiveQuests    []string `json:"active_quests"`
	RelevantNPCs    []string `json:"relevant_npcs"`
	Reminders       []string `json:"reminders"`
	LikelyScenarios []string `json:"likely_scenarios"`
}

// CharacterContext provides a lightweight view of a player character.
type CharacterContext struct {
	Name       string `json:"name"`
	Race       string `json:"race"`
	Class      string `json:"class"`
	Level      int    `json:"level"`
	Background string `json:"background"`
	Alignment  string `json:"alignment"`
	HP         HP     `json:"hp"`
	AC         int    `json:"ac"`
	Stats      Stats  `json:"stats"`
}

// Validate checks if the character context is valid.
func (c *CharacterContext) Validate() error {
	if c.Name == "" {
		return NewValidationError("name", "character name is required")
	}
	if c.Level < 1 || c.Level > 20 {
		return NewValidationError("level", "level must be between 1 and 20")
	}
	return nil
}

// AreaContext provides a lightweight view of a playable area.
type AreaContext struct {
	ID              string          `json:"id"`
	ChapterID       string          `json:"chapter_id"`
	AreaNumber      int             `json:"area_number"`
	Title           string          `json:"title"`
	Summary         string          `json:"summary"`
	PlayerReadAloud string          `json:"player_read_aloud,omitempty"`
	Encounters      []AreaEncounter `json:"encounters"`
	NPCs            []AreaNPC       `json:"npcs"`
	Treasure        []Treasure      `json:"treasure"`
}

// Validate checks if the area context is valid.
func (a *AreaContext) Validate() error {
	if a.ID == "" {
		return NewValidationError("id", "area ID is required")
	}
	if a.AreaNumber < 1 || a.AreaNumber > 15 {
		return NewValidationError("area_number", "area number must be between 1 and 15")
	}
	return nil
}

// NPCContext provides narrative and tactical context for an NPC.
type NPCContext struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Motivation    string   `json:"motivation"`
	Secret        string   `json:"secret,omitempty"`
	Faction       string   `json:"faction,omitempty"`
	DialogueVoice string   `json:"dialogue_voice"`
	Personality   []string `json:"personality_traits"`
	Stats         NPCStats `json:"stats,omitempty"`
	Tactics       string   `json:"tactics,omitempty"`
}

// Validate checks if the NPC context is valid.
func (n *NPCContext) Validate() error {
	if n.Name == "" {
		return NewValidationError("name", "NPC name is required")
	}
	return nil
}

// NPCStats provides combat-relevant stats for an NPC.
type NPCStats struct {
	HP int `json:"hp"`
	AC int `json:"ac"`
}

// MonsterContext provides combat-relevant context for a bestiary entry.
type MonsterContext struct {
	Name            string            `json:"name"`
	CR              string            `json:"cr"`
	AC              int               `json:"ac"`
	HP              int               `json:"hp"`
	Tactics         string            `json:"tactics"`
	DescriptiveCues map[string]string `json:"descriptive_cues"`
}

// Validate checks if the monster context is valid.
func (m *MonsterContext) Validate() error {
	if m.Name == "" {
		return NewValidationError("name", "monster name is required")
	}
	if m.CR == "" {
		return NewValidationError("cr", "challenge rating is required")
	}
	return nil
}

// ExpectedDescriptiveCueKeys are the standard HP threshold cue keys.
var ExpectedDescriptiveCueKeys = []string{"full_hp", "half_hp", "low_hp", "defeated"}

// HasAllDescriptiveCues checks if the monster has all expected cue keys.
func (m MonsterContext) HasAllDescriptiveCues() bool {
	for _, key := range ExpectedDescriptiveCueKeys {
		if _, ok := m.DescriptiveCues[key]; !ok {
			return false
		}
	}
	return true
}

// PrologueContext provides the narrative prologue for a campaign.
type PrologueContext struct {
	Tone  string                `json:"tone"`
	Parts []ProloguePartContext `json:"parts"`
}

// ProloguePartContext represents a single section of a narrative prologue.
type ProloguePartContext struct {
	Order       int    `json:"order"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsReadAloud bool   `json:"is_read_aloud"`
}

// FactionContext provides reputation and attitude context for a faction.
type FactionContext struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Reputation int8   `json:"reputation"`
	Status     string `json:"status"`
	Attitude   string `json:"attitude"`
}

// QuestContext provides a lightweight view of a quest.
type QuestContext struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Type   string `json:"type"`
	Giver  string `json:"giver"`
}

// Validate checks if the quest context is valid.
func (q *QuestContext) Validate() error {
	if q.ID == "" {
		return NewValidationError("id", "quest ID is required")
	}
	if q.Title == "" {
		return NewValidationError("title", "quest title is required")
	}
	return nil
}

// DMNotesContext provides DM-facing warnings and reminders.
type DMNotesContext struct {
	Warnings  []string `json:"warnings"`
	Reminders []string `json:"reminders"`
}
