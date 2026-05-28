package generators

import (
	"context"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository/mocks"
)

func TestNPCGenerator_GenerateFromCanon(t *testing.T) {
	t.Run("generates NPCs from canon entities", func(t *testing.T) {
		// Arrange
		canonRepo := &mocks.CanonRepositoryMock{
			LoadFunc: func(campaignID string) (*domain.CanonDocument, error) {
				return &domain.CanonDocument{
					SchemaVersion: domain.SchemaVersionV2,
					CampaignID:    "test-campaign",
					Entities: []domain.CanonEntity{
						{
							ID:          "npc-001",
							Name:        "Elara Shadowmere",
							Type:        domain.EntityTypeNPC,
							Role:        "merchant",
							Motivation:  "Vende información secreta a los aventureros",
							CanonState:  domain.EntityStateAlive,
							Properties: map[string]any{
								"faction": "Shadow Guild",
								"hp":      25,
								"ac":      14,
							},
							Connections: []string{"Shadow Guild", "City Guard"},
							Secret:      "Es espía doble para la guardia de la ciudad",
						},
						{
							ID:          "npc-002",
							Name:        "Theron Brightblade",
							Type:        domain.EntityTypeNPC,
							Role:        "ally",
							Motivation:  "Proteger a los débiles y luchar contra la tiranía",
							CanonState:  domain.EntityStateAlive,
							Properties: map[string]any{
								"hp": 45,
								"ac": 18,
							},
						},
					},
				}, nil
			},
		}

		generator := NewNPCGenerator(canonRepo)
		ctx := context.Background()

		// Act
		npcs, warnings, err := generator.GenerateFromCanon(ctx, "test-campaign")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}

		if len(npcs) != 2 {
			t.Fatalf("expected 2 NPCs, got %d", len(npcs))
		}

		// Check first NPC
		if npcs[0].ID != "npc-001" {
			t.Errorf("expected ID npc-001, got %s", npcs[0].ID)
		}
		if npcs[0].Name != "Elara Shadowmere" {
			t.Errorf("expected name Elara Shadowmere, got %s", npcs[0].Name)
		}
		if npcs[0].Role != "merchant" {
			t.Errorf("expected role merchant, got %s", npcs[0].Role)
		}
		if npcs[0].Faction != "Shadow Guild" {
			t.Errorf("expected faction Shadow Guild, got %s", npcs[0].Faction)
		}
		if npcs[0].Stats.HP != 25 {
			t.Errorf("expected HP 25, got %d", npcs[0].Stats.HP)
		}
		if npcs[0].Stats.AC != 14 {
			t.Errorf("expected AC 14, got %d", npcs[0].Stats.AC)
		}

		// Check second NPC
		if npcs[1].ID != "npc-002" {
			t.Errorf("expected ID npc-002, got %s", npcs[1].ID)
		}
		if npcs[1].Stats.HP != 45 {
			t.Errorf("expected HP 45, got %d", npcs[1].Stats.HP)
		}
	})

	t.Run("returns warning when no NPCs found", func(t *testing.T) {
		// Arrange
		canonRepo := &mocks.CanonRepositoryMock{
			LoadFunc: func(campaignID string) (*domain.CanonDocument, error) {
				return &domain.CanonDocument{
					SchemaVersion: domain.SchemaVersionV2,
					CampaignID:    "test-campaign",
					Entities: []domain.CanonEntity{
						{
							ID:         "item-001",
							Name:       "Magic Sword",
							Type:       domain.EntityTypeItem,
							CanonState: domain.EntityStateAlive,
						},
					},
				}, nil
			},
		}

		generator := NewNPCGenerator(canonRepo)
		ctx := context.Background()

		// Act
		npcs, warnings, err := generator.GenerateFromCanon(ctx, "test-campaign")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(warnings))
		}

		if len(npcs) != 0 {
			t.Fatalf("expected 0 NPCs, got %d", len(npcs))
		}
	})

	t.Run("generates description from entity properties", func(t *testing.T) {
		// Arrange
		canonRepo := &mocks.CanonRepositoryMock{
			LoadFunc: func(campaignID string) (*domain.CanonDocument, error) {
				return &domain.CanonDocument{
					SchemaVersion: domain.SchemaVersionV2,
					CampaignID:    "test-campaign",
					Entities: []domain.CanonEntity{
						{
							ID:          "npc-001",
							Name:        "Test NPC",
							Type:        domain.EntityTypeNPC,
							Role:        "guard",
							CanonState:  domain.EntityStateAlive,
							Connections: []string{"King", "Queen"},
							Secret:      "Knows about the assassination plot",
						},
					},
				}, nil
			},
		}

		generator := NewNPCGenerator(canonRepo)
		ctx := context.Background()

		// Act
		npcs, _, err := generator.GenerateFromCanon(ctx, "test-campaign")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(npcs) != 1 {
			t.Fatalf("expected 1 NPC, got %d", len(npcs))
		}

		desc := npcs[0].Description
		if !contains(desc, "guard") {
			t.Errorf("expected description to mention role, got: %s", desc)
		}
		if !contains(desc, "King") {
			t.Errorf("expected description to mention connections, got: %s", desc)
		}
		if !contains(desc, "assassination") {
			t.Errorf("expected description to mention secret, got: %s", desc)
		}
	})
}

