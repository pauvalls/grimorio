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
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()

	campaignService := services.NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo,
		"/tmp/test", "",
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

func TestHandleGenerateCharacter(t *testing.T) {
	_, charHandlers, _ := setupTestHandlers()

	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "char-test", Title: "Char Test"}); err != nil {
		t.Fatal(err)
	}

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
	if err := campaignRepo.Create(&domain.Campaign{Name: "quest-test", Title: "Quest Test"}); err != nil {
		t.Fatal(err)
	}

	handler := questHandlers.HandleCreateQuest()
	args := map[string]any{
		"campaign":    "quest-test",
		"quest_title": "Find the Sword",
		"quest_type":  "main",
		"hook":        "A stranger approaches...",
		"stakes":      "The kingdom's fate",
		"reward":      "1000 gold",
	}

	result, err := handler(context.Background(), newToolRequest("create_personal_quest", args))
	if err != nil {
		t.Fatalf("HandleCreateQuest() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleCreateQuest() returned error result: %v", result.Content)
	}
}

func TestHandleSaveLore(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "lore-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveLore()
	args := map[string]any{
		"campaign": "lore-test",
		"content":  "# World History\n\nLong ago...",
	}

	result, err := handler(context.Background(), newToolRequest("save_lore", args))
	if err != nil {
		t.Fatalf("HandleSaveLore() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveLore() returned error result: %v", result.Content)
	}
}

func TestHandleSaveLore_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveLore()
	args := map[string]any{
		"content": "Some lore",
	}

	result, err := handler(context.Background(), newToolRequest("save_lore", args))
	if err != nil {
		t.Fatalf("HandleSaveLore() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveLore() should return error for missing campaign")
	}
}

func TestHandleSaveNPCs(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "npc-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveNPCs()
	args := map[string]any{
		"campaign": "npc-test",
		"content":  "# NPCs\n\n## Gandalf\nA wizard...",
	}

	result, err := handler(context.Background(), newToolRequest("save_npcs", args))
	if err != nil {
		t.Fatalf("HandleSaveNPCs() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveNPCs() returned error result: %v", result.Content)
	}
}

func TestHandleSaveEncounters(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "enc-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveEncounters()
	args := map[string]any{
		"campaign": "enc-test",
		"content":  "# Encounters\n\n## Ambush\nA bandit ambush...",
	}

	result, err := handler(context.Background(), newToolRequest("save_encounters", args))
	if err != nil {
		t.Fatalf("HandleSaveEncounters() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveEncounters() returned error result: %v", result.Content)
	}
}

func TestHandleSaveBestiary(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "best-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveBestiary()
	args := map[string]any{
		"campaign": "best-test",
		"content":  "# Bestiary\n\n## Goblin\nA small creature...",
	}

	result, err := handler(context.Background(), newToolRequest("save_bestiary", args))
	if err != nil {
		t.Fatalf("HandleSaveBestiary() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveBestiary() returned error result: %v", result.Content)
	}
}

func TestHandleSaveMaps(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "map-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveMaps()
	args := map[string]any{
		"campaign": "map-test",
		"content":  "# Maps\n\n## Dungeon\nA dark dungeon...",
	}

	result, err := handler(context.Background(), newToolRequest("save_maps", args))
	if err != nil {
		t.Fatalf("HandleSaveMaps() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveMaps() returned error result: %v", result.Content)
	}
}

func TestHandleSaveIntroduction(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "intro-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveIntroduction()
	args := map[string]any{
		"campaign": "intro-test",
		"content":  "# Introduction\n\n## Story Overview\n...",
	}

	result, err := handler(context.Background(), newToolRequest("save_introduction", args))
	if err != nil {
		t.Fatalf("HandleSaveIntroduction() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveIntroduction() returned error result: %v", result.Content)
	}
}

