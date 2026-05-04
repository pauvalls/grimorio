package repository

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryCampaignRepository(t *testing.T) {
	repo := NewMemoryCampaignRepository()

	campaign := &domain.Campaign{
		Name:    "test-campaign",
		Title:   "Test Campaign",
		Setting: "A test setting",
	}

	// Test Create
	if err := repo.Create(campaign); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Test duplicate Create
	if err := repo.Create(campaign); err == nil {
		t.Error("Create() should error for duplicate campaign")
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

	// Test Update not found
	if err := repo.Update(&domain.Campaign{Name: "nonexistent"}); err == nil {
		t.Error("Update() should error for nonexistent campaign")
	}

	// Test Delete
	if err := repo.Delete("test-campaign"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if repo.Exists("test-campaign") {
		t.Error("Exists() should return false after delete")
	}
}

func TestMemoryCampaignRepository_CreateInvalid(t *testing.T) {
	repo := NewMemoryCampaignRepository()

	invalid := &domain.Campaign{Name: ""}
	if err := repo.Create(invalid); err == nil {
		t.Error("Create() should error for invalid campaign")
	}
}
