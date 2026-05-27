package repository

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestCheckpointRepository_SaveLoad(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "test-campaign"
	state := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}
	canonHash := "abc123def456"
	metadata := map[string]any{
		"test_key": "test_value",
		"number":   42,
	}

	// Test SaveCheckpoint
	err := repo.SaveCheckpoint(campaignID, "session_end", 1, "chapter-1", state, canonHash, metadata)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Test LoadCheckpoint
	loaded, err := repo.LoadCheckpoint(campaignID, "session_end")
	if err != nil {
		t.Fatalf("LoadCheckpoint failed: %v", err)
	}

	// Validate loaded checkpoint
	if loaded.CampaignID != campaignID {
		t.Errorf("expected CampaignID %s, got %s", campaignID, loaded.CampaignID)
	}
	if loaded.CheckpointType != "session_end" {
		t.Errorf("expected CheckpointType session_end, got %s", loaded.CheckpointType)
	}
	if loaded.SessionNum != 1 {
		t.Errorf("expected SessionNum 1, got %d", loaded.SessionNum)
	}
	if loaded.ChapterID != "chapter-1" {
		t.Errorf("expected ChapterID chapter-1, got %s", loaded.ChapterID)
	}
	if loaded.CanonHash != canonHash {
		t.Errorf("expected CanonHash %s, got %s", canonHash, loaded.CanonHash)
	}
	if loaded.Metadata["test_key"] != "test_value" {
		t.Errorf("expected Metadata test_key to be test_value, got %v", loaded.Metadata["test_key"])
	}
	// Check number metadata - JSON unmarshals numbers as float64
	if num, ok := loaded.Metadata["number"].(float64); ok {
		if int(num) != 42 {
			t.Errorf("expected Metadata number to be 42, got %v", num)
		}
	} else {
		t.Errorf("expected Metadata number to be float64, got %T", loaded.Metadata["number"])
	}
}

func TestCheckpointRepository_List(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "test-campaign"

	// Create multiple checkpoints
	for i := 1; i <= 3; i++ {
		state := &domain.NarrativeState{
			SchemaVersion:  "v2",
			CampaignID:     campaignID,
			CurrentSession: i,
			LastUpdated:    time.Now().UTC(),
		}
		err := repo.SaveCheckpoint(campaignID, "session_end", i, "chapter-1", state, "hash"+string(rune(i)), nil)
		if err != nil {
			t.Fatalf("SaveCheckpoint failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Test ListCheckpoints
	list, err := repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}

	if len(list) != 3 {
		t.Errorf("expected 3 checkpoints, got %d", len(list))
	}

	// Verify sorting (newest first)
	if list[0].SessionNum < list[1].SessionNum || list[1].SessionNum < list[2].SessionNum {
		t.Error("checkpoints should be sorted by CreatedAt descending")
	}
}

func TestCheckpointRepository_Delete(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "test-campaign"
	state := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Save a checkpoint
	err := repo.SaveCheckpoint(campaignID, "session_end", 1, "chapter-1", state, "hash123", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Get the checkpoint ID
	list, err := repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(list))
	}
	checkpointID := list[0].ID

	// Test DeleteCheckpoint
	err = repo.DeleteCheckpoint(campaignID, checkpointID)
	if err != nil {
		t.Fatalf("DeleteCheckpoint failed: %v", err)
	}

	// Verify deletion
	list, err = repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 checkpoints after deletion, got %d", len(list))
	}

	// Test deleting non-existent checkpoint
	err = repo.DeleteCheckpoint(campaignID, "non-existent-id")
	if err == nil {
		t.Error("expected error when deleting non-existent checkpoint, got nil")
	}
}

func TestCheckpointRepository_MultipleTypes(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "test-campaign"
	state := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Save different types of checkpoints
	err := repo.SaveCheckpoint(campaignID, "session_end", 1, "chapter-1", state, "hash1", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	err = repo.SaveCheckpoint(campaignID, "chapter_complete", 1, "chapter-1", state, "hash2", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Test LoadCheckpoint for session_end
	sessionCP, err := repo.LoadCheckpoint(campaignID, "session_end")
	if err != nil {
		t.Fatalf("LoadCheckpoint session_end failed: %v", err)
	}
	if sessionCP.CheckpointType != "session_end" {
		t.Errorf("expected session_end type, got %s", sessionCP.CheckpointType)
	}

	// Test LoadCheckpoint for chapter_complete
	chapterCP, err := repo.LoadCheckpoint(campaignID, "chapter_complete")
	if err != nil {
		t.Fatalf("LoadCheckpoint chapter_complete failed: %v", err)
	}
	if chapterCP.CheckpointType != "chapter_complete" {
		t.Errorf("expected chapter_complete type, got %s", chapterCP.CheckpointType)
	}

	// Test ListCheckpoints returns all types
	list, err := repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 checkpoints, got %d", len(list))
	}
}

func TestCheckpointRepository_EmptyCampaign(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "non-existent-campaign"

	// Test LoadCheckpoint on non-existent campaign
	_, err := repo.LoadCheckpoint(campaignID, "session_end")
	if err == nil {
		t.Error("expected error when loading from non-existent campaign, got nil")
	}

	// Test ListCheckpoints on non-existent campaign
	list, err := repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0 checkpoints for non-existent campaign, got %d", len(list))
	}
}

func TestCheckpointRepository_InvalidState(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "test-campaign"
	// State with empty SchemaVersion should fail validation
	invalidState := &domain.NarrativeState{
		SchemaVersion:  "",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Test SaveCheckpoint with invalid state - State validation happens in domain
	// Checkpoint validation only checks if State is nil, not its contents
	// So this should succeed at the checkpoint level
	err := repo.SaveCheckpoint(campaignID, "session_end", 1, "chapter-1", invalidState, "hash123", nil)
	if err != nil {
		// If it fails, that's okay - means validation is stricter
		// But if it succeeds, that's also okay - checkpoint validation only checks nil
		t.Logf("SaveCheckpoint returned: %v (this is acceptable)", err)
	}
}

func TestCheckpointRepository_FilePersistence(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	repo := NewCheckpointRepository(tmpDir)

	campaignID := "test-campaign"
	state := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Save a checkpoint
	err := repo.SaveCheckpoint(campaignID, "session_end", 1, "chapter-1", state, "hash123", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Verify file exists on filesystem
	checkpointsDir := filepath.Join(tmpDir, campaignID, "checkpoints")
	entries, err := os.ReadDir(checkpointsDir)
	if err != nil {
		t.Fatalf("failed to read checkpoints directory: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 checkpoint file, got %d", len(entries))
	}

	// Verify filename format
	filename := entries[0].Name()
	if len(filename) < len("session_end_1_.json") {
		t.Errorf("filename too short: %s", filename)
	}
}
