package domain

import (
	"testing"
)

func TestGetPHBMilestoneThresholds(t *testing.T) {
	thresholds := GetPHBMilestoneThresholds()

	// Verify we have thresholds for all 20 levels
	if len(thresholds) != 20 {
		t.Errorf("expected 20 thresholds, got %d", len(thresholds))
	}

	// Verify key thresholds match PHB
	expectedKeyThresholds := map[int]int{
		1:  0,      // Level 1
		2:  300,    // Level 2
		3:  900,    // Level 3
		5:  6500,   // Level 5
		10: 64000,  // Level 10
		15: 195000, // Level 15
		20: 630000, // Level 20
	}

	for level, expected := range expectedKeyThresholds {
		if thresholds[level-1] != expected {
			t.Errorf("level %d: expected XP %d, got %d", level, expected, thresholds[level-1])
		}
	}

	// Verify thresholds are strictly increasing
	for i := 1; i < len(thresholds); i++ {
		if thresholds[i] <= thresholds[i-1] {
			t.Errorf("thresholds not strictly increasing at index %d: %d <= %d", i, thresholds[i], thresholds[i-1])
		}
	}
}

func TestChapterXPTable_Validate(t *testing.T) {
	tests := []struct {
		name    string
		table   *ChapterXPTable
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid table",
			table: &ChapterXPTable{
				ChapterID:    "chapter_1",
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 1, Max: 5},
				Milestones: []MilestoneXP{
					{ChapterID: "chapter_1", SessionNumber: 1, XPThreshold: 0, CumulativeXP: 0, LevelAchieved: 1},
					{ChapterID: "chapter_1", SessionNumber: 2, XPThreshold: 300, CumulativeXP: 300, LevelAchieved: 2},
					{ChapterID: "chapter_1", SessionNumber: 3, XPThreshold: 600, CumulativeXP: 900, LevelAchieved: 3},
				},
				TotalSessions: 3,
			},
			wantErr: false,
		},
		{
			name: "missing chapter_id",
			table: &ChapterXPTable{
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 1, Max: 5},
				Milestones:   []MilestoneXP{{SessionNumber: 1}},
			},
			wantErr: true,
			errMsg:  "chapter_id is required",
		},
		{
			name: "missing chapter_title",
			table: &ChapterXPTable{
				ChapterID:  "chapter_1",
				LevelRange: LevelRange{Min: 1, Max: 5},
				Milestones: []MilestoneXP{{SessionNumber: 1}},
			},
			wantErr: true,
			errMsg:  "chapter_title is required",
		},
		{
			name: "invalid level range",
			table: &ChapterXPTable{
				ChapterID:    "chapter_1",
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 5, Max: 3}, // Invalid: min > max
				Milestones:   []MilestoneXP{{SessionNumber: 1}},
			},
			wantErr: true,
			errMsg:  "invalid level range",
		},
		{
			name: "empty milestones",
			table: &ChapterXPTable{
				ChapterID:    "chapter_1",
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 1, Max: 5},
				Milestones:   []MilestoneXP{},
			},
			wantErr: true,
			errMsg:  "milestones cannot be empty",
		},
		{
			name: "non-sequential session numbers",
			table: &ChapterXPTable{
				ChapterID:    "chapter_1",
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 1, Max: 5},
				Milestones: []MilestoneXP{
					{SessionNumber: 1, LevelAchieved: 1},
					{SessionNumber: 3, LevelAchieved: 2}, // Skips session 2
				},
			},
			wantErr: true,
			errMsg:  "expected session number 2",
		},
		{
			name: "negative XP threshold",
			table: &ChapterXPTable{
				ChapterID:    "chapter_1",
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 1, Max: 5},
				Milestones: []MilestoneXP{
					{SessionNumber: 1, XPThreshold: -100, LevelAchieved: 1},
				},
			},
			wantErr: true,
			errMsg:  "XP threshold cannot be negative",
		},
		{
			name: "cumulative XP not increasing",
			table: &ChapterXPTable{
				ChapterID:    "chapter_1",
				ChapterTitle: "The Beginning",
				LevelRange:   LevelRange{Min: 1, Max: 5},
				Milestones: []MilestoneXP{
					{SessionNumber: 1, XPThreshold: 300, CumulativeXP: 300, LevelAchieved: 2},
					{SessionNumber: 2, XPThreshold: 0, CumulativeXP: 300, LevelAchieved: 2}, // Same cumulative
				},
			},
			wantErr: true,
			errMsg:  "cumulative XP must be greater",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.table.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateLevelRange(t *testing.T) {
	tests := []struct {
		name    string
		lr      LevelRange
		wantErr bool
		errMsg  string
	}{
		{"valid range", LevelRange{Min: 1, Max: 5}, false, ""},
		{"valid range high levels", LevelRange{Min: 15, Max: 20}, false, ""},
		{"min below 1", LevelRange{Min: 0, Max: 5}, true, "level min must be between 1 and 20"},
		{"max above 20", LevelRange{Min: 1, Max: 21}, true, "level max must be between 1 and 20"},
		{"min equals max", LevelRange{Min: 5, Max: 5}, true, "level min (5) must be less than max (5)"},
		{"min greater than max", LevelRange{Min: 10, Max: 5}, true, "level min (10) must be less than max (5)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLevelRange(tt.lr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLevelRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateLevelRange() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestCalculateLevelFromXP(t *testing.T) {
	tests := []struct {
		name string
		xp   int
		want int
	}{
		{"level 1", 0, 1},
		{"level 2 exact", 300, 2},
		{"level 2 mid", 600, 2},
		{"level 3 exact", 900, 3},
		{"level 5 exact", 6500, 5},
		{"level 10 exact", 64000, 10},
		{"level 20 exact", 630000, 20},
		{"level 20 overflow", 1000000, 20},
		{"negative xp", -100, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateLevelFromXP(tt.xp)
			if got != tt.want {
				t.Errorf("CalculateLevelFromXP(%d) = %d, want %d", tt.xp, got, tt.want)
			}
		})
	}
}

func TestCalculateSessionsForLevelRange(t *testing.T) {
	tests := []struct {
		name string
		lr   LevelRange
		want int
	}{
		{"levels 1-5", LevelRange{Min: 1, Max: 5}, 10}, // 5 levels * 2 sessions
		{"levels 1-10", LevelRange{Min: 1, Max: 10}, 20},
		{"levels 5-10", LevelRange{Min: 5, Max: 10}, 12}, // 6 levels * 2 sessions
		{"single level", LevelRange{Min: 5, Max: 5}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSessionsForLevelRange(tt.lr)
			if got != tt.want {
				t.Errorf("CalculateSessionsForLevelRange(%v) = %d, want %d", tt.lr, got, tt.want)
			}
		})
	}
}

func TestGetXPNeededForLevel(t *testing.T) {
	tests := []struct {
		name  string
		level int
		want  int
	}{
		{"level 1", 1, 0},
		{"level 2", 2, 300},
		{"level 5", 5, 6500},
		{"level 10", 10, 64000},
		{"level 20", 20, 630000},
		{"level 0 (invalid)", 0, 0},
		{"level 21 (invalid)", 21, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetXPNeededForLevel(tt.level)
			if got != tt.want {
				t.Errorf("GetXPNeededForLevel(%d) = %d, want %d", tt.level, got, tt.want)
			}
		})
	}
}

func TestGetXPNeededForNextLevel(t *testing.T) {
	tests := []struct {
		name         string
		currentLevel int
		want         int
	}{
		{"level 1 to 2", 1, 300},
		{"level 2 to 3", 2, 600},
		{"level 4 to 5", 4, 3800},   // 6500 - 2700
		{"level 9 to 10", 9, 16000}, // 64000 - 48000
		{"level 19 to 20", 19, 140000},
		{"level 20 (no next)", 20, 0},
		{"level 0 (invalid)", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetXPNeededForNextLevel(tt.currentLevel)
			if got != tt.want {
				t.Errorf("GetXPNeededForNextLevel(%d) = %d, want %d", tt.currentLevel, got, tt.want)
			}
		})
	}
}

// Helper function to check if a string contains a substring

// Helper function to check if a string contains a substring
