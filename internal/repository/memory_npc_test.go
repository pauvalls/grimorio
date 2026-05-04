package repository

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryNPCRepository(t *testing.T) {
	repo := NewMemoryNPCRepository()

	npc := &domain.NPC{
		CampaignID: "campaign-1",
		Name:       "Test NPC",
		Role:       "merchant",
	}

	// Test Save
	if err := repo.Save(npc); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test Read
	read, err := repo.Read("campaign-1", "Test NPC")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Name != "Test NPC" {
		t.Errorf("Read() name = %s, want Test NPC", read.Name)
	}

	// Test Read not found
	_, err = repo.Read("campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent NPC")
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
	npc.Role = "villain"
	if err := repo.Save(npc); err != nil {
		t.Fatalf("Save() update error: %v", err)
	}
	updated, _ := repo.Read("campaign-1", "Test NPC")
	if updated.Role != "villain" {
		t.Errorf("Update failed, role = %s, want villain", updated.Role)
	}

	// Test Delete
	if err := repo.Delete("campaign-1", "Test NPC"); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List("campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}

func TestMemoryNPCRepository_Invalid(t *testing.T) {
	repo := NewMemoryNPCRepository()

	invalid := &domain.NPC{CampaignID: "", Name: ""}
	if err := repo.Save(invalid); err == nil {
		t.Error("Save() should error for invalid NPC")
	}
}
