package rules

import (
	"testing"
)

// TestXPTable_AllRows asserts the XP-for-CR table for every CR row.
// Source: DMG p. 282 / MM 2025 pp. 214-250 / SRD 5.1.
func TestXPTable_AllRows(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr   float64
		want int
	}{
		{0, 10},     // CR 0 with stat block (DMG p. 282)
		{0.125, 25}, // 1/8
		{0.25, 50},  // 1/4
		{0.5, 100},  // 1/2
		{1, 200},
		{2, 450},
		{3, 700},
		{4, 1100},
		{5, 1800},
		{6, 2300},
		{7, 2900},
		{8, 3900},
		{9, 5000},
		{10, 5900},
		{11, 7200},
		{12, 8400},
		{13, 10000},
		{14, 11500},
		{15, 13000},
		{16, 15000},
		{17, 18000},
		{18, 20000},
		{19, 22000},
		{20, 25000},
		{21, 33000},
		{22, 41000},
		{23, 50000},
		{24, 62000},
		{25, 75000},
		{26, 90000},
		{27, 105000},
		{28, 120000},
		{29, 135000},
		{30, 155000},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.cr), func(t *testing.T) {
			t.Parallel()
			if got := XPForCR(c.cr); got != c.want {
				t.Errorf("XPForCR(%v) = %d, want %d", c.cr, got, c.want)
			}
		})
	}
}

func TestXPForCR_OutOfRange(t *testing.T) {
	t.Parallel()
	if got := XPForCR(-1); got != 0 {
		t.Errorf("XPForCR(-1) = %d, want 0", got)
	}
	if got := XPForCR(31); got != 0 {
		t.Errorf("XPForCR(31) = %d, want 0", got)
	}
}

// TestPBTable_AllBands asserts PB for the 8 CR bands.
func TestPBTable_AllBands(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr   float64
		want int
	}{
		{0, 2}, {1, 2}, {2, 2}, {3, 2}, {4, 2}, // 0-4
		{5, 3}, {6, 3}, {7, 3}, {8, 3}, // 5-8
		{9, 4}, {10, 4}, {11, 4}, {12, 4}, // 9-12
		{13, 5}, {14, 5}, {15, 5}, {16, 5}, // 13-16
		{17, 6}, {18, 6}, {19, 6}, {20, 6}, // 17-20
		{21, 7}, {22, 7}, {23, 7}, {24, 7}, // 21-24
		{25, 8}, {26, 8}, {27, 8}, {28, 8}, // 25-28
		{29, 9}, {30, 9}, // 29-30
		{0.125, 2}, {0.25, 2}, {0.5, 2}, // sub-integer in band 0-4
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.cr), func(t *testing.T) {
			t.Parallel()
			if got := PBForCR(c.cr); got != c.want {
				t.Errorf("PBForCR(%v) = %d, want %d", c.cr, got, c.want)
			}
		})
	}
}

// TestHitDiceBySize asserts the hit die and average HP/die for every size.
func TestHitDiceBySize(t *testing.T) {
	t.Parallel()
	cases := []struct {
		size   Size
		die    int
		avgHPD float64
	}{
		{SizeTiny, 4, 2.5},
		{SizeSmall, 6, 3.5},
		{SizeMedium, 8, 4.5},
		{SizeLarge, 10, 5.5},
		{SizeHuge, 12, 6.5},
		{SizeGargantuan, 20, 10.5},
	}
	for _, c := range cases {
		c := c
		t.Run(string(c.size), func(t *testing.T) {
			t.Parallel()
			if got := HitDieForSize(c.size); got != c.die {
				t.Errorf("HitDieForSize(%s) = %d, want %d", c.size, got, c.die)
			}
			if got := AvgHPPerDie(c.size); got != c.avgHPD {
				t.Errorf("AvgHPPerDie(%s) = %v, want %v", c.size, got, c.avgHPD)
			}
		})
	}
}

