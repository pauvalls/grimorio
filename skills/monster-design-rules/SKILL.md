---
name: monster-design-rules
description: D&D 5e monster design rules (DMG cap. 9) — CR calculation, stat block format, modifiers. Use when generating, validating, or auditing monsters, or when designing encounters that must be CR-balanced.
license: OGL-1.0a
metadata:
  source: docs/dnd-monster-design-rules.md
  primary_source: DMG 5e cap. 9
  mm_format: 2025
---

# Monster Design Rules (D&D 5e)

> **Source of truth**: `docs/dnd-monster-design-rules.md` (810 lines). This skill
> is the actionable index. The MD has the full spec.

## When to use

- Generating a new monster with correct CR / VD
- Validating a pre-existing stat block
- Auditing a campaign's bestiary
- Designing encounters with CR balance
- Reconciling monster stats between the 2014 and 2025 MM formats

## Available MCP tools (use these — do not write your own)

- `validate_monster` — validate a single monster (by name or markdown)
- `suggest_monster_cr` — given a target CR, return a stat block skeleton
- `audit_monster_cr` — audit an entire campaign's bestiary

## Core algorithm (Defensive / Offensive / Final CR)

1. **Defensive CR** = CR by HP, then ±1 if AC differs by ≥2 from the canonical AC.
2. **Offensive CR** = CR by DPR, then ±1 if attack bonus or save DC differs by ≥2.
3. **Final CR** = average of Defensive and Offensive, rounded to the nearest valid CR.

See MD §3.2 for the full algorithm.

## Canonical tables (do not hardcode — call the helpers)

| What you need | Helper |
|---|---|
| CR → {PB, AC, HP, Atk, DPR, DC} | `internal/monster/rules.GetStatsForCR(cr)` |
| CR → XP | `internal/monster/rules.XPForCR(cr)` |
| CR → PB | `internal/monster/rules.PBForCR(cr)` |
| Size → Hit Die | `internal/monster/rules.HitDieForSize(size)` |
| Defensive CR | `internal/monster/rules.DefensiveCR(hp, ac)` |
| Offensive CR | `internal/monster/rules.OffensiveCR(dpr, attack)` |
| Final CR | `internal/monster/rules.FinalCR(defensive, offensive)` |
| Feature effect (e.g. Pack Tactics) | `internal/monster/rules.FeatureFor(name)` |

## Hard rules (NEVER violate)

- **HP, AC, attack bonus, save DC, DPR must fall within the canonical range of the monster's CR** (±1 band of tolerance). Beyond that, the monster is broken.
- **Ability scores must be in [1, 30]**. (DMG p. 277)
- **Damage notation: number OR dice, never both** in the same stat block. (MM 2025 p. 13)
- **Initiative in 2025 format**: `+X (+Y)` (mod + absolute score). (MM 2025 p. 9)
- **"None" entries are NOT emitted**. Omit empty sections. (MM 2025 p. 8)
- **Effective HP multipliers** apply when the monster has resistance or immunity to common damage types. (DMG p. 278)
- **Flying + ranged + CR ≤ 10** = +2 effective AC. (DMG p. 280)
- **3+ saving throw proficiencies** = +2 effective AC. (DMG p. 280)

## Workflow for generating a new monster

1. Pick a target CR (based on party level).
2. Call `suggest_monster_cr(target_cr, concept)` to get a skeleton.
3. Customize the skeleton's narrative, traits, actions.
4. Call `validate_monster(markdown=<your monster>)` to confirm the CR.
5. Only save if `severity=ok` or `severity=minor`. Regenerate if `major`.

## Workflow for designing an encounter

1. Pick a target CR for the encounter.
2. Pick the creatures.
3. Call `validate_monster(markdown=<each creature>)` for each.
4. All creatures must have `severity ≤ minor`.
5. Sum the XP and apply the encounter multiplier (DMG p. 82) to compute the encounter's final XP budget.

## See also

- `dnd-5e-srd` — broader D&D 5e system reference
- `grimorio-bestiary` — how to use these rules in bestiary generation
- `grimorio-encounters` — how to use these rules in encounter design
