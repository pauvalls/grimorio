package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func TestHandleFactionDashboard_HappyPath(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	narrativeStateRepo := repository.NewMemoryNarrativeStateRepository()

	// Seed canon with faction entities (one secret)
	canonRepo.Save("test-campaign", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "fac1", Name: "Council of Light", Type: domain.EntityTypeFaction, Properties: map[string]any{}},
			{ID: "fac2", Name: "Shadow Guild", Type: domain.EntityTypeFaction, Properties: map[string]any{}},
			{ID: "fac3", Name: "Secret Order", Type: domain.EntityTypeFaction, Properties: map[string]any{"is_secret": true}},
		},
	})

	// Seed reputation via faction service
	factionService := services.NewFactionService(canonRepo, factionRepo)
	factionService.UpdateReputation(context.Background(), "test-campaign", "fac1", "party1", 30, "Helped villagers", "manual")
	factionService.UpdateReputation(context.Background(), "test-campaign", "fac2", "party1", -40, "Stole from guild", "manual")

	canonService := services.NewCanonService(canonRepo, narrativeStateRepo)
	dash := NewDashboardHandlers(factionService, canonService)

	handler := dash.HandleFactionDashboard()
	result, err := handler(context.Background(), newToolRequest("faction_reputation_dashboard", map[string]any{
		"campaign_id": "test-campaign",
	}))
	if err != nil {
		t.Fatalf("HandleFactionDashboard() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleFactionDashboard() returned error: %v", result.Content)
	}

	text := extractText(result)
	if !strings.Contains(text, "<!DOCTYPE html") && !strings.Contains(text, "<html") {
		t.Error("Response should contain HTML")
	}

	if !strings.Contains(text, "Council of Light") {
		t.Error("HTML should contain 'Council of Light'")
	}
	if !strings.Contains(text, "Shadow Guild") {
		t.Error("HTML should contain 'Shadow Guild'")
	}

	if strings.Contains(text, "Secret Order") {
		t.Error("Secret faction should be excluded from dashboard")
	}

	if !strings.Contains(text, "30") {
		t.Error("HTML should contain score '30'")
	}
	if !strings.Contains(text, "-40") {
		t.Error("HTML should contain score '-40'")
	}
}

func TestHandleFactionDashboard_NoData(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	narrativeStateRepo := repository.NewMemoryNarrativeStateRepository()

	// Seed empty canon
	canonRepo.Save("empty-campaign", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "empty-campaign",
	})

	canonService := services.NewCanonService(canonRepo, narrativeStateRepo)
	dash := NewDashboardHandlers(
		services.NewFactionService(canonRepo, factionRepo),
		canonService,
	)

	handler := dash.HandleFactionDashboard()
	result, err := handler(context.Background(), newToolRequest("faction_reputation_dashboard", map[string]any{
		"campaign_id": "empty-campaign",
	}))
	if err != nil {
		t.Fatalf("HandleFactionDashboard() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleFactionDashboard() returned error: %v", result.Content)
	}

	text := extractText(result)
	if !strings.Contains(text, "No faction data") {
		t.Errorf("Expected 'No faction data', got: %s", text)
	}
}

func TestHandleFactionDashboard_MissingCampaignID(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	narrativeStateRepo := repository.NewMemoryNarrativeStateRepository()

	canonService := services.NewCanonService(canonRepo, narrativeStateRepo)
	dash := NewDashboardHandlers(
		services.NewFactionService(canonRepo, factionRepo),
		canonService,
	)

	handler := dash.HandleFactionDashboard()
	result, err := handler(context.Background(), newToolRequest("faction_reputation_dashboard", map[string]any{}))
	if err != nil {
		t.Fatalf("HandleFactionDashboard() error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing campaign_id")
	}
}

func TestRepScoreColor(t *testing.T) {
	tests := []struct {
		score int8
		color string
	}{
		{50, "#FFD700"},
		{30, "#27AE60"},
		{20, "#27AE60"},
		{0, "#95A5A6"},
		{-19, "#95A5A6"},
		{-20, "#F39C12"},
		{-49, "#F39C12"},
		{-50, "#E74C3C"},
		{-100, "#E74C3C"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := repScoreColor(tt.score); got != tt.color {
				t.Errorf("repScoreColor(%d) = %s, want %s", tt.score, got, tt.color)
			}
		})
	}
}

func TestRepStatusLabel(t *testing.T) {
	tests := []struct {
		score int8
		label string
	}{
		{50, "Allied"},
		{30, "Friendly"},
		{0, "Neutral"},
		{-30, "Unfriendly"},
		{-50, "Hostile"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if got := repStatusLabel(tt.score); got != tt.label {
				t.Errorf("repStatusLabel(%d) = %s, want %s", tt.score, got, tt.label)
			}
		})
	}
}
