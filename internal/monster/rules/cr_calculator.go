package rules

// GetStatsForCR returns the canonical stats (PB, AC, HP range, Attack bonus,
// DPR range, Save DC) for the given CR, per DMG p. 274.
//
// Returns ErrCROutOfRange when cr is < 0 or > 30.
func GetStatsForCR(cr float64) (CRStats, error) {
	if cr < 0 || cr > 30 {
		return CRStats{}, ErrCROutOfRange
	}
	stats, ok := crMasterTable[cr]
	if !ok {
		// Snap to nearest valid CR (handles sub-integer drift).
		snapped := roundToValidCR(cr)
		stats, ok = crMasterTable[snapped]
		if !ok {
			return CRStats{}, ErrCROutOfRange
		}
	}
	return stats, nil
}

// roundToValidCR snaps a CR to the nearest valid CR value, rounding UP
// on ties (e.g. 2.5 → CR 3). This matches the DMG p. 275 example
// FinalCR(2, 3) = 3.
func roundToValidCR(cr float64) float64 {
	best := validCRsList[0]
	bestDist := absFloat(cr - best)
	for _, v := range validCRsList[1:] {
		d := absFloat(cr - v)
		// Strictly less wins; equal distance prefers the higher value.
		if d < bestDist || (d == bestDist && v > best) {
			best = v
			bestDist = d
		}
	}
	return best
}

func absFloat(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

// XPForCR returns the XP value for a given CR per DMG p. 282.
// Returns 0 for CRs outside the supported range.
func XPForCR(cr float64) int {
	if cr < 0 || cr > 30 {
		return 0
	}
	xp, ok := xpTable[cr]
	if !ok {
		xp = xpTable[roundToValidCR(cr)]
	}
	return xp
}

// PBForCR returns the Proficiency Bonus for a given CR per DMG p. 274.
func PBForCR(cr float64) int {
	if cr < 0 || cr > 30 {
		return 2 // minimum PB
	}
	return pbForCR(cr)
}

// DefensiveCRFromHP finds the base CR whose HP range contains the given
// HP value (DMG p. 275 step 1).
func DefensiveCRFromHP(hp int) float64 {
	if hp < 0 {
		return 0
	}
	// Special-case CR 0 (1..6 HP).
	if hp <= 6 {
		return 0
	}
	for _, cr := range validCRsList {
		stats, ok := crMasterTable[cr]
		if !ok {
			continue
		}
		if hp >= stats.HPMin && hp <= stats.HPMax {
			return cr
		}
	}
	return 30
}

// DefensiveCR computes the Defensive CR by mapping HP→CR then adjusting ±1
// when |realAC − expectedAC| ≥ 2 (DMG p. 275).
func DefensiveCR(hp, ac int) float64 {
	base := DefensiveCRFromHP(hp)
	return AdjustCRByAC(base, ac)
}

// AdjustCRByAC shifts a CR by ±1 when the actual AC differs from the
// expected AC by ≥ 2 (DMG p. 275).
//
//   - AC is ≥ 2 lower than expected → CR - 1
//   - AC is ≥ 2 higher than expected → CR + 1
//   - Otherwise → unchanged
func AdjustCRByAC(baseCR float64, ac int) float64 {
	stats, err := GetStatsForCR(baseCR)
	if err != nil {
		return baseCR
	}
	diff := ac - stats.AC
	if diff >= 2 {
		return shiftCR(baseCR, +1)
	}
	if diff <= -2 {
		return shiftCR(baseCR, -1)
	}
	return baseCR
}

// OffensiveCRFromDPR finds the base CR whose DPR range contains the given
// DPR value (DMG p. 275 step 2).
func OffensiveCRFromDPR(dpr float64) float64 {
	if dpr < 0 {
		return 0
	}
	// CR 0 covers 0..1 DPR.
	if dpr <= 1 {
		return 0
	}
	for _, cr := range validCRsList {
		stats, ok := crMasterTable[cr]
		if !ok {
			continue
		}
		if dpr >= stats.DPRMin && dpr <= stats.DPRMax {
			return cr
		}
	}
	return 30
}

// OffensiveCR computes the Offensive CR from DPR and attack bonus.
func OffensiveCR(dpr float64, attackBonus int) float64 {
	base := OffensiveCRFromDPR(dpr)
	return AdjustCRByAttack(base, attackBonus)
}

// OffensiveCRFromDC computes the Offensive CR from DPR and save DC.
func OffensiveCRFromDC(dpr float64, saveDC int) float64 {
	base := OffensiveCRFromDPR(dpr)
	return AdjustCRBySaveDC(base, saveDC)
}

// AdjustCRByAttack shifts a CR by ±1 when the actual attack bonus differs
// from the expected attack bonus by ≥ 2 (DMG p. 275).
func AdjustCRByAttack(baseCR float64, attackBonus int) float64 {
	stats, err := GetStatsForCR(baseCR)
	if err != nil {
		return baseCR
	}
	diff := attackBonus - stats.AttackBonus
	if diff >= 2 {
		return shiftCR(baseCR, +1)
	}
	if diff <= -2 {
		return shiftCR(baseCR, -1)
	}
	return baseCR
}

// AdjustCRBySaveDC shifts a CR by ±1 when the actual save DC differs from
// the expected save DC by ≥ 2 (DMG p. 275).
func AdjustCRBySaveDC(baseCR float64, saveDC int) float64 {
	stats, err := GetStatsForCR(baseCR)
	if err != nil {
		return baseCR
	}
	diff := saveDC - stats.SaveDC
	if diff >= 2 {
		return shiftCR(baseCR, +1)
	}
	if diff <= -2 {
		return shiftCR(baseCR, -1)
	}
	return baseCR
}

// FinalCR averages the defensive and offensive CRs and rounds to the
// nearest valid CR per DMG p. 275. On a tie, rounds UP
// (DMG p. 275 literal example: FinalCR(2, 3) = 3).
//
//   - FinalCR(2, 3) == 3   (DMG p. 275 literal example)
//   - FinalCR(4, 4) == 4
//   - FinalCR(1, 1.5) == 1 (1.25 rounds down to 1)
func FinalCR(defensive, offensive float64) float64 {
	avg := (defensive + offensive) / 2
	return roundToValidCR(avg)
}

// shiftCR shifts a CR by ±1 valid-CR step.
func shiftCR(cr float64, delta int) float64 {
	idx := -1
	for i, v := range validCRsList {
		if v == cr {
			idx = i
			break
		}
	}
	if idx == -1 {
		cr = roundToValidCR(cr)
		for i, v := range validCRsList {
			if v == cr {
				idx = i
				break
			}
		}
	}
	if idx == -1 {
		return cr
	}
	newIdx := idx + delta
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(validCRsList) {
		newIdx = len(validCRsList) - 1
	}
	return validCRsList[newIdx]
}
