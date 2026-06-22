package renderer

import (
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/monster/rules"
	"github.com/pauvalls/grimorio/internal/monster/rules/parser"
)

func TestRenderStatBlock_AllSections(t *testing.T) {
	t.Parallel()
	src := `## Goblin

*Small humanoid (goblinoid), neutral evil*

**Armor Class** 15 (leather armor, shield)
**Hit Points** 7 (2d6 + 2)
**Speed** 30 ft.

|STR|DEX|CON|INT|WIS|CHA|
|---|---|---|---|---|---|
|8 (-1)|14 (+2)|10 (+0)|10 (+0)|8 (-1)|8 (-1)|

**Senses** darkvision 60 ft., passive Perception 9
**Languages** Common, Goblin
**Challenge** 1/4 (50 XP)

### Traits

**Nimble Escape.** The goblin can take the Disengage or Hide action as a bonus action on each of its turns.

### Actions

**Scimitar.** *Melee Weapon Attack:* +4 to hit, reach 5 ft., one target. *Hit:* 5 (1d6 + 2) slashing damage.
`
	m, err := parser.ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock returned error: %v", err)
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	// 8 sections (or critical subsections) must appear.
	wantSubs := []string{
		"## Goblin",
		"Small humanoid (goblinoid), neutral evil",
		"**Armor Class**",
		"**Hit Points**",
		"**Speed**",
		"**Senses**",
		"**Languages**",
		"**Challenge**",
		"### Traits",
		"### Actions",
		"Nimble Escape",
		"Scimitar",
	}
	for _, want := range wantSubs {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered output missing %q\n--- output ---\n%s", want, rendered)
		}
	}
}

