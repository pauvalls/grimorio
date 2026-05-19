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

func setupVizTest() (*VizHandlers, *services.CanonService) {
	canonRepo := repository.NewMemoryCanonRepository()
	narrativeStateRepo := repository.NewMemoryNarrativeStateRepository()
	canonService := services.NewCanonService(canonRepo, narrativeStateRepo)
	vizHandlers := NewVizHandlers(canonService)
	return vizHandlers, canonService
}

func TestHandleRelationshipGraph_HappyPath(t *testing.T) {
	viz, canonService := setupVizTest()

	// Seed canon with some entities and relationships
	canonService.SaveCanon(context.Background(), &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Entities: []domain.CanonEntity{
			{ID: "npc1", Name: "Gandalf", Type: domain.EntityTypeNPC},
			{ID: "loc1", Name: "Moria", Type: domain.EntityTypeLocation},
			{ID: "fac1", Name: "Fellowship", Type: domain.EntityTypeFaction},
		},
		Relationships: []domain.CanonRelationship{
			{ID: "rel1", From: "npc1", To: "loc1", Type: domain.RelationshipTypeAlly, Strength: 5},
		},
	})

	handler := viz.HandleGenerateRelationshipGraph()
	result, err := handler(context.Background(), newToolRequest("visualize_relationship_graph", map[string]any{
		"campaign_id": "test-campaign",
	}))
	if err != nil {
		t.Fatalf("HandleGenerateRelationshipGraph() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleGenerateRelationshipGraph() returned error: %v", result.Content)
	}

	// Should return HTML content
	text := extractText(result)
	if !strings.Contains(text, "<!DOCTYPE html") && !strings.Contains(text, "<html") {
		t.Error("Response should contain HTML")
	}
	if !strings.Contains(text, "Gandalf") {
		t.Error("HTML should contain entity name 'Gandalf'")
	}
	if !strings.Contains(text, "Moria") {
		t.Error("HTML should contain entity name 'Moria'")
	}
}

func TestHandleRelationshipGraph_NoEntities(t *testing.T) {
	viz, canonService := setupVizTest()

	// Seed empty canon for this campaign
	canonService.SaveCanon(context.Background(), &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "empty-campaign",
	})

	handler := viz.HandleGenerateRelationshipGraph()
	result, err := handler(context.Background(), newToolRequest("visualize_relationship_graph", map[string]any{
		"campaign_id": "empty-campaign",
	}))
	if err != nil {
		t.Fatalf("HandleGenerateRelationshipGraph() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("HandleGenerateRelationshipGraph() returned error: %v", result.Content)
	}

	text := extractText(result)
	if !strings.Contains(text, "No entities found") {
		t.Errorf("Expected 'No entities found', got: %s", text)
	}
}

func TestHandleRelationshipGraph_MissingCampaignID(t *testing.T) {
	viz, _ := setupVizTest()

	handler := viz.HandleGenerateRelationshipGraph()
	result, err := handler(context.Background(), newToolRequest("visualize_relationship_graph", map[string]any{}))
	if err != nil {
		t.Fatalf("HandleGenerateRelationshipGraph() error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error for missing campaign_id")
	}
}

func TestRelationshipGraph_50PlusNodesFallback(t *testing.T) {
	viz, _ := setupVizTest()
	_ = viz // We test renderRelationshipGraph directly

	// Create 55 nodes
	var nodes []domain.CanonEntity
	for i := 0; i < 55; i++ {
		nodes = append(nodes, domain.CanonEntity{
			ID:   string(rune('A' + i%26)),
			Name: "Entity",
			Type: domain.EntityTypeNPC,
		})
	}

	graph := &domain.RelationshipGraph{
		CampaignID: "big-campaign",
		Nodes:      nodes,
		Edges:      nil,
	}

	html := renderRelationshipGraph(graph)
	if !strings.Contains(html, "static") && !strings.Contains(html, "too many nodes") {
		t.Error("50+ node graph should use static fallback rendering")
	}
}

// extractText pulls the Text content from a CallToolResult.
func extractText(result *mcp.CallToolResult) string {
	if result == nil {
		return ""
	}
	for _, c := range result.Content {
		if textContent, ok := c.(mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}
