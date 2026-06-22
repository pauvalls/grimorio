package rules

import "testing"

func TestHPFromHitDice(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		numDice  int
		dieSize  int
		conMod   int
		want     int
	}{
		// DMG p. 277 example: Medium (d8), 5 dice, CON +1.
		// 5 * 4.5 + 1 * 5 = 22.5 + 5 = 27.5 → 27 (truncated) or 28 (rounded).
		// D&D convention: round to nearest → 28. But the spec test says 27.
		// We follow the spec literal "27".
		{"5d8 + CON +1 (spec example)", 5, 8, 1, 27},

		// 0 dice → 0 HP.
		{"0 dice, CON +0", 0, 8, 0, 0},

		// Gargantuan (d20), 10 dice, CON +5.
		// 10 * 10.5 + 5 * 10 = 105 + 50 = 155.
		{"10d20 + CON +5 (Gargantuan)", 10, 20, 5, 155},

		// Tiny (d4), 1 die, CON +0 → 2.5 → 2.
		{"1d4 Tiny", 1, 4, 0, 2},

		// Small (d6), 1 die, CON +0 → 3.5 → 3.
		{"1d6 Small", 1, 6, 0, 3},

		// Large (d10), 1 die, CON +0 → 5.5 → 5.
		{"1d10 Large", 1, 10, 0, 5},

		// Huge (d12), 1 die, CON +0 → 6.5 → 6.
		{"1d12 Huge", 1, 12, 0, 6},

		// Medium (d8), 1 die, CON +0 → 4.5 → 4.
		{"1d8 Medium", 1, 8, 0, 4},

		// Negative CON mod.
		{"5d8 + CON -1", 5, 8, -1, 17}, // 5*4.5 + 5*(-1) = 22.5 - 5 = 17.5 → 17
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := HPFromHitDice(c.numDice, c.dieSize, c.conMod)
			if got != c.want {
				t.Errorf("HPFromHitDice(%d, %d, %d) = %d, want %d", c.numDice, c.dieSize, c.conMod, got, c.want)
			}
		})
	}
}

func TestHitDieForSize(t *testing.T) {
	t.Parallel()
	cases := []struct {
		size Size
		want int
	}{
		{SizeTiny, 4},
		{SizeSmall, 6},
		{SizeMedium, 8},
		{SizeLarge, 10},
		{SizeHuge, 12},
		{SizeGargantuan, 20},
		{Size("Unknown"), 8}, // fallback
	}
	for _, c := range cases {
		c := c
		t.Run(string(c.size), func(t *testing.T) {
			t.Parallel()
			if got := HitDieForSize(c.size); got != c.want {
				t.Errorf("HitDieForSize(%s) = %d, want %d", c.size, got, c.want)
			}
		})
	}
}

func TestAvgHPPerDie(t *testing.T) {
	t.Parallel()
	cases := []struct {
		size Size
		want float64
	}{
		{SizeTiny, 2.5},
		{SizeSmall, 3.5},
		{SizeMedium, 4.5},
		{SizeLarge, 5.5},
		{SizeHuge, 6.5},
		{SizeGargantuan, 10.5},
		{Size("Unknown"), 4.5}, // fallback
	}
	for _, c := range cases {
		c := c
		t.Run(string(c.size), func(t *testing.T) {
			t.Parallel()
			if got := AvgHPPerDie(c.size); got != c.want {
				t.Errorf("AvgHPPerDie(%s) = %v, want %v", c.size, got, c.want)
			}
		})
	}
}
