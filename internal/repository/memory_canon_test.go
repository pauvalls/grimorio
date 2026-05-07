package repository

import (
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryCanonRepository_SaveLoad(t *testing.T) {
	repo := NewMemoryCanonRepository()
	campaignID := "test-campaign"

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    campaignID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Facts: []domain.CanonFact{
			{ID: "fact-1", Category: "lore", Statement: "The ancient god sleeps beneath the mountain", Immutable: true},
		},
		Entities: []domain.CanonEntity{
			{ID: "npc-1", Name: "Lord Vex", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
		},
	}

	if err := repo.Save(campaignID, doc); err != nil {
		t.Fatalf("failed to save canon: %v", err)
	}

	if !repo.Exists(campaignID) {
		t.Fatal("expected Exists to return true")
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

func TestMemoryCanonRepository_LoadNotFound(t *testing.T) {
	repo := NewMemoryCanonRepository()
	_, err := repo.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent canon")
	}
}

func TestMemoryCanonRepository_SaveInvalid(t *testing.T) {
	repo := NewMemoryCanonRepository()
	doc := &domain.CanonDocument{
		SchemaVersion: "1.0",
		CampaignID:    "test",
	}
	if err := repo.Save("test", doc); err == nil {
		t.Fatal("expected error for invalid schema version")
	}
}

func TestMemoryNarrativeStateRepository_SaveLoad(t *testing.T) {
	repo := NewMemoryNarrativeStateRepository()
	campaignID := "test-campaign"

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     campaignID,
		CurrentSession: 2,
		LastUpdated:    time.Now(),
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-1", Name: "El Informador", Session: 2, Cause: "combat"},
		},
		RevealedClues: []domain.RevealedClue{
			{ID: "clue-1", Description: "Password is MORBUS", SourceAct: "act_1", SessionRevealed: 1},
		},
	}

	if err := repo.Save(campaignID, state); err != nil {
		t.Fatalf("failed to save state: %v", err)
	}

	if !repo.Exists(campaignID) {
		t.Fatal("expected Exists to return true")
	}

	loaded, err := repo.Load(campaignID)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	if loaded.CurrentSession != 2 {
		t.Fatalf("expected current session 2, got %d", loaded.CurrentSession)
	}

	if len(loaded.DeadNPCs) != 1 {
		t.Fatalf("expected 1 dead NPC, got %d", len(loaded.DeadNPCs))
	}
}

func TestMemoryNarrativeStateRepository_LoadNotFound(t *testing.T) {
	repo := NewMemoryNarrativeStateRepository()
	_, err := repo.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent state")
	}
}