func TestNPCGenerator_GenerateMarkdown(t *testing.T) {
	t.Run("generates markdown for NPCs and factions", func(t *testing.T) {
		// Arrange
		npcs := []domain.NPC{
			{
				ID:          "npc-001",
				Name:        "Elara Shadowmere",
				Role:        "merchant",
				Faction:     "Shadow Guild",
				Description: "Vende información secreta",
				Stats:       &domain.StatBlock{HP: 25, AC: 14},
			},
		}

		factions := []domain.Faction{
			{
				ID:          "faction-001",
				Name:        "Shadow Guild",
				Description: "Gremio de ladrones",
				Agenda:      "Controlar el mercado negro",
				Tier:        3,
				Allies:      []string{"Thieves Guild"},
				Enemies:     []string{"City Guard"},
			},
		}

		generator := &NPCGenerator{}

		// Act
		markdown := generator.GenerateMarkdown(npcs, factions)

		// Assert
		if !contains(markdown, "# NPCs y Facciones") {
			t.Error("expected markdown to include header")
		}
		if !contains(markdown, "Elara Shadowmere") {
			t.Error("expected markdown to include NPC name")
		}
		if !contains(markdown, "Shadow Guild") {
			t.Error("expected markdown to include faction name")
		}
		if !contains(markdown, "PG:") {
			t.Error("expected markdown to include HP")
		}
	})
}

