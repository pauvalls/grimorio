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
	QuestStatusAbandoned QuestStatus = "abandoned" // V3
)

// QuestTier represents the tier/scope of a quest (V3)
type QuestTier string

const (
	QuestTierMinor    QuestTier = "minor"
	QuestTierMajor    QuestTier = "major"
	QuestTierChapter  QuestTier = "chapter"
	QuestTierCampaign QuestTier = "campaign"
)

// Quest represents a mission or objective (V3 enhanced)
type Quest struct {
	ID            string         `json:"id"`
	CampaignID    string         `json:"campaign_id"`
	Title         string         `json:"title"`
	Type          QuestType      `json:"type"`
	Tier          QuestTier      `json:"tier,omitempty"` // V3
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
	// V3 enhanced fields
	Approaches    []QuestApproach `json:"approaches,omitempty"`
	FailureStates []QuestFailure  `json:"failure_states,omitempty"`
	Clues         []QuestClue     `json:"clues,omitempty"`
	Areas         []string        `json:"areas,omitempty"` // Area IDs involved
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

// QuestApproach represents one approach to completing a quest (V3)
type QuestApproach struct {
	Type             string       `json:"type"` // combat, social, stealth, exploration, puzzle
	Title            string       `json:"title"`
	Requirements     []string     `json:"requirements"`
	Steps            []QuestStep  `json:"steps"`
	Challenges       []Challenge  `json:"challenges"`
	Rewards          []QuestReward `json:"rewards"`
	SuccessCondition string       `json:"success_condition"`
}

// QuestStep represents a step in a quest approach (V3)
type QuestStep struct {
	StepNumber int    `json:"step_number"`
	Description string `json:"description"`
	AreaRef     string `json:"area_ref,omitempty"`
	NPCRef      string `json:"npc_ref,omitempty"`
	DC          *int   `json:"dc,omitempty"`
	Skill       string `json:"skill,omitempty"`
}

// Challenge represents a challenge within a quest (V3)
type Challenge struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	DC          int    `json:"dc,omitempty"`
	Consequence string `json:"consequence,omitempty"`
}

// QuestReward represents a reward for completing a quest or approach (V3)
type QuestReward struct {
	Type        string `json:"type"` // xp, treasure, reputation, story
	Description string `json:"description"`
	Value       int    `json:"value,omitempty"` // XP amount or GP value
	FactionID   string `json:"faction_id,omitempty"`
}

// QuestFailure represents a failure state for a quest (V3)
type QuestFailure struct {
	Type           string   `json:"type"` // soft, hard, complication
	Trigger        string   `json:"trigger"`
	Consequences   []string `json:"consequences"`
	Continuation   string   `json:"continuation,omitempty"`
	NPCReactions   []string `json:"npc_reactions"`
}

// QuestClue represents a clue related to a quest (V3)
type QuestClue struct {
	ID            string   `json:"id"`
	Description   string   `json:"description"`
	Location      string   `json:"location"`
	DiscoveryDC   *int     `json:"discovery_dc,omitempty"`
	LeadsTo       []string `json:"leads_to"` // Quest step or area refs
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
		case QuestStatusActive, QuestStatusCompleted, QuestStatusFailed, QuestStatusOnHold, QuestStatusAbandoned:
			// valid
		default:
			return NewValidationError("status", "invalid quest status: "+string(q.Status))
		}
	}
	return nil
}

// Validate checks QuestApproach validity (V3)
func (qa *QuestApproach) Validate() error {
	if qa.Type == "" {
		return NewValidationError("type", "type is required")
	}
	if qa.Title == "" {
		return NewValidationError("title", "title is required")
	}
	if len(qa.Steps) < 2 {
		return NewValidationError("steps", "must have at least 2 steps")
	}
	return nil
}

// Validate checks QuestFailure validity (V3)
func (qf *QuestFailure) Validate() error {
	if qf.Type == "" {
		return NewValidationError("type", "type is required")
	}
	if qf.Trigger == "" {
		return NewValidationError("trigger", "trigger is required")
	}
	if len(qf.Consequences) == 0 {
		return NewValidationError("consequences", "must have at least 1 consequence")
	}
	return nil
}

// IsValidQuestType checks if a quest type is valid (V3)
func IsValidQuestType(t QuestType) bool {
	switch t {
	case QuestTypeMain, QuestTypeSide, QuestTypePersonal, QuestTypeGroup,
		QuestTypeRedencion, QuestTypeVenganza, QuestTypeDescubrimiento, QuestTypeProteccion:
		return true
	default:
		return false
	}
}

// IsValidQuestTier checks if a quest tier is valid (V3)
func IsValidQuestTier(t QuestTier) bool {
	switch t {
	case QuestTierMinor, QuestTierMajor, QuestTierChapter, QuestTierCampaign:
		return true
	default:
		return false
	}
}

// IsValidQuestStatus checks if a quest status is valid (V3)
func IsValidQuestStatus(s QuestStatus) bool {
	switch s {
	case QuestStatusActive, QuestStatusCompleted, QuestStatusFailed, QuestStatusOnHold, QuestStatusAbandoned:
		return true
	default:
		return false
	}
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

// Scene represents a possible scene for a session
type Scene struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"` // combat, roleplay, exploration, puzzle
	Location    string   `json:"location,omitempty"`
	NPCs        []string `json:"npcs,omitempty"`
}
