package repository

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryQuestRepository(t *testing.T) {
	repo := NewMemoryQuestRepository()

	charID := "char-1"
	quest := &domain.Quest{
		CampaignID:  "campaign-1",
		Title:       "Test Quest",
		Type:        domain.QuestTypeRedencion,
		Status:      domain.QuestStatusActive,
		CharacterID: &charID,
	}

	// Test Save
	if err := repo.Save(quest); err != nil {
		t.Fatalf("Save() error: %v", err)
	}
	if quest.ID == "" {
		t.Error("Save() should auto-generate ID")
	}

	// Test Read
	read, err := repo.Read("campaign-1", quest.ID)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Title != "Test Quest" {
		t.Errorf("Read() title = %s, want Test Quest", read.Title)
	}

	// Test Read not found
	_, err = repo.Read("campaign-1", "nonexistent")
	if err == nil {
		t.Error("Read() should error for nonexistent quest")
	}

	// Test List
	list, err := repo.List("campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("List() len = %d, want 1", len(list))
	}

	// Test ListByCharacter
	charQuests, err := repo.ListByCharacter("campaign-1", "char-1")
	if err != nil {
		t.Fatalf("ListByCharacter() error: %v", err)
	}
	if len(charQuests) != 1 {
		t.Errorf("ListByCharacter() len = %d, want 1", len(charQuests))
	}

	// Test ListByStatus
	activeQuests, err := repo.ListByStatus("campaign-1", domain.QuestStatusActive)
	if err != nil {
		t.Fatalf("ListByStatus() error: %v", err)
	}
	if len(activeQuests) != 1 {
		t.Errorf("ListByStatus() len = %d, want 1", len(activeQuests))
	}

	// Test Update
	quest.Title = "Updated Quest"
	if err := repo.Save(quest); err != nil {
		t.Fatalf("Save() update error: %v", err)
	}
	updated, _ := repo.Read("campaign-1", quest.ID)
	if updated.Title != "Updated Quest" {
		t.Errorf("Update failed, title = %s, want Updated Quest", updated.Title)
	}

	// Test Delete
	if err := repo.Delete("campaign-1", quest.ID); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List("campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}

func TestMemoryQuestRepository_Invalid(t *testing.T) {
	repo := NewMemoryQuestRepository()

	invalid := &domain.Quest{CampaignID: "", Title: ""}
	if err := repo.Save(invalid); err == nil {
		t.Error("Save() should error for invalid quest")
	}
}