func TestRenderStatBlock_Initiative2025(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{
		Name:       "Foo",
		Size:       rules.SizeMedium,
		Type:       rules.TypeHumanoid,
		Alignment:  rules.AlignNE,
		AC:         15,
		HP:         10,
		Speed:      map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		Initiative: 4,
		InitScore:  14,
		CR:         0.25,
		XP:         50,
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	// Per MM 2025 p. 9, Initiative is emitted as "+4 (+14)".
	if !strings.Contains(rendered, "+4 (+14)") {
		t.Errorf("rendered output missing '+4 (+14)' (2025 Initiative format)\n--- output ---\n%s", rendered)
	}
}

func TestRenderStatBlock_EmptySectionsOmitted(t *testing.T) {
	t.Parallel()
	// A monster with no Bonus Actions / Reactions / Legendary / Senses
	// should NOT emit "None" placeholders.
	m := &rules.Monster{
		Name:       "Commoner",
		Size:       rules.SizeMedium,
		Type:       rules.TypeHumanoid,
		Alignment:  rules.AlignTN,
		AC:         10,
		HP:         4,
		Speed:      map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		Initiative: 0,
		CR:         0,
		XP:         10,
		Senses:     rules.Senses{Special: []string{}},
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	if strings.Contains(rendered, "None") {
		t.Errorf("rendered output should not contain 'None' for empty sections\n--- output ---\n%s", rendered)
	}
	if strings.Contains(rendered, "### Bonus Actions") {
		t.Errorf("rendered output should omit ### Bonus Actions when empty\n--- output ---\n%s", rendered)
	}
	if strings.Contains(rendered, "### Reactions") {
		t.Errorf("rendered output should omit ### Reactions when empty\n--- output ---\n%s", rendered)
	}
	if strings.Contains(rendered, "### Legendary Actions") {
		t.Errorf("rendered output should omit ### Legendary Actions when empty\n--- output ---\n%s", rendered)
	}
}

func TestRenderStatBlock_RoundTrip(t *testing.T) {
	t.Parallel()
	src := `## Bugbear

*Medium humanoid (goblinoid), chaotic evil*

**Armor Class** 16 (hide armor, shield)
**Hit Points** 27 (5d8 + 5)
**Speed** 30 ft.

|STR|DEX|CON|INT|WIS|CHA|
|---|---|---|---|---|---|
|15 (+2)|14 (+2)|13 (+1)|8 (-1)|11 (+0)|9 (-1)|

**Skills** Stealth +6, Survival +3
**Senses** darkvision 60 ft., passive Perception 10
**Languages** Common, Goblin
**Challenge** 1 (200 XP)

### Traits

**Brute.** A melee weapon deals one extra die of its damage when the bugbear hits with it.

**Heart of Hruggek.** The bugbear has advantage on saving throws against being charmed, frightened, paralyzed, poisoned, stunned, or put to sleep.

### Actions

**Morningstar.** *Melee Weapon Attack:* +4 to hit, reach 5 ft., one target. *Hit:* 11 (2d8 + 2) piercing damage.

**Javelin.** *Melee or Ranged Weapon Attack:* +4 to hit, reach 5 ft. or range 30/120 ft., one target. *Hit:* 9 (2d6 + 2) piercing damage in melee or 5 (1d6 + 2) piercing damage at range.
`
	m1, err := parser.ParseStatBlock(src)
	if err != nil {
		t.Fatalf("ParseStatBlock #1 returned error: %v", err)
	}
	rendered, err := RenderStatBlock(m1)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	m2, err := parser.ParseStatBlock(rendered)
	if err != nil {
		t.Fatalf("ParseStatBlock #2 (round-trip) returned error: %v\n--- output ---\n%s", err, rendered)
	}
	// Compare key fields.
	if m2.Name != m1.Name {
		t.Errorf("Name: %q → %q", m1.Name, m2.Name)
	}
	if m2.Size != m1.Size {
		t.Errorf("Size: %q → %q", m1.Size, m2.Size)
	}
	if m2.Type != m1.Type {
		t.Errorf("Type: %q → %q", m1.Type, m2.Type)
	}
	if m2.AC != m1.AC {
		t.Errorf("AC: %d → %d", m1.AC, m2.AC)
	}
	if m2.HP != m1.HP {
		t.Errorf("HP: %d → %d", m1.HP, m2.HP)
	}
	if m2.CR != m1.CR {
		t.Errorf("CR: %v → %v", m1.CR, m2.CR)
	}
	if m2.XP != m1.XP {
		t.Errorf("XP: %d → %d", m1.XP, m2.XP)
	}
	if m2.Abilities != m1.Abilities {
		t.Errorf("Abilities: %+v → %+v", m1.Abilities, m2.Abilities)
	}
	if len(m2.Traits) != len(m1.Traits) {
		t.Errorf("Traits count: %d → %d", len(m1.Traits), len(m2.Traits))
	}
	if len(m2.Actions) != len(m1.Actions) {
		t.Errorf("Actions count: %d → %d", len(m1.Actions), len(m2.Actions))
	}
}

func TestRenderStatBlock_WithLegendary(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{
		Name:       "Ancient Black Dragon",
		Size:       rules.SizeGargantuan,
		Type:       rules.TypeDragon,
		Alignment:  rules.AlignCE,
		AC:         22,
		HP:         367,
		Speed:      map[rules.SpeedKind]int{rules.SpeedWalk: 40, rules.SpeedFly: 80, rules.SpeedSwim: 40},
		CR:         21,
		XP:         33000,
		Legendary: &rules.LegendaryGroup{
			Uses: 3,
			Actions: []rules.Action{
				{Name: "Detect", Description: "The dragon makes a Wisdom (Perception) check."},
				{Name: "Tail Attack", Description: "The dragon makes a tail attack."},
			},
		},
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	for _, want := range []string{"### Legendary Actions", "can take 3 legendary actions", "Detect", "Tail Attack"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered output missing %q", want)
		}
	}
}

func TestRenderStatBlock_WithReactions(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{
		Name:       "Foo",
		Size:       rules.SizeMedium,
		Type:       rules.TypeHumanoid,
		Alignment:  rules.AlignTN,
		AC:         12,
		HP:         10,
		Speed:      map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		CR:         0,
		XP:         10,
		Reactions: []rules.Reaction{
			{Action: rules.Action{Name: "Parry", Description: "Add +2 to AC against one attack."}},
		},
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	for _, want := range []string{"### Reactions", "Parry", "Add +2 to AC"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered output missing %q", want)
		}
	}
}

func TestRenderStatBlock_WithBonusActions(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{
		Name:        "Foo",
		Size:        rules.SizeMedium,
		Type:        rules.TypeHumanoid,
		Alignment:   rules.AlignTN,
		AC:          12,
		HP:          10,
		Speed:       map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		CR:          0,
		XP:          10,
		BonusActions: []rules.Action{{Name: "Hide", Description: "Hide as a bonus action."}},
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	if !strings.Contains(rendered, "### Bonus Actions") {
		t.Error("rendered output missing '### Bonus Actions'")
	}
}

func TestRenderStatBlock_WithAllResistances(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{
		Name:      "Foo",
		Size:      rules.SizeMedium,
		Type:      rules.TypeUndead,
		Alignment: rules.AlignLE,
		AC:        13,
		HP:        13,
		Speed:     map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		CR:        0.25,
		XP:        50,
		DamageVulnerabilities: []rules.DamageType{rules.DmgBludgeon},
		DamageResistances:     []rules.DamageType{rules.DmgPierce, rules.DmgSlash},
		DamageImmunities:      []rules.DamageType{rules.DmgPoison},
		ConditionImmunities:   []rules.Condition{rules.CondExhaustion, rules.CondPoisoned},
		Saves:                 []string{"Wis"},
		Skills:                map[string]int{"Perception": 2},
		Senses:                rules.Senses{PassivePerception: 9, Special: []string{"darkvision 60 ft."}},
		Languages:             []string{"Common"},
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	for _, want := range []string{
		"**Damage Vulnerabilities** bludgeoning",
		"**Damage Resistances** piercing, slashing",
		"**Damage Immunities** poison",
		"**Condition Immunities** exhaustion, poisoned",
		"**Saving Throws** Wis",
		"**Skills** Perception",
		"**Senses**",
		"**Languages** Common",
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("rendered output missing %q\n--- output ---\n%s", want, rendered)
		}
	}
}

func TestRenderStatBlock_CommaInt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{10, "10"},
		{999, "999"},
		{1000, "1,000"},
		{10000, "10,000"},
		{155000, "155,000"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.want, func(t *testing.T) {
			t.Parallel()
			if got := commaInt(c.in); got != c.want {
				t.Errorf("commaInt(%d) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestRenderStatBlock_NilMonster(t *testing.T) {
	t.Parallel()
	if _, err := RenderStatBlock(nil); err == nil {
		t.Error("RenderStatBlock(nil) returned no error, want error")
	}
}

func TestRenderStatBlock_NoSize(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{
		Name:      "Foo",
		Type:      rules.TypeHumanoid,
		Alignment: rules.AlignTN,
		AC:        10,
		HP:        5,
		Speed:     map[rules.SpeedKind]int{rules.SpeedWalk: 30},
		CR:        0,
		XP:        10,
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	if !strings.Contains(rendered, "## Foo") {
		t.Errorf("rendered output missing '## Foo'")
	}
}

func TestRenderStatBlock_SpeedZeroWalk(t *testing.T) {
	t.Parallel()
	// Beholder-style: walk 0, fly 20.
	m := &rules.Monster{
		Name:      "Beholder",
		Size:      rules.SizeLarge,
		Type:      rules.TypeAberration,
		Alignment: rules.AlignLE,
		AC:        18,
		HP:        180,
		Speed:     map[rules.SpeedKind]int{rules.SpeedWalk: 0, rules.SpeedFly: 20},
		CR:        13,
		XP:        10000,
	}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	if !strings.Contains(rendered, "0 ft.") {
		t.Errorf("rendered output missing '0 ft.' for Beholder walk speed\n--- output ---\n%s", rendered)
	}
	if !strings.Contains(rendered, "fly 20 ft.") {
		t.Errorf("rendered output missing 'fly 20 ft.'")
	}
}

func TestRenderStatBlock_CRSubInteger(t *testing.T) {
	t.Parallel()
	m := &rules.Monster{Name: "Foo", CR: 0.5, XP: 100}
	rendered, err := RenderStatBlock(m)
	if err != nil {
		t.Fatalf("RenderStatBlock returned error: %v", err)
	}
	if !strings.Contains(rendered, "1/2") {
		t.Errorf("rendered output missing '1/2' for CR 0.5\n--- output ---\n%s", rendered)
	}
}
