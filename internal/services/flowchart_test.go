package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestFlowchartService_GenerateMermaid(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	svc := NewFlowchartService(canonRepo, nil)

	seedCanon := func() {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "test-campaign",
			Entities: []domain.CanonEntity{
				{ID: "act-1", Name: "The Beginning", Type: domain.EntityTypeLocation, Role: "act"},
				{ID: "act-2", Name: "The Twist", Type: domain.EntityTypeLocation, Role: "act"},
				{ID: "act-3", Name: "The Climax", Type: domain.EntityTypeLocation, Role: "act"},
				{ID: "npc-giver", Name: "Eldrin", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
				{ID: "npc-dead", Name: "Villain", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateDead},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "act-1", To: "act-2", Type: domain.RelationshipTypeAlly, Notes: "leads to"},
				{ID: "rel-2", From: "act-2", To: "act-3", Type: domain.RelationshipTypeAlly, Notes: "leads to"},
				{ID: "rel-3", From: "npc-giver", To: "act-1", Type: domain.RelationshipTypeAlly, Notes: "introduces"},
			},
		}
		_ = canonRepo.Save("test-campaign", doc)
	}

	t.Run("overview level generates acts only", func(t *testing.T) {
		seedCanon()

		mermaid, err := svc.GenerateMermaid(ctx, "test-campaign", "overview")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(mermaid, "flowchart TD") {
			t.Fatalf("expected mermaid to start with 'flowchart TD', got: %s", mermaid)
		}
		if !strings.Contains(mermaid, "act-1") {
			t.Fatalf("expected act-1 in mermaid")
		}
		if !strings.Contains(mermaid, "act-2") {
			t.Fatalf("expected act-2 in mermaid")
		}
		// Overview should not have NPC nodes
		if strings.Contains(mermaid, "npc-giver") {
			t.Fatalf("overview should not contain NPC nodes")
		}
	})

	t.Run("act level includes quests", func(t *testing.T) {
		seedCanon()

		mermaid, err := svc.GenerateMermaid(ctx, "test-campaign", "act")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(mermaid, "act-1") {
			t.Fatalf("expected act-1 in mermaid")
		}
		// Act level may include related entities
		if !strings.Contains(mermaid, "-->") {
			t.Fatalf("expected arrows in flowchart")
		}
	})

	t.Run("decision level includes full detail", func(t *testing.T) {
		seedCanon()

		mermaid, err := svc.GenerateMermaid(ctx, "test-campaign", "decision")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(mermaid, "act-1") {
			t.Fatalf("expected act-1 in mermaid")
		}
	})

	t.Run("dead npcs excluded", func(t *testing.T) {
		seedCanon()

		mermaid, err := svc.GenerateMermaid(ctx, "test-campaign", "decision")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if strings.Contains(mermaid, "Villain") {
			t.Fatalf("dead NPC should not appear in flowchart")
		}
	})

	t.Run("invalid detail level returns error", func(t *testing.T) {
		seedCanon()

		_, err := svc.GenerateMermaid(ctx, "test-campaign", "invalid")
		if err == nil {
			t.Fatalf("expected error for invalid detail level")
		}
	})

	t.Run("missing campaign returns error", func(t *testing.T) {
		_, err := svc.GenerateMermaid(ctx, "nonexistent", "overview")
		if err == nil {
			t.Fatalf("expected error for missing campaign")
		}
	})

	t.Run("no acts returns error", func(t *testing.T) {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "no-acts",
			Entities: []domain.CanonEntity{
				{ID: "npc-1", Name: "Random NPC", Type: domain.EntityTypeNPC},
			},
		}
		_ = canonRepo.Save("no-acts", doc)

		_, err := svc.GenerateMermaid(ctx, "no-acts", "overview")
		if err == nil {
			t.Fatalf("expected error for campaign with no acts")
		}
	})

	t.Run("decision level includes timeline events", func(t *testing.T) {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "timeline-campaign",
			Entities: []domain.CanonEntity{
				{ID: "act-1", Name: "Act 1", Type: domain.EntityTypeLocation, Role: "act"},
				{ID: "npc-hero", Name: "Hero", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive},
			},
			Relationships: []domain.CanonRelationship{
				{ID: "rel-1", From: "npc-hero", To: "act-1", Type: domain.RelationshipTypeAlly},
			},
			Timeline: []domain.CanonTimelineEvent{
				{ID: "ev-1", Description: "Hero discovers truth", IsRevealed: true, Involved: []string{"npc-hero", "act-1"}},
				{ID: "ev-2", Description: "Secret not revealed", IsRevealed: false, Involved: []string{"npc-hero"}},
			},
		}
		_ = canonRepo.Save("timeline-campaign", doc)

		mermaid, err := svc.GenerateMermaid(ctx, "timeline-campaign", "decision")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(mermaid, "Hero discovers truth") {
			t.Fatalf("expected revealed timeline event in mermaid, got: %s", mermaid)
		}
		if strings.Contains(mermaid, "Secret not revealed") {
			t.Fatalf("unrevealed timeline event should not appear")
		}
	})
}

