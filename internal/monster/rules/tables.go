package rules

// CRStats captures the 6 canonical columns of DMG p. 274 for a single CR.
type CRStats struct {
	PB           int
	AC           int
	HPMin        int
	HPMax        int
	AttackBonus  int
	DPRMin       float64
	DPRMax       float64
	SaveDC       int
}

// crMasterTable encodes the full DMG p. 274 Monster Statistics by Challenge
// Rating table. CR 0 uses the "≤ 13 / ≤ +3 / 0-1 / ≤ 13" row (per DMG p. 274).
//
// Sub-integer CRs (0.125 / 0.25 / 0.5) are stored with their float key.
// All 31+sub-integer rows are present.
var crMasterTable = map[float64]CRStats{
	0:     {PB: 2, AC: 13, HPMin: 1, HPMax: 6, AttackBonus: 3, DPRMin: 0, DPRMax: 1, SaveDC: 13},
	0.125: {PB: 2, AC: 13, HPMin: 7, HPMax: 35, AttackBonus: 3, DPRMin: 2, DPRMax: 3, SaveDC: 13},
	0.25:  {PB: 2, AC: 13, HPMin: 36, HPMax: 49, AttackBonus: 3, DPRMin: 4, DPRMax: 5, SaveDC: 13},
	0.5:   {PB: 2, AC: 13, HPMin: 50, HPMax: 70, AttackBonus: 3, DPRMin: 6, DPRMax: 8, SaveDC: 13},
	1:     {PB: 2, AC: 13, HPMin: 71, HPMax: 85, AttackBonus: 3, DPRMin: 9, DPRMax: 14, SaveDC: 13},
	2:     {PB: 2, AC: 13, HPMin: 86, HPMax: 100, AttackBonus: 3, DPRMin: 15, DPRMax: 20, SaveDC: 13},
	3:     {PB: 2, AC: 13, HPMin: 101, HPMax: 115, AttackBonus: 4, DPRMin: 21, DPRMax: 26, SaveDC: 13},
	4:     {PB: 2, AC: 14, HPMin: 116, HPMax: 130, AttackBonus: 5, DPRMin: 27, DPRMax: 32, SaveDC: 14},
	5:     {PB: 3, AC: 15, HPMin: 131, HPMax: 145, AttackBonus: 6, DPRMin: 33, DPRMax: 38, SaveDC: 15},
	6:     {PB: 3, AC: 15, HPMin: 146, HPMax: 160, AttackBonus: 6, DPRMin: 39, DPRMax: 44, SaveDC: 15},
	7:     {PB: 3, AC: 15, HPMin: 161, HPMax: 175, AttackBonus: 6, DPRMin: 45, DPRMax: 50, SaveDC: 15},
	8:     {PB: 3, AC: 16, HPMin: 176, HPMax: 190, AttackBonus: 7, DPRMin: 51, DPRMax: 56, SaveDC: 16},
	9:     {PB: 4, AC: 16, HPMin: 191, HPMax: 205, AttackBonus: 7, DPRMin: 57, DPRMax: 62, SaveDC: 16},
	10:    {PB: 4, AC: 17, HPMin: 206, HPMax: 220, AttackBonus: 7, DPRMin: 63, DPRMax: 68, SaveDC: 16},
	11:    {PB: 4, AC: 17, HPMin: 221, HPMax: 235, AttackBonus: 8, DPRMin: 69, DPRMax: 74, SaveDC: 17},
	12:    {PB: 4, AC: 17, HPMin: 236, HPMax: 250, AttackBonus: 8, DPRMin: 75, DPRMax: 80, SaveDC: 17},
	13:    {PB: 5, AC: 18, HPMin: 251, HPMax: 265, AttackBonus: 8, DPRMin: 81, DPRMax: 86, SaveDC: 18},
	14:    {PB: 5, AC: 18, HPMin: 266, HPMax: 280, AttackBonus: 8, DPRMin: 87, DPRMax: 92, SaveDC: 18},
	15:    {PB: 5, AC: 18, HPMin: 281, HPMax: 295, AttackBonus: 8, DPRMin: 93, DPRMax: 98, SaveDC: 18},
	16:    {PB: 5, AC: 18, HPMin: 296, HPMax: 310, AttackBonus: 9, DPRMin: 99, DPRMax: 104, SaveDC: 18},
	17:    {PB: 6, AC: 19, HPMin: 311, HPMax: 325, AttackBonus: 10, DPRMin: 105, DPRMax: 110, SaveDC: 19},
	18:    {PB: 6, AC: 19, HPMin: 326, HPMax: 340, AttackBonus: 10, DPRMin: 111, DPRMax: 116, SaveDC: 19},
	19:    {PB: 6, AC: 19, HPMin: 341, HPMax: 355, AttackBonus: 10, DPRMin: 117, DPRMax: 122, SaveDC: 19},
	20:    {PB: 6, AC: 19, HPMin: 356, HPMax: 400, AttackBonus: 10, DPRMin: 123, DPRMax: 140, SaveDC: 19},
	21:    {PB: 7, AC: 19, HPMin: 401, HPMax: 445, AttackBonus: 11, DPRMin: 141, DPRMax: 158, SaveDC: 20},
	22:    {PB: 7, AC: 19, HPMin: 446, HPMax: 490, AttackBonus: 11, DPRMin: 159, DPRMax: 176, SaveDC: 20},
	23:    {PB: 7, AC: 19, HPMin: 491, HPMax: 535, AttackBonus: 11, DPRMin: 177, DPRMax: 194, SaveDC: 20},
	24:    {PB: 7, AC: 19, HPMin: 536, HPMax: 580, AttackBonus: 12, DPRMin: 195, DPRMax: 212, SaveDC: 21},
	25:    {PB: 8, AC: 19, HPMin: 581, HPMax: 625, AttackBonus: 12, DPRMin: 213, DPRMax: 230, SaveDC: 21},
	26:    {PB: 8, AC: 19, HPMin: 626, HPMax: 670, AttackBonus: 12, DPRMin: 231, DPRMax: 248, SaveDC: 21},
	27:    {PB: 8, AC: 19, HPMin: 671, HPMax: 715, AttackBonus: 13, DPRMin: 249, DPRMax: 266, SaveDC: 22},
	28:    {PB: 8, AC: 19, HPMin: 716, HPMax: 760, AttackBonus: 13, DPRMin: 267, DPRMax: 284, SaveDC: 22},
	29:    {PB: 9, AC: 19, HPMin: 761, HPMax: 805, AttackBonus: 13, DPRMin: 285, DPRMax: 302, SaveDC: 22},
	30:    {PB: 9, AC: 19, HPMin: 806, HPMax: 850, AttackBonus: 14, DPRMin: 303, DPRMax: 320, SaveDC: 23},
}

