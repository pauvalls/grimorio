package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// PlayerHookService generates personalized plot hooks for PCs
type PlayerHookService struct {
	charRepo  repository.CharacterRepository
	canonRepo repository.CanonRepository
}

// NewPlayerHookService creates a new player hook service
func NewPlayerHookService(charRepo repository.CharacterRepository, canonRepo repository.CanonRepository) *PlayerHookService {
	return &PlayerHookService{
		charRepo:  charRepo,
		canonRepo: canonRepo,
	}
}

// GenerateHooks creates personalized hooks for all characters in a campaign
func (s *PlayerHookService) GenerateHooks(ctx context.Context, campaignID string) ([]domain.CharacterHook, []string, error) {
	var warnings []string

	_, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, nil, fmt.Errorf("campaign not found: %w", err)
	}

	characters, err := s.charRepo.List(campaignID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list characters: %w", err)
	}

	if len(characters) == 0 {
		return []domain.CharacterHook{}, []string{"no characters in campaign"}, nil
	}

	doc, _ := s.canonRepo.Load(campaignID)

	var hooks []domain.CharacterHook
	for _, char := range characters {
		hook, warn := s.generateHookForCharacter(char, doc)
		if warn != "" {
			warnings = append(warnings, warn)
		}
		hooks = append(hooks, hook)
	}

	return hooks, warnings, nil
}

func (s *PlayerHookService) generateHookForCharacter(char domain.Character, doc *domain.CanonDocument) (domain.CharacterHook, string) {
	archetype := classArchetype(char.Class)
	bg := normalizeBackground(char.Background)

	if bg == "" {
		return s.genericHook(char, doc), fmt.Sprintf("character %s has no background, using generic hook", char.Name)
	}

	template := s.selectTemplate(archetype, bg)
	canonRef := s.pickCanonReference(doc)

	hookText := fmt.Sprintf(template, char.Name, canonRef)
	connection := fmt.Sprintf("%s's %s background ties them to %s through their past experiences.", char.Name, char.Background, canonRef)

	return domain.CharacterHook{
		CharacterID:      char.Name, // using name as ID since Character doesn't have ID field
		CharacterName:    char.Name,
		Background:       char.Background,
		Class:            char.Class,
		Hook:             hookText,
		ConnectionToPlot: connection,
	}, ""
}

func classArchetype(class string) string {
	class = strings.ToLower(class)
	switch class {
	case "fighter", "barbarian", "paladin", "monk", "ranger",
		"guerrero", "barbaro", "monje", "explorador":
		return "martial"
	case "wizard", "sorcerer", "warlock", "cleric", "druid", "bard",
		"mago", "hechicero", "brujo", "clerigo", "druida", "bardo":
		return "caster"
	case "rogue", "artificer",
		"picaro", "artifice":
		return "skill"
	default:
		return "martial"
	}
}

func normalizeBackground(bg string) string {
	bg = strings.ToLower(strings.TrimSpace(bg))
	switch bg {
	case "soldier", "knight", "mercenary", "soldado", "gladiador":
		return "soldier"
	case "sage", "scholar", "acolyte", "sabio", "acolito":
		return "sage"
	case "criminal", "spy", "urchin", "charlatan":
		return "criminal"
	case "folk hero", "outlander", "hermit", "heroe del pueblo", "ermitano":
		return "folk_hero"
	default:
		return bg
	}
}

func (s *PlayerHookService) selectTemplate(archetype, background string) string {
	// 12 templates: 4 backgrounds × 3 archetypes
	templates := map[string]map[string]string{
		"soldier": {
			"martial": "%s's military past has drawn the attention of %s. During a past campaign, they saw something they weren't meant to see — a symbol that now appears in connection with %s.",
			"caster":  "%s once served as a battle mage in a unit that was wiped out defending %s. The sole survivor besides %s speaks of a prophecy that ties their fate to %s.",
			"skill":   "%s's regiment once hired a mysterious informant who warned of %s. That informant has now vanished, leaving only a coded message addressed to %s.",
		},
		"sage": {
			"martial": "%s studied ancient warfare at the Academy where a forbidden text about %s was kept. When the text was stolen, %s was the last person to see it.",
			"caster":  "%s's research into arcane lore has revealed a connection between their magical lineage and %s. An old journal speaks of %s's ancestors and their pact with %s.",
			"skill":   "%s once catalogued artifacts for the Grand Library, including a strange relic tied to %s. The relic's description matches events now unfolding around %s.",
		},
		"criminal": {
			"martial": "%s's underground fighting ring was shut down by agents of %s. The leader left a message: 'When you're ready for real work, find %s.'",
			"caster":  "%s once smuggled magical components for a client obsessed with %s. That client has resurfaced, demanding %s complete one last job.",
			"skill":   "%s's old crew pulled a heist targeting a noble with ties to %s. They got more than gold — they got a map that %s now carries, marked with locations connected to %s.",
		},
		"folk_hero": {
			"martial": "%s once saved a village from bandits who worked for %s. The villagers whisper that %s is destined to stand against %s when darkness rises.",
			"caster":  "%s healed a stranger in the woods who spoke of %s in their fever dreams. That stranger left behind a talisman that pulses with energy whenever %s is near.",
			"skill":   "%s's heroic act of rescuing children from a burning temple revealed a hidden cellar with records about %s. The temple's abbot begged %s to keep the secret — until now.",
		},
	}

	bgTemplates, ok := templates[background]
	if !ok {
		// Fallback for unknown background
		return "%s's unique past has intertwined with the fate of %s. A recent dream or omen has shown %s visions of %s, drawing them into the unfolding story."
	}

	tmpl, ok := bgTemplates[archetype]
	if !ok {
		return bgTemplates["martial"]
	}

	return tmpl
}

func (s *PlayerHookService) pickCanonReference(doc *domain.CanonDocument) string {
	if doc == nil || len(doc.Entities) == 0 {
		return "the main plot"
	}

	// Prefer mcguffin, villain, or ally roles
	var candidates []string
	for _, e := range doc.Entities {
		if e.Role == "mcguffin" || e.Role == "villain" || e.Role == "ally" {
			candidates = append(candidates, e.Name)
		}
	}

	if len(candidates) == 0 {
		// Fall back to any named entity
		for _, e := range doc.Entities {
			if e.Name != "" {
				candidates = append(candidates, e.Name)
			}
		}
	}

	if len(candidates) == 0 {
		return "the main plot"
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	return candidates[rng.Intn(len(candidates))]
}

func (s *PlayerHookService) genericHook(char domain.Character, doc *domain.CanonDocument) domain.CharacterHook {
	canonRef := s.pickCanonReference(doc)
	hookText := fmt.Sprintf("A strange twist of fate has brought %s to the path of %s. Whether by chance or destiny, %s now holds a piece of the puzzle that others seek.", char.Name, canonRef, char.Name)
	connection := fmt.Sprintf("%s's journey as a %s connects them to %s through the unfolding events of the campaign.", char.Name, char.Class, canonRef)

	return domain.CharacterHook{
		CharacterID:      char.Name,
		CharacterName:    char.Name,
		Background:       char.Background,
		Class:            char.Class,
		Hook:             hookText,
		ConnectionToPlot: connection,
	}
}
