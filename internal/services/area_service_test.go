package services

import (
	"context"
	"os"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

// mockAreaRepoV3 implements AreaRepositoryV3 for testing.
type mockAreaRepoV3 struct {
	areas map[string]*domain.Area
}

func newMockAreaRepoV3() *mockAreaRepoV3 {
	return &mockAreaRepoV3{areas: make(map[string]*domain.Area)}
}

func (r *mockAreaRepoV3) Create(ctx context.Context, campaignID string, area *domain.Area) error {
	r.areas[area.ID] = area
	return nil
}

func (r *mockAreaRepoV3) Read(ctx context.Context, campaignID string, areaID string) (*domain.Area, error) {
	area, ok := r.areas[areaID]
	if !ok {
		return nil, os.ErrNotExist
	}
	return area, nil
}

func (r *mockAreaRepoV3) Update(ctx context.Context, campaignID string, area *domain.Area) error {
	r.areas[area.ID] = area
	return nil
}

func (r *mockAreaRepoV3) Delete(ctx context.Context, campaignID string, areaID string) error {
	delete(r.areas, areaID)
	return nil
}

func (r *mockAreaRepoV3) GetByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.Area, error) {
	var result []*domain.Area
	for _, a := range r.areas {
		if a.ChapterID == chapterID {
			result = append(result, a)
		}
	}
	return result, nil
}

func TestAreaService_GetAreaByNumber(t *testing.T) {
	repo := newMockAreaRepoV3()
	svc := NewAreaService(repo)

	// Create areas for chapter 1
	for i := 1; i <= 3; i++ {
		area := &domain.Area{
			ID:         "test_area",
			ChapterID:  "chapter_1",
			AreaNumber: i,
			Title:      "Area",
			Summary:    "Summary",
			Description: "Description",
			LevelRange: domain.LevelRange{Min: 1, Max: 3},
			Features: []domain.AreaFeature{
				{Type: "room", Name: "Room", Description: "A room", Hidden: false},
			},
			Encounters: []domain.AreaEncounter{
				{EncounterID: "enc_1", Trigger: "Enter", CRTotal: 1.0, XPValue: 200},
			},
		}
		// Give each area a unique ID
		area.ID = "area_ch1_" + string(rune('0'+i))
		if err := repo.Create(context.Background(), "campaign-1", area); err != nil {
			t.Fatalf("failed to seed area %d: %v", i, err)
		}
	}

	// Test: find area number 2
	found, err := svc.GetAreaByNumber(context.Background(), "campaign-1", 1, 2)
	if err != nil {
		t.Fatalf("GetAreaByNumber() error: %v", err)
	}
	if found == nil {
		t.Fatal("GetAreaByNumber() returned nil, expected area")
	}
	if found.AreaNumber != 2 {
		t.Errorf("AreaNumber = %d, want 2", found.AreaNumber)
	}
}

func TestAreaService_GetAreaByNumber_NotFound(t *testing.T) {
	repo := newMockAreaRepoV3()
	svc := NewAreaService(repo)

	// Test: no match returns nil without error
	found, err := svc.GetAreaByNumber(context.Background(), "campaign-1", 1, 99)
	if err != nil {
		t.Fatalf("GetAreaByNumber() error: %v", err)
	}
	if found != nil {
		t.Errorf("expected nil for non-existent area, got %v", found)
	}
}

func TestAreaService_GetAreasByChapter(t *testing.T) {
	repo := newMockAreaRepoV3()
	svc := NewAreaService(repo)

	// Create a single area for simplicity
	area := &domain.Area{
		ID:          "area_1",
		ChapterID:   "chapter_1",
		AreaNumber:  1,
		Title:       "Test",
		Summary:     "Test",
		Description: "Test area",
		LevelRange:  domain.LevelRange{Min: 1, Max: 3},
		Features: []domain.AreaFeature{
			{Type: "room", Name: "Room", Description: "A room", Hidden: false},
		},
		Encounters: []domain.AreaEncounter{
			{EncounterID: "enc_1", Trigger: "Enter", CRTotal: 1.0, XPValue: 200},
		},
	}
	if err := repo.Create(context.Background(), "campaign-1", area); err != nil {
		t.Fatalf("failed to seed area: %v", err)
	}

	areas, err := svc.GetAreasByChapter(context.Background(), "campaign-1", "chapter_1")
	if err != nil {
		t.Fatalf("GetAreasByChapter() error: %v", err)
	}
	if len(areas) != 1 {
		t.Errorf("len = %d, want 1", len(areas))
	}
}
