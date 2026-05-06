package domain

import (
	"fmt"
	"time"
)

// EventType represents the type of session event
type EventType string

const (
	EventTypeAction       EventType = "action"
	EventTypeDiceRoll     EventType = "dice_roll"
	EventTypeNarration    EventType = "narration"
	EventTypeCombatStart  EventType = "combat_start"
	EventTypeCombatEnd    EventType = "combat_end"
	EventTypeHPChange     EventType = "hp_change"
	EventTypeCondition    EventType = "condition"
	EventTypeTurnChange   EventType = "turn_change"
	EventTypeSceneChange  EventType = "scene_change"
)

// ConditionType represents a condition that can affect a character
type ConditionType string

const (
	ConditionBlinded      ConditionType = "blinded"
	ConditionCharmed      ConditionType = "charmed"
	ConditionDeafened     ConditionType = "deafened"
	ConditionFrightened   ConditionType = "frightened"
	ConditionGrappled     ConditionType = "grappled"
	ConditionIncapacitated ConditionType = "incapacitated"
	ConditionInvisible    ConditionType = "invisible"
	ConditionParalyzed    ConditionType = "paralyzed"
	ConditionPetrified    ConditionType = "petrified"
	ConditionPoisoned     ConditionType = "poisoned"
	ConditionProne        ConditionType = "prone"
	ConditionRestrained   ConditionType = "restrained"
	ConditionStunned      ConditionType = "stunned"
	ConditionUnconscious  ConditionType = "unconscious"
)

// Point represents a 2D position
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// PlayerState represents the current state of a player character in a session
type PlayerState struct {
	CharacterID string        `json:"character_id"`
	CurrentHP   int           `json:"current_hp"`
	MaxHP       int           `json:"max_hp"`
	TempHP      int           `json:"temp_hp"`
	AC          int           `json:"ac"`
	Position    Point         `json:"position"`
	Initiative  int           `json:"initiative"`
	IsActive    bool          `json:"is_active"`
	Conditions  []Condition   `json:"conditions"`
}

// Condition represents a temporary condition affecting a character
type Condition struct {
	Type      ConditionType `json:"type"`
	Duration  int           `json:"duration"`  // in rounds, 0 = permanent
	AppliedAt time.Time     `json:"applied_at"`
	Source    string        `json:"source"`    // who/what applied it
}

// IsExpired checks if the condition has expired given the current round
func (c Condition) IsExpired(round int) bool {
	if c.Duration == 0 {
		return false // permanent
	}
	return round >= c.Duration
}

// ApplyDamage applies damage to the player, considering temp HP
func (p *PlayerState) ApplyDamage(damage int) {
	if damage <= 0 {
		return
	}

	// Apply to temp HP first
	if p.TempHP > 0 {
		if damage <= p.TempHP {
			p.TempHP -= damage
			return
		}
		damage -= p.TempHP
		p.TempHP = 0
	}

	// Apply to current HP
	p.CurrentHP -= damage
	if p.CurrentHP < 0 {
		p.CurrentHP = 0
	}
}

// Heal restores HP up to max
func (p *PlayerState) Heal(amount int) {
	if amount <= 0 {
		return
	}
	p.CurrentHP += amount
	if p.CurrentHP > p.MaxHP {
		p.CurrentHP = p.MaxHP
	}
}

// IsAlive returns true if the player has HP > 0
func (p *PlayerState) IsAlive() bool {
	return p.CurrentHP > 0
}

// AddCondition adds a condition to the player
func (p *PlayerState) AddCondition(c Condition) {
	p.Conditions = append(p.Conditions, c)
}

// RemoveCondition removes a condition by type
func (p *PlayerState) RemoveCondition(conditionType ConditionType) {
	var filtered []Condition
	for _, c := range p.Conditions {
		if c.Type != conditionType {
			filtered = append(filtered, c)
		}
	}
	p.Conditions = filtered
}

// HasCondition checks if the player has a specific condition
func (p *PlayerState) HasCondition(conditionType ConditionType) bool {
	for _, c := range p.Conditions {
		if c.Type == conditionType {
			return true
		}
	}
	return false
}

// Session represents an active game session
type Session struct {
	ID           string         `json:"id"`
	CampaignID   string         `json:"campaign_id"`
	Players      []PlayerState  `json:"players"`
	CurrentScene *Scene         `json:"current_scene,omitempty"`
	InCombat     bool           `json:"in_combat"`
	StartedAt    time.Time      `json:"started_at"`
	EndedAt      *time.Time     `json:"ended_at,omitempty"`
}

// Validate checks if the session is valid
func (s *Session) Validate() error {
	if s.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if len(s.Players) == 0 {
		return NewValidationError("players", "at least one player is required")
	}
	for i, player := range s.Players {
		if player.CharacterID == "" {
			return NewValidationError("players", fmt.Sprintf("player %d missing character ID", i))
		}
		if player.CurrentHP < 0 {
			return NewValidationError("players", fmt.Sprintf("player %s has negative HP", player.CharacterID))
		}
		if player.CurrentHP > player.MaxHP {
			return NewValidationError("players", fmt.Sprintf("player %s current HP exceeds max HP", player.CharacterID))
		}
	}
	return nil
}

