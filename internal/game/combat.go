package game

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/pauvalls/grimorio/internal/domain"
)

// CalculateModifier is a convenience alias for domain.CalculateModifier
func CalculateModifier(score int) int {
	return domain.CalculateModifier(score)
}

// CombatResolver handles D&D 5e combat resolution
type CombatResolver struct{}

// NewCombatResolver creates a new combat resolver
func NewCombatResolver() *CombatResolver {
	return &CombatResolver{}
}

// AttackResult represents the result of an attack resolution
type AttackResult struct {
	Hit          bool   `json:"hit"`
	CriticalHit  bool   `json:"critical_hit"`
	CriticalFail bool   `json:"critical_fail"`
	AttackRoll   int    `json:"attack_roll"`
	AC           int    `json:"ac"`
	Damage       int    `json:"damage,omitempty"`
	Description  string `json:"description"`
}

// SaveResult represents the result of a saving throw
type SaveResult struct {
	Success  bool   `json:"success"`
	Roll     int    `json:"roll"`
	DC       int    `json:"dc"`
	Total    int    `json:"total"`
	Ability  string `json:"ability,omitempty"`
}

// SkillCheckResult represents the result of a skill check
type SkillCheckResult struct {
	Success bool   `json:"success"`
	Roll    int    `json:"roll"`
	DC      int    `json:"dc"`
	Total   int    `json:"total"`
	Skill   string `json:"skill,omitempty"`
}

// ResolveAttack resolves an attack roll against a target
func (r *CombatResolver) ResolveAttack(attacker, target *domain.PlayerState, attackRoll, attackBonus int) *AttackResult {
	result := &AttackResult{
		AttackRoll: attackRoll,
		AC:         target.AC,
	}

	// Natural 20 = automatic critical hit
	if attackRoll == 20 {
		result.Hit = true
		result.CriticalHit = true
		result.Description = fmt.Sprintf("Critical hit! Natural 20 against AC %d", target.AC)
		return result
	}

	// Natural 1 = automatic miss
	if attackRoll == 1 {
		result.Hit = false
		result.CriticalFail = true
		result.Description = fmt.Sprintf("Critical miss! Natural 1 against AC %d", target.AC)
		return result
	}

	// Normal attack: roll + bonus >= AC
	total := attackRoll + attackBonus
	if total >= target.AC {
		result.Hit = true
		result.Description = fmt.Sprintf("Hit! Attack roll %d + %d = %d meets AC %d", 
			attackRoll, attackBonus, total, target.AC)
	} else {
		result.Hit = false
		result.Description = fmt.Sprintf("Miss! Attack roll %d + %d = %d vs AC %d", 
			attackRoll, attackBonus, total, target.AC)
	}

	return result
}

// CalculateDamage rolls damage dice and returns the total
func (r *CombatResolver) CalculateDamage(dice string, isCritical bool) (int, error) {
	spec, err := ParseDice(dice)
	if err != nil {
		return 0, fmt.Errorf("invalid damage dice: %w", err)
	}

	if isCritical {
		// Double the dice on critical hit
		spec.Count *= 2
	}

	result := Roll(spec)
	return result.Total, nil
}

// ApplyDamageToPlayer applies damage to a player character
func (r *CombatResolver) ApplyDamageToPlayer(player *domain.PlayerState, damage int) {
	player.ApplyDamage(damage)
}

// HealPlayer heals a player character
func (r *CombatResolver) HealPlayer(player *domain.PlayerState, amount int) {
	player.Heal(amount)
}

// InitiativeActor represents an actor in initiative calculation
type InitiativeActor struct {
	CharacterID string
	DEXModifier int
}

// CalculateInitiative calculates initiative order for a list of actors
// Returns the actors sorted by initiative roll (highest first)
func (r *CombatResolver) CalculateInitiative(actors []InitiativeActor) []InitiativeActor {
	type initiativeRoll struct {
		actor InitiativeActor
		roll  int
	}

	rolls := make([]initiativeRoll, len(actors))
	for i, actor := range actors {
		roll := rand.Intn(20) + 1 + actor.DEXModifier // d20 + DEX modifier
		rolls[i] = initiativeRoll{actor: actor, roll: roll}
	}

	// Sort by initiative roll (highest first)
	sort.Slice(rolls, func(i, j int) bool {
		return rolls[i].roll > rolls[j].roll
	})

	result := make([]InitiativeActor, len(rolls))
	for i, roll := range rolls {
		result[i] = roll.actor
	}

	return result
}

// CalculateInitiativeOrder returns just the ordered character IDs
func (r *CombatResolver) CalculateInitiativeOrder(actors []InitiativeActor) []string {
	ordered := r.CalculateInitiative(actors)
	result := make([]string, len(ordered))
	for i, actor := range ordered {
		result[i] = actor.CharacterID
	}
	return result
}

// ResolveSaveThrow resolves a saving throw
func (r *CombatResolver) ResolveSaveThrow(abilityMod int, isProficient bool, proficiencyBonus int, DC int, roll int) *SaveResult {
	total := roll + abilityMod
	if isProficient {
		total += proficiencyBonus
	}

	return &SaveResult{
		Success: total >= DC,
		Roll:    roll,
		DC:      DC,
		Total:   total,
	}
}

// ResolveSkillCheck resolves a skill check
func (r *CombatResolver) ResolveSkillCheck(abilityMod int, isProficient bool, proficiencyBonus int, DC int, roll int) *SkillCheckResult {
	total := roll + abilityMod
	if isProficient {
		total += proficiencyBonus
	}

	return &SkillCheckResult{
		Success: total >= DC,
		Roll:    roll,
		DC:      DC,
		Total:   total,
	}
}

// GetProficiencyBonus returns the proficiency bonus for a given level
func GetProficiencyBonus(level int) int {
	switch {
	case level >= 17:
		return 6
	case level >= 13:
		return 5
	case level >= 9:
		return 4
	case level >= 5:
		return 3
	default:
		return 2
	}
}

// DeathSavingThrowResult represents the result of a death saving throw
type DeathSavingThrowResult struct {
	Success   bool `json:"success"`
	Critical  bool `json:"critical"`
	Failures  int  `json:"failures"`
	Successes int  `json:"successes"`
	Stabilized bool `json:"stabilized"`
	Died      bool `json:"died"`
}

// ResolveDeathSavingThrow resolves a death saving throw
func (r *CombatResolver) ResolveDeathSavingThrow(currentSuccesses, currentFailures int, roll int) *DeathSavingThrowResult {
	result := &DeathSavingThrowResult{
		Successes: currentSuccesses,
		Failures:  currentFailures,
	}

	if roll == 20 {
		// Natural 20: regain 1 HP
		result.Critical = true
		result.Success = true
		result.Stabilized = true
		return result
	}

	if roll == 1 {
		// Natural 1: two failures
		result.Failures += 2
	} else if roll >= 10 {
		result.Successes++
		result.Success = true
	} else {
		result.Failures++
	}

	// Check for stabilization or death
	if result.Successes >= 3 {
		result.Stabilized = true
	}
	if result.Failures >= 3 {
		result.Died = true
	}

	return result
}
