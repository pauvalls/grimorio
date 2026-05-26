package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestValidateChapterProgression_NoRule(t *testing.T) {
	allowed, warnings := ValidateChapterProgression("chapter-1", "chapter-2", []string{}, 1, []domain.ChapterProgressionRule{})
	
	if !allowed {
		t.Error("Expected allowed when no rule defined")
	}
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(warnings))
	}
}

func TestValidateChapterProgression_MissingQuests(t *testing.T) {
	rules := []domain.ChapterProgressionRule{
		{
			ChapterID:      "chapter-2",
			Title:          "Chapter 2",
			RequiredQuests: []string{"quest-1", "quest-2"},
			MinPartyLevel:  1,
		},
	}

	// No quests completed
	allowed, warnings := ValidateChapterProgression("chapter-1", "chapter-2", []string{}, 1, rules)
	
	if !allowed {
		t.Error("Expected allowed (DM is god)")
	}
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !contains(warnings[0], "quest-1") || !contains(warnings[0], "quest-2") {
		t.Errorf("Expected warning to mention missing quests, got: %s", warnings[0])
	}
}

func TestValidateChapterProgression_CompletedQuests(t *testing.T) {
	rules := []domain.ChapterProgressionRule{
		{
			ChapterID:      "chapter-2",
			Title:          "Chapter 2",
			RequiredQuests: []string{"quest-1", "quest-2"},
			MinPartyLevel:  1,
		},
	}

	// All quests completed
	allowed, warnings := ValidateChapterProgression("chapter-1", "chapter-2", []string{"quest-1", "quest-2"}, 1, rules)
	
	if !allowed {
		t.Error("Expected allowed")
	}
	if len(warnings) > 0 {
		t.Errorf("Expected no warnings, got %v", warnings)
	}
}

func TestValidateChapterProgression_LowLevel(t *testing.T) {
	rules := []domain.ChapterProgressionRule{
		{
			ChapterID:      "chapter-2",
			Title:          "Chapter 2",
			RequiredQuests: []string{},
			MinPartyLevel:  5,
		},
	}

	// Party level too low
	allowed, warnings := ValidateChapterProgression("chapter-1", "chapter-2", []string{}, 2, rules)
	
	if !allowed {
		t.Error("Expected allowed (DM is god)")
	}
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(warnings))
	}
	if !contains(warnings[0], "level 5") || !contains(warnings[0], "level 2") {
		t.Errorf("Expected warning to mention level mismatch, got: %s", warnings[0])
	}
}

func TestValidateChapterProgression_HighLevel(t *testing.T) {
	rules := []domain.ChapterProgressionRule{
		{
			ChapterID:      "chapter-2",
			Title:          "Chapter 2",
			RequiredQuests: []string{},
			MinPartyLevel:  1,
			MaxPartyLevel:  5,
		},
	}

	// Party level too high
	allowed, warnings := ValidateChapterProgression("chapter-1", "chapter-2", []string{}, 7, rules)
	
	if !allowed {
		t.Error("Expected allowed (DM is god)")
	}
	if len(warnings) != 1 {
		t.Fatalf("Expected 1 warning, got %d", len(warnings))
	}
	if !contains(warnings[0], "5") || !contains(warnings[0], "7") {
		t.Errorf("Expected warning to mention level mismatch, got: %s", warnings[0])
	}
}

func TestCalculatePartyLevel(t *testing.T) {
	tests := []struct {
		xpTotal    int
		expectLevel int
	}{
		{0, 1},
		{299, 1},
		{300, 2},
		{899, 2},
		{900, 3},
		{2700, 4},
		{6500, 5},
		{14000, 6},
		{355000, 20},
		{400000, 20}, // Over max
	}

	for _, tt := range tests {
		level := CalculatePartyLevel(tt.xpTotal)
		if level != tt.expectLevel {
			t.Errorf("CalculatePartyLevel(%d) = %d, want %d", tt.xpTotal, level, tt.expectLevel)
		}
	}
}

func TestCountCompletedObjectives(t *testing.T) {
	objectives := []string{"quest-1", "quest-2", "quest-3"}
	completed := []string{"quest-1", "quest-3"}

	count := CountCompletedObjectives(completed, objectives)
	if count != 2 {
		t.Errorf("Expected 2 completed, got %d", count)
	}
}

func TestCountCompletedObjectives_None(t *testing.T) {
	objectives := []string{"quest-1", "quest-2"}
	completed := []string{"quest-3"}

	count := CountCompletedObjectives(completed, objectives)
	if count != 0 {
		t.Errorf("Expected 0 completed, got %d", count)
	}
}

func TestCountCompletedObjectives_All(t *testing.T) {
	objectives := []string{"quest-1", "quest-2"}
	completed := []string{"quest-1", "quest-2", "quest-3"}

	count := CountCompletedObjectives(completed, objectives)
	if count != 2 {
		t.Errorf("Expected 2 completed, got %d", count)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
