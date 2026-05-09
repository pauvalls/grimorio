package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// TacticsService generates enemy tactics for encounters.
type TacticsService struct {
	monsterRepo   MonsterRepository
	encounterRepo EncounterRepository
	areaRepo      AreaRepository
}

// MonsterRepository defines repository interface for monsters.
type MonsterRepository interface {
	Read(ctx context.Context, campaignID string, monsterID string) (*domain.Monster, error)
}

// EncounterRepository defines repository interface for encounters.
type EncounterRepository interface {
	Read(ctx context.Context, campaignID string, encounterID string) (*domain.Encounter, error)
}

// AreaRepository defines repository interface for areas.
type AreaRepository interface {
	Read(ctx context.Context, campaignID string, areaID string) (*domain.Area, error)
}

// NewTacticsService creates a new TacticsService.
func NewTacticsService(
	monsterRepo MonsterRepository,
	encounterRepo EncounterRepository,
	areaRepo AreaRepository,
) *TacticsService {
	return &TacticsService{
		monsterRepo:   monsterRepo,
		encounterRepo: encounterRepo,
		areaRepo:      areaRepo,
	}
}

// GenerateMonsterTactics generates tactics for a single monster.
func (s *TacticsService) GenerateMonsterTactics(
	ctx context.Context,
	monster *domain.Monster,
	encounter *domain.Encounter,
	area *domain.Area,
) (*domain.Tactics, error) {
	if monster == nil {
		return nil, fmt.Errorf("monster cannot be nil")
	}

	tier := domain.GetIntelligenceTierFromScore(monster.Stats.INT)

	tactics := &domain.Tactics{
		MonsterID:        monster.ID,
		EncounterID:      encounter.ID,
		IntelligenceTier: tier,
		OpeningMove:      generateOpeningMove(monster, tier),
		TargetPriority:   generateTargetPriorities(tier),
		AbilityUsage:     generateAbilityUsage(monster, tier),
		RetreatConditions: generateRetreatConditions(monster, tier),
	}

	if area != nil {
		tactics.EnvironmentalTactics = generateEnvironmentalTactics(area, tier)
	}

	if hasPackTactics(monster) {
		tactics.PackBehavior = &domain.PackTactic{
			Type:        "pack_tactics",
			Description: "Gains advantage on attack rolls against enemies adjacent to allies",
		}
	}

	if err := tactics.Validate(); err != nil {
		return nil, fmt.Errorf("generated tactics validation failed: %w", err)
	}

	return tactics, nil
}

// GenerateEncounterTactics generates tactics for all monsters in an encounter.
func (s *TacticsService) GenerateEncounterTactics(
	ctx context.Context,
	encounter *domain.Encounter,
	area *domain.Area,
) ([]*domain.Tactics, error) {
	if encounter == nil {
		return nil, fmt.Errorf("encounter cannot be nil")
	}

	var allTactics []*domain.Tactics

	for _, monsterRef := range encounter.Monsters {
		monster, err := s.monsterRepo.Read(ctx, encounter.CampaignID, monsterRef.Name)
		if err != nil {
			continue // Skip unavailable monsters
		}

		tactics, err := s.GenerateMonsterTactics(ctx, monster, encounter, area)
		if err != nil {
			return nil, err
		}
		allTactics = append(allTactics, tactics)
	}

	return allTactics, nil
}

// GetTacticsByEncounter retrieves tactics for a specific encounter.
func (s *TacticsService) GetTacticsByEncounter(ctx context.Context, encounterID string) ([]*domain.Tactics, error) {
	// TODO: Implement repository method
	return nil, nil
}

// EvaluateTacticComplexity returns a complexity rating for tactics.
func (s *TacticsService) EvaluateTacticComplexity(ctx context.Context, monster *domain.Monster) (string, error) {
	if monster == nil {
		return "", fmt.Errorf("monster cannot be nil")
	}
	tier := domain.GetIntelligenceTierFromScore(monster.Stats.INT)
	return domain.GetTacticalComplexity(tier), nil
}

// Helper functions

func generateOpeningMove(monster *domain.Monster, tier domain.IntelligenceTier) string {
	switch tier {
	case domain.TierInstinctive:
		return "Charges at the nearest visible enemy"
	case domain.TierSimple:
		return "Moves to engage the closest threat"
	case domain.TierTactical:
		return "Assesses party formation and targets weak points"
	case domain.TierStrategic:
		return "Analyzes party composition and plans multi-round strategy"
	default:
		return "Attacks"
	}
}

