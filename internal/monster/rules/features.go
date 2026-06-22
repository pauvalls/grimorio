package rules

// FeatureEffect describes the CR-adjustment effect of a named monster
// feature from the DMG pp. 280-281 Monster Features table.
//
// The Type field encodes the kind of adjustment:
//
//   - "ac"        → +Amount to effective AC
//   - "attack"    → +Amount to effective attack bonus
//   - "damage"    → +Amount to effective per-round damage
//   - "hp"        → +Amount to effective HP (HPRule may be nil for flat)
//   - "dpr"       → +Amount to effective DPR
//   - "ac_attack" → applies to BOTH AC and attack bonus (e.g. Nimble Escape)
//
// Amount is a signed integer; positive values raise CR, negative values
// would lower it (none defined in the DMG table). HPRule is set for
// CR-band-dependent HP features (Undead Fortitude, Relentless,
// Legendary Resistance) and takes the monster's expected CR.
type FeatureEffect struct {
	Type    string
	Amount  int
	HPRule  func(cr float64) int
	Note    string
}

// undeadFortitudeHP is the canonical CR-band HP bonus for
// Undead Fortitude and Relentless (DMG p. 280).
func undeadFortitudeHP(cr float64) int {
	switch {
	case cr <= 4:
		return 7
	case cr <= 10:
		return 14
	case cr <= 16:
		return 21
	default:
		return 28
	}
}

// legendaryResistanceHP is the canonical CR-band HP bonus for
// Legendary Resistance (DMG p. 280).
func legendaryResistanceHP(cr float64) int {
	switch {
	case cr <= 4:
		return 10
	case cr <= 10:
		return 20
	default:
		return 30
	}
}

// monsterFeatures is the lookup table for the DMG pp. 280-281 Monster
// Features table. Features with "—" (no CR effect) are intentionally
// absent — FeatureFor returns ok=false for them.
var monsterFeatures = map[string]FeatureEffect{
	// Damage / DPR features.
	"Aggressive":           {Type: "damage", Amount: 2, Note: "+2 effective per-round damage (DMG p. 280)"},
	"Rampage":              {Type: "damage", Amount: 2, Note: "+2 effective per-round damage (DMG p. 281)"},
	"Angelic Weapons":      {Type: "dpr", Amount: 0, Note: "+X per-round damage (per trait)"},
	"Brute":                {Type: "dpr", Amount: 0, Note: "+X per-round damage (per trait)"},
	"Charge":               {Type: "dpr", Amount: 0, Note: "+X damage on one attack (per trait)"},
	"Death Burst":          {Type: "dpr", Amount: 0, Note: "+X damage for 1 round; assume 2 creatures"},
	"Dive":                 {Type: "dpr", Amount: 0, Note: "+X damage on one attack (per trait)"},
	"Elemental Body":       {Type: "dpr", Amount: 0, Note: "+X per-round damage (per trait)"},
	"Enlarge":              {Type: "dpr", Amount: 0, Note: "+X per-round damage (per trait)"},
	"Martial Advantage":    {Type: "dpr", Amount: 0, Note: "+X damage on one attack (per trait)"},
	"Pounce":               {Type: "dpr", Amount: 0, Note: "+X damage for 1 round (per trait)"},
	"Surprise Attack":      {Type: "dpr", Amount: 0, Note: "+X damage for 1 round (per trait)"},
	"Wounded Fury":         {Type: "dpr", Amount: 0, Note: "+X damage for 1 round (per trait)"},

	// Attack bonus features.
	"Ambusher":             {Type: "attack", Amount: 1, Note: "+1 effective attack bonus (DMG p. 280)"},
	"Blood Frenzy":         {Type: "attack", Amount: 4, Note: "+4 effective attack bonus (DMG p. 280)"},
	"Pack Tactics":         {Type: "attack", Amount: 1, Note: "+1 effective attack bonus (DMG p. 281)"},

	// AC features.
	"Avoidance":            {Type: "ac", Amount: 1, Note: "+1 effective AC (DMG p. 280)"},
	"Constrict":            {Type: "ac", Amount: 1, Note: "+1 effective AC (DMG p. 280)"},
	"Magic Resistance":     {Type: "ac", Amount: 2, Note: "+2 effective AC (DMG p. 280)"},
	"Parry":                {Type: "ac", Amount: 1, Note: "+1 effective AC (DMG p. 281)"},
	"Shadow Stealth":       {Type: "ac", Amount: 4, Note: "+4 effective AC (DMG p. 281)"},
	"Stench":               {Type: "ac", Amount: 1, Note: "+1 effective AC (DMG p. 281)"},
	"Superior Invisibility": {Type: "ac", Amount: 2, Note: "+2 effective AC (DMG p. 281)"},
	"Web":                  {Type: "ac", Amount: 1, Note: "+1 effective AC (DMG p. 281)"},

	// Both AC and attack.
	"Nimble Escape":        {Type: "ac_attack", Amount: 4, Note: "+4 effective AC and +4 effective attack bonus (DMG p. 281)"},

	// HP features with CR-dependent values.
	"Undead Fortitude":     {Type: "hp", HPRule: undeadFortitudeHP, Note: "+7/14/21/28 effective HP (CR 1-4/5-10/11-16/17+) — DMG p. 281"},
	"Relentless":           {Type: "hp", HPRule: undeadFortitudeHP, Note: "+7/14/21/28 effective HP (CR 1-4/5-10/11-16/17+) — DMG p. 281"},
	"Legendary Resistance": {Type: "hp", HPRule: legendaryResistanceHP, Note: "+10/20/30 effective HP per use (CR 1-4/5-10/11+) — DMG p. 280"},

	// "Punch" effects encoded as the "—" entries in MD §4.4 (kept for completeness).
	"Breath Weapon":        {Type: "special", Note: "Assume 2 targets and both fail the save (DMG p. 280)"},
	"Damage Transfer":      {Type: "hp", Amount: 0, Note: "×2 effective HP; +1/3 HP to per-round damage (DMG p. 280)"},
	"Fiendish Blessing":    {Type: "special", Note: "Apply CHA mod to real AC (DMG p. 280)"},
	"Frightful Presence":   {Type: "hp", Amount: 0, Note: "+25% effective HP if CR ≤ 10 (DMG p. 280)"},
	"Possession":           {Type: "hp", Amount: 0, Note: "×2 effective HP (DMG p. 281)"},
	"Psychic Defense":      {Type: "special", Note: "Apply WIS mod to real AC (DMG p. 281)"},
	"Regeneration":         {Type: "hp", Amount: 0, Note: "+3×HP_regen_per_round effective HP (DMG p. 281)"},
	"Swallow":              {Type: "special", Note: "Assume 1 creature swallowed + 2 rounds of acid damage (DMG p. 281)"},
}

// FeatureFor returns the FeatureEffect for the named feature. The
// second return value is false when the feature has no CR effect
// (the "—" entries in DMG p. 281).
func FeatureFor(name string) (FeatureEffect, bool) {
	f, ok := monsterFeatures[name]
	return f, ok
}

// AllFeatures returns the list of every feature name known to the
// engine (i.e. features with a CR effect). Order is not guaranteed.
func AllFeatures() []string {
	names := make([]string, 0, len(monsterFeatures))
	for name := range monsterFeatures {
		names = append(names, name)
	}
	return names
}
