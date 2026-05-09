package domain

import (
	"errors"
	"fmt"
)

// IntelligenceTier represents monster tactical sophistication based on INT score.
type IntelligenceTier string

const (
	TierInstinctive IntelligenceTier = "instinctive" // INT 1-4
	TierSimple      IntelligenceTier = "simple"      // INT 5-9
	TierTactical    IntelligenceTier = "tactical"    // INT 10-14
	TierStrategic   IntelligenceTier = "strategic"   // INT 15+
)

// Tactics represents tactical behavior for a monster in combat.
type Tactics struct {
	MonsterID          string              `json:"monster_id"`
	EncounterID        string              `json:"encounter_id"`
	IntelligenceTier   IntelligenceTier    `json:"intelligence_tier"`
	OpeningMove        string              `json:"opening_move"`
	TargetPriority     []TargetPriority    `json:"target_priority"`
	AbilityUsage       []AbilityTactic     `json:"ability_usage"`
	RetreatConditions  []RetreatCondition  `json:"retreat_conditions"`
	EnvironmentalTactics []EnvironmentalTactic `json:"environmental_tactics"`
	PackBehavior       *PackTactic         `json:"pack_behavior,omitempty"`
}

// TargetPriority defines targeting preference for monsters.
type TargetPriority struct {
	Priority   int    `json:"priority"` // 1 = highest
	TargetType string `json:"target_type"` // squishy, threat, healer, tank, nearest
	Reasoning  string `json:"reasoning"`
}

// AbilityTactic defines when/how to use a monster ability.
type AbilityTactic struct {
	AbilityName    string `json:"ability_name"`
	UsageCondition string `json:"usage_condition"` // "when 3+ enemies clustered"
	CooldownTurns  int    `json:"cooldown_turns,omitempty"`
	Priority       string `json:"priority"` // always, situational, last_resort
}

// RetreatCondition defines when a monster flees combat.
type RetreatCondition struct {
	Trigger string `json:"trigger"` // "HP < 25%", "allies defeated"
	Method  string `json:"method"`  // "disengage and flee", "fight to death"
}

// EnvironmentalTactic defines use of terrain features.
type EnvironmentalTactic struct {
	Feature string `json:"feature"` // "high ground", "chokepoint"
	Tactic  string `json:"tactic"`
	Bonus   string `json:"bonus,omitempty"` // "+2 AC", "advantage on ranged"
}

// PackTactic defines pack behavior for social monsters.
type PackTactic struct {
	Type             string `json:"type"` // pack_tactics, swarm, coordinated
	Description      string `json:"description"`
	BonusWhenAdjacent string `json:"bonus_when_adjacent,omitempty"`
}

// Validate checks tactics completeness according to WotC standards.
func (t *Tactics) Validate() error {
	if t.MonsterID == "" {
		return errors.New("monster_id is required")
	}
	if t.EncounterID == "" {
		return errors.New("encounter_id is required")
	}
	if t.IntelligenceTier == "" {
		return errors.New("intelligence_tier is required")
	}
	if !isValidIntelligenceTier(t.IntelligenceTier) {
		return fmt.Errorf("invalid intelligence tier: %s", t.IntelligenceTier)
	}
	if len(t.TargetPriority) < 2 {
		return errors.New("must have at least 2 target priorities")
	}
	if len(t.RetreatConditions) == 0 {
		return errors.New("must have at least 1 retreat condition")
	}
	if t.OpeningMove == "" {
		return errors.New("opening_move is required")
	}
	// Validate target priorities are sequential
	for i, tp := range t.TargetPriority {
		if tp.Priority != i+1 {
			return fmt.Errorf("target priority %d: expected priority %d, got %d", i+1, i+1, tp.Priority)
		}
		if tp.TargetType == "" {
			return fmt.Errorf("target priority %d: target_type is required", i+1)
		}
	}
	return nil
}

// isValidIntelligenceTier checks if a tier is valid.
func isValidIntelligenceTier(tier IntelligenceTier) bool {
	switch tier {
	case TierInstinctive, TierSimple, TierTactical, TierStrategic:
		return true
	default:
		return false
	}
}

// GetIntelligenceTierFromScore returns tier from INT score per MM guidelines.
func GetIntelligenceTierFromScore(intScore int) IntelligenceTier {
	switch {
	case intScore <= 4:
		return TierInstinctive
	case intScore <= 9:
		return TierSimple
	case intScore <= 14:
		return TierTactical
	default:
		return TierStrategic
	}
}

// GetTacticalComplexity returns a description of tactical complexity for a tier.
func GetTacticalComplexity(tier IntelligenceTier) string {
	switch tier {
	case TierInstinctive:
		return "Acts on instinct; no tactics beyond basic aggression"
	case TierSimple:
		return "Simple tactics; focuses on nearest threat"
	case TierTactical:
		return "Coordinated tactics; uses environment and abilities strategically"
	case TierStrategic:
		return "Advanced strategy; anticipates party moves and adapts mid-combat"
	default:
		return "Unknown"
	}
}

// HasPackTactics checks if tactics include pack behavior.
func (t *Tactics) HasPackTactics() bool {
	return t.PackBehavior != nil
}

// GetPrimaryTarget returns the highest priority target type.
func (t *Tactics) GetPrimaryTarget() string {
	if len(t.TargetPriority) == 0 {
		return ""
	}
	return t.TargetPriority[0].TargetType
}

// ShouldRetreat checks if retreat conditions are met based on HP percentage.
func (t *Tactics) ShouldRetreat(hpPercent int) (bool, string) {
	for _, rc := range t.RetreatConditions {
		if hpPercent <= 25 && rc.Trigger == "HP < 25%" {
			return true, rc.Method
		}
		if hpPercent <= 50 && rc.Trigger == "HP < 50%" {
			return true, rc.Method
		}
	}
	return false, ""
}
