package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func TestSessionPrepHandler(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	svc := services.NewSessionPrepService(canonRepo, stateRepo, nil)
	handler := NewSessionPrepHandlers(svc)

	// Seed data
	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "npc-1", Name: "Eldrin", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
		SessionLog: []domain.SessionRecord{
			{SessionNum: 1, Date: time.Now(), Summary: "Started the journey."},
			{SessionNum: 2, Date: time.Now(), Summary: "Found the key."},
		},
		ActiveQuests: []domain.QuestState{
			{ID: "q1", Name: "Main Quest", Status: "active", SourceAct: "act-1", GiverNPC: "npc-1"},
		},
	}
	_ = stateRepo.Save("test-campaign", state)

	t.Run("valid campaign returns prep", func(t *testing.T) {
		args := map[string]any{
			"campaign_id": "test-campaign",
			"session_num": 3,
		}
		result, err := handler.HandleGenerateSessionPrep()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_session_prep",
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
		if !strings.Contains(text, `"session_num": 3`) {
			t.Fatalf("expected session 3 in result, got: %s", text)
		}
	})

	t.Run("missing campaign_id returns error", func(t *testing.T) {
		args := map[string]any{}
		result, err := handler.HandleGenerateSessionPrep()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_session_prep",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatalf("expected error result for missing campaign_id")
		}
	})

	t.Run("invalid campaign creates initial state and returns prep", func(t *testing.T) {
		args := map[string]any{
			"campaign_id": "nonexistent",
		}
		result, err := handler.HandleGenerateSessionPrep()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_session_prep",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success for invalid campaign (initial state created), got error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "nonexistent") {
			t.Fatalf("expected campaign_id in result, got: %s", text)
		}
	})

	t.Run("with_scenarios true returns encounter recommendations", func(t *testing.T) {
		args := map[string]any{
			"campaign_id":     "test-campaign",
			"session_num":     3,
			"with_scenarios":  true,
		}
		result, err := handler.HandleGenerateSessionPrep()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_session_prep",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.IsError {
			t.Fatalf("expected success, got error: %v", result.Content)
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "session_num") {
			t.Fatalf("expected session_num in result, got: %s", text)
		}
		// The result should be valid JSON containing prep data
		if !strings.Contains(text, "prep") {
			t.Fatalf("expected prep in result, got: %s", text)
		}
	})
}
