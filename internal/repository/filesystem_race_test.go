package repository

import (
	"sync"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestFilesystemCanonRepository_Race(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemCanonRepository(tmpDir)
	campaignID := "race-campaign"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doc := &domain.CanonDocument{
				SchemaVersion: domain.SchemaVersionV2,
				CampaignID:    campaignID,
				Entities: []domain.CanonEntity{
					{ID: "act-1", Name: "Act One", Type: domain.EntityTypeLocation, Role: "act"},
				},
			}
			_ = repo.Save(campaignID, doc)
			_, _ = repo.Load(campaignID)
			_ = repo.Exists(campaignID)
		}(i)
	}
	wg.Wait()
}

func TestFilesystemNarrativeStateRepository_Race(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemNarrativeStateRepository(tmpDir)
	campaignID := "race-campaign"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			state := &domain.NarrativeState{
				SchemaVersion:  domain.SchemaVersionV2,
				CampaignID:     campaignID,
				CurrentSession: i,
			}
			_ = repo.Save(campaignID, state)
			_, _ = repo.Load(campaignID)
			_ = repo.Exists(campaignID)
		}(i)
	}
	wg.Wait()
}

func TestFilesystemFactionRepository_Race(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemFactionRepository(tmpDir)
	campaignID := "race-campaign"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			matrix := &domain.FactionReputationMatrix{
				CampaignID: campaignID,
				Entries: []domain.ReputationEntry{
					{FactionID: "faction-1", PartyID: "party-1", Score: int8(i)},
				},
			}
			_ = repo.Save(campaignID, matrix)
			_, _ = repo.Load(campaignID)
		}(i)
	}
	wg.Wait()
}

func TestFilesystemQuestRepository_Race(t *testing.T) {
	tmpDir := t.TempDir()
	repo := NewFilesystemQuestRepository(tmpDir)
	campaignID := "race-campaign"

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			quest := &domain.Quest{
				ID:         "quest-1",
				CampaignID: campaignID,
				Title:      "Test Quest",
				Status:     domain.QuestStatusActive,
			}
			_ = repo.Save(quest)
			_, _ = repo.Read(campaignID, "quest-1")
			_, _ = repo.List(campaignID)
			_, _ = repo.ListByStatus(campaignID, domain.QuestStatusActive)
		}(i)
	}
	wg.Wait()
}
