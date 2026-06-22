package rules

import "testing"

func TestHPMultipliers_AllBands(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr              float64
		wantResistances float64
		wantImmunities  float64
	}{
		// Edge of each band.
		{1, 2, 2},     // 1-4 band
		{4, 2, 2},     // 1-4 band upper edge
		{5, 1.5, 2},   // 5-10 band
		{10, 1.5, 2},
		{11, 1.25, 1.5}, // 11-16 band
		{16, 1.25, 1.5},
		{17, 1, 1.25}, // 17+ band
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

func TestEffectiveHP(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name           string
		cr             float64
		hp             int
		hasResist      bool
		hasImmune      bool
		want           int
	}{
		// DMG p. 278 example: CR 6, 150 HP, B/P/S resistance → ×1.5 = 225.
		{"CR 6 150HP resist", 6, 150, true, false, 225},
		{"CR 6 150HP immune", 6, 150, false, true, 300}, // ×2
		{"CR 6 150HP no flags", 6, 150, false, false, 150},
		// 0 HP edge case.
		{"0 HP", 6, 0, true, false, 0},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := EffectiveHP(c.cr, c.hp, c.hasResist, c.hasImmune)
			if got != c.want {
				t.Errorf("EffectiveHP(%v, %d, %v, %v) = %d, want %d", c.cr, c.hp, c.hasResist, c.hasImmune, got, c.want)
			}
		})
	}
}

func TestIsFlyingRangedUnderCR10(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		m    *Monster
		want bool
	}{
		{
			name: "flying + ranged + CR 5 → true",
			m: &Monster{
				CR: 5,
				Speed: map[SpeedKind]int{SpeedWalk: 30, SpeedFly: 60},
				Actions: []Action{
					{Description: "Ranged Weapon Attack: +5 to hit, range 80/320 ft., one target."},
				},
			},
			want: true,
		},
		{
			name: "flying but melee → false",
			m: &Monster{
				CR: 5,
				Speed: map[SpeedKind]int{SpeedWalk: 30, SpeedFly: 60},
				Actions: []Action{
					{Description: "Melee Weapon Attack: +5 to hit, reach 5 ft."},
				},
			},
			want: false,
		},
		{
			name: "flying + ranged but CR 15 → false",
			m: &Monster{
				CR: 15,
				Speed: map[SpeedKind]int{SpeedFly: 80},
				Actions: []Action{
					{Description: "Ranged Weapon Attack: +8 to hit, range 120 ft."},
				},
			},
			want: false,
		},
		{
			name: "no fly → false",
			m: &Monster{
				CR: 5,
				Speed: map[SpeedKind]int{SpeedWalk: 30},
				Actions: []Action{
					{Description: "Ranged Weapon Attack: +5 to hit, range 80 ft."},
				},
			},
			want: false,
		},
		{
			name: "nil monster",
			m:    nil,
			want: false,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := IsFlyingRangedUnderCR10(c.m); got != c.want {
				t.Errorf("IsFlyingRangedUnderCR10(%s) = %v, want %v", c.name, got, c.want)
			}
		})
	}
}

func TestFlyingBonusAC(t *testing.T) {
	t.Parallel()
	m := &Monster{
		CR: 5,
		Speed: map[SpeedKind]int{SpeedFly: 60},
		Actions: []Action{
			{Description: "Ranged Weapon Attack: +5 to hit, range 80 ft."},
		},
	}
	if got := FlyingBonusAC(m); got != 2 {
		t.Errorf("FlyingBonusAC for CR 5 flying ranged = %d, want 2", got)
	}
	// CR 15 with the same setup → 0
	m.CR = 15
	if got := FlyingBonusAC(m); got != 0 {
		t.Errorf("FlyingBonusAC for CR 15 flying ranged = %d, want 0", got)
	}
}

func TestSTBonusACAdj(t *testing.T) {
	t.Parallel()
	cases := []struct {
		stBonuses int
		want      int
	}{
		{0, 0},
		{1, 0},
		{2, 0},
		{3, 2},
		{4, 2},
		{5, 4},
		{6, 4},
		{7, 4}, // capped
		{100, 4},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(float64(c.stBonuses)), func(t *testing.T) {
			t.Parallel()
			if got := STBonusACAdj(c.stBonuses); got != c.want {
				t.Errorf("STBonusACAdj(%d) = %d, want %d", c.stBonuses, got, c.want)
			}
		})
	}
}
