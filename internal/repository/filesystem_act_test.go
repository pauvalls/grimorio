package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemActRepository(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemActRepository(tmpDir)

	// Create campaign directory first
	campaignDir := filepath.Join(tmpDir, "campaign-1")
	if err := os.MkdirAll(filepath.Join(campaignDir, "areas"), 0755); err != nil {
		t.Fatal(err)
	}

	act := &domain.Act{
		CampaignID:        "campaign-1",
		Number:            1,
		Title:             "The Beginning",
		Content:           "Once upon a time...",
		GameMode:          "investigacion",
		ChapterObjectives: []string{"Descubrir la identidad del traidor", "Recuperar el artefacto robado"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "mystery",
		RunningGuidance:   strings.Repeat("palabra ", 700),
		AssetHandoff:      "La carta encontrada revela la ubicación del almacén en el Acto 2",
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
	if read.Number != 1 {
		t.Errorf("Read() number = %d, want 1", read.Number)
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

	// Test Save with header prefix
	act2 := &domain.Act{
		CampaignID:        "campaign-1",
		Number:            2,
		Title:             "The Middle",
		Content:           "# The Middle\n\nContent here",
		GameMode:          "sandbox_urbano",
		GameModeSecondary: "intriga",
		ChapterObjectives: []string{"Establecer contacto con la facción rebelde", "Evitar la guerra entre gremios"},
		EstimatedDuration: "2-3 sesiones",
		Tone:              "political",
		RunningGuidance:   strings.Repeat("palabra ", 700),
		AssetHandoff:      "El sello del gremio obtenido permite acceso al distrito noble en el Acto 3",
	}
	if err := repo.Save(act2); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Test Delete
	if err := repo.Delete("campaign-1", 1); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	list, _ = repo.List("campaign-1")
	if len(list) != 1 {
		t.Errorf("List() after delete len = %d, want 1", len(list))
	}
}

func TestFilesystemActRepository_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemActRepository(tmpDir)

	invalid := &domain.Act{CampaignID: "", Number: 0}
	if err := repo.Save(invalid); err == nil {
		t.Error("Save() should error for invalid act")
	}
}

func TestFilesystemActRepository_ListEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemActRepository(tmpDir)

	// Create campaign directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "campaign-1", "areas"), 0755); err != nil {
		t.Fatal(err)
	}

	list, err := repo.List("campaign-1")
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("List() len = %d, want 0", len(list))
	}
}