func generateTargetPriorities(tier domain.IntelligenceTier) []domain.TargetPriority {
	switch tier {
	case domain.TierInstinctive:
		return []domain.TargetPriority{
			{Priority: 1, TargetType: "nearest", Reasoning: "Acts on instinct"},
			{Priority: 2, TargetType: "squishy", Reasoning: "Senses vulnerability"},
		}
	case domain.TierSimple:
		return []domain.TargetPriority{
			{Priority: 1, TargetType: "nearest", Reasoning: "Simple threat assessment"},
			{Priority: 2, TargetType: "tank", Reasoning: "Biggest threat"},
		}
	case domain.TierTactical:
		return []domain.TargetPriority{
			{Priority: 1, TargetType: "healer", Reasoning: "Prevent healing"},
			{Priority: 2, TargetType: "squishy", Reasoning: "Eliminate damage dealers"},
			{Priority: 3, TargetType: "tank", Reasoning: "Deal with last"},
		}
	case domain.TierStrategic:
		return []domain.TargetPriority{
			{Priority: 1, TargetType: "controller", Reasoning: "Disable battlefield control"},
			{Priority: 2, TargetType: "healer", Reasoning: "Prevent recovery"},
			{Priority: 3, TargetType: "damage", Reasoning: "Reduce incoming damage"},
		}
	default:
		return []domain.TargetPriority{{Priority: 1, TargetType: "nearest"}}
	}
}

func generateAbilityUsage(monster *domain.Monster, tier domain.IntelligenceTier) []domain.AbilityTactic {
	tactics := []domain.AbilityTactic{}

	// Add ability tactics based on tier
	if tier == domain.TierTactical || tier == domain.TierStrategic {
		tactics = append(tactics, domain.AbilityTactic{
			AbilityName:    "Special Ability",
			UsageCondition: "When 3+ enemies are clustered",
			Priority:       "situational",
		})
	}

	if tier == domain.TierStrategic {
		tactics = append(tactics, domain.AbilityTactic{
			AbilityName:    "Recharge Ability",
			UsageCondition: "Save for critical moment",
			CooldownTurns:  3,
			Priority:       "last_resort",
		})
	}

	return tactics
}

func generateRetreatConditions(monster *domain.Monster, tier domain.IntelligenceTier) []domain.RetreatCondition {
	switch tier {
	case domain.TierInstinctive:
		return []domain.RetreatCondition{
			{Trigger: "HP < 25%", Method: "Flee in random direction"},
		}
	case domain.TierSimple:
		return []domain.RetreatCondition{
			{Trigger: "HP < 25%", Method: "Disengage and flee"},
		}
	case domain.TierTactical:
		return []domain.RetreatCondition{
			{Trigger: "HP < 25%", Method: "Covering retreat with allies"},
			{Trigger: "Leader defeated", Method: "Surrender or flee"},
		}
	case domain.TierStrategic:
		return []domain.RetreatCondition{
			{Trigger: "Tactical disadvantage", Method: "Orderly withdrawal to regroup"},
			{Trigger: "Mission impossible", Method: "Escape to fight another day"},
		}
	default:
		return []domain.RetreatCondition{{Trigger: "HP < 10%", Method: "Flee"}}
	}
}

func generateEnvironmentalTactics(area *domain.Area, tier domain.IntelligenceTier) []domain.EnvironmentalTactic {
	tactics := []domain.EnvironmentalTactic{}

	if tier == domain.TierInstinctive {
		return tactics // No environmental awareness
	}

	// Generate tactics based on area features
	for _, feature := range area.Features {
		if feature.Type == "hazard" || feature.Type == "trap" {
			tactics = append(tactics, domain.EnvironmentalTactic{
				Feature: feature.Name,
				Tactic:  "Lure enemies toward hazard",
				Bonus:   "Environmental damage",
			})
		}
		if tier >= domain.TierTactical {
			tactics = append(tactics, domain.EnvironmentalTactic{
				Feature: "High ground",
				Tactic:  "Take elevated position",
				Bonus:   "+2 AC against ranged attacks",
			})
		}
	}

	return tactics
}

func hasPackTactics(monster *domain.Monster) bool {
	// Check if monster has pack tactics in abilities
	for _, ability := range monster.Abilities {
		if containsIgnoreCase(ability, "pack tactics") {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && findSubstringIgnoreCase(s, substr)
}

func findSubstringIgnoreCase(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