func TestMonsterGenerator_GenerateFromCanon(t *testing.T) {
	t.Run("generates monsters from canon entities", func(t *testing.T) {
		// Arrange
		canonRepo := &mocks.CanonRepositoryMock{
			LoadFunc: func(campaignID string) (*domain.CanonDocument, error) {
				return &domain.CanonDocument{
					SchemaVersion: domain.SchemaVersionV2,
					CampaignID:    "test-campaign",
					Entities: []domain.CanonEntity{
						{
							ID:          "monster-001",
							Name:        "Shadow Drake",
							Type:        domain.EntityTypeMonster,
							Role:        "monster",
							Motivation:  "Guarda el tesoro del dragón anciano",
							CanonState:  domain.EntityStateAlive,
							Properties: map[string]any{
								"cr":   "5",
								"type": "dragon",
								"size": "Large",
								"hp":   120,
								"ac":   17,
							},
						},
					},
				}, nil
			},
		}

		generator := NewMonsterGenerator(canonRepo)
		ctx := context.Background()

		// Act
		monsters, warnings, err := generator.GenerateFromCanon(ctx, "test-campaign")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}

		if len(monsters) != 1 {
			t.Fatalf("expected 1 monster, got %d", len(monsters))
		}

		monster := monsters[0]
		if monster.ID != "monster-001" {
			t.Errorf("expected ID monster-001, got %s", monster.ID)
		}
		if monster.Name != "Shadow Drake" {
			t.Errorf("expected name Shadow Drake, got %s", monster.Name)
		}
		if monster.CR != "5" {
			t.Errorf("expected CR 5, got %s", monster.CR)
		}
		if monster.Type != "dragon" {
			t.Errorf("expected type dragon, got %s", monster.Type)
		}
		if monster.Size != "Large" {
			t.Errorf("expected size Large, got %s", monster.Size)
		}
		if monster.Stats.HP != 120 {
			t.Errorf("expected HP 120, got %d", monster.Stats.HP)
		}
		if monster.Stats.AC != 17 {
			t.Errorf("expected AC 17, got %d", monster.Stats.AC)
		}
	})

	t.Run("uses defaults when properties missing", func(t *testing.T) {
		// Arrange
		canonRepo := &mocks.CanonRepositoryMock{
			LoadFunc: func(campaignID string) (*domain.CanonDocument, error) {
				return &domain.CanonDocument{
					SchemaVersion: domain.SchemaVersionV2,
					CampaignID:    "test-campaign",
					Entities: []domain.CanonEntity{
						{
							ID:         "monster-001",
							Name:       "Generic Monster",
							Type:       domain.EntityTypeMonster,
							CanonState: domain.EntityStateAlive,
						},
					},
				}, nil
			},
		}

		generator := NewMonsterGenerator(canonRepo)
		ctx := context.Background()

		// Act
		monsters, _, err := generator.GenerateFromCanon(ctx, "test-campaign")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(monsters) != 1 {
			t.Fatalf("expected 1 monster, got %d", len(monsters))
		}

		monster := monsters[0]
		if monster.CR != "1" {
			t.Errorf("expected default CR 1, got %s", monster.CR)
		}
		if monster.Type != "humanoid" {
			t.Errorf("expected default type humanoid, got %s", monster.Type)
		}
		if monster.Size != "Medium" {
			t.Errorf("expected default size Medium, got %s", monster.Size)
		}
	})

	t.Run("returns warning when no monsters found", func(t *testing.T) {
		// Arrange
		canonRepo := &mocks.CanonRepositoryMock{
			LoadFunc: func(campaignID string) (*domain.CanonDocument, error) {
				return &domain.CanonDocument{
					SchemaVersion: domain.SchemaVersionV2,
					CampaignID:    "test-campaign",
					Entities:      []domain.CanonEntity{},
				}, nil
			},
		}

		generator := NewMonsterGenerator(canonRepo)
		ctx := context.Background()

		// Act
		monsters, warnings, err := generator.GenerateFromCanon(ctx, "test-campaign")

		// Assert
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(warnings))
		}

		if len(monsters) != 0 {
			t.Fatalf("expected 0 monsters, got %d", len(monsters))
		}
	})
}

func TestMonsterGenerator_GenerateMarkdown(t *testing.T) {
	t.Run("generates markdown for monsters", func(t *testing.T) {
		// Arrange
		monsters := []domain.Monster{
			{
				ID:          "monster-001",
				Name:        "Shadow Drake",
				CR:          "5",
				Type:        "dragon",
				Size:        "Large",
				Description: "Un dragón de las sombras",
				Stats:       domain.StatBlock{HP: 120, AC: 17},
				Abilities:   []string{"Shadow Breath", "Darkvision"},
			},
		}

		generator := &MonsterGenerator{}

		// Act
		markdown := generator.GenerateMarkdown(monsters)

		// Assert
		if !contains(markdown, "# Bestiario") {
			t.Error("expected markdown to include header")
		}
		if !contains(markdown, "Shadow Drake") {
			t.Error("expected markdown to include monster name")
		}
		if !contains(markdown, "**CR:** 5") {
			t.Error("expected markdown to include CR")
		}
		if !contains(markdown, "**Clase de Armadura:** 17") {
			t.Error("expected markdown to include AC")
		}
		if !contains(markdown, "Shadow Breath") {
			t.Error("expected markdown to include abilities")
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure time package is used
var _ = time.Now
