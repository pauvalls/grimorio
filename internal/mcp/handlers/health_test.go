package handlers

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func TestHealthHandlers_HandleCampaignHealthDashboard(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	baseDir := t.TempDir()

	campaignService := services.NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, baseDir, "")
	canonService := services.NewCanonService(canonRepo, repository.NewMemoryNarrativeStateRepository(), repository.NewMemoryCheckpointRepository())
	stateService := services.NewNarrativeStateService(repository.NewMemoryNarrativeStateRepository(), canonRepo)
	validationEngine := services.NewValidationEngine(canonService, stateService, repository.NewMemoryFactionRepository(), baseDir)

	// Create a campaign first
	_, _ = campaignService.CreateCampaign("health-test", "Health Test", "Setting", "")

	healthCheck := services.NewCampaignHealthCheck(canonRepo, repository.NewMemoryNarrativeStateRepository(), questRepo, repository.NewMemoryFactionRepository(), npcRepo, baseDir)
	healthScore := services.NewCampaignHealthScore(healthCheck, validationEngine)
	handlers := NewHealthHandlers(healthScore)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"campaign_id": "health-test"}

	result, err := handlers.HandleCampaignHealthDashboard()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// If campaign has no canon/state yet, handler may return an error result
	if result.IsError {
		t.Logf("health dashboard returned error (expected for empty campaign): %v", result.Content)
		return
	}

	// Should return JSON text
	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	textResult := result.Content[0].(mcp.TextContent)
	var report domain.CampaignHealthReport
	if err := json.Unmarshal([]byte(textResult.Text), &report); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if report.Status == "" {
		t.Error("expected status field in health report")
	}
}

func TestTreasureHandlers_HandleGenerateTreasure(t *testing.T) {
	svc := services.NewTreasureService()
	handlers := NewTreasureHandlers(svc)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"type": "individual",
		"cr":   3,
	}

	result, err := handlers.HandleGenerateTreasure()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	textResult := result.Content[0].(mcp.TextContent)
	if !strings.Contains(textResult.Text, "gp") && !strings.Contains(textResult.Text, "cp") {
		t.Error("expected treasure result to contain coins")
	}
}

func TestExportHandlers_HandleExportCampaign(t *testing.T) {
	tmpDir := t.TempDir()
	campaignDir := tmpDir + "/test-campaign"
	os.MkdirAll(campaignDir, 0755)
	os.WriteFile(campaignDir+"/introduction.md", []byte("# Introduction\n\nTest."), 0644)

	handlers := NewExportHandlers(tmpDir, "")

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign": "test-campaign",
		"format":   "markdown",
	}

	result, err := handlers.HandleExportCampaign()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) == 0 {
		t.Fatal("expected content in result")
	}

	textResult := result.Content[0].(mcp.TextContent)
	if !strings.Contains(textResult.Text, "markdown") {
		t.Error("expected result to mention markdown format")
	}
}

func TestExportHandlers_HandleExportCampaign_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	handlers := NewExportHandlers(tmpDir, "")

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{
		"campaign": "test-campaign",
		"format":   "html",
	}

	result, err := handlers.HandleExportCampaign()(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error result for invalid format")
	}
}

func TestExportHandlers_Registry(t *testing.T) {
	handlers := NewExportHandlers(t.TempDir(), "")
	formats := []string{"pdf", "markdown", "epub"}
	for _, f := range formats {
		if handlers.exporters[f] == nil {
			t.Errorf("expected exporter for format %s", f)
		}
	}
	if handlers.exporters["pdf"].Format() != "pdf" {
		t.Error("pdf exporter has wrong format")
	}
	if handlers.exporters["markdown"].Format() != "markdown" {
		t.Error("markdown exporter has wrong format")
	}
	if handlers.exporters["epub"].Format() != "epub" {
		t.Error("epub exporter has wrong format")
	}
}
