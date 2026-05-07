package handlers

import (
	"context"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
)

func TestFlowchartHandler(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	svc := services.NewFlowchartService(canonRepo)
	handler := NewFlowchartHandlers(svc)

	// Seed data
	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "act-1", Name: "The Beginning", Type: domain.EntityTypeLocation, Role: "act"},
			{ID: "act-2", Name: "The Twist", Type: domain.EntityTypeLocation, Role: "act"},
		},
		Relationships: []domain.CanonRelationship{
			{ID: "rel-1", From: "act-1", To: "act-2", Type: domain.RelationshipTypeAlly},
		},
	}
	_ = canonRepo.Save("test-campaign", doc)

	t.Run("valid campaign returns mermaid", func(t *testing.T) {
		args := map[string]any{
			"campaign_id":  "test-campaign",
			"detail_level": "overview",
		}
		result, err := handler.HandleGenerateFlowchart()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_flowchart",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result == nil {
			t.Fatalf("expected result, got nil")
		}
		text := result.Content[0].(mcp.TextContent).Text
		if !strings.Contains(text, "flowchart TD") {
			t.Fatalf("expected mermaid flowchart in result, got: %s", text)
		}
	})

	t.Run("invalid detail level returns error", func(t *testing.T) {
		args := map[string]any{
			"campaign_id":  "test-campaign",
			"detail_level": "invalid",
		}
		result, err := handler.HandleGenerateFlowchart()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_flowchart",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatalf("expected error for invalid detail_level")
		}
	})

	t.Run("missing campaign_id returns error", func(t *testing.T) {
		args := map[string]any{}
		result, err := handler.HandleGenerateFlowchart()(context.Background(), mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Arguments: args,
				Name:      "generate_flowchart",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatalf("expected error for missing campaign_id")
		}
	})
}
