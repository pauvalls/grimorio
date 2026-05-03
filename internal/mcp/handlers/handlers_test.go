package handlers

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupTestHandlers() (*CampaignHandlers, *CharacterHandlers, *QuestHandlers) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()

	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		"/tmp/test", "wkhtmltopdf",
	)
	characterService := services.NewCharacterService(charRepo)
	questService := services.NewQuestService(questRepo)

	return NewCampaignHandlers(campaignService),
		NewCharacterHandlers(characterService),
		NewQuestHandlers(questService)
}

func newToolRequest(name string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

func TestHandleCreateCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()
	handler := handlers.HandleCreateCampaign()

	args := map[string]any{
		"name":    "test-campaign",
		"title":   "Test Campaign",
		"setting": "Forgotten Realms",
	}

	result, err := handler(context.Background(), newToolRequest("create_campaign", args))
	if err != nil {
		t.Fatalf("HandleCreateCampaign() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleCreateCampaign() returned error result: %v", result.Content)
	}
}

func TestHandleCreateCampaign_InvalidArgs(t *testing.T) {
	handlers, _, _ := setupTestHandlers()
	handler := handlers.HandleCreateCampaign()

	result, err := handler(context.Background(), newToolRequest("create_campaign", nil))
	if err != nil {
		t.Fatalf("HandleCreateCampaign() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleCreateCampaign() should return error for invalid args")
	}
}

func TestHandleSaveAct(t *testing.T) {
	handlers, _, _ := setupTestHandlers()
	
	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "act-test",
	}))

	// Now save act
	actHandler := handlers.HandleSaveAct()
	args := map[string]any{
		"campaign":   "act-test",
		"act_number": float64(1),
		"title":      "The Beginning",
		"content":    "Once upon a time...",
	}

	result, err := actHandler(context.Background(), newToolRequest("save_act", args))
	if err != nil {
		t.Fatalf("HandleSaveAct() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveAct() returned error result: %v", result.Content)
	}
}

func TestHandleGenerateCharacter(t *testing.T) {
	_, charHandlers, _ := setupTestHandlers()
	
	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	campaignRepo.Create(&domain.Campaign{Name: "char-test", Title: "Char Test"})
	
	handler := charHandlers.HandleGenerateCharacter()
	args := map[string]any{
		"campaign":   "char-test",
		"name":       "Gandalf",
		"race":       "humano",
		"class":      "mago",
		"level":      float64(5),
		"background": "sabio",
		"alignment":  "LG",
	}

	result, err := handler(context.Background(), newToolRequest("generate_character", args))
	if err != nil {
		t.Fatalf("HandleGenerateCharacter() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleGenerateCharacter() returned error result: %v", result.Content)
	}
}

func TestHandleCreateQuest(t *testing.T) {
	_, _, questHandlers := setupTestHandlers()
	
	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	campaignRepo.Create(&domain.Campaign{Name: "quest-test", Title: "Quest Test"})
	
	handler := questHandlers.HandleCreateQuest()
	args := map[string]any{
		"campaign":      "quest-test",
		"quest_title":   "Find the Sword",
		"quest_type":    "main",
		"hook":          "A stranger approaches...",
		"stakes":        "The kingdom's fate",
		"reward":        "1000 gold",
	}

	result, err := handler(context.Background(), newToolRequest("create_personal_quest", args))
	if err != nil {
		t.Fatalf("HandleCreateQuest() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleCreateQuest() returned error result: %v", result.Content)
	}
}
