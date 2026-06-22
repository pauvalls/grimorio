package rules

import "testing"

func TestFeatureFor_PackTactics(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Pack Tactics")
	if !ok {
		t.Fatal("FeatureFor(\"Pack Tactics\") returned ok=false")
	}
	if f.Type != "attack" {
		t.Errorf("Type = %q, want attack", f.Type)
	}
	if f.Amount != 1 {
		t.Errorf("Amount = %d, want 1", f.Amount)
	}
}

func TestFeatureFor_MagicResistance(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Magic Resistance")
	if !ok {
		t.Fatal("FeatureFor(\"Magic Resistance\") returned ok=false")
	}
	if f.Type != "ac" {
		t.Errorf("Type = %q, want ac", f.Type)
	}
	if f.Amount != 2 {
		t.Errorf("Amount = %d, want 2", f.Amount)
	}
}

func TestFeatureFor_Aggressive(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Aggressive")
	if !ok {
		t.Fatal("FeatureFor(\"Aggressive\") returned ok=false")
	}
	if f.Type != "damage" && f.Type != "dpr" {
		t.Errorf("Type = %q, want damage or dpr", f.Type)
	}
	if f.Amount != 2 {
		t.Errorf("Amount = %d, want 2", f.Amount)
	}
}

func TestFeatureFor_UndeadFortitude_CR4Bands(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr   float64
		want int
	}{
		{1, 7},   // CR 1-4 band
		{3, 7},
		{4, 7},
		{5, 14},  // CR 5-10 band
		{8, 14},
		{10, 14},
		{11, 21}, // CR 11-16 band
		{14, 21},
		{16, 21},
		{17, 28}, // CR 17+ band
		{25, 28},
		{30, 28},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.cr), func(t *testing.T) {
			t.Parallel()
			f, ok := FeatureFor("Undead Fortitude")
			if !ok {
				t.Fatal("FeatureFor(\"Undead Fortitude\") returned ok=false")
			}
			if f.HPRule == nil {
				t.Fatal("Undead Fortitude has no HPRule")
			}
			got := f.HPRule(c.cr)
			if got != c.want {
				t.Errorf("Undead Fortitude HP for CR %v = %d, want %d", c.cr, got, c.want)
			}
		})
	}
}

func TestFeatureFor_LegendaryResistance_CR3Bands(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr   float64
		want int
	}{
		{1, 10},  // CR 1-4
		{4, 10},
		{5, 20},  // CR 5-10
		{10, 20},
		{11, 30}, // CR 11+
		{30, 30},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.cr), func(t *testing.T) {
			t.Parallel()
			f, ok := FeatureFor("Legendary Resistance")
			if !ok {
				t.Fatal("FeatureFor(\"Legendary Resistance\") returned ok=false")
			}
			got := f.HPRule(c.cr)
			if got != c.want {
				t.Errorf("Legendary Resistance HP for CR %v = %d, want %d", c.cr, got, c.want)
			}
		})
	}
}

func TestFeatureFor_Relentless_CR4Bands(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cr   float64
		want int
	}{
		{1, 7},
		{4, 7},
		{5, 14},
		{10, 14},
		{11, 21},
		{16, 21},
		{17, 28},
		{30, 28},
	}
	for _, c := range cases {
		c := c
		t.Run(formatCR(c.cr), func(t *testing.T) {
			t.Parallel()
			f, ok := FeatureFor("Relentless")
			if !ok {
				t.Fatal("FeatureFor(\"Relentless\") returned ok=false")
			}
			got := f.HPRule(c.cr)
			if got != c.want {
				t.Errorf("Relentless HP for CR %v = %d, want %d", c.cr, got, c.want)
			}
		})
	}
}

func TestFeatureFor_NimbleEscape(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Nimble Escape")
	if !ok {
		t.Fatal("FeatureFor(\"Nimble Escape\") returned ok=false")
	}
	// Nimble Escape: +4 effective AC and +4 effective attack bonus.
	if f.Type != "ac_attack" {
		t.Errorf("Type = %q, want ac_attack (Nimble Escape hits both AC and attack)", f.Type)
	}
	if f.Amount != 4 {
		t.Errorf("Amount = %d, want 4", f.Amount)
	}
}

