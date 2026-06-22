package rules

import "math"

// HPMultiplierForResistances returns the HP multiplier applied when a
// creature has damage resistances, per the Effective HP table (DMG p. 278).
//
//   - CR 1-4   → 2
//   - CR 5-10  → 1.5
//   - CR 11-16 → 1.25
//   - CR 17+   → 1
func HPMultiplierForResistances(cr float64) float64 {
	return effectiveHPResistMults[crBandForHP(cr)]
}

// HPMultiplierForImmunities returns the HP multiplier applied when a
// creature has damage immunities, per the Effective HP table (DMG p. 278).
//
//   - CR 1-4   → 2
//   - CR 5-10  → 2
//   - CR 11-16 → 1.5
//   - CR 17+   → 1.25
func HPMultiplierForImmunities(cr float64) float64 {
	return effectiveHPImmMults[crBandForHP(cr)]
}

// EffectiveHP applies the resistance/immunity multipliers to a base HP
// value, per the Effective HP table (DMG p. 278). The multipliers stack
// additively (resistances + immunities), then are applied to base HP and
// rounded.
func EffectiveHP(cr float64, hp int, hasResistances, hasImmunities bool) int {
	if hp <= 0 {
		return 0
	}
	mult := 1.0
	if hasResistances {
		mult += HPMultiplierForResistances(cr) - 1
	}
	if hasImmunities {
		mult += HPMultiplierForImmunities(cr) - 1
	}
	if mult < 1 {
		mult = 1
	}
	return int(math.Floor(float64(hp)*mult + 0.5))
}

// IsFlyingRangedUnderCR10 returns true when the monster can fly AND deal
// damage at range AND its CR is ≤ 10 (DMG p. 280).
func IsFlyingRangedUnderCR10(m *Monster) bool {
	if m == nil {
		return false
	}
	if m.CR > 10 {
		return false
	}
	canFly := m.Speed[SpeedFly] > 0
	if !canFly {
		return false
	}
	return hasRangedDamage(m)
}

// FlyingBonusAC returns +2 if the monster qualifies for the flying AC
// bonus (DMG p. 280), otherwise 0.
func FlyingBonusAC(m *Monster) int {
	if IsFlyingRangedUnderCR10(m) {
		return 2
	}
	return 0
}

// hasRangedDamage inspects the monster's actions to detect a ranged attack.
// We look for "ranged" or "range" in the description (lowercase match).
func hasRangedDamage(m *Monster) bool {
	all := append([]Action{}, m.Actions...)
	all = append(all, m.BonusActions...)
	for _, t := range m.Traits {
		_ = t
	}
	for _, a := range all {
		desc := lower(a.Description)
		if containsWord(desc, "ranged") || containsWord(desc, "range ") {
			return true
		}
	}
	return false
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func containsWord(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// STBonusACAdj returns the effective AC adjustment for the number of ST
// proficiencies the monster has, per the Saving Throw Bonuses table
// (DMG p. 280).
//
//   - 0..2 bonuses → +0
//   - 3..4 bonuses → +2
//   - 5..6 bonuses → +4
//   - ≥ 7 bonuses  → +4 (capped)
func STBonusACAdj(stBonuses int) int {
	switch {
	case stBonuses <= 2:
		return 0
	case stBonuses <= 4:
		return 2
	default:
		return 4
	}
}
