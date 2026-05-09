package domain

import (
	"errors"
	"fmt"
)

// MilestoneXP represents a single milestone in XP progression following PHB thresholds.
type MilestoneXP struct {
	ChapterID       string `json:"chapter_id"`
	SessionNumber   int    `json:"session_number"`
	XPThreshold     int    `json:"xp_threshold"`
	CumulativeXP    int    `json:"cumulative_xp"`
	LevelAchieved   int    `json:"level_achieved"`
	SessionsToLevel int    `json:"sessions_to_level,omitempty"`
}

// ChapterXPTable represents XP progression for a chapter with level range.
type ChapterXPTable struct {
	ChapterID     string       `json:"chapter_id"`
	ChapterTitle  string       `json:"chapter_title"`
	LevelRange    LevelRange   `json:"level_range"`
	Milestones    []MilestoneXP `json:"milestones"`
	TotalSessions int          `json:"total_sessions"`
}

// LevelRange represents a level range for a chapter.
type LevelRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// GetPHBMilestoneThresholds returns standard PHB milestone XP thresholds for levels 1-20.
// These thresholds represent the cumulative XP needed to reach each level.
func GetPHBMilestoneThresholds() []int {
	return []int{
		0,       // Level 1
		300,     // Level 2
		900,     // Level 3
		2700,    // Level 4
		6500,    // Level 5
		14000,   // Level 6
		23000,   // Level 7
		34000,   // Level 8
		48000,   // Level 9
		64000,   // Level 10
		85000,   // Level 11
		100000,  // Level 12
		120000,  // Level 13
		155000,  // Level 14
		195000,  // Level 15
		240000,  // Level 16
		305000,  // Level 17
		385000,  // Level 18
		490000,  // Level 19
		630000,  // Level 20
	}
}

// Validate checks if the ChapterXPTable is valid according to PHB milestone progression.
func (t *ChapterXPTable) Validate() error {
	if t.ChapterID == "" {
		return errors.New("chapter_id is required")
	}
	if t.ChapterTitle == "" {
		return errors.New("chapter_title is required")
	}
	if err := ValidateLevelRange(t.LevelRange); err != nil {
		return fmt.Errorf("invalid level range: %w", err)
	}
	if len(t.Milestones) == 0 {
		return errors.New("milestones cannot be empty")
	}

	for i, m := range t.Milestones {
		if m.SessionNumber != i+1 {
			return fmt.Errorf("session %d: expected session number %d, got %d", i+1, i+1, m.SessionNumber)
		}
		if m.XPThreshold < 0 {
			return fmt.Errorf("session %d: XP threshold cannot be negative", i+1)
		}
		if m.LevelAchieved < 1 || m.LevelAchieved > 20 {
			return fmt.Errorf("session %d: level must be between 1 and 20", i+1)
		}
	}

	// Validate cumulative XP is increasing
	for i := 1; i < len(t.Milestones); i++ {
		if t.Milestones[i].CumulativeXP <= t.Milestones[i-1].CumulativeXP {
			return fmt.Errorf("session %d: cumulative XP must be greater than previous session", i+1)
		}
	}

	return nil
}

// ValidateLevelRange checks if a level range is valid.
func ValidateLevelRange(lr LevelRange) error {
	if lr.Min < 1 || lr.Min > 20 {
		return fmt.Errorf("level min must be between 1 and 20, got %d", lr.Min)
	}
	if lr.Max < 1 || lr.Max > 20 {
		return fmt.Errorf("level max must be between 1 and 20, got %d", lr.Max)
	}
	if lr.Min >= lr.Max {
		return fmt.Errorf("level min (%d) must be less than max (%d)", lr.Min, lr.Max)
	}
	return nil
}

// CalculateLevelFromXP calculates the character level from cumulative XP.
func CalculateLevelFromXP(xp int) int {
	thresholds := GetPHBMilestoneThresholds()
	for i := len(thresholds) - 1; i >= 0; i-- {
		if xp >= thresholds[i] {
			return i + 1
		}
	}
	return 1
}

// CalculateSessionsForLevelRange calculates the number of sessions needed for a level range.
// This is an estimate based on typical milestone pacing (1-2 sessions per level).
func CalculateSessionsForLevelRange(lr LevelRange) int {
	levels := lr.Max - lr.Min + 1
	// Typical milestone campaign: 1-3 sessions per level, average 2
	return levels * 2
}

// GetXPNeededForLevel returns the XP needed to reach a specific level.
func GetXPNeededForLevel(level int) int {
	if level < 1 || level > 20 {
		return 0
	}
	thresholds := GetPHBMilestoneThresholds()
	return thresholds[level-1]
}

// GetXPNeededForNextLevel returns the XP needed to reach the next level from current level.
func GetXPNeededForNextLevel(currentLevel int) int {
	if currentLevel < 1 || currentLevel >= 20 {
		return 0
	}
	thresholds := GetPHBMilestoneThresholds()
	return thresholds[currentLevel] - thresholds[currentLevel-1]
}
