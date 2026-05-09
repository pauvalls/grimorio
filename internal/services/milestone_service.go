package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// MilestoneService handles milestone XP tracking and table generation.
type MilestoneService struct {
	xpRepo MilestoneXPRepository
}

// MilestoneXPRepository defines the repository interface for milestone XP.
type MilestoneXPRepository interface {
	Create(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error
	Read(ctx context.Context, campaignID string, chapterID string) (*domain.ChapterXPTable, error)
	Update(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error
	Delete(ctx context.Context, campaignID string, chapterID string) error
	GetTotalXP(ctx context.Context, campaignID string, partyID string) (int, error)
}

// NewMilestoneService creates a new MilestoneService.
func NewMilestoneService(xpRepo MilestoneXPRepository) *MilestoneService {
	return &MilestoneService{xpRepo: xpRepo}
}

// GenerateChapterTable generates an XP table for a chapter based on level range.
func (s *MilestoneService) GenerateChapterTable(ctx context.Context, chapterID string, chapterTitle string, levelRange domain.LevelRange) (*domain.ChapterXPTable, error) {
	if err := domain.ValidateLevelRange(levelRange); err != nil {
		return nil, fmt.Errorf("invalid level range: %w", err)
	}

	thresholds := domain.GetPHBMilestoneThresholds()
	sessionsNeeded := domain.CalculateSessionsForLevelRange(levelRange)

	// Adjust sessions to fit within level range
	levelsToCover := levelRange.Max - levelRange.Min + 1
	if sessionsNeeded > levelsToCover*3 {
		sessionsNeeded = levelsToCover * 3 // Max 3 sessions per level
	}

	milestones := make([]domain.MilestoneXP, 0, sessionsNeeded)
	cumulative := 0
	startLevel := levelRange.Min

	for i := 0; i < sessionsNeeded && (startLevel+i-1) < len(thresholds); i++ {
		currentLevel := startLevel + i
		if currentLevel > 20 {
			break
		}

		xpThreshold := 0
		if i > 0 && currentLevel-1 < len(thresholds) {
			xpThreshold = thresholds[currentLevel-1] - thresholds[startLevel-1]
		}

		if currentLevel <= len(thresholds) {
			cumulative = thresholds[currentLevel-1]
		}

		sessionsToLevel := 1
		if i < sessionsNeeded-1 {
			sessionsToLevel = 1
		}

		milestones = append(milestones, domain.MilestoneXP{
			ChapterID:       chapterID,
			SessionNumber:   i + 1,
			XPThreshold:     xpThreshold,
			CumulativeXP:    cumulative,
			LevelAchieved:   currentLevel,
			SessionsToLevel: sessionsToLevel,
		})
	}

	table := &domain.ChapterXPTable{
		ChapterID:     chapterID,
		ChapterTitle:  chapterTitle,
		LevelRange:    levelRange,
		Milestones:    milestones,
		TotalSessions: len(milestones),
	}

	if err := table.Validate(); err != nil {
		return nil, fmt.Errorf("generated table validation failed: %w", err)
	}

	return table, nil
}

// CalculatePartyLevel calculates the current party level from cumulative XP.
func (s *MilestoneService) CalculatePartyLevel(ctx context.Context, campaignID string, partyID string) (int, error) {
	totalXP, err := s.xpRepo.GetTotalXP(ctx, campaignID, partyID)
	if err != nil {
		return 0, fmt.Errorf("failed to get total XP: %w", err)
	}
	return domain.CalculateLevelFromXP(totalXP), nil
}

// GetNextMilestone returns the next milestone for a party.
func (s *MilestoneService) GetNextMilestone(ctx context.Context, campaignID string, currentXP int) (*domain.MilestoneXP, error) {
	currentLevel := domain.CalculateLevelFromXP(currentXP)
	if currentLevel >= 20 {
		return nil, nil // No more milestones at level 20
	}

	nextLevelXP := domain.GetXPNeededForLevel(currentLevel + 1)
	xpNeeded := nextLevelXP - currentXP

	return &domain.MilestoneXP{
		ChapterID:     campaignID,
		XPThreshold:   xpNeeded,
		CumulativeXP:  nextLevelXP,
		LevelAchieved: currentLevel + 1,
	}, nil
}

// UpdateSessionXP updates XP for a session (placeholder for future implementation).
func (s *MilestoneService) UpdateSessionXP(ctx context.Context, sessionID string, xpAwarded int) error {
	// TODO: Implement session XP tracking
	return nil
}
