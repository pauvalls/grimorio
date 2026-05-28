package handlers

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupPrologueTest(t *testing.T) *PrologueHandlers {
	t.Helper()

	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()

	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo,
		"/tmp/test-prologue", "",
	)
	prologueService := services.NewPrologueService("/tmp/test-prologue", canonRepo)

	return NewPrologueHandlers(prologueService, campaignService)
}

func TestHandleGeneratePrologue_InvalidArgs(t *testing.T) {
	handlers := setupPrologueTest(t)
	handler := handlers.HandleGeneratePrologue()

	// Missing campaign should error
	result, err := handler(context.Background(), newToolRequest("grimorio_generate_prologue", nil))
	if err != nil {
		t.Fatalf("HandleGeneratePrologue() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleGeneratePrologue() should return error for invalid args")
	}
}

func TestHandleGeneratePrologue_MissingCampaign(t *testing.T) {
	handlers := setupPrologueTest(t)
	handler := handlers.HandleGeneratePrologue()

	args := map[string]any{
		"campaign": "",
	}

	result, err := handler(context.Background(), newToolRequest("grimorio_generate_prologue", args))
	if err != nil {
		t.Fatalf("HandleGeneratePrologue() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleGeneratePrologue() should return error for empty campaign")
	}
}

func TestHandleGeneratePrologue_NonexistentCampaign(t *testing.T) {
	handlers := setupPrologueTest(t)
	handler := handlers.HandleGeneratePrologue()

	args := map[string]any{
		"campaign": "nonexistent-campaign",
	}

	result, err := handler(context.Background(), newToolRequest("grimorio_generate_prologue", args))
	if err != nil {
		t.Fatalf("HandleGeneratePrologue() error: %v", err)
	}

	// Should error because campaign dir doesn't exist on filesystem
	if !result.IsError {
		t.Errorf("HandleGeneratePrologue() should return error for nonexistent campaign")
	}
}
