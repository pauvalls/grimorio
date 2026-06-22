// Package rules implements pure D&D 5e monster-design helpers derived from
// DMG 5e cap. 9, the 2025 Monster Manual stat block format, and the SRD 5.1
// bestiary. The package is the single source of truth for CR (Challenge
// Rating) calculation, HP formulas, monster feature modifiers, and the
// markdown stat-block parser/renderer.
//
// The package is organized as seven sibling files so each can be unit-tested
// in isolation:
//
//   - domain.go         : typed enum-like constants and the Monster struct
//   - tables.go         : private canonical lookup tables (XP, PB, CR master, hit dice)
//   - cr_calculator.go  : public CR computation helpers (Defensive, Offensive, Final, Adjust)
//   - hp_calculator.go  : public HP formulas (HitDieForSize, AvgHPPerDie, HPFromHitDice)
//   - modifiers.go      : public CR modifiers (EffectiveHP, flying bonus, ST bonus, ParseCR)
//   - features.go       : the Monster Features table (DMG pp. 280-281)
//   - parser/           : markdown stat-block parser
//   - renderer/         : 2025 MM stat-block renderer
//
// Source of truth: docs/dnd-monster-design-rules.md
package rules
