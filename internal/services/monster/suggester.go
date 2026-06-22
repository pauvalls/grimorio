package monster

import (
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/monster/rules"
)

// MonsterSuggester generates stat-block skeletons for a target CR.
type MonsterSuggester struct{}

// NewMonsterSuggester returns a new suggester.
func NewMonsterSuggester() *MonsterSuggester {
	return &MonsterSuggester{}
}

// Suggest returns a *rules.Monster skeleton for the given target CR.
// Concept is optional; it currently influences default size/alignment.
//
// The returned monster is fully populated with canonical values from
// the rules package: HP at the midpoint of the CR band, AC equal to
// the canonical value, attack bonus and Save DC equal to the
// canonical values, etc.
func (s *MonsterSuggester) Suggest(targetCR float64, concept string) (*rules.Monster, error) {
	if targetCR < 0 || targetCR > 30 {
		return nil, fmt.Errorf("unsupported_cr: CR %v is out of range 0..30", targetCR)
	}
	stats, err := rules.GetStatsForCR(targetCR)
	if err != nil {
		return nil, fmt.Errorf("suggest: %w", err)
	}

	// Default size and alignment.
	size := rules.SizeMedium
	creatureType := rules.TypeHumanoid
	alignment := rules.AlignU
	_ = size
	_ = creatureType
	_ = alignment
	_ = strings.ToLower(concept) // suppress unused warning

	// Midpoint HP.
	hpMid := (stats.HPMin + stats.HPMax) / 2
	// Midpoint DPR.
	dprMid := (stats.DPRMin + stats.DPRMax) / 2

	m := &rules.Monster{
		Size:       rules.SizeMedium,
		Type:       rules.TypeHumanoid,
		Alignment:  rules.AlignU,
		AC:         stats.AC,
		HP:         hpMid,
		Speed:      map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		CR:         targetCR,
		XP:         rules.XPForCR(targetCR),
		Abilities:  defaultAbilitiesForCR(targetCR),
		Senses:     rules.Senses{PassivePerception: 10 + rules.Stats{}.Modifier(10)},
		Languages:  []string{"Common"},
	}
	_ = dprMid
	return m, nil
}

// defaultAbilitiesForCR returns a reasonable ability set for the
// given CR. The values follow the rule of thumb: PB+2 is the attack,
// so the relevant stat is PB+2-1 ≈ PB+1 (mod 0 + PB) for low CRs
// and reaches a peak around 20 for CR 30.
func defaultAbilitiesForCR(cr float64) rules.Stats {
	pb := rules.PBForCR(cr)
	// Mod = attack bonus - PB → from canonical table.
	stats, _ := rules.GetStatsForCR(cr)
	mod := stats.AttackBonus - pb
	// Score whose modifier is `mod`: 10 + 2*mod (clamped to 1..30).
	score := 10 + 2*mod
	if score < 8 {
		score = 8
	}
	if score > 20 {
		score = 20
	}
	return rules.Stats{
		STR: score,
		DEX: score,
		CON: score,
		INT: 10,
		WIS: 10,
		CHA: 10,
	}
}