// TestCRMasterTable_AllRows asserts every column for every CR row.
// This is the canonical 31-row table from DMG p. 274.
func TestCRMasterTable_AllRows(t *testing.T) {
	t.Parallel()
	type row struct {
		cr   float64
		pb   int
		ac   int
		hpLo int
		hpHi int
		atk  int
		dprLo float64
		dprHi float64
		dc   int
	}
	rows := []row{
		{0, 2, 13, 1, 6, 3, 0, 1, 13},
		{0.125, 2, 13, 7, 35, 3, 2, 3, 13},
		{0.25, 2, 13, 36, 49, 3, 4, 5, 13},
		{0.5, 2, 13, 50, 70, 3, 6, 8, 13},
		{1, 2, 13, 71, 85, 3, 9, 14, 13},
		{2, 2, 13, 86, 100, 3, 15, 20, 13},
		{3, 2, 13, 101, 115, 4, 21, 26, 13},
		{4, 2, 14, 116, 130, 5, 27, 32, 14},
		{5, 3, 15, 131, 145, 6, 33, 38, 15},
		{6, 3, 15, 146, 160, 6, 39, 44, 15},
		{7, 3, 15, 161, 175, 6, 45, 50, 15},
		{8, 3, 16, 176, 190, 7, 51, 56, 16},
		{9, 4, 16, 191, 205, 7, 57, 62, 16},
		{10, 4, 17, 206, 220, 7, 63, 68, 16},
		{11, 4, 17, 221, 235, 8, 69, 74, 17},
		{12, 4, 17, 236, 250, 8, 75, 80, 17},
		{13, 5, 18, 251, 265, 8, 81, 86, 18},
		{14, 5, 18, 266, 280, 8, 87, 92, 18},
		{15, 5, 18, 281, 295, 8, 93, 98, 18},
		{16, 5, 18, 296, 310, 9, 99, 104, 18},
		{17, 6, 19, 311, 325, 10, 105, 110, 19},
		{18, 6, 19, 326, 340, 10, 111, 116, 19},
		{19, 6, 19, 341, 355, 10, 117, 122, 19},
		{20, 6, 19, 356, 400, 10, 123, 140, 19},
		{21, 7, 19, 401, 445, 11, 141, 158, 20},
		{22, 7, 19, 446, 490, 11, 159, 176, 20},
		{23, 7, 19, 491, 535, 11, 177, 194, 20},
		{24, 7, 19, 536, 580, 12, 195, 212, 21},
		{25, 8, 19, 581, 625, 12, 213, 230, 21},
		{26, 8, 19, 626, 670, 12, 231, 248, 21},
		{27, 8, 19, 671, 715, 13, 249, 266, 22},
		{28, 8, 19, 716, 760, 13, 267, 284, 22},
		{29, 9, 19, 761, 805, 13, 285, 302, 22},
		{30, 9, 19, 806, 850, 14, 303, 320, 23},
	}
	if len(rows) != 34 {
		t.Fatalf("expected 34 CR rows (0, 1/8, 1/4, 1/2, 1..30), got %d", len(rows))
	}
	for _, r := range rows {
		r := r
		t.Run(formatCR(r.cr), func(t *testing.T) {
			t.Parallel()
			stats, err := GetStatsForCR(r.cr)
			if err != nil {
				t.Fatalf("GetStatsForCR(%v) returned error: %v", r.cr, err)
			}
			if stats.PB != r.pb {
				t.Errorf("PB = %d, want %d", stats.PB, r.pb)
			}
			if stats.AC != r.ac {
				t.Errorf("AC = %d, want %d", stats.AC, r.ac)
			}
			if stats.HPMin != r.hpLo {
				t.Errorf("HPMin = %d, want %d", stats.HPMin, r.hpLo)
			}
			if stats.HPMax != r.hpHi {
				t.Errorf("HPMax = %d, want %d", stats.HPMax, r.hpHi)
			}
			if stats.AttackBonus != r.atk {
				t.Errorf("AttackBonus = %d, want %d", stats.AttackBonus, r.atk)
			}
			if stats.DPRMin != r.dprLo {
				t.Errorf("DPRMin = %v, want %v", stats.DPRMin, r.dprLo)
			}
			if stats.DPRMax != r.dprHi {
				t.Errorf("DPRMax = %v, want %v", stats.DPRMax, r.dprHi)
			}
			if stats.SaveDC != r.dc {
				t.Errorf("SaveDC = %d, want %d", stats.SaveDC, r.dc)
			}
		})
	}
}

func TestGetStatsForCR_OutOfRange(t *testing.T) {
	t.Parallel()
	if _, err := GetStatsForCR(-1); err == nil {
		t.Error("GetStatsForCR(-1) returned no error, want error")
	}
	if _, err := GetStatsForCR(31); err == nil {
		t.Error("GetStatsForCR(31) returned no error, want error")
	}
}

// TestHPMultipliers asserts the Effective HP table (DMG p. 278) for all 4
// CR bands and both multiplier types.
func TestHPMultipliers(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr              float64
		wantResistances float64
		wantImmunities  float64
	}{
		{1, 2, 2},   // CR 1-4 band
		{4, 2, 2},
		{5, 1.5, 2}, // CR 5-10 band
		{6, 1.5, 2},
		{10, 1.5, 2},
		{11, 1.25, 1.5}, // CR 11-16 band
		{16, 1.25, 1.5},
		{17, 1, 1.25}, // CR 17+ band
		{20, 1, 1.25},
		{30, 1, 1.25},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.cr), func(t *testing.T) {
			t.Parallel()
			if got := HPMultiplierForResistances(c.cr); got != c.wantResistances {
				t.Errorf("HPMultiplierForResistances(%v) = %v, want %v", c.cr, got, c.wantResistances)
			}
			if got := HPMultiplierForImmunities(c.cr); got != c.wantImmunities {
				t.Errorf("HPMultiplierForImmunities(%v) = %v, want %v", c.cr, got, c.wantImmunities)
			}
		})
	}
}

// formatCR returns a stable test name regardless of sub-integer form.
func formatCR(cr float64) string {
	switch cr {
	case 0.125:
		return "1/8"
	case 0.25:
		return "1/4"
	case 0.5:
		return "1/2"
	}
	if cr == float64(int(cr)) {
		return intToStr(int(cr))
	}
	return floatToStr(cr)
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

func floatToStr(f float64) string {
	return intToStr(int(f*1000)) + "/1000"
}