func TestFeatureFor_Constrict(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Constrict")
	if !ok {
		t.Fatal("FeatureFor(\"Constrict\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 1 {
		t.Errorf("Constrict = %+v, want {ac, 1}", f)
	}
}

func TestFeatureFor_BloodFrenzy(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Blood Frenzy")
	if !ok {
		t.Fatal("FeatureFor(\"Blood Frenzy\") returned ok=false")
	}
	if f.Type != "attack" || f.Amount != 4 {
		t.Errorf("Blood Frenzy = %+v, want {attack, 4}", f)
	}
}

func TestFeatureFor_Ambusher(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Ambusher")
	if !ok {
		t.Fatal("FeatureFor(\"Ambusher\") returned ok=false")
	}
	if f.Type != "attack" || f.Amount != 1 {
		t.Errorf("Ambusher = %+v, want {attack, 1}", f)
	}
}

func TestFeatureFor_Avoidance(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Avoidance")
	if !ok {
		t.Fatal("FeatureFor(\"Avoidance\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 1 {
		t.Errorf("Avoidance = %+v, want {ac, 1}", f)
	}
}

func TestFeatureFor_Parry(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Parry")
	if !ok {
		t.Fatal("FeatureFor(\"Parry\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 1 {
		t.Errorf("Parry = %+v, want {ac, 1}", f)
	}
}

func TestFeatureFor_Stench(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Stench")
	if !ok {
		t.Fatal("FeatureFor(\"Stench\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 1 {
		t.Errorf("Stench = %+v, want {ac, 1}", f)
	}
}

func TestFeatureFor_Web(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Web")
	if !ok {
		t.Fatal("FeatureFor(\"Web\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 1 {
		t.Errorf("Web = %+v, want {ac, 1}", f)
	}
}

func TestFeatureFor_ShadowStealth(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Shadow Stealth")
	if !ok {
		t.Fatal("FeatureFor(\"Shadow Stealth\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 4 {
		t.Errorf("Shadow Stealth = %+v, want {ac, 4}", f)
	}
}

func TestFeatureFor_SuperiorInvisibility(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Superior Invisibility")
	if !ok {
		t.Fatal("FeatureFor(\"Superior Invisibility\") returned ok=false")
	}
	if f.Type != "ac" || f.Amount != 2 {
		t.Errorf("Superior Invisibility = %+v, want {ac, 2}", f)
	}
}

func TestFeatureFor_Rampage(t *testing.T) {
	t.Parallel()
	f, ok := FeatureFor("Rampage")
	if !ok {
		t.Fatal("FeatureFor(\"Rampage\") returned ok=false")
	}
	if (f.Type != "damage" && f.Type != "dpr") || f.Amount != 2 {
		t.Errorf("Rampage = %+v, want {damage, 2}", f)
	}
}

func TestFeatureFor_Unknown(t *testing.T) {
	t.Parallel()
	_, ok := FeatureFor("ThisFeatureDoesNotExist")
	if ok {
		t.Error("FeatureFor(unknown) returned ok=true, want false")
	}
}

func TestFeatureFor_DashOnly(t *testing.T) {
	t.Parallel()
	// Features marked with "—" in the DMG table should return ok=false.
	// Test a handful of known dash-only features.
	for _, name := range []string{"Amorphous", "Charm", "Etherealness", "Keen Senses", "Mimicry", "Sunlight Sensitivity"} {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, ok := FeatureFor(name)
			if ok {
				t.Errorf("FeatureFor(%q) returned ok=true, want false (dash-only feature)", name)
			}
		})
	}
}

func TestAllFeatures(t *testing.T) {
	t.Parallel()
	all := AllFeatures()
	if len(all) < 15 {
		t.Errorf("AllFeatures() returned %d features, want at least 15", len(all))
	}
	// Every feature returned must be retrievable via FeatureFor.
	for _, name := range all {
		_, ok := FeatureFor(name)
		if !ok {
			t.Errorf("AllFeatures contains %q, but FeatureFor(%q) returned ok=false", name, name)
		}
	}
}
