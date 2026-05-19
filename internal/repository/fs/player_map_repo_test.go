package fs

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemPlayerMapRepository(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	repo := NewFilesystemPlayerMapRepository(tmpDir)

	pm := &domain.PlayerMap{
		ID:          "pm_test_1",
		SourceMapID: "dm_map_1",
		CampaignID:  "campaign-1",
		AreaID:      "area_1",
		Title:       "Test Player Map",
		Description: "Player-facing map",
		IncludedFeatures: []domain.MapFeature{
			{ID: "f1", Type: "room", Label: "Entry", IsSecret: false},
		},
		ExportFormats: []string{"svg", "pdf", "png"},
		GridVisible:   true,
		Scale:         "1 square = 5 feet",
	}

	// Create
	if err := repo.Create(ctx, "campaign-1", pm); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Read
	read, err := repo.Read(ctx, "campaign-1", "pm_test_1")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.ID != "pm_test_1" {
		t.Errorf("Read() ID = %s, want pm_test_1", read.ID)
	}

	// Read not found
	_, err = repo.Read(ctx, "campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent")
	}

	// Update
	pm.Title = "Updated Player Map"
	if err := repo.Update(ctx, "campaign-1", pm); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	updated, _ := repo.Read(ctx, "campaign-1", "pm_test_1")
	if updated.Title != "Updated Player Map" {
		t.Errorf("Update() Title = %s, want Updated Player Map", updated.Title)
	}

	// GetByArea
	byArea, err := repo.GetByArea(ctx, "campaign-1", "area_1")
	if err != nil {
		t.Fatalf("GetByArea() error: %v", err)
	}
	if len(byArea) != 1 {
		t.Errorf("GetByArea() len = %d, want 1", len(byArea))
	}

	// Delete
	if err := repo.Delete(ctx, "campaign-1", "pm_test_1"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	_, err = repo.Read(ctx, "campaign-1", "pm_test_1")
	if err == nil {
		t.Error("Read() after delete should error")
	}

	// Delete nonexistent should not error
	if err := repo.Delete(ctx, "campaign-1", "nonexistent"); err != nil {
		t.Errorf("Delete() nonexistent should not error: %v", err)
	}
}
