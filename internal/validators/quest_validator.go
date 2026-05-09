package validators

import (
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// QuestValidator validates quest content according to WotC standards.
type QuestValidator struct{}

// NewQuestValidator creates a new QuestValidator.
func NewQuestValidator() *QuestValidator {
	return &QuestValidator{}
}

// ValidateApproachCount checks if quest has at least 3 approaches.
func (v *QuestValidator) ValidateApproachCount(approaches []domain.QuestApproach) error {
	if len(approaches) < 3 {
		return fmt.Errorf("quest must have at least 3 approaches, got %d", len(approaches))
	}
	return nil
}

// ValidateApproachSteps checks if each approach has at least 2 steps.
func (v *QuestValidator) ValidateApproachSteps(approaches []domain.QuestApproach) error {
	for i, approach := range approaches {
		if len(approach.Steps) < 2 {
			return fmt.Errorf("approach %d must have at least 2 steps, got %d", i+1, len(approach.Steps))
		}
	}
	return nil
}

// ValidateFailureStates checks if quest has soft and hard failure states.
func (v *QuestValidator) ValidateFailureStates(failures []domain.QuestFailure) error {
	hasSoft := false
	hasHard := false
	
	for _, failure := range failures {
		if failure.Type == "soft" {
			hasSoft = true
		}
		if failure.Type == "hard" {
			hasHard = true
		}
	}
	
	if !hasSoft {
		return fmt.Errorf("quest must have soft failure state")
	}
	if !hasHard {
		return fmt.Errorf("quest must have hard failure state")
	}
	
	return nil
}

// ValidateQuestType checks if quest type is valid.
func (v *QuestValidator) ValidateQuestType(qt domain.QuestType) error {
	if !domain.IsValidQuestType(qt) {
		return fmt.Errorf("invalid quest type: %s", qt)
	}
	return nil
}

// ValidateQuestTier checks if quest tier is valid.
func (v *QuestValidator) ValidateQuestTier(tier domain.QuestTier) error {
	if !domain.IsValidQuestTier(tier) {
		return fmt.Errorf("invalid quest tier: %s", tier)
	}
	return nil
}

// ValidateRewardsXP checks if XP reward is appropriate for quest tier.
func (v *QuestValidator) ValidateRewardsXP(xp int, tier domain.QuestTier) error {
	minXP, maxXP := getXPRangeForTier(tier)
	if xp < minXP || xp > maxXP {
		return fmt.Errorf("XP %d not appropriate for tier %s (expected %d-%d)", xp, tier, minXP, maxXP)
	}
	return nil
}

func getXPRangeForTier(tier domain.QuestTier) (int, int) {
	switch tier {
	case domain.QuestTierMinor:
		return 50, 200
	case domain.QuestTierMajor:
		return 201, 1000
	case domain.QuestTierChapter:
		return 1001, 5000
	case domain.QuestTierCampaign:
		return 5001, 20000
	default:
		return 0, 0
	}
}
