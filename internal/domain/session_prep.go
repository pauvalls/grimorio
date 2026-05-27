package domain

import "time"

// SessionPrep is a synthesized prep sheet for the next session
type SessionPrep struct {
	CampaignID               string                    `json:"campaign_id"`
	SessionNum               int                       `json:"session_num"`
	PreviouslyOn             string                    `json:"previously_on"`
	LikelyScenarios          []string                  `json:"likely_scenarios"`
	RelevantNPCs             []string                  `json:"relevant_npcs"`
	ActiveQuests             []string                  `json:"active_quests"`
	Reminders                []string                  `json:"reminders"`
	PrepDate                 time.Time                 `json:"prep_date"`
	EncounterRecommendations []EncounterRecommendation `json:"encounter_recommendations,omitempty"`
	LootSuggestions          []LootSuggestion          `json:"loot_suggestions,omitempty"`
	NPCAppearances           []NPCAppearance           `json:"npc_appearances,omitempty"`
	PendingEffects           []DelayedEffect           `json:"pending_effects,omitempty"`
	FactionSnapshot          []ReputationEntry         `json:"faction_snapshot,omitempty"`
}

// EncounterRecommendation represents a recommended encounter for a session.
type EncounterRecommendation struct {
	Name        string   `json:"name"`
	CR          string   `json:"cr"`
	Type        string   `json:"type"` // combat, social, exploration, mixed
	Description string   `json:"description"`
	Context     string   `json:"context,omitempty"` // Narrative context for this encounter
}

// LootSuggestion represents a suggested loot reward.
type LootSuggestion struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // weapon, armor, consumable, magical, gold
	Rarity      string `json:"rarity"` // common, uncommon, rare, very_rare, legendary
	Description string `json:"description"`
	Context     string `json:"context,omitempty"` // Why this loot is appropriate
}

// NPCAppearance represents an NPC that may appear in a session.
type NPCAppearance struct {
	NPCID       string `json:"npc_id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Context     string `json:"context"` // Why/how they appear in this session
	Importance  string `json:"importance"` // major, minor, cameo
}

// FlowchartNode represents a node in a campaign flowchart
type FlowchartNode struct {
	ID           string   `json:"id"`
	Label        string   `json:"label"`
	Type         string   `json:"type"` // act, decision, event
	Dependencies []string `json:"dependencies"`
}

// RosterNPC represents an NPC entry in the adventure roster
type RosterNPC struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Location string `json:"location"`
	Act      string `json:"act"`
	PageRef  string `json:"page_ref"`
}

// RosterMonster represents a monster entry in the adventure roster
type RosterMonster struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	CR        string   `json:"cr"`
	Locations []string `json:"locations"`
	Act       string   `json:"act"`
	PageRef   string   `json:"page_ref"`
}

// RosterEncounter represents an encounter entry in the adventure roster
type RosterEncounter struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Act     string `json:"act"`
	Area    string `json:"area"`
	PageRef string `json:"page_ref"`
}

// AdventureRoster is a master index of all campaign entities
type AdventureRoster struct {
	CampaignID string            `json:"campaign_id"`
	NPCs       []RosterNPC       `json:"npcs"`
	Monsters   []RosterMonster   `json:"monsters"`
	Encounters []RosterEncounter `json:"encounters"`
}

// CharacterHook is a personalized plot hook for a PC
type CharacterHook struct {
	CharacterID      string `json:"character_id"`
	CharacterName    string `json:"character_name"`
	Background       string `json:"background"`
	Class            string `json:"class"`
	Hook             string `json:"hook"`
	ConnectionToPlot string `json:"connection_to_plot"`
}
