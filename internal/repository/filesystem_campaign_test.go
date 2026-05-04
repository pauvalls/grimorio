package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemCampaignRepository(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCampaignRepository(tmpDir)

	campaign := &domain.Campaign{
		Name:    "test-campaign",
		Title:   "Test Campaign",
		Setting: "A test setting",
	}

	// Test Create
	if err := repo.Create(campaign); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Verify directories created
	campaignDir := filepath.Join(tmpDir, "test-campaign")
	for _, sub := range []string{"acts", "npcs", "bestiary", "encounters", "maps", "assets", "characters", "quests"} {
		if _, err := os.Stat(filepath.Join(campaignDir, sub)); err != nil {
			t.Errorf("Create() did not create %s directory: %v", sub, err)
		}
	}

	// Test Read
	read, err := repo.Read("test-campaign")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Name != "test-campaign" {
		t.Errorf("Read() name = %s, want test-campaign", read.Name)
	}

	// Test Read not found
	_, err = repo.Read("nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent campaign")
	}

	// Test Exists
	if !repo.Exists("test-campaign") {
		t.Error("Exists() should return true for existing campaign")
	}
	if repo.Exists("nonexistent") {
		t.Error("Exists() should return false for nonexistent campaign")
	}

	// Test List
	list, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List() len = %d, want 1", len(list))
	}

	// Test Update
	campaign.Title = "Updated Title"
	if err := repo.Update(campaign); err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	updated, _ := repo.Read("test-campaign")
	if updated.Title != "Updated Title" {
		t.Errorf("Update failed, title = %s, want Updated Title", updated.Title)
	}

	// Test Delete
	if err := repo.Delete("test-campaign"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if repo.Exists("test-campaign") {
		t.Error("Exists() should return false after delete")
	}
}

func TestFilesystemCampaignRepository_CreateInvalid(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCampaignRepository(tmpDir)

	invalid := &domain.Campaign{Name: ""}
	if err := repo.Create(invalid); err == nil {
		t.Error("Create() should error for invalid campaign")
	}
}

func TestFilesystemCampaignRepository_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCampaignRepository(tmpDir)

	list, err := repo.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List() len = %d, want 0", len(list))
	}
}
