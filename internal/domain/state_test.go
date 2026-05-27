package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestXPEntry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entry   XPEntry
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid entry",
			entry: XPEntry{
				ID:         "party1-xp-1",
				SessionNum: 1,
				Amount:     300,
				Reason:     "Combat encounter",
				Timestamp:  time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			entry: XPEntry{
				SessionNum: 1,
				Amount:     300,
				Reason:     "Combat encounter",
			},
			wantErr: true,
			errMsg:  "id",
		},
		{
			name: "negative session",
			entry: XPEntry{
				ID:         "party1-xp-1",
				SessionNum: -1,
				Amount:     300,
				Reason:     "Combat encounter",
			},
			wantErr: true,
			errMsg:  "session_num",
		},
		{
			name: "negative XP amount",
			entry: XPEntry{
				ID:         "party1-xp-1",
				SessionNum: 1,
				Amount:     -100,
				Reason:     "Penalty",
			},
			wantErr: true,
			errMsg:  "amount",
		},
		{
			name: "missing reason",
			entry: XPEntry{
				ID:         "party1-xp-1",
				SessionNum: 1,
				Amount:     300,
			},
			wantErr: true,
			errMsg:  "reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("XPEntry.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("XPEntry.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestChapterProgressionRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rule    ChapterProgressionRule
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid rule",
			rule: ChapterProgressionRule{
				ChapterID:     "chapter-1",
				Title:         "The Beginning",
				MinPartyLevel: 1,
				MaxPartyLevel: 5,
			},
			wantErr: false,
		},
		{
			name: "missing chapter ID",
			rule: ChapterProgressionRule{
				Title:         "The Beginning",
				MinPartyLevel: 1,
				MaxPartyLevel: 5,
			},
			wantErr: true,
			errMsg:  "chapter_id",
		},
		{
			name: "missing title",
			rule: ChapterProgressionRule{
				ChapterID:     "chapter-1",
				MinPartyLevel: 1,
				MaxPartyLevel: 5,
			},
			wantErr: true,
			errMsg:  "title",
		},
		{
			name: "invalid min level",
			rule: ChapterProgressionRule{
				ChapterID:     "chapter-1",
				Title:         "The Beginning",
				MinPartyLevel: 0,
				MaxPartyLevel: 5,
			},
			wantErr: true,
			errMsg:  "min_party_level",
		},
		{
			name: "invalid max level",
			rule: ChapterProgressionRule{
				ChapterID:     "chapter-1",
				Title:         "The Beginning",
				MinPartyLevel: 1,
				MaxPartyLevel: 21,
			},
			wantErr: true,
			errMsg:  "max_party_level",
		},
		{
			name: "min greater than max",
			rule: ChapterProgressionRule{
				ChapterID:     "chapter-1",
				Title:         "The Beginning",
				MinPartyLevel: 10,
				MaxPartyLevel: 5,
			},
			wantErr: true,
			errMsg:  "min_party_level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ChapterProgressionRule.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ChapterProgressionRule.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestPartyState_Validate(t *testing.T) {
	tests := []struct {
		name    string
		state   PartyState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid state",
			state: PartyState{
				PartyID:      "party-1",
				CurrentLevel: 5,
				XPTotal:      6500,
			},
			wantErr: false,
		},
		{
			name: "missing party ID",
			state: PartyState{
				CurrentLevel: 5,
				XPTotal:      6500,
			},
			wantErr: true,
			errMsg:  "party_id",
		},
		{
			name: "invalid level",
			state: PartyState{
				PartyID:      "party-1",
				CurrentLevel: 0,
				XPTotal:      6500,
			},
			wantErr: true,
			errMsg:  "current_level",
		},
		{
			name: "negative XP",
			state: PartyState{
				PartyID:      "party-1",
				CurrentLevel: 5,
				XPTotal:      -100,
			},
			wantErr: true,
			errMsg:  "xp_total",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PartyState.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("PartyState.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestPartyState_AddXP(t *testing.T) {
	state := &PartyState{
		PartyID:      "party-1",
		CurrentLevel: 1,
		XPTotal:      0,
		XPLedger:     []XPEntry{},
	}

	// Add XP for first level
	state.AddXP(300, "Combat encounter", 1, "chapter-1")

	if state.XPTotal != 300 {
		t.Errorf("Expected XPTotal=300, got %d", state.XPTotal)
	}
	if state.CurrentLevel != 2 {
		t.Errorf("Expected CurrentLevel=2, got %d", state.CurrentLevel)
	}
	if len(state.XPLedger) != 1 {
		t.Errorf("Expected 1 ledger entry, got %d", len(state.XPLedger))
	}
	if state.XPLedger[0].Amount != 300 {
		t.Errorf("Expected ledger entry amount=300, got %d", state.XPLedger[0].Amount)
	}
	if state.XPLedger[0].ChapterID != "chapter-1" {
		t.Errorf("Expected ledger entry chapterID=chapter-1, got %s", state.XPLedger[0].ChapterID)
	}

	// Add more XP
	state.AddXP(600, "Quest completion", 2, "chapter-1")

	if state.XPTotal != 900 {
		t.Errorf("Expected XPTotal=900, got %d", state.XPTotal)
	}
	if state.CurrentLevel != 3 {
		t.Errorf("Expected CurrentLevel=3, got %d", state.CurrentLevel)
	}
	if len(state.XPLedger) != 2 {
		t.Errorf("Expected 2 ledger entries, got %d", len(state.XPLedger))
	}
}

func TestCanonDocument_Validate_WithChapterProgression(t *testing.T) {
	doc := &CanonDocument{
		SchemaVersion: SchemaVersionV2,
		CampaignID:    "test-campaign",
		ChapterProgression: []ChapterProgressionRule{
			{
				ChapterID:     "chapter-1",
				Title:         "The Beginning",
				MinPartyLevel: 1,
				MaxPartyLevel: 5,
			},
		},
		PartyState: &PartyState{
			PartyID:      "party-1",
			CurrentLevel: 1,
			XPTotal:      0,
		},
	}

	err := doc.Validate()
	if err != nil {
		t.Errorf("CanonDocument.Validate() with chapter progression error = %v", err)
	}
}

func TestXPEntry_Serialization(t *testing.T) {
	entry := XPEntry{
		ID:         "party1-xp-1",
		SessionNum: 1,
		Amount:     300,
		Reason:     "Combat encounter",
		Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ChapterID:  "chapter-1",
	}

	// Test JSON marshaling
	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal XPEntry: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled XPEntry
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal XPEntry: %v", err)
	}

	// Verify fields
	if unmarshaled.ID != entry.ID {
		t.Errorf("ID mismatch: expected %s, got %s", entry.ID, unmarshaled.ID)
	}
	if unmarshaled.SessionNum != entry.SessionNum {
		t.Errorf("SessionNum mismatch: expected %d, got %d", entry.SessionNum, unmarshaled.SessionNum)
	}
	if unmarshaled.Amount != entry.Amount {
		t.Errorf("Amount mismatch: expected %d, got %d", entry.Amount, unmarshaled.Amount)
	}
	if unmarshaled.Reason != entry.Reason {
		t.Errorf("Reason mismatch: expected %s, got %s", entry.Reason, unmarshaled.Reason)
	}
	if !unmarshaled.Timestamp.Equal(entry.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", entry.Timestamp, unmarshaled.Timestamp)
	}
	if unmarshaled.ChapterID != entry.ChapterID {
		t.Errorf("ChapterID mismatch: expected %s, got %s", entry.ChapterID, unmarshaled.ChapterID)
	}
}

func TestPartyState_Serialization(t *testing.T) {
	state := PartyState{
		PartyID:      "party-1",
		CurrentLevel: 5,
		XPTotal:      6500,
		XPLedger: []XPEntry{
			{
				ID:         "party1-xp-1",
				SessionNum: 1,
				Amount:     300,
				Reason:     "Combat",
				Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			{
				ID:         "party1-xp-2",
				SessionNum: 2,
				Amount:     600,
				Reason:     "Quest completion",
				Timestamp:  time.Date(2024, 1, 22, 14, 0, 0, 0, time.UTC),
			},
		},
		CurrentChapterID:  "chapter-1",
		CompletedChapters: []string{"prologue"},
		CreatedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:         time.Date(2024, 1, 22, 14, 0, 0, 0, time.UTC),
	}

	// Test JSON marshaling
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("Failed to marshal PartyState: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled PartyState
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal PartyState: %v", err)
	}

	// Verify fields
	if unmarshaled.PartyID != state.PartyID {
		t.Errorf("PartyID mismatch: expected %s, got %s", state.PartyID, unmarshaled.PartyID)
	}
	if unmarshaled.CurrentLevel != state.CurrentLevel {
		t.Errorf("CurrentLevel mismatch: expected %d, got %d", state.CurrentLevel, unmarshaled.CurrentLevel)
	}
	if unmarshaled.XPTotal != state.XPTotal {
		t.Errorf("XPTotal mismatch: expected %d, got %d", state.XPTotal, unmarshaled.XPTotal)
	}
	if len(unmarshaled.XPLedger) != len(state.XPLedger) {
		t.Errorf("XPLedger length mismatch: expected %d, got %d", len(state.XPLedger), len(unmarshaled.XPLedger))
	}
	if unmarshaled.CurrentChapterID != state.CurrentChapterID {
		t.Errorf("CurrentChapterID mismatch: expected %s, got %s", state.CurrentChapterID, unmarshaled.CurrentChapterID)
	}
	if len(unmarshaled.CompletedChapters) != len(state.CompletedChapters) {
		t.Errorf("CompletedChapters length mismatch: expected %d, got %d", len(state.CompletedChapters), len(unmarshaled.CompletedChapters))
	}
}

func TestChapterProgressionRule_Serialization(t *testing.T) {
	rule := ChapterProgressionRule{
		ChapterID:         "chapter-1",
		Title:             "The Beginning",
		RequiredQuests:    []string{"quest-1", "quest-2"},
		OptionalQuests:    []string{"side-quest-1"},
		MinPartyLevel:     1,
		MaxPartyLevel:     5,
		XPThreshold:       5000,
		RequiredLocations: []string{"area-1", "area-2"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(rule)
	if err != nil {
		t.Fatalf("Failed to marshal ChapterProgressionRule: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled ChapterProgressionRule
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ChapterProgressionRule: %v", err)
	}

	// Verify fields
	if unmarshaled.ChapterID != rule.ChapterID {
		t.Errorf("ChapterID mismatch: expected %s, got %s", rule.ChapterID, unmarshaled.ChapterID)
	}
	if unmarshaled.Title != rule.Title {
		t.Errorf("Title mismatch: expected %s, got %s", rule.Title, unmarshaled.Title)
	}
	if !sliceEqual(unmarshaled.RequiredQuests, rule.RequiredQuests) {
		t.Errorf("RequiredQuests mismatch: expected %v, got %v", rule.RequiredQuests, unmarshaled.RequiredQuests)
	}
	if unmarshaled.MinPartyLevel != rule.MinPartyLevel {
		t.Errorf("MinPartyLevel mismatch: expected %d, got %d", rule.MinPartyLevel, unmarshaled.MinPartyLevel)
	}
	if unmarshaled.MaxPartyLevel != rule.MaxPartyLevel {
		t.Errorf("MaxPartyLevel mismatch: expected %d, got %d", rule.MaxPartyLevel, unmarshaled.MaxPartyLevel)
	}
}

// Helper function for slice comparison

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
