package domain

// Faction represents a faction within a campaign
type Faction struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Agenda      string   `json:"agenda"`
	ContactNPC  string   `json:"contact_npc"`
	Tier        int      `json:"tier"`
	Territory   []string `json:"territory"`
	Enemies     []string `json:"enemies"`
	Allies      []string `json:"allies"`
	IsSecret    bool     `json:"is_secret"`
}

// Validate checks if the faction is valid
func (f *Faction) Validate() error {
	if f.ID == "" {
		return NewValidationError("id", "faction ID is required")
	}
	if !IsValidKebabCase(f.ID) {
		return NewValidationError("id", "faction ID must be kebab-case")
	}
	if f.Name == "" {
		return NewValidationError("name", "faction name is required")
	}
	if f.Tier < 1 || f.Tier > 5 {
		return NewValidationError("tier", "tier must be between 1 and 5")
	}
	return nil
}

// FactionReputationStatus represents the status label for a reputation score
type FactionReputationStatus string

const (
	FactionStatusHostile    FactionReputationStatus = "hostile"
	FactionStatusUnfriendly FactionReputationStatus = "unfriendly"
	FactionStatusNeutral    FactionReputationStatus = "neutral"
	FactionStatusFriendly   FactionReputationStatus = "friendly"
	FactionStatusAllied     FactionReputationStatus = "allied"
)

// ScoreToStatus converts a reputation score to its status label
func ScoreToStatus(score int8) string {
	switch {
	case score <= -80:
		return string(FactionStatusHostile)
	case score <= -30:
		return string(FactionStatusUnfriendly)
	case score < 30:
		return string(FactionStatusNeutral)
	case score < 80:
		return string(FactionStatusFriendly)
	default:
		return string(FactionStatusAllied)
	}
}

// ApplyDelta applies a bounded reputation delta and records the event
func (r *ReputationEntry) ApplyDelta(delta int8, session int, reason, actionType string) {
	oldScore := r.Score
	newScore := oldScore + delta
	if newScore > 100 {
		newScore = 100
	}
	if newScore < -100 {
		newScore = -100
	}
	actualDelta := newScore - oldScore
	r.Score = newScore
	r.Status = ScoreToStatus(r.Score)
	r.History = append(r.History, ReputationEvent{
		Session:    session,
		Delta:      actualDelta,
		Reason:     reason,
		ActionType: actionType,
	})
}

// GetEntry returns the reputation entry for a faction-party pair, creating a neutral default if missing
func (m *FactionReputationMatrix) GetEntry(factionID, partyID string) *ReputationEntry {
	for i := range m.Entries {
		if m.Entries[i].FactionID == factionID && m.Entries[i].PartyID == partyID {
			return &m.Entries[i]
		}
	}
	m.Entries = append(m.Entries, ReputationEntry{
		FactionID: factionID,
		PartyID:   partyID,
		Score:     0,
		Status:    ScoreToStatus(0),
	})
	return &m.Entries[len(m.Entries)-1]
}

// ReputationUpdateResult captures direct and propagated changes
type ReputationUpdateResult struct {
	DirectChange      ReputationEntry   `json:"direct_change"`
	PropagatedChanges []ReputationEntry `json:"propagated_changes"`
}
