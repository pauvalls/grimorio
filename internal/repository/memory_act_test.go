package repository

import (
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryActRepository(t *testing.T) {
	repo := NewMemoryActRepository()

	act := &domain.Act{
		CampaignID:        "campaign-1",
		Number:            1,
		Title:             "Act One",
		Content:           "Content here",
		GameMode:          "investigacion",
		ChapterObjectives: []string{"Investigar el crimen", "Identificar al culpable"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "mystery",
		RunningGuidance:   strings.Repeat("palabra ", 700),
		AssetHandoff:      "La evidencia encontrada apunta al Acto 2",
	}

	// Test Save
	if err := repo.Save(act); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test Read
	read, err := repo.Read("campaign-1", 1)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if read.Title != "Act One" {
		t.Errorf("Read() title = %s, want Act One", read.Title)
	}

	// Test Read not found
	_, err = repo.Read("campaign-1", 999)
	if err == nil {
		t.Error("Read() should error for nonexistent act")
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
	act.Title = "Updated Act"
	if err := repo.Save(act); err != nil {
		t.Fatalf("Save() update error: %v", err)
	}
	updated, _ := repo.Read("campaign-1", 1)
	if updated.Title != "Updated Act" {
		t.Errorf("Update failed, title = %s, want Updated Act", updated.Title)
	}

	// Test Delete
	if err := repo.Delete("campaign-1", 1); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List("campaign-1")
	if len(list) != 0 {
		t.Errorf("List() after delete len = %d, want 0", len(list))
	}
}

func TestMemoryActRepository_Invalid(t *testing.T) {
	repo := NewMemoryActRepository()

	invalid := &domain.Act{CampaignID: "", Number: 0}
	if err := repo.Save(invalid); err == nil {
		t.Error("Save() should error for invalid act")
	}
}
