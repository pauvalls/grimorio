package handlers

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func setupCharacterHandlers() *CharacterHandlers {
	charRepo := repository.NewMemoryCharacterRepository()
	charService := services.NewCharacterService(charRepo)
	return NewCharacterHandlers(charService)
}

func TestHandleSaveCharacters(t *testing.T) {
	handlers := setupCharacterHandlers()

	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "save-chars-test", Title: "Save Chars Test"}); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveCharacters()
	args := map[string]any{
		"campaign": "save-chars-test",
		"characters": []any{
			map[string]any{
				"name":       "Gandalf",
				"race":       "humano",
				"class":      "mago",
				"level":      float64(5),
				"background": "sabio",
				"alignment":  "LG",
			},
			map[string]any{
				"name":       "Aragorn",
				"race":       "humano",
				"class":      "guerrero",
				"level":      float64(3),
				"background": "soldado",
				"alignment":  "NG",
			},
		},
	}

	result, err := handler(context.Background(), newToolRequest("save_characters", args))
	if err != nil {
		t.Fatalf("HandleSaveCharacters() error: %v", err)
	}

	if result.IsError {
		t.Errorf("HandleSaveCharacters() returned error result: %v", result.Content)
	}

	// Verify characters were saved by listing them
	chars, err := handlers.service.ListCharacters("save-chars-test")
	if err != nil {
		t.Fatalf("ListCharacters() error: %v", err)
	}
	if len(chars) != 2 {
		t.Errorf("expected 2 characters saved, got %d", len(chars))
	}
}

func TestHandleSaveCharacters_MissingCampaign(t *testing.T) {
	handlers := setupCharacterHandlers()

	handler := handlers.HandleSaveCharacters()
	args := map[string]any{
		"characters": []any{
			map[string]any{"name": "Gandalf"},
		},
	}

	result, err := handler(context.Background(), newToolRequest("save_characters", args))
	if err != nil {
		t.Fatalf("HandleSaveCharacters() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveCharacters() should return error for missing campaign")
	}
}

func TestHandleSaveCharacters_EmptyCharacters(t *testing.T) {
	handlers := setupCharacterHandlers()

	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "empty-chars-test", Title: "Empty Chars Test"}); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveCharacters()
	args := map[string]any{
		"campaign":   "empty-chars-test",
		"characters": []any{},
	}

	result, err := handler(context.Background(), newToolRequest("save_characters", args))
	if err != nil {
		t.Fatalf("HandleSaveCharacters() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveCharacters() should return error for empty characters array")
	}
}

func TestHandleSaveCharacters_InvalidCharacterData(t *testing.T) {
	handlers := setupCharacterHandlers()

	// Create campaign first
	campaignRepo := repository.NewMemoryCampaignRepository()
	if err := campaignRepo.Create(&domain.Campaign{Name: "invalid-char-test", Title: "Invalid Char Test"}); err != nil {
		t.Fatal(err)
	}

	handler := handlers.HandleSaveCharacters()
	args := map[string]any{
		"campaign": "invalid-char-test",
		"characters": []any{
			map[string]any{
				"name":  "",
				"race":  "humano",
				"class": "mago",
				"level": float64(5),
			},
		},
	}

	result, err := handler(context.Background(), newToolRequest("save_characters", args))
	if err != nil {
		t.Fatalf("HandleSaveCharacters() error: %v", err)
	}

	if !result.IsError {
		t.Errorf("HandleSaveCharacters() should return error for invalid character data")
	}
}
