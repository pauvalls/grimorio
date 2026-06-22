// Package rules implements pure D&D 5e monster-design helpers derived from
// DMG 5e cap. 9, the 2025 Monster Manual stat block format, and the SRD 5.1
// bestiary. The package is the single source of truth for CR (Challenge
// Rating) calculation, HP formulas, monster feature modifiers, and the
// markdown stat-block parser/renderer.
//
// Source of truth: docs/dnd-monster-design-rules.md
package rules

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Size follows DMG p. 277 (PHB 5e).
type Size string

const (
	SizeTiny       Size = "Tiny"
	SizeSmall      Size = "Small"
	SizeMedium     Size = "Medium"
	SizeLarge      Size = "Large"
	SizeHuge       Size = "Huge"
	SizeGargantuan Size = "Gargantuan"
)

// CreatureType follows MM 2025 p. 7 (14 types, humanoids included).
type CreatureType string

const (
	TypeAberration CreatureType = "aberration"
	TypeBeast      CreatureType = "beast"
	TypeCelestial  CreatureType = "celestial"
	TypeConstruct  CreatureType = "construct"
	TypeDragon     CreatureType = "dragon"
	TypeElemental  CreatureType = "elemental"
	TypeFey        CreatureType = "fey"
	TypeFiend      CreatureType = "fiend"
	TypeGiant      CreatureType = "giant"
	TypeHumanoid   CreatureType = "humanoid"
	TypeMonstrosity CreatureType = "monstrosity"
	TypeOoze       CreatureType = "ooze"
	TypePlant      CreatureType = "plant"
	TypeUndead     CreatureType = "undead"
)

// Alignment follows PHB 5e plus the "unaligned" value for creatures with
// no moral compass.
type Alignment string

const (
	AlignLG Alignment = "lawful good"
	AlignNG Alignment = "neutral good"
	AlignCG Alignment = "chaotic good"
	AlignLN Alignment = "lawful neutral"
	AlignTN Alignment = "true neutral"
	AlignCN Alignment = "chaotic neutral"
	AlignLE Alignment = "lawful evil"
	AlignNE Alignment = "neutral evil"
	AlignCE Alignment = "chaotic evil"
	AlignU  Alignment = "unaligned"
)

// DamageType is the 12 official D&D 5e damage types (MM 2025 p. 12).
type DamageType string

const (
	DmgAcid      DamageType = "acid"
	DmgBludgeon  DamageType = "bludgeoning"
	DmgCold      DamageType = "cold"
	DmgFire      DamageType = "fire"
	DmgForce     DamageType = "force"
	DmgLightning DamageType = "lightning"
	DmgNecrotic  DamageType = "necrotic"
	DmgPierce    DamageType = "piercing"
	DmgPsychic   DamageType = "psychic"
	DmgRadiant   DamageType = "radiant"
	DmgSlash     DamageType = "slashing"
	DmgThunder   DamageType = "thunder"
)

// Condition represents one of the 15 PHB 5e conditions a creature may be
// immune to.
type Condition string

const (
	CondBlinded       Condition = "blinded"
	CondCharmed       Condition = "charmed"
	CondDeafened      Condition = "deafened"
	CondFrightened    Condition = "frightened"
	CondGrappled      Condition = "grappled"
	CondIncapacitated Condition = "incapacitated"
	CondInvisible     Condition = "invisible"
	CondParalyzed     Condition = "paralyzed"
	CondPetrified     Condition = "petrified"
	CondPoisoned      Condition = "poisoned"
	CondProne         Condition = "prone"
	CondRestrained    Condition = "restrained"
	CondStunned       Condition = "stunned"
	CondUnconscious   Condition = "unconscious"
	CondExhaustion    Condition = "exhaustion"
)

// SpeedKind enumerates the 5 movement modes a monster can possess.
type SpeedKind string

const (
	SpeedWalk   SpeedKind = "walk"
	SpeedFly    SpeedKind = "fly"
	SpeedSwim   SpeedKind = "swim"
	SpeedBurrow SpeedKind = "burrow"
	SpeedClimb  SpeedKind = "climb"
)

// Stats holds the six ability scores. Scores must be 1-30 per DMG p. 277.
type Stats struct {
	STR int
	DEX int
	CON int
	INT int
	WIS int
	CHA int
}

