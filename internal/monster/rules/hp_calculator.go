package rules

import "math"

// HPFromHitDice computes the average HP from `numDice` of `dieSize` plus
// the per-die CON modifier (DMG p. 277). The average HP per die is
// floored (rounded DOWN) per D&D 5e convention.
//
//   - HPFromHitDice(5, 8, 1)   == 27  (Medium, 5d8, CON +1; 22.5 + 5 = 27.5 → 27)
//   - HPFromHitDice(0, 8, 0)   == 0
//   - HPFromHitDice(10, 20, 5) == 155 (Gargantuan, 10d20, CON +5)
func HPFromHitDice(numDice, dieSize, conMod int) int {
	if numDice <= 0 {
		return 0
	}
	avg, ok := avgPerDie[dieSize]
	if !ok {
		avg = float64(dieSize+1) / 2.0
	}
	total := float64(numDice)*avg + float64(numDice*conMod)
	return int(math.Floor(total))
}

// avgPerDie maps a die size to its average roll.
var avgPerDie = map[int]float64{
	4:  2.5,
	6:  3.5,
	8:  4.5,
	10: 5.5,
	12: 6.5,
	20: 10.5,
}

// HitDieForSize returns the hit die size for a given creature size
// (DMG p. 277). Returns 8 (Medium) for unknown sizes.
func HitDieForSize(s Size) int {
	d, ok := hitDiceBySize[s]
	if !ok {
		return 8
	}
	return d
}

// AvgHPPerDie returns the average HP per die for a given creature size
// (DMG p. 277).
func AvgHPPerDie(s Size) float64 {
	v, ok := avgHPPerDieMap[s]
	if !ok {
		return 4.5 // Medium default
	}
	return v
}
