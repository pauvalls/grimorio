package services

import (
	"context"
	"os"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

// mockMilestoneRepo implements MilestoneXPRepository for testing.
type mockMilestoneRepo struct {
	tables map[string]*domain.ChapterXPTable
}

func newMockMilestoneRepo() *mockMilestoneRepo {
	return &mockMilestoneRepo{tables: make(map[string]*domain.ChapterXPTable)}
}

func (r *mockMilestoneRepo) Create(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error {
	r.tables[table.ChapterID] = table
	return nil
}

func (r *mockMilestoneRepo) Read(ctx context.Context, campaignID string, chapterID string) (*domain.ChapterXPTable, error) {
	table, ok := r.tables[chapterID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return table, nil
}

func (r *mockMilestoneRepo) Update(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error {
	r.tables[table.ChapterID] = table
	return nil
}

func (r *mockMilestoneRepo) Delete(ctx context.Context, campaignID string, chapterID string) error {
	delete(r.tables, chapterID)
	return nil
}

func (r *mockMilestoneRepo) GetTotalXP(ctx context.Context, campaignID string, partyID string) (int, error) {
	total := 0
	for _, table := range r.tables {
		for _, m := range table.Milestones {
			total += m.CumulativeXP
		}
	}
	return total, nil
}

func TestMilestoneService_UpdateSessionXP(t *testing.T) {
	repo := newMockMilestoneRepo()
	svc := NewMilestoneService(repo)

	table := &domain.ChapterXPTable{
		ChapterID:    "chapter_1",
		ChapterTitle: "The Beginning",
		LevelRange:   domain.LevelRange{Min: 1, Max: 3},
		Milestones: []domain.MilestoneXP{
			{ChapterID: "chapter_1", SessionNumber: 1, XPThreshold: 300, CumulativeXP: 300, LevelAchieved: 2},
			{ChapterID: "chapter_1", SessionNumber: 2, XPThreshold: 600, CumulativeXP: 900, LevelAchieved: 3},
		},
		TotalSessions: 2,
	}

	if err := repo.Create(context.Background(), "campaign-1", table); err != nil {
		t.Fatalf("failed to seed table: %v", err)
	}

	// Update XP for session 1
	err := svc.UpdateSessionXP(context.Background(), "campaign-1", "chapter_1", 100)
	if err != nil {
		t.Fatalf("UpdateSessionXP() error: %v", err)
	}

	// Verify update was persisted
	updated, _ := repo.Read(context.Background(), "campaign-1", "chapter_1")
	if updated.Milestones[0].XPThreshold != 400 {
		t.Errorf("Milestone[0].XPThreshold = %d, want 400", updated.Milestones[0].XPThreshold)
	}
}

func TestMilestoneService_UpdateSessionXP_Negative(t *testing.T) {
	repo := newMockMilestoneRepo()
	svc := NewMilestoneService(repo)

	err := svc.UpdateSessionXP(context.Background(), "campaign-1", "chapter_1", -50)
	if err == nil {
		t.Error("expected error for negative xp_awarded, got nil")
	}
}

func TestMilestoneService_UpdateSessionXP_TableNotFound(t *testing.T) {
	repo := newMockMilestoneRepo()
	svc := NewMilestoneService(repo)

	err := svc.UpdateSessionXP(context.Background(), "campaign-1", "nonexistent", 100)
	if err == nil {
		t.Error("expected error for nonexistent table, got nil")
	}
}

func TestMilestoneService_CalculatePartyLevel(t *testing.T) {
	repo := newMockMilestoneRepo()
	svc := NewMilestoneService(repo)

	table := &domain.ChapterXPTable{
		ChapterID:    "chapter_1",
		ChapterTitle: "Test",
		LevelRange:   domain.LevelRange{Min: 1, Max: 3},
		Milestones: []domain.MilestoneXP{
			{ChapterID: "chapter_1", SessionNumber: 1, XPThreshold: 300, CumulativeXP: 300, LevelAchieved: 2},
		},
		TotalSessions: 1,
	}
	if err := repo.Create(context.Background(), "campaign-1", table); err != nil {
		t.Fatalf("failed to seed table: %v", err)
	}

	level, err := svc.CalculatePartyLevel(context.Background(), "campaign-1", "party_1")
	if err != nil {
		t.Fatalf("CalculatePartyLevel() error: %v", err)
	}
	if level != 2 {
		t.Errorf("level = %d, want 2", level)
	}
}