func TestHandleSaveChapter(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "chapter-handler-test",
	})); err != nil {
		t.Fatal(err)
	}

	chapterContent := `# Capítulo 1: El Comienzo

## Áreas

### Área 1: Entrada

Descripción del área.

### Área 2: Taberna

Descripción de la taberna.

## Consecuencias

Fin.
`

	handler := handlers.HandleSaveChapter()
	args := map[string]any{
		"campaign":       "chapter-handler-test",
		"chapter_number": float64(1),
		"title":          "El Comienzo",
		"content":        chapterContent,
	}

	result, err := handler(context.Background(), newToolRequest("save_chapter", args))
	if err != nil {
		t.Fatalf("HandleSaveChapter() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveChapter() returned error result: %v", result.Content)
	}
}

func TestHandleSaveChapter_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveChapter()
	args := map[string]any{
		"chapter_number": float64(1),
		"title":          "Title",
		"content":        "content",
	}

	result, err := handler(context.Background(), newToolRequest("save_chapter", args))
	if err != nil {
		t.Fatalf("HandleSaveChapter() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveChapter() should return error for missing campaign")
	}
}

func TestHandleSaveChapter_InvalidChapterNumber(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "chapter-invalid-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveChapter()
	args := map[string]any{
		"campaign":       "chapter-invalid-test",
		"chapter_number": float64(0),
		"title":          "Title",
		"content":        "content",
	}

	result, err := handler(context.Background(), newToolRequest("save_chapter", args))
	if err != nil {
		t.Fatalf("HandleSaveChapter() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveChapter() should return error for invalid chapter_number")
	}
}

func TestHandleSaveChapter_NoAreas(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "chapter-no-areas-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveChapter()
	args := map[string]any{
		"campaign":       "chapter-no-areas-test",
		"chapter_number": float64(1),
		"title":          "Title",
		"content":        "# Capítulo 1\n\nSin áreas.",
	}

	result, err := handler(context.Background(), newToolRequest("save_chapter", args))
	if err != nil {
		t.Fatalf("HandleSaveChapter() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveChapter() should return error for content without areas")
	}
}

func TestHandleSaveSettingGuide(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "setting-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveSettingGuide()
	args := map[string]any{
		"campaign": "setting-test",
		"content":  "# Setting Guide\n\n## Geography\n...",
	}

	result, err := handler(context.Background(), newToolRequest("save_setting_guide", args))
	if err != nil {
		t.Fatalf("HandleSaveSettingGuide() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveSettingGuide() returned error result: %v", result.Content)
	}
}

func TestHandleSaveSettingGuide_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveSettingGuide()
	args := map[string]any{
		"content": "Some content",
	}

	result, err := handler(context.Background(), newToolRequest("save_setting_guide", args))
	if err != nil {
		t.Fatalf("HandleSaveSettingGuide() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveSettingGuide() should return error for missing campaign")
	}
}

func TestHandleSaveAppendices(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	// Create campaign first
	createHandler := handlers.HandleCreateCampaign()
	if _, err := createHandler(context.Background(), newToolRequest("create_campaign", map[string]any{
		"name": "appendix-test",
	})); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveAppendices()
	args := map[string]any{
		"campaign": "appendix-test",
		"content":  "# Appendices\n\n## Appendix A: Magic Items\n...",
	}

	result, err := handler(context.Background(), newToolRequest("save_appendices", args))
	if err != nil {
		t.Fatalf("HandleSaveAppendices() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveAppendices() returned error result: %v", result.Content)
	}
}

func TestHandleSaveAppendices_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveAppendices()
	args := map[string]any{
		"content": "Some content",
	}

	result, err := handler(context.Background(), newToolRequest("save_appendices", args))
	if err != nil {
		t.Fatalf("HandleSaveAppendices() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveAppendices() should return error for missing campaign")
	}
}

func TestHandleSaveIntroduction_MissingCampaign(t *testing.T) {
	handlers, _, _ := setupTestHandlers()

	handler := handlers.HandleSaveIntroduction()
	args := map[string]any{
		"content": "Some content",
	}

	result, err := handler(context.Background(), newToolRequest("save_introduction", args))
	if err != nil {
		t.Fatalf("HandleSaveIntroduction() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveIntroduction() should return error for missing campaign")
	}
}
