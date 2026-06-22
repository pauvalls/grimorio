package rules

import (
	"math"
	"testing"
)

func TestStats_Modifier(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		score int
		want  int
	}{
		{"score 1 (min)", 1, -5},
		{"score 3 (very low)", 3, -4},
		{"score 8", 8, -1},
		{"score 9", 9, -1},
		{"score 10 (average)", 10, 0},
		{"score 11", 11, 0},
		{"score 12", 12, 1},
		{"score 13", 13, 1},
		{"score 14", 14, 2},
		{"score 18", 18, 4},
		{"score 20 (max normal)", 20, 5},
		{"score 30 (max legendary)", 30, 10},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := Stats{}.Modifier(c.score)
			if got != c.want {
				t.Errorf("Modifier(%d) = %d, want %d", c.score, got, c.want)
			}
		})
	}
}

func TestStats_Abilities(t *testing.T) {
	t.Parallel()
	stats := Stats{STR: 18, DEX: 14, CON: 16, INT: 8, WIS: 12, CHA: 10}
	if got := stats.Modifier(stats.STR); got != 4 {
		t.Errorf("STR modifier = %d, want 4", got)
	}
	if got := stats.Modifier(stats.DEX); got != 2 {
		t.Errorf("DEX modifier = %d, want 2", got)
	}
	if got := stats.Modifier(stats.CON); got != 3 {
		t.Errorf("CON modifier = %d, want 3", got)
	}
	if got := stats.Modifier(stats.INT); got != -1 {
		t.Errorf("INT modifier = %d, want -1", got)
	}
	if got := stats.Modifier(stats.WIS); got != 1 {
		t.Errorf("WIS modifier = %d, want 1", got)
	}
	if got := stats.Modifier(stats.CHA); got != 0 {
		t.Errorf("CHA modifier = %d, want 0", got)
	}
}

func TestParseCR(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"0", 0, false},
		{"1/8", 0.125, false},
		{"1/4", 0.25, false},
		{"1/2", 0.5, false},
		{"1", 1, false},
		{"2", 2, false},
		{"13", 13, false},
		{"30", 30, false},
		{"", 0, true},
		{"-1", 0, true},
		{"31", 0, true},
		{"abc", 0, true},
		{"1/3", 0, true},   // invalid denominator
		{"1/0", 0, true},   // zero denominator
		{"1/16", 0, true},  // too small
		{"100", 0, true},   // out of range
		{"0.5", 0.5, false}, // canonical float form accepted
	}
	for _, c := range cases {
		c := c
		t.Run(c.input, func(t *testing.T) {
			t.Parallel()
			got, err := ParseCR(c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("ParseCR(%q) = %v, want error", c.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCR(%q) returned error: %v", c.input, err)
			}
			if math.Abs(got-c.want) > 1e-9 {
				t.Errorf("ParseCR(%q) = %v, want %v", c.input, got, c.want)
			}
		})
	}
}

func TestSizeConstants(t *testing.T) {
	t.Parallel()
	want := map[Size]string{
		SizeTiny: "Tiny", SizeSmall: "Small", SizeMedium: "Medium",
		SizeLarge: "Large", SizeHuge: "Huge", SizeGargantuan: "Gargantuan",
	}
	for size, label := range want {
		if string(size) != label {
			t.Errorf("size %q, want %q", string(size), label)
		}
	}
	if len(want) != 6 {
		t.Errorf("expected 6 sizes, got %d", len(want))
	}
}

func TestCreatureTypeConstants(t *testing.T) {
	t.Parallel()
	want := map[CreatureType]string{
		TypeAberration: "aberration", TypeBeast: "beast", TypeCelestial: "celestial",
		TypeConstruct: "construct", TypeDragon: "dragon", TypeElemental: "elemental",
		TypeFey: "fey", TypeFiend: "fiend", TypeGiant: "giant", TypeHumanoid: "humanoid",
		TypeMonstrosity: "monstrosity", TypeOoze: "ooze", TypePlant: "plant", TypeUndead: "undead",
	}
	if len(want) != 14 {
		t.Errorf("expected 14 creature types, got %d", len(want))
	}
	for ct, label := range want {
		if string(ct) != label {
			t.Errorf("creature type %q, want %q", string(ct), label)
		}
	}
}

func TestDamageTypeConstants(t *testing.T) {
	t.Parallel()
	want := map[DamageType]string{
		DmgAcid: "acid", DmgBludgeon: "bludgeoning", DmgCold: "cold",
		DmgFire: "fire", DmgForce: "force", DmgLightning: "lightning",
		DmgNecrotic: "necrotic", DmgPierce: "piercing", DmgPoison: "poison",
		DmgPsychic: "psychic", DmgRadiant: "radiant", DmgSlash: "slashing", DmgThunder: "thunder",
	}
	if len(want) != 13 {
		t.Errorf("expected 13 damage types, got %d", len(want))
	}
	for dt, label := range want {
		if string(dt) != label {
			t.Errorf("damage type %q, want %q", string(dt), label)
		}
	}
}

func TestAlignmentConstants(t *testing.T) {
	t.Parallel()
	all := []Alignment{
		AlignLG, AlignNG, AlignCG,
		AlignLN, AlignTN, AlignCN,
		AlignLE, AlignNE, AlignCE,
		AlignU,
	}
	if len(all) != 10 {
		t.Errorf("expected 10 alignments, got %d", len(all))
	}
	if string(AlignU) != "unaligned" {
		t.Errorf("AlignU = %q, want unaligned", string(AlignU))
	}
}

func TestSpeedKindConstants(t *testing.T) {
	t.Parallel()
	all := []SpeedKind{SpeedWalk, SpeedFly, SpeedSwim, SpeedBurrow, SpeedClimb}
	if len(all) != 5 {
		t.Errorf("expected 5 speed kinds, got %d", len(all))
	}
}