// mockActRepo is a simple mock for ActRepository
type mockActRepo struct {
	acts []domain.Act
}

func (m *mockActRepo) Save(act *domain.Act) error           { return nil }
func (m *mockActRepo) Read(campaignID string, number int) (*domain.Act, error) {
	return nil, fmt.Errorf("not found")
}
func (m *mockActRepo) List(campaignID string) ([]domain.Act, error) { return m.acts, nil }
func (m *mockActRepo) Delete(campaignID string, number int) error   { return nil }

func TestFlowchartService_GenerateMermaid_FromActRepo(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()
	actRepo := &mockActRepo{
		acts: []domain.Act{
			{
				Number:            1,
				Title:             "The Beginning",
				Content:           "# The Beginning\nFirst act content.",
				GameMode:          "investigacion",
				ChapterObjectives: []string{"Descubrir el misterio inicial"},
				EstimatedDuration: "2 sesiones",
				Tone:              "mystery",
				RunningGuidance:   "Este capítulo introduce el misterio. Los PJs deben encontrar tres pistas clave que los lleven al siguiente capítulo. El ritmo debe ser moderado, permitiendo exploración y roleo. Si los PJs se estancan, un NPC proporciona información crucial. Asegúrate de que cada sesión termine con una revelación que motive a continuar.",
				AssetHandoff:      "La pista final revela la ubicación del Acto 2",
			},
			{
				Number:            2,
				Title:             "The Twist",
				Content:           "# The Twist\nSecond act content.",
				GameMode:          "intriga",
				ChapterObjectives: []string{"Confrontar la verdad", "Decidir el próximo paso"},
				EstimatedDuration: "2-3 sesiones",
				Tone:              "political",
				RunningGuidance:   "Este capítulo presenta un giro inesperado. Los PJs deben procesar la revelación y decidir cómo actuar. Hay múltiples caminos posibles y cada decisión tiene consecuencias. El tono es de tensión política. Asegúrate de que los PJs sientan el peso de su elección.",
				AssetHandoff:      "La decisión determina qué facciones ayudan en el Acto 3",
			},
		},
	}

	svc := NewFlowchartService(canonRepo, actRepo)

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "act-repo-campaign",
		Entities:      []domain.CanonEntity{},
	}
	_ = canonRepo.Save("act-repo-campaign", doc)

	mermaid, err := svc.GenerateMermaid(ctx, "act-repo-campaign", "overview")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(mermaid, "The Beginning") {
		t.Fatalf("expected act title from act repo, got: %s", mermaid)
	}
	if !strings.Contains(mermaid, "The Twist") {
		t.Fatalf("expected second act title from act repo, got: %s", mermaid)
	}
}

func TestFlowchartService_GenerateSVG(t *testing.T) {
	ctx := context.Background()
	canonRepo := repository.NewMemoryCanonRepository()

	svc := NewFlowchartService(canonRepo, nil)

	doc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "svg-campaign",
		Entities: []domain.CanonEntity{
			{ID: "act-1", Name: "Act 1", Type: domain.EntityTypeLocation, Role: "act"},
			{ID: "act-2", Name: "Act 2", Type: domain.EntityTypeLocation, Role: "act"},
		},
		Relationships: []domain.CanonRelationship{
			{ID: "rel-1", From: "act-1", To: "act-2", Type: domain.RelationshipTypeAlly},
		},
	}
	_ = canonRepo.Save("svg-campaign", doc)

	t.Run("generates valid SVG XML", func(t *testing.T) {
		svg, err := svc.GenerateSVG(ctx, "svg-campaign", "overview")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(svg, "<svg") {
			t.Fatalf("expected SVG to start with '<svg', got: %s", svg)
		}
		if !strings.Contains(svg, "</svg>") {
			t.Fatalf("expected SVG to contain closing tag")
		}
	})

	t.Run("no acts returns error in SVG", func(t *testing.T) {
		doc := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "svg-no-acts",
			Entities: []domain.CanonEntity{
				{ID: "npc-1", Name: "Random NPC", Type: domain.EntityTypeNPC},
			},
		}
		_ = canonRepo.Save("svg-no-acts", doc)

		_, err := svc.GenerateSVG(ctx, "svg-no-acts", "overview")
		if err == nil {
			t.Fatalf("expected error for campaign with no acts in SVG")
		}
	})
}
