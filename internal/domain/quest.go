package domain

import "time"

// QuestType represents the type of quest
type QuestType string

const (
	QuestTypeMain           QuestType = "main"
	QuestTypeSide           QuestType = "side"
	QuestTypePersonal       QuestType = "personal"
	QuestTypeGroup          QuestType = "group"
	QuestTypeRedencion      QuestType = "redencion"
	QuestTypeVenganza       QuestType = "venganza"
	QuestTypeDescubrimiento QuestType = "descubrimiento"
	QuestTypeProteccion     QuestType = "proteccion"
)

// QuestStatus represents the status of a quest
type QuestStatus string

const (
	QuestStatusActive    QuestStatus = "active"
	QuestStatusCompleted QuestStatus = "completed"
	QuestStatusFailed    QuestStatus = "failed"
	QuestStatusOnHold    QuestStatus = "on_hold"
)

// Quest represents a mission or objective
type Quest struct {
	ID            string         `json:"id"`
	CampaignID    string         `json:"campaign_id"`
	Title         string         `json:"title"`
	Type          QuestType      `json:"type"`
	Status        QuestStatus    `json:"status"`
	Hook          string         `json:"hook"` // How it's introduced
	Description   string         `json:"description"`
	Objectives    []Objective    `json:"objectives"`
	Stakes        string         `json:"stakes"` // What's at risk
	Reward        Reward         `json:"reward"`
	CharacterID   *string        `json:"character_id,omitempty"` // For personal quests
	RelatedNPCs   []string       `json:"related_npcs"`           // NPC names involved
	RelatedActs   []int          `json:"related_acts"`           // Act numbers where it appears
	ProgressNotes []ProgressNote `json:"progress_notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Objective represents a step in a quest
type Objective struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	Order       int    `json:"order"`
}

// ProgressNote tracks quest progress
type ProgressNote struct {
	Date    time.Time `json:"date"`
	Note    string    `json:"note"`
	Session int       `json:"session,omitempty"` // Session number when noted
}

// Validate checks if the quest is valid
func (q *Quest) Validate() error {
	if q.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if q.Title == "" {
		return NewValidationError("title", "quest title is required")
	}
	if q.Type != "" {
		switch q.Type {
		case QuestTypeMain, QuestTypeSide, QuestTypePersonal, QuestTypeGroup,
			QuestTypeRedencion, QuestTypeVenganza, QuestTypeDescubrimiento, QuestTypeProteccion:
			// valid
		default:
			return NewValidationError("type", "invalid quest type: "+string(q.Type))
		}
	}
	if q.Status != "" {
		switch q.Status {
		case QuestStatusActive, QuestStatusCompleted, QuestStatusFailed, QuestStatusOnHold:
			// valid
		default:
			return NewValidationError("status", "invalid quest status: "+string(q.Status))
		}
	}
	return nil
}

// CampaignState represents the complete state of a campaign for reading
type CampaignState struct {
	Campaign          Campaign        `json:"campaign"`
	Acts              []Act           `json:"acts"`
	Characters        []Character     `json:"characters"`
	NPCs              []NPC           `json:"npcs"`
	Monsters          []Monster       `json:"monsters"`
	Encounters        []Encounter     `json:"encounters"`
	Maps              []Map           `json:"maps"`
	Quests            []Quest         `json:"quests"`
	ActiveQuests      []Quest         `json:"active_quests"`
	CompletedQuests   []Quest         `json:"completed_quests"`
	Relationships     []Relationship  `json:"relationships"`
	ThreadsUnresolved []string        `json:"threads_unresolved"`
	Timeline          []TimelineEvent `json:"timeline"`
}

// TimelineEvent represents an event in the campaign timeline
type TimelineEvent struct {
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // session, plot, character, combat
	Session     int       `json:"session,omitempty"`
	RelatedIDs  []string  `json:"related_ids,omitempty"`
}

// SessionPrep contains preparation material for a session
type SessionPrep struct {
	SessionNumber      int         `json:"session_number"`
	PreviouslyOn       string      `json:"previously_on"`
	ActiveQuests       []Quest     `json:"active_quests"`
	RelevantNPCs       []NPC       `json:"relevant_npcs"`
	PossibleScenes     []Scene     `json:"possible_scenes"`
	PreparedEncounters []Encounter `json:"prepared_encounters"`
	Hooks              []string    `json:"hooks"`
}

// Scene represents a possible scene for a session
type Scene struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // combat, roleplay, exploration, puzzle
	Location    string   `json:"location,omitempty"`
	NPCs        []string `json:"npcs,omitempty"`
}
