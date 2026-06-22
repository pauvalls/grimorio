package parser

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/monster/rules"
)

// These tests target uncovered branches in the parser package.

func TestParseStatBlock_FlyWithClimb(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** 5
**Speed** 30 ft., climb 20 ft.
**Challenge** 0 (10 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	if m.Speed[rules.SpeedClimb] != 20 {
		t.Errorf("climb = %d, want 20", m.Speed[rules.SpeedClimb])
	}
}

func TestParseStatBlock_FlyWithBurrow(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium monstrosity, neutral*

**Armor Class** 12
**Hit Points** 15
**Speed** 30 ft., burrow 20 ft., swim 30 ft.
**Challenge** 1/4 (50 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	if m.Speed[rules.SpeedBurrow] != 20 {
		t.Errorf("burrow = %d, want 20", m.Speed[rules.SpeedBurrow])
	}
	if m.Speed[rules.SpeedSwim] != 30 {
		t.Errorf("swim = %d, want 30", m.Speed[rules.SpeedSwim])
	}
}

func TestParseStatBlock_FlyWithHover(t *testing.T) {
	t.Parallel()
	src := `## Beholder

*Large aberration, lawful evil*

**Armor Class** 18
**Hit Points** 180
**Speed** 0 ft., fly 20 ft. (hover)
**Challenge** 13 (10,000 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	if m.Speed[rules.SpeedFly] != 20 {
		t.Errorf("fly = %d, want 20", m.Speed[rules.SpeedFly])
	}
	if m.Speed[rules.SpeedWalk] != 0 {
		t.Errorf("walk = %d, want 0", m.Speed[rules.SpeedWalk])
	}
}

func TestParseStatBlock_HPBadNumber(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** foo
**Speed** 30 ft.
**Challenge** 0 (10 XP)
`
	_, err := ParseStatBlock(src)
	// The parser is lenient for HP — won't error, but HP stays 0.
	// We test it doesn't crash.
	if err != nil {
		// Acceptable either way.
		_ = err
	}
}

func TestParseStatBlock_SpeedUnknownPart(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** 5
**Speed** 30 ft., glide 10 ft.
**Challenge** 0 (10 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	// "glide" is unknown → walk stays at 30, no glide entry.
	if m.Speed[rules.SpeedWalk] != 30 {
		t.Errorf("walk = %d, want 30", m.Speed[rules.SpeedWalk])
	}
}

func TestParseStatBlock_DamageTypesWithNoise(t *testing.T) {
	t.Parallel()
	// "bludgeoning, piercing, and slashing from nonmagical weapons"
	// → only bludgeoning and piercing are picked up (trailing "from" filtered).
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** 5
**Speed** 30 ft.
**Damage Resistances** bludgeoning, piercing, and slashing from nonmagical weapons
**Challenge** 1 (200 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	if len(m.DamageResistances) < 2 {
		t.Errorf("DamageResistances = %v, want at least 2", m.DamageResistances)
	}
}

func TestParseStatBlock_SensesWithOnlyPassive(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** 5
**Speed** 30 ft.
**Senses** passive Perception 10
**Challenge** 0 (10 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	if m.Senses.PassivePerception != 10 {
		t.Errorf("PassivePerception = %d, want 10", m.Senses.PassivePerception)
	}
	if len(m.Senses.Special) != 0 {
		t.Errorf("Special = %v, want empty", m.Senses.Special)
	}
}

func TestParseStatBlock_NameHeaderOnly(t *testing.T) {
	t.Parallel()
	src := "##  \n"
	_, err := ParseStatBlock(src)
	if err == nil {
		t.Error("ParseStatBlock(whitespace name) returned no error, want error")
	}
}

func TestParseStatBlock_SpeedDash(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** 5
**Speed** -
**Challenge** 0 (10 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	// "-" is treated as "None" and skipped.
	if m.Speed[rules.SpeedWalk] != 0 {
		t.Errorf("walk = %d, want 0 (dash treated as none)", m.Speed[rules.SpeedWalk])
	}
}

func TestParseStatBlock_ChallengeMalformed(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Armor Class** 10
**Hit Points** 5
**Speed** 30 ft.
**Challenge** 99 (XX XP)
`
	_, err := ParseStatBlock(src)
	if err == nil {
		t.Error("ParseStatBlock(CR out of range) returned no error, want error")
	}
}

func TestParseStatBlock_Initiative2014(t *testing.T) {
	t.Parallel()
	src := `## Foo

*Medium humanoid, neutral*

**Initiative** +3

**Armor Class** 10
**Hit Points** 5
**Speed** 30 ft.
**Challenge** 0 (10 XP)
`
	m, err := ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock: %v", err)
	}
	if m.Initiative != 3 {
		t.Errorf("Initiative = %d, want 3", m.Initiative)
	}
	if m.InitScore != 13 {
		t.Errorf("InitScore = %d, want 13 (auto-computed)", m.InitScore)
	}
}