// GetActivePlayer returns the player whose turn it is (in combat)
func (s *Session) GetActivePlayer() *PlayerState {
	for i := range s.Players {
		if s.Players[i].IsActive {
			return &s.Players[i]
		}
	}
	return nil
}

// GetPlayer returns a player by character ID
func (s *Session) GetPlayer(characterID string) *PlayerState {
	for i := range s.Players {
		if s.Players[i].CharacterID == characterID {
			return &s.Players[i]
		}
	}
	return nil
}

// IsPlayerAlive checks if a player is alive
func (s *Session) IsPlayerAlive(characterID string) bool {
	player := s.GetPlayer(characterID)
	if player == nil {
		return false
	}
	return player.IsAlive()
}

// SessionEvent represents an event that occurred during a session
type SessionEvent struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Type      EventType `json:"type"`
	Actor     string    `json:"actor"`
	Content   string    `json:"content"`
	Result    JSON      `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// Validate checks if the event is valid
func (e *SessionEvent) Validate() error {
	if e.SessionID == "" {
		return NewValidationError("session_id", "session ID is required")
	}
	if e.Type == "" {
		return NewValidationError("type", "event type is required")
	}
	switch e.Type {
	case EventTypeAction, EventTypeDiceRoll, EventTypeNarration,
		EventTypeCombatStart, EventTypeCombatEnd, EventTypeHPChange,
		EventTypeCondition, EventTypeTurnChange, EventTypeSceneChange:
		// valid
	default:
		return NewValidationError("type", "invalid event type: "+string(e.Type))
	}
	if e.Actor == "" {
		return NewValidationError("actor", "actor is required")
	}
	return nil
}

// CombatState represents the current state of combat
type CombatState struct {
	SessionID       string   `json:"session_id"`
	Round           int      `json:"round"`
	InitiativeOrder []string `json:"initiative_order"`
	ActiveIndex     int      `json:"active_index"`
	MapID           string   `json:"map_id,omitempty"`
}

// Validate checks if the combat state is valid
func (c *CombatState) Validate() error {
	if c.SessionID == "" {
		return NewValidationError("session_id", "session ID is required")
	}
	if len(c.InitiativeOrder) == 0 {
		return NewValidationError("initiative_order", "initiative order cannot be empty")
	}
	if c.ActiveIndex < 0 || c.ActiveIndex >= len(c.InitiativeOrder) {
		return NewValidationError("active_index", "active index out of bounds")
	}
	if c.Round < 1 {
		return NewValidationError("round", "round must be at least 1")
	}
	return nil
}

// GetActiveActor returns the ID of the actor whose turn it is
func (c *CombatState) GetActiveActor() string {
	if c.ActiveIndex < 0 || c.ActiveIndex >= len(c.InitiativeOrder) {
		return ""
	}
	return c.InitiativeOrder[c.ActiveIndex]
}

// NextTurn advances to the next turn in the initiative order
func (c *CombatState) NextTurn() {
	c.ActiveIndex++
	if c.ActiveIndex >= len(c.InitiativeOrder) {
		c.ActiveIndex = 0
		c.Round++
	}
}

// IsActorTurn checks if it's the given actor's turn
func (c *CombatState) IsActorTurn(actorID string) bool {
	return c.GetActiveActor() == actorID
}

// DiceResult represents the result of a dice roll
type DiceResult struct {
	Dice     string `json:"dice"`
	Results  []int  `json:"results"`
	Total    int    `json:"total"`
	Modifier int    `json:"modifier"`
}

// ActionResult represents the result of a player action
type ActionResult struct {
	Success     bool        `json:"success"`
	Description string      `json:"description"`
	DiceRolls   []DiceResult  `json:"dice_rolls,omitempty"`
	Damage      int         `json:"damage,omitempty"`
	StateChange StateChange `json:"state_change,omitempty"`
}

// StateChange represents a change to the game state
type StateChange struct {
	TargetID    string `json:"target_id,omitempty"`
	HPChange    int    `json:"hp_change,omitempty"`
	NewHP       int    `json:"new_hp,omitempty"`
	Condition   string `json:"condition,omitempty"`
	Position    *Point `json:"position,omitempty"`
}

// GameState represents the complete current state of a game session
type GameState struct {
	SessionID    string         `json:"session_id"`
	CampaignID   string         `json:"campaign_id"`
	InCombat     bool           `json:"in_combat"`
	CurrentScene *Scene         `json:"current_scene,omitempty"`
	Players      []PlayerState  `json:"players"`
	Combat       *CombatState   `json:"combat,omitempty"`
	ActiveActor  string         `json:"active_actor,omitempty"`
	Round        int            `json:"round,omitempty"`
}

// SessionSummary represents a summary of a completed session
type SessionSummary struct {
	SessionID     string    `json:"session_id"`
	CampaignID    string    `json:"campaign_id"`
	Duration      string    `json:"duration"`
	RoundsPlayed  int       `json:"rounds_played"`
	EventsCount   int       `json:"events_count"`
	PlayersAlive  int       `json:"players_alive"`
	PlayersDead   int       `json:"players_dead"`
	KeyEvents     []string  `json:"key_events"`
	EndedAt       time.Time `json:"ended_at"`
}

// JSON is a helper type for JSON fields
type JSON map[string]interface{}
