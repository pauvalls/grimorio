package game

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// DiceSpec represents a parsed dice notation
type DiceSpec struct {
	Count    int
	Sides    int
	Modifier int
}

// String returns the canonical dice notation string
func (d *DiceSpec) String() string {
	if d.Count == 1 && d.Modifier == 0 {
		return fmt.Sprintf("d%d", d.Sides)
	}
	if d.Modifier == 0 {
		return fmt.Sprintf("%dd%d", d.Count, d.Sides)
	}
	if d.Modifier > 0 {
		return fmt.Sprintf("%dd%d+%d", d.Count, d.Sides, d.Modifier)
	}
	return fmt.Sprintf("%dd%d%d", d.Count, d.Sides, d.Modifier)
}



// diceNotationRegex matches dice notation like "2d6+3" or "d20"
// Groups: 1=count (optional), 2=sides, 3=sign+modifier (optional)
var diceNotationRegex = regexp.MustCompile(`^(\d+)?[dD](\d+)([+-]?\d+)?$`)

// ParseDice parses a dice notation string and returns a DiceSpec
// Supported formats: "d20", "2d6", "1d8+3", "3d10-2"
func ParseDice(notation string) (*DiceSpec, error) {
	notation = strings.TrimSpace(notation)
	
	if notation == "" {
		return nil, fmt.Errorf("empty dice notation")
	}

	matches := diceNotationRegex.FindStringSubmatch(notation)
	if matches == nil {
		return nil, fmt.Errorf("invalid dice notation: %s (expected format like '2d6+3')", notation)
	}

	// Parse count (default to 1 if not specified)
	count := 1
	if matches[1] != "" {
		var err error
		count, err = strconv.Atoi(matches[1])
		if err != nil {
			return nil, fmt.Errorf("invalid dice count: %s", matches[1])
		}
	}

	// Parse sides
	sides, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid dice sides: %s", matches[2])
	}

	// Parse modifier (default to 0)
	modifier := 0
	if matches[3] != "" {
		modifier, err = strconv.Atoi(matches[3])
		if err != nil {
			return nil, fmt.Errorf("invalid dice modifier: %s", matches[3])
		}
	}

	// Validation
	if count <= 0 {
		return nil, fmt.Errorf("dice count must be positive, got %d", count)
	}
	if sides <= 0 {
		return nil, fmt.Errorf("dice sides must be positive, got %d", sides)
	}

	return &DiceSpec{
		Count:    count,
		Sides:    sides,
		Modifier: modifier,
	}, nil
}

// Roll rolls dice according to the specification and returns the result
func Roll(spec *DiceSpec) *domain.DiceResult {
	if spec == nil || spec.Count <= 0 {
		return &domain.DiceResult{
			Dice:     spec.String(),
			Results:  []int{},
			Total:    spec.Modifier,
			Modifier: spec.Modifier,
		}
	}

	results := make([]int, spec.Count)
	sum := 0

	for i := 0; i < spec.Count; i++ {
		result := rand.Intn(spec.Sides) + 1 // 1 to sides (inclusive)
		results[i] = result
		sum += result
	}

	total := sum + spec.Modifier

	return &domain.DiceResult{
		Dice:     spec.String(),
		Results:  results,
		Total:    total,
		Modifier: spec.Modifier,
	}
}

// RollWithAdvantage rolls twice and returns the higher result
func RollWithAdvantage(spec *DiceSpec) (*domain.DiceResult, *domain.DiceResult, int) {
	r1 := Roll(spec)
	r2 := Roll(spec)
	
	if r1.Total >= r2.Total {
		return r1, r2, r1.Total
	}
	return r1, r2, r2.Total
}

// RollWithDisadvantage rolls twice and returns the lower result
func RollWithDisadvantage(spec *DiceSpec) (*domain.DiceResult, *domain.DiceResult, int) {
	r1 := Roll(spec)
	r2 := Roll(spec)
	
	if r1.Total <= r2.Total {
		return r1, r2, r1.Total
	}
	return r1, r2, r2.Total
}

// RollAbilityCheck rolls a d20 with ability modifier and proficiency bonus
func RollAbilityCheck(abilityMod int, proficiencyBonus int, isProficient bool) (*domain.DiceResult, int) {
	spec := &DiceSpec{Count: 1, Sides: 20, Modifier: abilityMod}
	if isProficient {
		spec.Modifier += proficiencyBonus
	}
	return Roll(spec), spec.Modifier
}

// RollAttack rolls an attack (d20 + attack bonus vs AC)
func RollAttack(attackBonus int, advantage bool, disadvantage bool) (*domain.DiceResult, int) {
	spec := &DiceSpec{Count: 1, Sides: 20, Modifier: attackBonus}
	
	var result *domain.DiceResult
	if advantage && !disadvantage {
		r1, r2, _ := RollWithAdvantage(spec)
		result = r1
		if r2.Total > r1.Total {
			result = r2
		}
		result.Results = append(result.Results, r2.Results...)
	} else if disadvantage && !advantage {
		r1, r2, _ := RollWithDisadvantage(spec)
		result = r1
		if r2.Total < r1.Total {
			result = r2
		}
		result.Results = append(result.Results, r2.Results...)
	} else {
		result = Roll(spec)
	}
	
	return result, result.Total
}

// RollDamage rolls damage dice
func RollDamage(dice string, isCritical bool) (*domain.DiceResult, error) {
	spec, err := ParseDice(dice)
	if err != nil {
		return nil, err
	}
	
	if isCritical {
		// Double the dice on critical hit
		spec.Count *= 2
	}
	
	return Roll(spec), nil
}
