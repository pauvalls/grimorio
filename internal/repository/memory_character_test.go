package repository

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryCharacterRepository(t *testing.T) {
	repo := NewMemoryCharacterRepository()

	char := &domain.Character{
		CampaignID: "campaign-1",
		Name:       "Test Character",
		Race:       "humano",
		Class:      "guerrero",
		Level:      1,
	}

	// Test Save
	if err := repo.Save(char); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test Read
	read, err := repo.Read("campaign-1", "Test Character")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Name != "Test Character" {
		t.Errorf("Read() name = %s, want Test Character", read.Name)
	}

	// Test Read not found
	_, err = repo.Read("campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent character")
	}

	// Test List
	list, err := repo.List("campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List() len = %d, want 1", len(list))
	}

	// Test Update
	char.Level = 5
	if err := repo.Save(char); err != nil {
		t.Fatalf("Save() update error: %v", err)
	}
	updated, _ := repo.Read("campaign-1", "Test Character")
	if updated.Level != 5 {
		t.Errorf("Update failed, level = %d, want 5", updated.Level)
	}

	// Test Delete
	if err := repo.Delete("campaign-1", "Test Character"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List("campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}

func TestMemoryCharacterRepository_Invalid(t *testing.T) {
	repo := NewMemoryCharacterRepository()

	invalid := &domain.Character{CampaignID: "", Name: ""}
	if err := repo.Save(invalid); err == nil {
		t.Error("Save() should error for invalid character")
	}
}
