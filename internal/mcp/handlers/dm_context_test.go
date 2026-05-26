package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	fsrepo "github.com/pauvalls/grimorio/internal/repository/fs"
	"github.com/pauvalls/grimorio/internal/services"
)

func TestDMContextHandler(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create repositories
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	monsterRepo := fsrepo.NewFilesystemMonsterRepository(tmpDir)
	areaRepo := fsrepo.NewFilesystemAreaRepositoryV3(tmpDir)

	sessionPrepSvc := services.NewSessionPrepService(canonRepo, stateRepo, nil)
	dmContextSvc := services.NewDMContextService(
		canonRepo, stateRepo, charRepo, npcRepo, questRepo,
		monsterRepo, areaRepo, factionRepo, sessionPrepSvc, tmpDir,
	)
	handler := NewDMContextHandlers(dmContextSvc)

	// Seed data
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "npc-1", Name: "Eldrin", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
		},
	}
	_ = canonRepo.Save("test-campaign", canonDoc)

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 3,
		SessionLog: []domain.SessionRecord{
			{SessionNum: 1, Date: time.Now(), Summary: "Started the journey."},
			{SessionNum: 2, Date: time.Now(), Summary: "Found the key."},
		},
		ActiveQuests: []domain.QuestState{
			{ID: "q1", Name: "Main Quest", Status: "active", SourceAct: "act-1", GiverNPC: "npc-1"},
		},
	}
	_ = stateRepo.Save("test-campaign", state)

	_ = charRepo.Save(&domain.Character{
		CampaignID: "test-campaign",
		Name:       "Aric", Race: "humano", Class: "guerrero", Level: 3,
		Background: "soldado", Alignment: "LG",
		HP:    domain.HP{Current: 25, Maximum: 30},
		AC:    16,
		Stats: domain.Stats{STR: 16, DEX: 12, CON: 14, INT: 10, WIS: 13, CHA: 8},
	})

	_ = questRepo.Save(&domain.Quest{
		CampaignID:  "test-campaign",
		ID:          "q1",
		Title:       "Main Quest",
		Status:      domain.QuestStatusActive,
		Type:        domain.QuestTypeMain,
		RelatedNPCs: []string{"Eldrin"},
	})

	t.Run("valid campaign returns context", func(t *testing.T) {
		args := map[string]any{
			"campaign_id":      "test-campaign",
			"session_num":      3,
			"include_prologue": false,
		}
		result, err := handler.HandleDMContext()(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "dm_session_context",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatalf("expected result, got nil")
		}
		if len(result.Content) == 0 {
			t.Fatalf("expected content in result")
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, `"campaign_id": "test-campaign"`) {
			t.Fatalf("expected campaign_id in result, got: %s", text)
		}
		if !strings.Contains(text, `"session_num": 3`) {
			t.Fatalf("expected session 3 in result, got: %s", text)
		}
		if !strings.Contains(text, `"name": "Aric"`) {
			t.Fatalf("expected character Aric in result, got: %s", text)
		}
	})

	t.Run("default session_num from narrative state", func(t *testing.T) {
		args := map[string]any{
			"campaign_id": "test-campaign",
		}
		result, err := handler.HandleDMContext()(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "dm_session_context",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, `"session_num": 3`) {
			t.Fatalf("expected default session 3 from narrative state, got: %s", text)
		}
	})

	t.Run("missing campaign_id returns error", func(t *testing.T) {
		args := map[string]any{}
		result, err := handler.HandleDMContext()(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "dm_session_context",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatalf("expected error result for missing campaign_id")
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "campaign_id is required") {
			t.Fatalf("expected 'campaign_id is required' error, got: %s", text)
		}
	})

	t.Run("invalid campaign_id format returns error", func(t *testing.T) {
		args := map[string]any{
			"campaign_id": "La Llave",
		}
		result, err := handler.HandleDMContext()(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "dm_session_context",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatalf("expected error result for invalid campaign_id format")
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "campaign_id must be kebab-case") {
			t.Fatalf("expected 'campaign_id must be kebab-case' error, got: %s", text)
		}
	})

	t.Run("empty repos returns minimal payload", func(t *testing.T) {
		// Create a minimal campaign with only canon and state
		_ = canonRepo.Save("minimal-campaign", &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "minimal-campaign",
		})
		_ = stateRepo.Save("minimal-campaign", &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     "minimal-campaign",
			CurrentSession: 1,
		})

		args := map[string]any{
			"campaign_id": "minimal-campaign",
		}
		result, err := handler.HandleDMContext()(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "dm_session_context",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success for minimal campaign, got error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, `"campaign_id": "minimal-campaign"`) {
			t.Fatalf("expected campaign_id in result, got: %s", text)
		}
		// Should not contain nil fields
		if strings.Contains(text, "null") {
			t.Fatalf("expected no null values in payload, got: %s", text)
		}
	})

	t.Run("payload under size cap", func(t *testing.T) {
		args := map[string]any{
			"campaign_id": "test-campaign",
		}
		result, err := handler.HandleDMContext()(ctx, mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "dm_session_context",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if len(text) > 100*1024 {
			t.Fatalf("payload exceeds 100KB: %d bytes", len(text))
		}
	})
}
