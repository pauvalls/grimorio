package services

import (
	"context"
	"os"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

// mockHandoutRepoV3 implements HandoutRepositoryV3 for testing.
type mockHandoutRepoV3 struct {
	handouts map[string]*domain.Handout
}

func newMockHandoutRepoV3() *mockHandoutRepoV3 {
	return &mockHandoutRepoV3{handouts: make(map[string]*domain.Handout)}
}

func (r *mockHandoutRepoV3) Create(ctx context.Context, campaignID string, h *domain.Handout) error {
	r.handouts[h.ID] = h
	return nil
}

func (r *mockHandoutRepoV3) Read(ctx context.Context, campaignID string, handoutID string) (*domain.Handout, error) {
	h, ok := r.handouts[handoutID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return h, nil
}

func (r *mockHandoutRepoV3) Update(ctx context.Context, campaignID string, h *domain.Handout) error {
	r.handouts[h.ID] = h
	return nil
}

func (r *mockHandoutRepoV3) Delete(ctx context.Context, campaignID string, handoutID string) error {
	delete(r.handouts, handoutID)
	return nil
}

func (r *mockHandoutRepoV3) GetByQuest(ctx context.Context, campaignID string, questID string) ([]*domain.Handout, error) {
	var result []*domain.Handout
	for _, h := range r.handouts {
		for _, ref := range h.QuestRefs {
			if ref == questID {
				result = append(result, h)
				break
			}
		}
	}
	return result, nil
}

func (r *mockHandoutRepoV3) GetByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.Handout, error) {
	var result []*domain.Handout
	for _, h := range r.handouts {
		for _, ref := range h.AreaRefs {
			if ref == areaID {
				result = append(result, h)
				break
			}
		}
	}
	return result, nil
}

func TestHandoutServiceV3_ExportHandout_Text(t *testing.T) {
	repo := newMockHandoutRepoV3()
	svc := NewHandoutServiceV3(repo)

	handout := &domain.Handout{
		ID:         "h_1",
		CampaignID: "campaign-1",
		Type:       domain.HandoutTypeLetter,
		Title:      "Test Letter",
		Content:    "Dear adventurer, your quest awaits...",
		Format:     domain.FormatText,
	}

	if err := repo.Create(context.Background(), "campaign-1", handout); err != nil {
		t.Fatalf("failed to seed handout: %v", err)
	}

	content, err := svc.ExportHandout(context.Background(), "campaign-1", "h_1", "text")
	if err != nil {
		t.Fatalf("ExportHandout() error: %v", err)
	}

	if string(content) != "Dear adventurer, your quest awaits..." {
		t.Errorf("content = %q, want %q", string(content), "Dear adventurer, your quest awaits...")
	}
}

func TestHandoutServiceV3_ExportHandout_UnsupportedFormat(t *testing.T) {
	repo := newMockHandoutRepoV3()
	svc := NewHandoutServiceV3(repo)

	handout := &domain.Handout{
		ID:         "h_2",
		CampaignID: "campaign-1",
		Type:       domain.HandoutTypeLetter,
		Title:      "Test",
		Content:    "Some content",
	}

	if err := repo.Create(context.Background(), "campaign-1", handout); err != nil {
		t.Fatalf("failed to seed handout: %v", err)
	}

	_, err := svc.ExportHandout(context.Background(), "campaign-1", "h_2", "pdf")
	if err == nil {
		t.Error("expected error for unsupported format 'pdf', got nil")
	}
}

func TestHandoutServiceV3_ExportHandout_NotFound(t *testing.T) {
	repo := newMockHandoutRepoV3()
	svc := NewHandoutServiceV3(repo)

	_, err := svc.ExportHandout(context.Background(), "campaign-1", "nonexistent", "text")
	if err == nil {
		t.Error("expected error for nonexistent handout, got nil")
	}
}
