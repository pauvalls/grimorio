package services

import (
	"context"
	"testing"
	"github.com/pauvalls/grimorio/internal/domain"
)

// MockMilestoneXPRepository for testing
type MockMilestoneXPRepository struct {
	totalXP int
	err     error
}

func (m *MockMilestoneXPRepository) Create(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error {
	return nil
}

func (m *MockMilestoneXPRepository) Read(ctx context.Context, campaignID string, chapterID string) (*domain.ChapterXPTable, error) {
	return nil, nil
}

func (m *MockMilestoneXPRepository) Update(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error {
	return nil
}

func (m *MockMilestoneXPRepository) Delete(ctx context.Context, campaignID string, chapterID string) error {
	return nil
}

func (m *MockMilestoneXPRepository) GetTotalXP(ctx context.Context, campaignID string, partyID string) (int, error) {
	return m.totalXP, m.err
}

func TestMilestoneService_GenerateChapterTable_ValidLevelRange(t *testing.T) {
	repo := &MockMilestoneXPRepository{}
	service := NewMilestoneService(repo)

	table, err := service.GenerateChapterTable(context.Background(), "chapter_1", "The Beginning", domain.LevelRange{Min: 1, Max: 5})
	if err != nil {
		t.Fatalf("GenerateChapterTable() error = %v", err)
	}

	if table.ChapterID != "chapter_1" {
		t.Errorf("ChapterID = %s, want chapter_1", table.ChapterID)
	}
	if table.LevelRange.Min != 1 || table.LevelRange.Max != 5 {
		t.Errorf("LevelRange = %v, want {1 5}", table.LevelRange)
	}
	if len(table.Milestones) == 0 {
		t.Error("Expected milestones, got none")
	}
}

func TestMilestoneService_GenerateChapterTable_InvalidLevelRange(t *testing.T) {
	repo := &MockMilestoneXPRepository{}
	service := NewMilestoneService(repo)

	_, err := service.GenerateChapterTable(context.Background(), "chapter_1", "Invalid", domain.LevelRange{Min: 5, Max: 3})
	if err == nil {
		t.Error("Expected error for invalid level range, got nil")
	}
}

func TestMilestoneService_CalculatePartyLevel_ExactMilestone(t *testing.T) {
	repo := &MockMilestoneXPRepository{totalXP: 300} // Level 2
	service := NewMilestoneService(repo)

	level, err := service.CalculatePartyLevel(context.Background(), "campaign_1", "party_1")
	if err != nil {
		t.Fatalf("CalculatePartyLevel() error = %v", err)
	}
	if level != 2 {
		t.Errorf("Level = %d, want 2", level)
	}
}

func TestMilestoneService_CalculatePartyLevel_BetweenMilestones(t *testing.T) {
	repo := &MockMilestoneXPRepository{totalXP: 600} // Between level 2 and 3
	service := NewMilestoneService(repo)

	level, err := service.CalculatePartyLevel(context.Background(), "campaign_1", "party_1")
	if err != nil {
		t.Fatalf("CalculatePartyLevel() error = %v", err)
	}
	if level != 2 {
		t.Errorf("Level = %d, want 2 (between milestones)", level)
	}
}

func TestMilestoneService_GetNextMilestone(t *testing.T) {
	repo := &MockMilestoneXPRepository{}
	service := NewMilestoneService(repo)

	milestone, err := service.GetNextMilestone(context.Background(), "campaign_1", 0) // Level 1
	if err != nil {
		t.Fatalf("GetNextMilestone() error = %v", err)
	}
	if milestone == nil {
		t.Fatal("Expected milestone, got nil")
	}
	if milestone.LevelAchieved != 2 {
		t.Errorf("LevelAchieved = %d, want 2", milestone.LevelAchieved)
	}
}
