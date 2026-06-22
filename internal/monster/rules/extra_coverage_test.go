package rules

import "testing"

func TestDefensiveCRFromHP(t *testing.T) {
	t.Parallel()
	cases := []struct {
		hp   int
		want float64
	}{
		// CR 0 (1..6 HP).
		{1, 0},
		{6, 0},
		// CR 1/8 (7..35 HP).
		{7, 0.125},
		{35, 0.125},
		// CR 1/4 (36..49 HP).
		{36, 0.25},
		{49, 0.25},
		// CR 1/2 (50..70 HP).
		{50, 0.5},
		{70, 0.5},
		// CR 1 (71..85 HP).
		{71, 1},
		{85, 1},
		// CR 5 (131..145 HP).
		{131, 5},
		{145, 5},
		// CR 30 (806..850 HP).
		{806, 30},
		{850, 30},
		// Beyond table → CR 30.
		{100000, 30},
		// Negative HP → CR 0.
		{-1, 0},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(float64(c.hp)), func(t *testing.T) {
			t.Parallel()
			got := DefensiveCRFromHP(c.hp)
			if got != c.want {
				t.Errorf("DefensiveCRFromHP(%d) = %v, want %v", c.hp, got, c.want)
			}
		})
	}
}

func TestOffensiveCRFromDPR(t *testing.T) {
	t.Parallel()
	cases := []struct {
		dpr  float64
		want float64
	}{
		{0, 0},     // CR 0 (0..1)
		{1, 0},     // CR 0
		{2, 0.125}, // 1/8 (2..3)
		{3, 0.125},
		{4, 0.25},  // 1/4 (4..5)
		{5, 0.25},
		{9, 1},     // CR 1 (9..14)
		{14, 1},
		{15, 2},    // CR 2 (15..20)
		{33, 5},    // CR 5 (33..38)
		{38, 5},
		{303, 30},  // CR 30 (303..320)
		{320, 30},
		{1000, 30}, // way beyond table
		{-1, 0},    // negative
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.dpr), func(t *testing.T) {
			t.Parallel()
			got := OffensiveCRFromDPR(c.dpr)
			if got != c.want {
				t.Errorf("OffensiveCRFromDPR(%v) = %v, want %v", c.dpr, got, c.want)
			}
		})
	}
}

func TestGetStatsForCR_SnapToNearest(t *testing.T) {
	t.Parallel()
	// 0.1 is not in the table; should snap to 0.
	stats, err := GetStatsForCR(0.1)
	if err != nil {
		t.Fatalf("GetStatsForCR(0.1) returned error: %v", err)
	}
	if stats.PB != 2 {
		t.Errorf("PB = %d, want 2 (snapped to CR 0)", stats.PB)
	}
}

func TestPBForCR_OutOfRange(t *testing.T) {
	t.Parallel()
	// Out-of-range returns the minimum PB (2).
	if got := PBForCR(-1); got != 2 {
		t.Errorf("PBForCR(-1) = %d, want 2", got)
	}
	if got := PBForCR(31); got != 2 {
		t.Errorf("PBForCR(31) = %d, want 2", got)
	}
}

func TestShiftCR_Boundaries(t *testing.T) {
	t.Parallel()
	// CR 0 + 1 → 0.125.
	if got := shiftCR(0, +1); got != 0.125 {
		t.Errorf("shiftCR(0, +1) = %v, want 0.125", got)
	}
	// CR 0 - 1 → CR 0 (saturates).
	if got := shiftCR(0, -1); got != 0 {
		t.Errorf("shiftCR(0, -1) = %v, want 0", got)
	}
	// CR 30 + 1 → 30 (saturates).
	if got := shiftCR(30, +1); got != 30 {
		t.Errorf("shiftCR(30, +1) = %v, want 30", got)
	}
	// CR 30 - 1 → 29.
	if got := shiftCR(30, -1); got != 29 {
		t.Errorf("shiftCR(30, -1) = %v, want 29", got)
	}
	// CR 1 + 1 → 2.
	if got := shiftCR(1, +1); got != 2 {
		t.Errorf("shiftCR(1, +1) = %v, want 2", got)
	}
	// CR 1 - 1 → 0.5.
	if got := shiftCR(1, -1); got != 0.5 {
		t.Errorf("shiftCR(1, -1) = %v, want 0.5", got)
	}
	// Unknown CR snaps to nearest first, then shifts.
	if got := shiftCR(0.3, +1); got != 0.5 {
		t.Errorf("shiftCR(0.3, +1) = %v, want 0.5", got)
	}
}

func TestHPFromHitDice_UnknownDie(t *testing.T) {
	t.Parallel()
	// Unknown die size (d2, d3) → fallback to (size+1)/2.
	// d2 → avg 1.5; 1d2 = 1; +0 CON = 1.
	if got := HPFromHitDice(1, 2, 0); got != 1 {
		t.Errorf("HPFromHitDice(1, 2, 0) = %d, want 1", got)
	}
	// d3 → avg 2.0; 2d3 = 4.0; +1 CON = 2*1 = 2 → 4+2 = 6.
	if got := HPFromHitDice(2, 3, 1); got != 6 {
		t.Errorf("HPFromHitDice(2, 3, 1) = %d, want 6", got)
	}
}

func TestEffectiveHP_Stacked(t *testing.T) {
	t.Parallel()
	// Both resistance and immunity multipliers stack.
	// CR 6, 100 HP: resistance (×1.5) + immunity (×2) → (1.5-1)+(2-1) = 1.5 added → ×2.5.
	got := EffectiveHP(6, 100, true, true)
	want := 250
	if got != want {
		t.Errorf("EffectiveHP(6, 100, true, true) = %d, want %d", got, want)
	}
}

func TestHasRangedDamage_BonusAction(t *testing.T) {
	t.Parallel()
	m := &Monster{
		CR: 5,
		Speed: map[SpeedKind]int{SpeedFly: 60},
		Actions: []Action{
			{Description: "Melee Weapon Attack: +5 to hit, reach 5 ft."},
		},
		BonusActions: []Action{
			{Description: "Ranged Weapon Attack: +5 to hit, range 60 ft."},
		},
	}
	if !IsFlyingRangedUnderCR10(m) {
		t.Error("expected flying bonus on a monster with ranged bonus action")
	}
}

func TestHasRangedDamage_NoActions(t *testing.T) {
	t.Parallel()
	m := &Monster{
		CR:           5,
		Speed:        map[SpeedKind]int{SpeedFly: 60},
		Actions:      nil,
		BonusActions: nil,
	}
	if IsFlyingRangedUnderCR10(m) {
		t.Error("expected no flying bonus on a monster with no actions")
	}
}

func TestContainsWord(t *testing.T) {
	t.Parallel()
	cases := []struct {
		haystack, needle string
		want             bool
	}{
		{"ranged weapon", "ranged", true},
		{"range 80 ft.", "range ", true},
		{"ranged", "ranged", true},
		{"ranger", "ranged", false},
		{"", "ranged", false},
		{"ranged", "", true},
		{"abc", "z", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.haystack+"_"+c.needle, func(t *testing.T) {
			t.Parallel()
			if got := containsWord(c.haystack, c.needle); got != c.want {
				t.Errorf("containsWord(%q, %q) = %v, want %v", c.haystack, c.needle, got, c.want)
			}
		})
	}
}

func TestLower(t *testing.T) {
	t.Parallel()
	if got := lower("HELLO World"); got != "hello world" {
		t.Errorf("lower(\"HELLO World\") = %q, want hello world", got)
	}
	if got := lower(""); got != "" {
		t.Errorf("lower(\"\") = %q, want empty", got)
	}
}
