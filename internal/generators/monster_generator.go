package generators

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// MonsterGenerator generates monsters from canon data
type MonsterGenerator struct {
	canonRepo repository.CanonRepository
}

// NewMonsterGenerator creates a new monster generator
func NewMonsterGenerator(canonRepo repository.CanonRepository) *MonsterGenerator {
	return &MonsterGenerator{
		canonRepo: canonRepo,
	}
}

// GenerateFromCanon generates monsters based on canon entities
func (g *MonsterGenerator) GenerateFromCanon(ctx context.Context, campaignID string) ([]domain.Monster, []string, error) {
	doc, err := g.canonRepo.Load(campaignID)
	if err != nil {
		return nil, []string{fmt.Sprintf("failed to load canon: %v", err)}, nil
	}

	var monsters []domain.Monster
	var warnings []string
	now := time.Now()

	for _, entity := range doc.Entities {
		if entity.Type != domain.EntityTypeMonster {
			continue
		}

		monster := domain.Monster{
			ID:          entity.ID,
			CampaignID:  campaignID,
			Name:        entity.Name,
			Description: entity.Motivation,
			CreatedAt:   now,
		}

		// Extract CR from properties if available
		if cr, ok := entity.Properties["cr"].(string); ok && cr != "" {
			monster.CR = cr
		} else {
			monster.CR = "1" // Default CR
		}

		// Extract type from properties if available
		if typ, ok := entity.Properties["type"].(string); ok && typ != "" {
			monster.Type = typ
		} else {
			monster.Type = "humanoid" // Default type
		}

		// Extract size from properties if available
		if size, ok := entity.Properties["size"].(string); ok && size != "" {
			monster.Size = size
		} else {
			monster.Size = "Medium" // Default size
		}

		// Extract stats from properties if available
		if hp, ok := entity.Properties["hp"].(int); ok {
			monster.Stats.HP = hp
		}
		if ac, ok := entity.Properties["ac"].(int); ok {
			monster.Stats.AC = ac
		}

		// Generate description from properties if not set
		if monster.Description == "" {
			monster.Description = g.generateDescription(entity)
		}

		monsters = append(monsters, monster)
	}

	if len(monsters) == 0 {
		warnings = append(warnings, "no monster entities found in canon document")
	}

	return monsters, warnings, nil
}

// generateDescription creates a description from entity properties
func (g *MonsterGenerator) generateDescription(entity domain.CanonEntity) string {
	var parts []string

	// Extract type and size for description
	typ := "criatura"
	if t, ok := entity.Properties["type"].(string); ok && t != "" {
		typ = t
	}

	size := "mediana"
	if s, ok := entity.Properties["size"].(string); ok && s != "" {
		size = s
	}

	parts = append(parts, fmt.Sprintf("Una %s %s que habita en este mundo.", size, typ))

	if cr, ok := entity.Properties["cr"].(string); ok && cr != "" {
		parts = append(parts, fmt.Sprintf("Nivel de desafío: %s.", cr))
	}

	if connections := entity.Connections; len(connections) > 0 {
		parts = append(parts, fmt.Sprintf("Relacionado con: %s.", strings.Join(connections, ", ")))
	}

	if len(parts) == 0 {
		return "Una criatura misteriosa con habilidades y comportamientos únicos."
	}

	return strings.Join(parts, " ")
}

// GenerateMarkdown generates markdown content for monsters
func (g *MonsterGenerator) GenerateMarkdown(monsters []domain.Monster) string {
	var sb strings.Builder

	sb.WriteString("# Bestiario\n\n")

	for _, monster := range monsters {
		sb.WriteString(fmt.Sprintf("## %s\n\n", monster.Name))
		sb.WriteString(fmt.Sprintf("*%s %s*\n\n", monster.Size, monster.Type))
		sb.WriteString(fmt.Sprintf("- **ID:** %s\n", monster.ID))
		sb.WriteString(fmt.Sprintf("- **CR:** %s\n", monster.CR))
		sb.WriteString(fmt.Sprintf("- **Descripción:** %s\n\n", monster.Description))

		if monster.Stats.HP > 0 || monster.Stats.AC > 0 {
			sb.WriteString("### Estadísticas Base\n\n")
			if monster.Stats.AC > 0 {
				sb.WriteString(fmt.Sprintf("**Clase de Armadura:** %d\n", monster.Stats.AC))
			}
			if monster.Stats.HP > 0 {
				sb.WriteString(fmt.Sprintf("**Puntos de Golpe:** %d\n", monster.Stats.HP))
			}
			sb.WriteString("\n")
		}

		if len(monster.Abilities) > 0 {
			sb.WriteString("### Habilidades Especiales\n\n")
			for _, ability := range monster.Abilities {
				sb.WriteString(fmt.Sprintf("- %s\n", ability))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
