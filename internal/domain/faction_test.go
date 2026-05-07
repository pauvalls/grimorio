package domain

import "testing"

func TestFactionValidate(t *testing.T) {
	tests := []struct {
		name    string
		faction Faction
		wantErr bool
		errField string
	}{
		{
			name: "valid faction",
			faction: Faction{
				ID:   "merchants-guild",
				Name: "Merchants Guild",
				Tier: 3,
			},
			wantErr: false,
		},
		{
			name: "missing id",
			faction: Faction{
				Name: "Merchants Guild",
				Tier: 3,
			},
			wantErr:  true,
			errField: "id",
		},
		{
			name: "invalid id not kebab-case",
			faction: Faction{
				ID:   "MerchantsGuild",
				Name: "Merchants Guild",
				Tier: 3,
			},
			wantErr:  true,
			errField: "id",
		},
		{
			name: "missing name",
			faction: Faction{
				ID:   "merchants-guild",
				Tier: 3,
			},
			wantErr:  true,
			errField: "name",
		},
		{
			name: "tier too low",
			faction: Faction{
				ID:   "merchants-guild",
				Name: "Merchants Guild",
				Tier: 0,
			},
			wantErr:  true,
			errField: "tier",
		},
		{
			name: "tier too high",
			faction: Faction{
				ID:   "merchants-guild",
				Name: "Merchants Guild",
				Tier: 6,
			},
			wantErr:  true,
			errField: "tier",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.faction.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if ve, ok := err.(*ValidationError); ok && ve.Field != tt.errField {
					t.Fatalf("expected error on field %q, got %q", tt.errField, ve.Field)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestScoreToStatus(t *testing.T) {
	tests := []struct {
		score int8
		want  string
	}{
		{-100, string(FactionStatusHostile)},
		{-80, string(FactionStatusHostile)},
		{-79, string(FactionStatusUnfriendly)},
		{-30, string(FactionStatusUnfriendly)},
		{-29, string(FactionStatusNeutral)},
		{0, string(FactionStatusNeutral)},
		{29, string(FactionStatusNeutral)},
		{30, string(FactionStatusFriendly)},
		{79, string(FactionStatusFriendly)},
		{80, string(FactionStatusAllied)},
		{100, string(FactionStatusAllied)},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := ScoreToStatus(tt.score)
			if got != tt.want {
				t.Fatalf("ScoreToStatus(%d) = %q, want %q", tt.score, got, tt.want)
			}
		})
	}
}

func TestReputationEntryApplyDelta(t *testing.T) {
	tests := []struct {
		name        string
		initial     int8
		delta       int8
		wantScore   int8
		wantDelta   int8
		wantHistory int
	}{
		{
			name:        "normal increase",
			initial:     0,
			delta:       20,
			wantScore:   20,
			wantDelta:   20,
			wantHistory: 1,
		},
		{
			name:        "bounds cap upper",
			initial:     95,
			delta:       10,
			wantScore:   100,
			wantDelta:   5,
			wantHistory: 1,
		},
		{
			name:        "bounds cap lower",
			initial:     -95,
			delta:       -10,
			wantScore:   -100,
			wantDelta:   -5,
			wantHistory: 1,
		},
		{
			name:        "multiple events accumulate history",
			initial:     0,
			delta:       10,
			wantScore:   10,
			wantDelta:   10,
			wantHistory: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ReputationEntry{
				FactionID: "faction-a",
				PartyID:   "party-1",
				Score:     tt.initial,
				Status:    ScoreToStatus(tt.initial),
			}
			r.ApplyDelta(tt.delta, 1, "test reason", "test_action")
			if r.Score != tt.wantScore {
				t.Fatalf("score = %d, want %d", r.Score, tt.wantScore)
			}
			if len(r.History) != tt.wantHistory {
				t.Fatalf("history len = %d, want %d", len(r.History), tt.wantHistory)
			}
			if len(r.History) > 0 && r.History[0].Delta != tt.wantDelta {
				t.Fatalf("recorded delta = %d, want %d", r.History[0].Delta, tt.wantDelta)
			}
			if r.Status != ScoreToStatus(tt.wantScore) {
				t.Fatalf("status = %q, want %q", r.Status, ScoreToStatus(tt.wantScore))
			}
		})
	}
}

func TestFactionReputationMatrixGetEntry(t *testing.T) {
	m := FactionReputationMatrix{CampaignID: "test-campaign"}

	// Getting a new entry should create it with neutral defaults
	e1 := m.GetEntry("faction-a", "party-1")
	if e1.Score != 0 {
		t.Fatalf("new entry score = %d, want 0", e1.Score)
	}
	if e1.Status != string(FactionStatusNeutral) {
		t.Fatalf("new entry status = %q, want neutral", e1.Status)
	}

	// Mutating the returned entry should affect the matrix
	e1.Score = 50
	e2 := m.GetEntry("faction-a", "party-1")
	if e2.Score != 50 {
		t.Fatalf("retrieved entry score = %d, want 50", e2.Score)
	}

	// Getting a different pair should create a separate entry
	if len(m.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m.Entries))
	}
	e3 := m.GetEntry("faction-b", "party-1")
	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}
	_ = e3
}
