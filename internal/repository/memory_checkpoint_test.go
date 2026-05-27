package repository

import (
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestMemoryCheckpointRepository(t *testing.T) {
	// Setup
	repo := NewMemoryCheckpointRepository()

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
	if loaded.Metadata["number"] != 42 {
		t.Errorf("expected Metadata number to be 42, got %v", loaded.Metadata["number"])
	}

	// Test ListCheckpoints
	list, err := repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 checkpoint, got %d", len(list))
	}

	// Test DeleteCheckpoint
	err = repo.DeleteCheckpoint(campaignID, loaded.ID)
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
}

func TestMemoryCheckpointRepository_Invalid(t *testing.T) {
	// Setup
	repo := NewMemoryCheckpointRepository()

	campaignID := "test-campaign"
	// State with empty SchemaVersion should fail validation
	invalidState := &domain.NarrativeState{
		SchemaVersion:  "",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Test SaveCheckpoint with invalid state - Checkpoint validation only checks nil
	// So this test focuses on other invalid scenarios
	err := repo.SaveCheckpoint(campaignID, "", 1, "chapter-1", invalidState, "hash123", nil)
	if err == nil {
		t.Error("expected error when saving checkpoint with empty checkpoint type, got nil")
	}

	// Test LoadCheckpoint on non-existent campaign
	_, err = repo.LoadCheckpoint("non-existent-campaign", "session_end")
	if err == nil {
		t.Error("expected error when loading from non-existent campaign, got nil")
	}

	// Test DeleteCheckpoint on non-existent checkpoint
	err = repo.DeleteCheckpoint(campaignID, "non-existent-id")
	if err == nil {
		t.Error("expected error when deleting non-existent checkpoint, got nil")
	}
}

func TestMemoryCheckpointRepository_MultipleCheckpoints(t *testing.T) {
	// Setup
	repo := NewMemoryCheckpointRepository()

	campaignID := "test-campaign"
	state := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Create multiple checkpoints
	for i := 1; i <= 3; i++ {
		err := repo.SaveCheckpoint(campaignID, "session_end", i, "chapter-1", state, "hash"+string(rune(i)), nil)
		if err != nil {
			t.Fatalf("SaveCheckpoint failed: %v", err)
		}
		time.Sleep(time.Millisecond) // Ensure different timestamps
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

	// Test LoadCheckpoint returns latest
	loaded, err := repo.LoadCheckpoint(campaignID, "session_end")
	if err != nil {
		t.Fatalf("LoadCheckpoint failed: %v", err)
	}
	if loaded.SessionNum != 3 {
		t.Errorf("expected SessionNum 3 (latest), got %d", loaded.SessionNum)
	}
}

func TestMemoryCheckpointRepository_MultipleCampaigns(t *testing.T) {
	// Setup
	repo := NewMemoryCheckpointRepository()

	state1 := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     "campaign-1",
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	state2 := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     "campaign-2",
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Save checkpoints for different campaigns
	err := repo.SaveCheckpoint("campaign-1", "session_end", 1, "chapter-1", state1, "hash1", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	err = repo.SaveCheckpoint("campaign-2", "session_end", 1, "chapter-1", state2, "hash2", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Test isolation between campaigns
	list1, err := repo.ListCheckpoints("campaign-1")
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list1) != 1 {
		t.Errorf("expected 1 checkpoint for campaign-1, got %d", len(list1))
	}

	list2, err := repo.ListCheckpoints("campaign-2")
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list2) != 1 {
		t.Errorf("expected 1 checkpoint for campaign-2, got %d", len(list2))
	}
}

func TestMemoryCheckpointRepository_MultipleTypes(t *testing.T) {
	// Setup
	repo := NewMemoryCheckpointRepository()

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

func TestMemoryCheckpointRepository_DeleteSpecific(t *testing.T) {
	// Setup
	repo := NewMemoryCheckpointRepository()

	campaignID := "test-campaign"
	state := &domain.NarrativeState{
		SchemaVersion:  "v2",
		CampaignID:     campaignID,
		CurrentSession: 1,
		LastUpdated:    time.Now().UTC(),
	}

	// Create multiple checkpoints
	err := repo.SaveCheckpoint(campaignID, "session_end", 1, "chapter-1", state, "hash1", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	err = repo.SaveCheckpoint(campaignID, "session_end", 2, "chapter-1", state, "hash2", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	err = repo.SaveCheckpoint(campaignID, "session_end", 3, "chapter-1", state, "hash3", nil)
	if err != nil {
		t.Fatalf("SaveCheckpoint failed: %v", err)
	}

	// Get the checkpoint to delete (session 2)
	list, err := repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}

	var checkpointToDelete string
	for _, cp := range list {
		if cp.SessionNum == 2 {
			checkpointToDelete = cp.ID
			break
		}
	}

	if checkpointToDelete == "" {
		t.Fatal("could not find checkpoint to delete")
	}

	// Delete specific checkpoint
	err = repo.DeleteCheckpoint(campaignID, checkpointToDelete)
	if err != nil {
		t.Fatalf("DeleteCheckpoint failed: %v", err)
	}

	// Verify only one was deleted
	list, err = repo.ListCheckpoints(campaignID)
	if err != nil {
		t.Fatalf("ListCheckpoints failed: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2 checkpoints after deletion, got %d", len(list))
	}

	// Verify session 2 is gone
	for _, cp := range list {
		if cp.SessionNum == 2 {
			t.Error("session 2 should have been deleted")
		}
	}
}
