package generators

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// NPCGenerator generates NPCs from canon data
type NPCGenerator struct {
	canonRepo repository.CanonRepository
}

// NewNPCGenerator creates a new NPC generator
func NewNPCGenerator(canonRepo repository.CanonRepository) *NPCGenerator {
	return &NPCGenerator{
		canonRepo: canonRepo,
	}
}

// GenerateFromCanon generates NPCs based on canon entities
func (g *NPCGenerator) GenerateFromCanon(ctx context.Context, campaignID string) ([]domain.NPC, []string, error) {
	doc, err := g.canonRepo.Load(campaignID)
	if err != nil {
		return nil, []string{fmt.Sprintf("failed to load canon: %v", err)}, nil
	}

	var npcs []domain.NPC
	var warnings []string
	now := time.Now()

	for _, entity := range doc.Entities {
		if entity.Type != domain.EntityTypeNPC {
			continue
		}

		npc := domain.NPC{
			ID:          entity.ID,
			CampaignID:  campaignID,
			Name:        entity.Name,
			Role:        entity.Role,
			Description: entity.Motivation,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Extract faction from properties if available
		if faction, ok := entity.Properties["faction"].(string); ok && faction != "" {
			npc.Faction = faction
		}

		// Extract stats from properties if available
		if hp, ok := entity.Properties["hp"].(int); ok {
			if npc.Stats == nil {
				npc.Stats = &domain.StatBlock{}
			}
			npc.Stats.HP = hp
		}
		if ac, ok := entity.Properties["ac"].(int); ok {
			if npc.Stats == nil {
				npc.Stats = &domain.StatBlock{}
			}
			npc.Stats.AC = ac
		}

		// Generate description from properties if not set
		if npc.Description == "" {
			npc.Description = g.generateDescription(entity)
		}

		npcs = append(npcs, npc)
	}

	if len(npcs) == 0 {
		warnings = append(warnings, "no NPC entities found in canon document")
	}

	return npcs, warnings, nil
}

// generateDescription creates a description from entity properties
func (g *NPCGenerator) generateDescription(entity domain.CanonEntity) string {
	var parts []string

	if entity.Role != "" {
		parts = append(parts, fmt.Sprintf("Un %s que juega un papel importante en la campaña.", entity.Role))
	}

	if connections := entity.Connections; len(connections) > 0 {
		parts = append(parts, fmt.Sprintf("Está conectado con: %s.", strings.Join(connections, ", ")))
	}

	if secret := entity.Secret; secret != "" {
		parts = append(parts, fmt.Sprintf("Secreto: %s", secret))
	}

	if len(parts) == 0 {
		return "Un personaje no jugador con motivaciones y objetivos propios."
	}

	return strings.Join(parts, " ")
}

// GenerateMarkdown generates markdown content for NPCs
func (g *NPCGenerator) GenerateMarkdown(npcs []domain.NPC, factions []domain.Faction) string {
	var sb strings.Builder

	sb.WriteString("# NPCs y Facciones\n\n")
	sb.WriteString("## NPCs Principales\n\n")

	for _, npc := range npcs {
		fmt.Fprintf(&sb, "### %s\n\n", npc.Name)
		fmt.Fprintf(&sb, "- **ID:** %s\n", npc.ID)
		fmt.Fprintf(&sb, "- **Rol:** %s\n", npc.Role)
		if npc.Faction != "" {
			fmt.Fprintf(&sb, "- **Facción:** %s\n", npc.Faction)
		}
		fmt.Fprintf(&sb, "- **Descripción:** %s\n\n", npc.Description)

		if npc.Stats != nil && (npc.Stats.HP > 0 || npc.Stats.AC > 0) {
			fmt.Fprintf(&sb, "#### Estadísticas de Combate\n\n")
			if npc.Stats.HP > 0 {
				fmt.Fprintf(&sb, "- **PG:** %d\n", npc.Stats.HP)
			}
			if npc.Stats.AC > 0 {
				fmt.Fprintf(&sb, "- **CA:** %d\n", npc.Stats.AC)
			}
			fmt.Fprintf(&sb, "\n")
		}

		fmt.Fprintf(&sb, "---\n\n")
	}

	if len(factions) > 0 {
		sb.WriteString("## Facciones\n\n")
		for _, faction := range factions {
			fmt.Fprintf(&sb, "### %s\n\n", faction.Name)
			fmt.Fprintf(&sb, "- **ID:** %s\n", faction.ID)
			fmt.Fprintf(&sb, "- **Descripción:** %s\n", faction.Description)
			fmt.Fprintf(&sb, "- **Objetivo:** %s\n", faction.Agenda)
			if faction.Tier > 0 {
				fmt.Fprintf(&sb, "- **Tier:** %d\n", faction.Tier)
			}
			if len(faction.Allies) > 0 {
				fmt.Fprintf(&sb, "- **Aliados:** %s\n", strings.Join(faction.Allies, ", "))
			}
			if len(faction.Enemies) > 0 {
				fmt.Fprintf(&sb, "- **Enemigos:** %s\n", strings.Join(faction.Enemies, ", "))
			}
			fmt.Fprintf(&sb, "\n---\n\n")
		}
	}

	return sb.String()
}
