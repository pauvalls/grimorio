// Package monster provides service-layer orchestration for D&D 5e
// monster CR validation, audit, and suggestion. It glues the pure
// helpers in internal/monster/rules to the rest of the engine
// (campaign service, MCP handlers, etc.) and provides the advisory
// reports that surface CR drift without ever blocking a save.
package monster

import (
	"math"
	"strconv"
	"strings"

	"github.com/pauvalls/grimorio/internal/monster/rules"
)

// Severity is the bucket a validation finding falls into.
type Severity string

const (
	SeverityOK    Severity = "ok"
	SeverityMinor Severity = "minor"
	SeverityMajor Severity = "major"
)

// Finding is a single advisory reported by the validator. It is
// non-blocking — the engine never fails a save because of a Finding.
type Finding struct {
	Field      string   `json:"field"`
	Expected   string   `json:"expected,omitempty"`
	Actual     string   `json:"actual,omitempty"`
	Severity   Severity `json:"severity"`
	Suggestion string   `json:"suggestion,omitempty"`
}

// ValidationResult is the full output of ValidateMonster for one
// monster. The engine never modifies the input monster; it only
// reports.
type ValidationResult struct {
	Monster      *rules.Monster `json:"monster,omitempty"`
	OfficialCR   float64        `json:"official_cr"`
	CalculatedCR float64        `json:"calculated_cr"`
	DefensiveCR  float64        `json:"defensive_cr"`
	OffensiveCR  float64        `json:"offensive_cr"`
	Delta        float64        `json:"delta"`
	Severity     Severity       `json:"severity"`
	EffectiveHP  int            `json:"effective_hp"`
	Findings     []Finding      `json:"findings"`
	Suggestions  []string       `json:"suggestions"`
}

// MonsterValidator computes the expected CR for a parsed monster
// using the rules package and emits an advisory ValidationResult.
type MonsterValidator struct{}

// NewMonsterValidator returns a new validator. The struct is
// stateless; this constructor exists for symmetry with the other
// services and to allow future DI (e.g. loggers).
func NewMonsterValidator() *MonsterValidator {
	return &MonsterValidator{}
}

// Validate computes the expected CR for m. It applies the rules
// helpers and emits a ValidationResult with severity, findings, and
// suggestions. It never mutates m.
//
// Severity thresholds (per spec):
//   - |delta| <= 0.5 → OK
//   - |delta| <= 1.0 → Minor
//   - |delta|  > 1.0 → Major
func (v *MonsterValidator) Validate(m *rules.Monster) *ValidationResult {
	if m == nil {
		return &ValidationResult{Severity: SeverityMajor, Findings: []Finding{{Field: "monster", Severity: SeverityMajor, Suggestion: "monster is nil"}}}
	}
	r := &ValidationResult{
		Monster:    m,
		OfficialCR: m.CR,
		Findings:   []Finding{},
		Suggestions: []string{},
	}

	// 1. Ability score range check (DMG p. 277: 1-30).
	checkAbilityRange(m, r)

	// 2. Compute Defensive CR from HP + AC.
	hp := m.HP
	hasResist := len(m.DamageResistances) > 0
	hasImmune := len(m.DamageImmunities) > 0
	effectiveHP := rules.EffectiveHP(m.CR, hp, hasResist, hasImmune)
	r.EffectiveHP = effectiveHP

	defCR := rules.DefensiveCR(effectiveHP, m.AC)
	// Apply AC modifiers from flying and ST bonuses.
	extraAC := rules.FlyingBonusAC(m) + rules.STBonusACAdj(len(m.Saves))
	if extraAC != 0 {
		defCR = rules.AdjustCRByAC(defCR, m.AC+extraAC)
	}
	r.DefensiveCR = defCR

	// 3. Compute Offensive CR.
	// We don't have explicit DPR/attack input, so use canonical stats
	// for the declared CR as a proxy: assume the monster's DPR is
	// within range, so Offensive CR ≈ declared CR (or shifted by ±1
	// from its attack bonus relative to canonical).
	stats, err := rules.GetStatsForCR(m.CR)
	if err != nil {
		// CR is out of range; fall back to using HP-based CR.
		offCR := rules.OffensiveCR(0, 0)
		r.OffensiveCR = offCR
	} else {
		// Estimate attack bonus from ability mods + PB.
		attackBonus := estimatedAttackBonus(m, stats.PB)
		offCR := rules.OffensiveCR(float64(stats.DPRMin), attackBonus)
		r.OffensiveCR = offCR
	}

	// 4. Final CR.
	r.CalculatedCR = rules.FinalCR(r.DefensiveCR, r.OffensiveCR)
	r.Delta = math.Abs(r.CalculatedCR - r.OfficialCR)

	// 5. Severity per spec.
	switch {
	case r.Delta <= 0.5:
		r.Severity = SeverityOK
	case r.Delta <= 1.0:
		r.Severity = SeverityMinor
	default:
		r.Severity = SeverityMajor
	}

	// 6. Suggestions.
	if r.Severity != SeverityOK {
		r.Suggestions = append(r.Suggestions,
			"Consider adjusting HP, AC, attack bonus, or declared CR to match the computed value.")
		if r.Delta > 1.0 && r.CalculatedCR > r.OfficialCR {
			r.Suggestions = append(r.Suggestions,
				"Increase the declared CR, or reduce HP/AC to match the band.")
		}
		if r.Delta > 1.0 && r.CalculatedCR < r.OfficialCR {
			r.Suggestions = append(r.Suggestions,
				"Decrease the declared CR, or add HP/AC to match the band.")
		}
	}
	for _, f := range r.Findings {
		if f.Suggestion != "" {
			r.Suggestions = append(r.Suggestions, f.Suggestion)
		}
	}

	return r
}

// estimatedAttackBonus returns a rough attack bonus based on the
// monster's highest STR/DEX mod + PB.
func estimatedAttackBonus(m *rules.Monster, pb int) int {
	mod := rules.Stats{}.Modifier(m.Abilities.STR)
	dexMod := rules.Stats{}.Modifier(m.Abilities.DEX)
	if dexMod > mod {
		mod = dexMod
	}
	return mod + pb
}

// checkAbilityRange adds a Major finding for any ability score outside
// the 1..30 range (DMG p. 277).
func checkAbilityRange(m *rules.Monster, r *ValidationResult) {
	abilities := map[string]int{
		"abilities.STR": m.Abilities.STR,
		"abilities.DEX": m.Abilities.DEX,
		"abilities.CON": m.Abilities.CON,
		"abilities.INT": m.Abilities.INT,
		"abilities.WIS": m.Abilities.WIS,
		"abilities.CHA": m.Abilities.CHA,
	}
	for field, score := range abilities {
		if score < 1 || score > 30 {
			r.Findings = append(r.Findings, Finding{
				Field:      field,
				Actual:     strconv.Itoa(score),
				Expected:   "1..30",
				Severity:   SeverityMajor,
				Suggestion: "Ability score " + strings.TrimPrefix(field, "abilities.") + "=" + strconv.Itoa(score) + " is outside the DMG p. 277 range (1..30).",
			})
		}
	}
}
