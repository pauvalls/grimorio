package domain

import (
	"time"
)

// XPEntry represents a single XP transaction in the ledger
type XPEntry struct {
	ID          string    `json:"id"`
	SessionNum  int       `json:"session_num"`
	Amount      int       `json:"amount"`
	Reason      string    `json:"reason"`
	Timestamp   time.Time `json:"timestamp"`
	ChapterID   string    `json:"chapter_id,omitempty"`
}

// Validate checks if the XP entry is valid
func (x *XPEntry) Validate() error {
	if x.ID == "" {
		return NewValidationError("id", "XP entry ID is required")
	}
	if x.SessionNum < 0 {
		return NewValidationError("session_num", "session number cannot be negative")
	}
	if x.Amount < 0 {
		return NewValidationError("amount", "XP amount cannot be negative")
	}
	if x.Reason == "" {
		return NewValidationError("reason", "XP reason is required")
	}
	return nil
}

// ChapterProgressionRule defines the rules for progressing between chapters
type ChapterProgressionRule struct {
	ChapterID         string   `json:"chapter_id"`
	Title             string   `json:"title"`
	RequiredQuests    []string `json:"required_quests,omitempty"`    // Quest IDs that must be completed
	OptionalQuests    []string `json:"optional_quests,omitempty"`    // Quest IDs that are optional
	MinPartyLevel     int      `json:"min_party_level"`              // Minimum party level to start
	MaxPartyLevel     int      `json:"max_party_level"`              // Maximum party level to start
	XPThreshold       int      `json:"xp_threshold,omitempty"`       // Optional XP threshold
	RequiredLocations []string `json:"required_locations,omitempty"` // Area IDs that must be visited
}

// Validate checks if the chapter progression rule is valid
func (r *ChapterProgressionRule) Validate() error {
	if r.ChapterID == "" {
		return NewValidationError("chapter_id", "chapter ID is required")
	}
	if r.Title == "" {
		return NewValidationError("title", "chapter title is required")
	}
	if r.MinPartyLevel < 1 || r.MinPartyLevel > 20 {
		return NewValidationError("min_party_level", "min party level must be between 1 and 20")
	}
	if r.MaxPartyLevel < 1 || r.MaxPartyLevel > 20 {
		return NewValidationError("max_party_level", "max party level must be between 1 and 20")
	}
	if r.MinPartyLevel > r.MaxPartyLevel {
		return NewValidationError("min_party_level", "min level cannot be greater than max level")
	}
	return nil
}

// PartyState tracks the party's progression through the campaign
type PartyState struct {
	PartyID           string     `json:"party_id"`
	CurrentLevel      int        `json:"current_level"`
	XPTotal           int        `json:"xp_total"`
	XPLedger          []XPEntry  `json:"xp_ledger"`
	CurrentChapterID  string     `json:"current_chapter_id,omitempty"`
	CompletedChapters []string   `json:"completed_chapters"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Validate checks if the party state is valid
func (p *PartyState) Validate() error {
	if p.PartyID == "" {
		return NewValidationError("party_id", "party ID is required")
	}
	if p.CurrentLevel < 1 || p.CurrentLevel > 20 {
		return NewValidationError("current_level", "current level must be between 1 and 20")
	}
	if p.XPTotal < 0 {
		return NewValidationError("xp_total", "XP total cannot be negative")
	}
	return nil
}

// AddXP adds XP to the party state and updates level
func (p *PartyState) AddXP(amount int, reason string, sessionNum int, chapterID string) {
	entry := XPEntry{
		ID:         generateXPEntryID(p.PartyID, len(p.XPLedger)),
		SessionNum: sessionNum,
		Amount:     amount,
		Reason:     reason,
		Timestamp:  time.Now(),
		ChapterID:  chapterID,
	}
	p.XPLedger = append(p.XPLedger, entry)
	p.XPTotal += amount
	p.CurrentLevel = CalculateLevelFromXP(p.XPTotal)
	p.UpdatedAt = time.Now()
}

// generateXPEntryID generates a unique ID for an XP entry
func generateXPEntryID(partyID string, index int) string {
	return partyID + "-xp-" + string(rune('A'+index%26))
}
