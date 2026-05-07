package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemCanonRepository_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCanonRepository(tmpDir)
	campaignID := "test-campaign"

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    campaignID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Facts: []domain.CanonFact{
			{ID: "fact-1", Category: "lore", Statement: "The ancient god sleeps beneath the mountain", Immutable: true},
		},
	}

	if err := repo.Save(campaignID, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	if !repo.Exists(campaignID) {
		t.Fatal("expected Exists to return true")
	}

	// Verify file exists
	canonPath := filepath.Join(tmpDir, campaignID, "canon", "canon.json")
	if _, err := os.Stat(canonPath); os.IsNotExist(err) {
		t.Fatal("expected canon.json to exist on filesystem")
	}

	loaded, err := repo.Load(campaignID)
	if err != nil {
		t.Fatalf("failed to load canon: %v", err)
	}

	if loaded.CampaignID != campaignID {
		t.Fatalf("expected campaign ID %q, got %q", campaignID, loaded.CampaignID)
	}

	if len(loaded.Facts) != 1 {
		t.Fatalf("expected 1 fact, got %d", len(loaded.Facts))
	}
}

func TestFilesystemCanonRepository_LoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCanonRepository(tmpDir)
	_, err := repo.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent canon")
	}
}

func TestFilesystemCanonRepository_SchemaVersionRejection(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCanonRepository(tmpDir)
	campaignID := "test-campaign"

	// Write a file with missing schema version directly
	dir := filepath.Join(tmpDir, campaignID, "canon")
	os.MkdirAll(dir, 0755)
	badJSON := []byte(`{"campaign_id":"test-campaign","facts":[]}`)
	os.WriteFile(filepath.Join(dir, "canon.json"), badJSON, 0644)

	_, err := repo.Load(campaignID)
	if err == nil {
		t.Fatal("expected error for missing schema version")
	}
}

func TestFilesystemNarrativeStateRepository_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemNarrativeStateRepository(tmpDir)
	campaignID := "test-campaign"

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     campaignID,
		CurrentSession: 3,
		LastUpdated:    time.Now(),
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-1", Name: "El Informador", Session: 2, Cause: "combat"},
		},
	}

	if err := repo.Save(campaignID, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	if !repo.Exists(campaignID) {
		t.Fatal("expected Exists to return true")
	}

	// Verify file exists
	statePath := filepath.Join(tmpDir, campaignID, "canon", "narrative_state.json")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("expected narrative_state.json to exist on filesystem")
	}

	loaded, err := repo.Load(campaignID)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.CurrentSession != 3 {
		t.Fatalf("expected current session 3, got %d", loaded.CurrentSession)
	}
}

func TestFilesystemNarrativeStateRepository_LoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemNarrativeStateRepository(tmpDir)
	_, err := repo.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent state")
	}
}
