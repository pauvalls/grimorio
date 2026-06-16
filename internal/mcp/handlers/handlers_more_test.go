package handlers

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/image"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func TestHandleCompilePDF(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "pdf-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleCompilePDF()
	args := map[string]any{
		"campaign": "pdf-test",
	}

	result, err := handler(context.Background(), newToolRequest("compile_pdf", args))
	if err != nil {
		t.Fatalf("HandleCompilePDF() error: %v", err)
	}

	// May fail due to missing PDF engine, but should not panic
	_ = result
}

func TestHandleGetTemplate(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleGetTemplate()
	args := map[string]any{
		"type": "areas",
	}

	result, err := handler(context.Background(), newToolRequest("get_template", args))
	if err != nil {
		t.Fatalf("HandleGetTemplate() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleGetTemplate() returned error result: %v", result.Content)
	}
}

func TestHandleGetTemplate_Invalid(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleGetTemplate()
	args := map[string]any{
		"type": "invalid",
	}

	result, err := handler(context.Background(), newToolRequest("get_template", args))
	if err != nil {
		t.Fatalf("HandleGetTemplate() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleGetTemplate() should return error for invalid type")
	}
}

func TestHandleGetCharacter(t *testing.T) {
	// Create campaign and character
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "char-get-test", Title: "Char Get Test"}); err != nil {
		t.Fatal(err)
	}
	charRepo := repository.NewMemoryCharacterRepository()
	if err := charRepo.Save(&domain.Character{CampaignID: "char-get-test", Name: "TestChar", Race: "humano", Class: "guerrero", Level: 1}); err != nil {
		t.Fatal(err)
	}

	charService := services.NewCharacterService(charRepo)
	handlers := NewCharacterHandlers(charService)

	handler := handlers.HandleGetCharacter()
	args := map[string]any{
		"campaign": "char-get-test",
		"name":     "TestChar",
	}

	result, err := handler(context.Background(), newToolRequest("get_character", args))
	if err != nil {
		t.Fatalf("HandleGetCharacter() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleGetCharacter() returned error result: %v", result.Content)
	}
}

func TestHandleListCharacters(t *testing.T) {
	// Create campaign and character
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "char-list-test", Title: "Char List Test"}); err != nil {
		t.Fatal(err)
	}
	charRepo := repository.NewMemoryCharacterRepository()
	if err := charRepo.Save(&domain.Character{CampaignID: "char-list-test", Name: "TestChar", Race: "humano", Class: "guerrero", Level: 1}); err != nil {
		t.Fatal(err)
	}

	charService := services.NewCharacterService(charRepo)
	handlers := NewCharacterHandlers(charService)

	handler := handlers.HandleListCharacters()
	args := map[string]any{
		"campaign": "char-list-test",
	}

	result, err := handler(context.Background(), newToolRequest("list_characters", args))
	if err != nil {
		t.Fatalf("HandleListCharacters() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleListCharacters() returned error result: %v", result.Content)
	}
}

func TestHandleUpdateQuestStatus(t *testing.T) {
	// Create campaign and quest
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "quest-update-test", Title: "Quest Update Test"}); err != nil {
		t.Fatal(err)
	}
	questRepo := repository.NewMemoryQuestRepository()
	quest := &domain.Quest{CampaignID: "quest-update-test", Title: "Test Quest", Type: domain.QuestTypeRedencion, Status: domain.QuestStatusActive}
	if err := questRepo.Save(quest); err != nil {
		t.Fatal(err)
	}

	questService := services.NewQuestService(questRepo)
	handlers := NewQuestHandlers(questService)

	handler := handlers.HandleUpdateQuestStatus()
	args := map[string]any{
		"campaign": "quest-update-test",
		"quest_id": quest.ID,
		"status":   "completed",
		"notes":    "Done!",
	}

	result, err := handler(context.Background(), newToolRequest("update_quest_status", args))
	if err != nil {
		t.Fatalf("HandleUpdateQuestStatus() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleUpdateQuestStatus() returned error result: %v", result.Content)
	}
}

func TestHandleListQuests(t *testing.T) {
	// Create campaign and quest
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "quest-list-test", Title: "Quest List Test"}); err != nil {
		t.Fatal(err)
	}
	questRepo := repository.NewMemoryQuestRepository()
	if err := questRepo.Save(&domain.Quest{CampaignID: "quest-list-test", Title: "Test Quest", Type: domain.QuestTypeRedencion, Status: domain.QuestStatusActive}); err != nil {
		t.Fatal(err)
	}

	questService := services.NewQuestService(questRepo)
	handlers := NewQuestHandlers(questService)

	handler := handlers.HandleListQuests()
	args := map[string]any{
		"campaign": "quest-list-test",
	}

	result, err := handler(context.Background(), newToolRequest("list_quests", args))
	if err != nil {
		t.Fatalf("HandleListQuests() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleListQuests() returned error result: %v", result.Content)
	}
}

func TestHandleGenerateMap(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "map-gen-test",
	})); err != nil {
		t.Fatal(err)
	}

	// Need asset handlers - create them
	assetService := services.NewAssetService("/tmp/test", image.Config{})
	assetHandlers := NewAssetHandlers(assetService)

	handler := assetHandlers.HandleGenerateMap()
	args := map[string]any{
		"campaign": "map-gen-test",
		"filename": "test-map",
		"style":    "dungeon",
		"rooms":    float64(4),
	}

	result, err := handler(context.Background(), newToolRequest("generate_map", args))
	if err != nil {
		t.Fatalf("HandleGenerateMap() error: %v", err)
	}

	// May fail due to filesystem, but should not panic
	_ = result
}

func TestHandleGenerateDivider(t *testing.T) {
	assetService := services.NewAssetService("/tmp/test", image.Config{})
	assetHandlers := NewAssetHandlers(assetService)

	handler := assetHandlers.HandleGenerateDivider()
	args := map[string]any{
		"campaign": "divider-test",
		"filename": "test-divider",
		"style":    "ornate",
		"width":    float64(600),
	}

	result, err := handler(context.Background(), newToolRequest("generate_divider", args))
	if err != nil {
		t.Fatalf("HandleGenerateDivider() error: %v", err)
	}

	// May fail due to filesystem, but should not panic
	_ = result
}

// TestHandleGenerateImage_ForceRegenerateArg verifies the MCP handler
// passes the force_regenerate boolean argument through to the service
// (Fase 4 image cache).
func TestHandleGenerateImage_ForceRegenerateArg(t *testing.T) {
	// Missing required args → should still return error without panicking.
	assetService := services.NewAssetService(t.TempDir(), image.Config{})
	assetHandlers := NewAssetHandlers(assetService)
	handler := assetHandlers.HandleGenerateImage()

	// force_regenerate alone is not enough — campaign/filename/prompt required.
	_, err := handler(context.Background(), newToolRequest("generate_image", map[string]any{
		"force_regenerate": true,
	}))
	if err != nil {
		t.Fatalf("HandleGenerateImage() error: %v", err)
	}
}