// Modifier computes the D&D 5e ability modifier for a given score using
// the standard formula: floor((score - 10) / 2).
//
// Examples:
//
//	Modifier(10) == 0
//	Modifier(14) == 2
//	Modifier(9)  == -1
//	Modifier(1)  == -5
func (Stats) Modifier(score int) int {
	return int(math.Floor(float64(score-10) / 2.0))
}

// Trait is a passive or conditional characteristic (MM 2025 p. 11).
type Trait struct {
	Name        string
	Description string
}

// LimitedUse describes a usage-limited action (Recharge, 1/Day, 3/Day).
// PerDay is 0 for recharge-only abilities. RechargeMin/RechargeMax are 0
// for non-recharge abilities.
type LimitedUse struct {
	PerDay       int
	RechargeMin  int
	RechargeMax  int
}

// Action is a standard monster action (attack, ability, etc.) per MM 2025 p. 12.
type Action struct {
	Name        string
	Description string
	Limited     *LimitedUse
}

// Reaction is a triggered monster action (MM 2025 p. 14).
type Reaction struct {
	Action
}

// Spell is one spell the monster knows (MM 2025 p. 13).
type Spell struct {
	Name  string
	Level int
	Notes string
}

// Spellcasting describes a monster's spellcasting ability and repertoire
// (MM 2025 p. 13). Spells is keyed by spell level, 0 = cantrips.
type Spellcasting struct {
	Ability      string
	SaveDC       int
	AttackBonus  int
	Spells       map[int][]Spell
}

// LegendaryGroup describes a monster's legendary actions (MM 2025 p. 15).
type LegendaryGroup struct {
	Uses    int
	Actions []Action
}

// Senses captures the creature's sensory capabilities (MM 2025 p. 11).
// Special lists special senses like "darkvision 60 ft." or
// "blindsight 30 ft. (blind beyond this radius)".
type Senses struct {
	PassivePerception int
	Special           []string
}

// Monster is the engine's canonical monster representation. It is what the
// parser produces and what the validator consumes. The struct intentionally
// lives in this package (rather than internal/domain) because it is the
// engine's internal data model; the domain package persists a slimmer
// representation for the campaign canon.
//
// The CR_Defensive / CR_Offensive / EffectiveHP fields are populated by
// the validator and are metadata only — they are NOT part of the input
// the validator operates on.
type Monster struct {
	Name        string
	Size        Size
	Type        CreatureType
	Tags        []string
	Alignment   Alignment

	AC         int
	ACSource   string
	HP         int
	HPDice     string
	Speed      map[SpeedKind]int
	Initiative int // modifier (2014 + 2025)
	InitScore  int // absolute score (2025 only)

	Abilities Stats
	Saves     []string
	Skills    map[string]int

	Senses    Senses
	Languages []string
	CR        float64
	XP        int

	DamageVulnerabilities []DamageType
	DamageResistances     []DamageType
	DamageImmunities      []DamageType
	ConditionImmunities   []Condition

	Gear []string

	Traits        []Trait
	Actions       []Action
	BonusActions  []Action
	Reactions     []Reaction
	Legendary     *LegendaryGroup
	Spellcasting  *Spellcasting

	// Metadata populated by the validator (omitempty on JSON).
	CRDefensive float64
	CROffensive float64
	EffectiveHP int
}

// ErrCROutOfRange is returned by helpers when the CR is < 0 or > 30.
var ErrCROutOfRange = fmt.Errorf("CR out of range: must be 0..30 (0, 1/8, 1/4, 1/2, 1..30)")

// validCRs is the set of valid CR strings per DMG p. 274.
var validCRs = map[string]float64{
	"0":   0,
	"1/8": 0.125,
	"1/4": 0.25,
	"1/2": 0.5,
}

// ParseCR converts a CR string ("0", "1/8", "1/4", "1/2", "1".."30") to a
// float64. Accepts the canonical float form ("0.5", "0.25") for inputs
// that come from API callers. Returns ErrCROutOfRange for any value
// outside 0..30.
func ParseCR(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("ParseCR: empty string")
	}

	if v, ok := validCRs[s]; ok {
		if v < 0 || v > 30 {
			return 0, ErrCROutOfRange
		}
		return v, nil
	}

	// Try the canonical float form ("0.5", "0.25", "10", "30").
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		if f < 0 || f > 30 {
			return 0, ErrCROutOfRange
		}
		return f, nil
	}

	return 0, fmt.Errorf("ParseCR: invalid CR string %q", s)
}
