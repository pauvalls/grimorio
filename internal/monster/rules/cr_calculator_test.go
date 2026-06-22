package rules

import (
	"math"
	"testing"
)

func TestDefensiveCR(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		hp   int
		ac   int
		want float64
	}{
		// HP and AC both align to CR 4.
		{"CR 4 baseline", 130, 14, 4},
		// AC +2 → CR +1
		{"AC +2", 130, 16, 5},
		// AC -2 → CR -1
		{"AC -2", 130, 12, 3},
		// AC ±1 → no change
		{"AC +1", 130, 15, 4},
		{"AC -1", 130, 13, 4},
		// CR 0 (Commoner: 4 HP, 10 AC)
		{"CR 0 commoner", 4, 10, 0},
		// CR 5 baseline (138 HP, AC 15)
		{"CR 5 baseline", 138, 15, 5},
		// CR 17 baseline (Adult Red Dragon: 256 HP, AC 19)
		{"CR 17", 318, 19, 17},
		// CR 24 baseline (Ancient Red Dragon: 546 HP, AC 19 expected, real 22)
		{"CR 24 with AC +3", 558, 22, 25},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := DefensiveCR(c.hp, c.ac)
			if got != c.want {
				t.Errorf("DefensiveCR(hp=%d, ac=%d) = %v, want %v", c.hp, c.ac, got, c.want)
			}
		})
	}
}

func TestOffensiveCR(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		dpr    float64
		atk    int
		saveDC int
		want   float64
	}{
		// DPR 50 → CR 7 baseline
		{"DPR 50 baseline atk 6", 50, 6, 0, 7},
		// attack -2 → CR -1
		{"DPR 50 atk -2", 50, 4, 0, 6},
		// attack +2 → CR +1
		{"DPR 50 atk +8", 50, 8, 0, 8},
		// attack ±1 → no change
		{"DPR 50 atk +1", 50, 7, 0, 7},
		{"DPR 50 atk -1", 50, 5, 0, 7},
		// DC variant
		{"DPR 50 DC 15", 50, 0, 15, 7},
		{"DPR 50 DC +2 (17)", 50, 0, 17, 8},
		{"DPR 50 DC -2 (13)", 50, 0, 13, 6},
		// CR 24 (Ancient Red Dragon: DPR ~204, atk +13 vs expected +12 → diff 1 = no shift)
		{"CR 24 baseline", 204, 13, 0, 24},
		// CR 24 with atk +14 (diff 2 → +1 → CR 25)
		{"CR 24 with atk +2", 204, 14, 0, 25},
		// CR 0 (Commoner: no attack, DPR 0)
		{"CR 0 commoner", 0, 0, 0, 0},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var got float64
			if c.saveDC != 0 {
				got = OffensiveCRFromDC(c.dpr, c.saveDC)
			} else {
				got = OffensiveCR(c.dpr, c.atk)
			}
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestFinalCR(t *testing.T) {
	t.Parallel()
	cases := []struct {
		def, off, want float64
	}{
		// DMG p. 275 literal example.
		{2, 3, 3},
		{4, 4, 4},
		// 1.5 between CR 1 and CR 2 → round to CR 1 (closer).
		{1, 1.5, 1},
		{1.5, 1.5, 1.5}, // 1.5 is not a valid CR → round to nearest (CR 1 or CR 2)
		{5, 5, 5},
		{0, 0, 0},
		// 2.5 average → snap to CR 2 or 3, depends on round.
		{2, 3, 3},
	}
	for i, c := range cases {
		c := c
		t.Run(formatIndex(i), func(t *testing.T) {
			t.Parallel()
			got := FinalCR(c.def, c.off)
			// 1.5 is not a valid CR — it must round to one of {1, 2}.
			if c.def == 1.5 && c.off == 1.5 {
				if math.Abs(got-1) > 0.01 && math.Abs(got-2) > 0.01 {
					t.Errorf("FinalCR(1.5, 1.5) = %v, want 1 or 2", got)
				}
				return
			}
			if got != c.want {
				t.Errorf("FinalCR(%v, %v) = %v, want %v", c.def, c.off, got, c.want)
			}
		})
	}
}

func TestAdjustCRByAC_DeltaCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		base     float64
		ac       int
		expected float64
	}{
		// CR 5 expects AC 15.
		{5, 15, 5},  // match
		{5, 17, 6},  // +2 → +1
		{5, 13, 4},  // -2 → -1
		{5, 16, 5},  // +1 → no change
		{5, 14, 5},  // -1 → no change
		// CR 10 expects AC 17.
		{10, 19, 11}, // +2 → +1
		{10, 15, 9},  // -2 → -1
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.base)+"_AC"+intToStr(c.ac), func(t *testing.T) {
			t.Parallel()
			got := AdjustCRByAC(c.base, c.ac)
			if got != c.expected {
				t.Errorf("AdjustCRByAC(%v, %d) = %v, want %v", c.base, c.ac, got, c.expected)
			}
		})
	}
}

func TestAdjustCRByAttack_DeltaCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		base     float64
		atk      int
		expected float64
	}{
		// CR 5 expects +6.
		{5, 6, 5},
		{5, 8, 6},
		{5, 4, 4},
		{5, 7, 5}, // +1 → no change
		{5, 5, 5}, // -1 → no change
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.base), func(t *testing.T) {
			t.Parallel()
			got := AdjustCRByAttack(c.base, c.atk)
			if got != c.expected {
				t.Errorf("AdjustCRByAttack(%v, %d) = %v, want %v", c.base, c.atk, got, c.expected)
			}
		})
	}
}

func TestAdjustCRBySaveDC_DeltaCases(t *testing.T) {
	t.Parallel()
	cases := []struct {
		base     float64
		dc       int
		expected float64
	}{
		// CR 5 expects DC 15.
		{5, 15, 5},
		{5, 17, 6},
		{5, 13, 4},
		{5, 16, 5},
		{5, 14, 5},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.base), func(t *testing.T) {
			t.Parallel()
			got := AdjustCRBySaveDC(c.base, c.dc)
			if got != c.expected {
				t.Errorf("AdjustCRBySaveDC(%v, %d) = %v, want %v", c.base, c.dc, got, c.expected)
			}
		})
	}
}

func TestCRStabilityFromMonster(t *testing.T) {
	t.Parallel()
	// Goblin (CR 1/4, HP 7, AC 15).
	// HP 7 → CR 1/8 (band 7-35). Expected AC for CR 1/8 = 13. Real AC = 15 → +2 → CR 1/4. ✓
	goblin := DefensiveCR(7, 15)
	if goblin != 0.25 {
		t.Errorf("DefensiveCR for Goblin = %v, want 0.25", goblin)
	}
}

func formatIndex(i int) string {
	return "case" + intToStr(i)
}