// validCRs is the canonical set of supported CR values (used to validate
// the input to PBForCR / XPForCR and to compute the average for FinalCR).
var validCRsList = []float64{
	0, 0.125, 0.25, 0.5,
	1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
	11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
}

// xpTable is the XP-for-CR table from DMG p. 282 / MM 2025 pp. 214-250.
// CR 0 returns 10 (the "with stat block" value, per DMG p. 282).
var xpTable = map[float64]int{
	0:     10,
	0.125: 25,
	0.25:  50,
	0.5:   100,
	1:     200,
	2:     450,
	3:     700,
	4:     1100,
	5:     1800,
	6:     2300,
	7:     2900,
	8:     3900,
	9:     5000,
	10:    5900,
	11:    7200,
	12:    8400,
	13:    10000,
	14:    11500,
	15:    13000,
	16:    15000,
	17:    18000,
	18:    20000,
	19:    22000,
	20:    25000,
	21:    33000,
	22:    41000,
	23:    50000,
	24:    62000,
	25:    75000,
	26:    90000,
	27:    105000,
	28:    120000,
	29:    135000,
	30:    155000,
}

// pbForCR returns the Proficiency Bonus for a given CR (DMG p. 274).
// 8 bands: 0-4, 5-8, 9-12, 13-16, 17-20, 21-24, 25-28, 29-30.
func pbForCR(cr float64) int {
	switch {
	case cr < 5:
		return 2
	case cr < 9:
		return 3
	case cr < 13:
		return 4
	case cr < 17:
		return 5
	case cr < 21:
		return 6
	case cr < 25:
		return 7
	case cr < 29:
		return 8
	default:
		return 9
	}
}

// hitDiceBySize maps each size to its hit die size (DMG p. 277).
var hitDiceBySize = map[Size]int{
	SizeTiny:       4,
	SizeSmall:      6,
	SizeMedium:     8,
	SizeLarge:      10,
	SizeHuge:       12,
	SizeGargantuan: 20,
}

// avgHPPerDie maps each size to the average HP per die (DMG p. 277).
var avgHPPerDieMap = map[Size]float64{
	SizeTiny:       2.5,
	SizeSmall:      3.5,
	SizeMedium:     4.5,
	SizeLarge:      5.5,
	SizeHuge:       6.5,
	SizeGargantuan: 10.5,
}

// effectiveHPResistMults maps CR bands to the resistance HP multiplier
// (DMG p. 278).
var effectiveHPResistMults = map[string]float64{
	"1-4":   2,
	"5-10":  1.5,
	"11-16": 1.25,
	"17+":   1,
}

// effectiveHPImmMults maps CR bands to the immunity HP multiplier
// (DMG p. 278).
var effectiveHPImmMults = map[string]float64{
	"1-4":   2,
	"5-10":  2,
	"11-16": 1.5,
	"17+":   1.25,
}

// crBandForHP returns the resistance/immunity CR band label for a given CR.
func crBandForHP(cr float64) string {
	switch {
	case cr <= 4:
		return "1-4"
	case cr <= 10:
		return "5-10"
	case cr <= 16:
		return "11-16"
	default:
		return "17+"
	}
}
