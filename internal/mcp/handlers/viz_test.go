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
	canonService := services.NewCanonService(canonRepo, narrativeStateRepo, repository.NewMemoryCheckpointRepository())
	vizHandlers := NewVizHandlers(canonService)
	return vizHandlers, canonService
}

func TestHandleRelationshipGraph_HappyPath(t *testing.T) {
	viz, canonService := setupVizTest()

	// Seed canon with some entities and relationships
	_ = canonService.SaveCanon(context.Background(), &domain.CanonDocument{
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
	_ = canonService.SaveCanon(context.Background(), &domain.CanonDocument{
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

// TestRenderRelationshipGraph_LangIsSpanish is a regression test for Fase 4 i18n:
// visualization HTMLs must declare `lang="es"` for Spanish campaigns.
func TestRenderRelationshipGraph_LangIsSpanish(t *testing.T) {
	graph := &domain.RelationshipGraph{
		CampaignID: "es-test",
		Nodes: []domain.CanonEntity{
			{ID: "npc1", Name: "Gandalf", Type: domain.EntityTypeNPC},
		},
		Edges: nil,
	}

	d3 := renderRelationshipGraph(graph)
	if !strings.Contains(d3, `<html lang="es">`) {
		t.Errorf("D3 relationship graph must declare lang=\"es\"; got:\n%s", d3)
	}
	if strings.Contains(d3, `<html lang="en">`) {
		t.Errorf("D3 relationship graph must NOT declare lang=\"en\"; got:\n%s", d3)
	}

	// 50+ nodes triggers the static SVG path — also a visualization HTML.
	big := &domain.RelationshipGraph{CampaignID: "es-big"}
	for i := 0; i < 55; i++ {
		big.Nodes = append(big.Nodes, domain.CanonEntity{
			ID:   string(rune('A' + i%26)),
			Name: "Ent",
			Type: domain.EntityTypeNPC,
		})
	}
	static := renderRelationshipGraph(big)
	if !strings.Contains(static, `<html lang="es">`) {
		t.Errorf("static SVG relationship graph must declare lang=\"es\"; got:\n%s", static)
	}
	if strings.Contains(static, `<html lang="en">`) {
		t.Errorf("static SVG relationship graph must NOT declare lang=\"en\"; got:\n%s", static)
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
