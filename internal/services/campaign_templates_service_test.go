package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

func TestCampaignService_CreateCampaign_WithTemplate(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, t.TempDir(), "")

	campaign, err := service.CreateCampaign("urban-test", "Urban Test", "A dark city", "Urban Fantasy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if campaign.Template != "Urban Fantasy" {
		t.Errorf("expected Template 'Urban Fantasy', got %q", campaign.Template)
	}
	if campaign.Status != "active" {
		t.Errorf("expected Status 'active', got %q", campaign.Status)
	}
}

func TestCampaignService_CreateCampaign_WithInvalidTemplate(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, t.TempDir(), "")

	// Invalid template should be ignored and campaign still created
	campaign, err := service.CreateCampaign("no-template", "No Template", "Setting", "Nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if campaign.Template != "" {
		t.Errorf("expected empty Template for invalid template name, got %q", campaign.Template)
	}
	if campaign.Status != "active" {
		t.Errorf("expected Status 'active', got %q", campaign.Status)
	}
}
