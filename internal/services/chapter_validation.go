package services

import (
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
)

// ValidateChapterProgression validates if a party can transition to target chapter
// Returns (allowed, warnings) - allows transition but warns if prerequisites not met
func ValidateChapterProgression(
	currentChapter string,
	targetChapter string,
	completedQuests []string,
	partyLevel int,
	rules []domain.ChapterProgressionRule,
) (bool, []string) {
	var warnings []string

	// Find rule for target chapter
	var targetRule *domain.ChapterProgressionRule
	for i := range rules {
		if rules[i].ChapterID == targetChapter {
			targetRule = &rules[i]
			break
		}
	}

	// No rule found for target chapter - allow with warning
	if targetRule == nil {
		warnings = append(warnings, fmt.Sprintf("⚠️ No progression rule for %s - allowing transition", targetChapter))
		return true, warnings
	}

	// Check required quests
	missingQuests := []string{}
	for _, questID := range targetRule.RequiredQuests {
		found := false
		for _, completed := range completedQuests {
			if completed == questID {
				found = true
				break
			}
		}
		if !found {
			missingQuests = append(missingQuests, questID)
		}
	}

	if len(missingQuests) > 0 {
		warnings = append(warnings, fmt.Sprintf(
			"⚠️ Chapter %s requires quests: %s (not completed)",
			targetChapter,
			strings.Join(missingQuests, ", "),
		))
	}

	// Check party level
	if partyLevel < targetRule.MinPartyLevel {
		warnings = append(warnings, fmt.Sprintf(
			"⚠️ Chapter %s recommends level %d+, party is level %d",
			targetChapter,
			targetRule.MinPartyLevel,
			partyLevel,
		))
	}

	// Check max level (optional - some chapters are designed for specific level ranges)
	if targetRule.MaxPartyLevel > 0 && partyLevel > targetRule.MaxPartyLevel {
		warnings = append(warnings, fmt.Sprintf(
			"⚠️ Chapter %s is designed for levels up to %d, party is level %d (may be too easy)",
			targetChapter,
			targetRule.MaxPartyLevel,
			partyLevel,
		))
	}

	// Always allow (DM is god), but warn
	return true, warnings
}

// CalculatePartyLevel calculates party level from total XP using PHB table
func CalculatePartyLevel(xpTotal int) int {
	// PHB Experience Points Thresholds by Level
	thresholds := []int{
		0,      // Level 1
		300,    // Level 2
		900,    // Level 3
		2700,   // Level 4
		6500,   // Level 5
		14000,  // Level 6
		23000,  // Level 7
		34000,  // Level 8
		48000,  // Level 9
		64000,  // Level 10
		85000,  // Level 11
		100000, // Level 12
		120000, // Level 13
		140000, // Level 14
		165000, // Level 15
		195000, // Level 16
		225000, // Level 17
		265000, // Level 18
		305000, // Level 19
		355000, // Level 20
	}

	for i := len(thresholds) - 1; i >= 0; i-- {
		if xpTotal >= thresholds[i] {
			return i + 1
		}
	}

	return 1 // Default to level 1
}

// GetChapterObjectives extracts quest IDs that are objectives for a chapter
// This reads from canon entities tagged with chapter objectives
func GetChapterObjectives(canon *domain.CanonDocument, chapterID string) []string {
	objectives := []string{}

	for _, entity := range canon.Entities {
		// Check if entity is tagged as chapter objective
		if chapterTag, ok := entity.Properties["chapter_objective"].(string); ok {
			if chapterTag == chapterID {
				objectives = append(objectives, entity.ID)
			}
		}
	}

	return objectives
}

// CountCompletedObjectives counts how many chapter objectives are completed
func CountCompletedObjectives(completedQuests []string, objectives []string) int {
	count := 0
	for _, objective := range objectives {
		for _, completed := range completedQuests {
			if completed == objective {
				count++
				break
			}
		}
	}
	return count
}
