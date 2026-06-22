// Package renderer emits a *rules.Monster as a 2025 MM-compliant markdown
// stat block. It follows the spec rules:
//
//   - Initiative is emitted as "+X (+Y)" (mod + score).
//   - Damage entries are emitted EITHER as a number OR a die expression,
//     never both.
//   - Empty sections (Bonus Actions, Reactions, Legendary Actions) are
//     omitted rather than rendered as "None".
//   - "Senses None" is omitted.
//
// The renderer is the round-trip partner of the parser package: for any
// monster m, ParseStatBlock(RenderStatBlock(m)) produces an equivalent
// monster (modulo whitespace and section ordering).
package renderer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pauvalls/grimorio/internal/monster/rules"
)

// RenderStatBlock renders a *rules.Monster in 2025 MM format.
func RenderStatBlock(m *rules.Monster) (string, error) {
	if m == nil {
		return "", fmt.Errorf("RenderStatBlock: nil monster")
	}
	var b strings.Builder
	writeHeader(&b, m)
	writeCombatHighlights(&b, m)
	writeAbilityScores(&b, m)
	writeOtherDetails(&b, m)
	writeTraits(&b, m)
	writeActions(&b, m, "### Actions", m.Actions)
	writeActions(&b, m, "### Bonus Actions", m.BonusActions)
	writeReactions(&b, m)
	writeLegendary(&b, m)
	return b.String(), nil
}

func writeHeader(b *strings.Builder, m *rules.Monster) {
	fmt.Fprintf(b, "## %s\n\n", m.Name)
	// Italic line: Size Type (tags), alignment.
	tags := ""
	if len(m.Tags) > 0 {
		tags = " (" + strings.Join(m.Tags, ", ") + ")"
	}
	fmt.Fprintf(b, "*%s %s%s, %s*\n\n", m.Size, m.Type, tags, m.Alignment)
}

func writeCombatHighlights(b *strings.Builder, m *rules.Monster) {
	// Initiative (2025: mod + score).
	if m.Initiative != 0 || m.InitScore != 0 {
		if m.InitScore == 0 {
			m.InitScore = 10 + m.Initiative
		}
		fmt.Fprintf(b, "**Initiative** %+d (%+d) (Dexterity)\n\n", m.Initiative, m.InitScore)
	}
	// Armor Class.
	if m.ACSource != "" {
		fmt.Fprintf(b, "**Armor Class** %d (%s)\n", m.AC, m.ACSource)
	} else if m.AC > 0 {
		fmt.Fprintf(b, "**Armor Class** %d\n", m.AC)
	}
	// Hit Points.
	if m.HP > 0 {
		if m.HPDice != "" {
			fmt.Fprintf(b, "**Hit Points** %d (%s)\n", m.HP, m.HPDice)
		} else {
			fmt.Fprintf(b, "**Hit Points** %d\n", m.HP)
		}
	}
	// Speed.
	if len(m.Speed) > 0 {
		parts := speedParts(m.Speed)
		fmt.Fprintf(b, "**Speed** %s\n", strings.Join(parts, ", "))
	}
	fmt.Fprint(b, "\n")
}

func speedParts(s map[rules.SpeedKind]int) []string {
	keys := make([]rules.SpeedKind, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	// Stable order: walk first, then other kinds alphabetically.
	order := map[rules.SpeedKind]int{
		rules.SpeedWalk: 0, rules.SpeedFly: 1, rules.SpeedSwim: 2,
		rules.SpeedBurrow: 3, rules.SpeedClimb: 4,
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return order[keys[i]] < order[keys[j]]
	})
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := s[k]
		if v == 0 && k == rules.SpeedWalk {
			parts = append(parts, "0 ft.")
		} else if k == rules.SpeedWalk {
			parts = append(parts, fmt.Sprintf("%d ft.", v))
		} else {
			parts = append(parts, fmt.Sprintf("%s %d ft.", k, v))
		}
	}
	return parts
}

func writeAbilityScores(b *strings.Builder, m *rules.Monster) {
	a := m.Abilities
	if a.STR == 0 && a.DEX == 0 && a.CON == 0 && a.INT == 0 && a.WIS == 0 && a.CHA == 0 {
		return
	}
	fmt.Fprint(b, "|STR|DEX|CON|INT|WIS|CHA|\n")
	fmt.Fprint(b, "|---|---|---|---|---|---|\n")
	fmt.Fprintf(b, "|%d (%+d)|%d (%+d)|%d (%+d)|%d (%+d)|%d (%+d)|%d (%+d)|\n\n",
		a.STR, scoreMod(a.STR), a.DEX, scoreMod(a.DEX), a.CON, scoreMod(a.CON),
		a.INT, scoreMod(a.INT), a.WIS, scoreMod(a.WIS), a.CHA, scoreMod(a.CHA))
}

func scoreMod(score int) int {
	return rules.Stats{}.Modifier(score)
}

func writeOtherDetails(b *strings.Builder, m *rules.Monster) {
	if len(m.Saves) > 0 {
		fmt.Fprintf(b, "**Saving Throws** %s\n", strings.Join(m.Saves, ", "))
	}
	if len(m.Skills) > 0 {
		keys := make([]string, 0, len(m.Skills))
		for k := range m.Skills {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s %+d", k, m.Skills[k]))
		}
		fmt.Fprintf(b, "**Skills** %s\n", strings.Join(parts, ", "))
	}
	if len(m.DamageVulnerabilities) > 0 {
		fmt.Fprintf(b, "**Damage Vulnerabilities** %s\n", joinDamageTypes(m.DamageVulnerabilities))
	}
	if len(m.DamageResistances) > 0 {
		fmt.Fprintf(b, "**Damage Resistances** %s\n", joinDamageTypes(m.DamageResistances))
	}
	if len(m.DamageImmunities) > 0 {
		fmt.Fprintf(b, "**Damage Immunities** %s\n", joinDamageTypes(m.DamageImmunities))
	}
	if len(m.ConditionImmunities) > 0 {
		parts := make([]string, 0, len(m.ConditionImmunities))
		for _, c := range m.ConditionImmunities {
			parts = append(parts, string(c))
		}
		fmt.Fprintf(b, "**Condition Immunities** %s\n", strings.Join(parts, ", "))
	}
	if m.Senses.PassivePerception > 0 || len(m.Senses.Special) > 0 {
		parts := make([]string, 0, len(m.Senses.Special)+1)
		parts = append(parts, m.Senses.Special...)
		if m.Senses.PassivePerception > 0 {
			parts = append(parts, fmt.Sprintf("passive Perception %d", m.Senses.PassivePerception))
		}
		fmt.Fprintf(b, "**Senses** %s\n", strings.Join(parts, ", "))
	}
	if len(m.Languages) > 0 {
		fmt.Fprintf(b, "**Languages** %s\n", strings.Join(m.Languages, ", "))
	}
	if m.CR > 0 || m.XP > 0 {
		crStr := crToString(m.CR)
		if m.XP > 0 {
			fmt.Fprintf(b, "**Challenge** %s (%s XP)\n", crStr, commaInt(m.XP))
		} else {
			fmt.Fprintf(b, "**Challenge** %s\n", crStr)
		}
	}
	fmt.Fprint(b, "\n")
}

func writeTraits(b *strings.Builder, m *rules.Monster) {
	if len(m.Traits) == 0 {
		return
	}
	fmt.Fprint(b, "### Traits\n\n")
	for _, t := range m.Traits {
		fmt.Fprintf(b, "**%s.** %s\n\n", t.Name, t.Description)
	}
}

func writeActions(b *strings.Builder, _ *rules.Monster, header string, actions []rules.Action) {
	if len(actions) == 0 {
		return
	}
	// Normalise "## Actions" / "### Actions" — we always emit "### " for
	// the section header.
	normalized := strings.TrimSpace(strings.TrimLeft(header, "#"))
	fmt.Fprintf(b, "### %s\n\n", normalized)
	for _, a := range actions {
		fmt.Fprintf(b, "**%s.** %s\n\n", a.Name, a.Description)
	}
}

func writeReactions(b *strings.Builder, m *rules.Monster) {
	if len(m.Reactions) == 0 {
		return
	}
	fmt.Fprint(b, "### Reactions\n\n")
	for _, r := range m.Reactions {
		fmt.Fprintf(b, "**%s.** %s\n\n", r.Name, r.Description)
	}
}

func writeLegendary(b *strings.Builder, m *rules.Monster) {
	if m.Legendary == nil {
		return
	}
	fmt.Fprint(b, "### Legendary Actions\n\n")
	fmt.Fprintf(b, "The %s can take %d legendary actions, choosing from the options below.\n\n", m.Name, m.Legendary.Uses)
	for _, a := range m.Legendary.Actions {
		fmt.Fprintf(b, "**%s.** %s\n\n", a.Name, a.Description)
	}
}

func joinDamageTypes(types []rules.DamageType) string {
	parts := make([]string, 0, len(types))
	for _, t := range types {
		parts = append(parts, string(t))
	}
	return strings.Join(parts, ", ")
}

func crToString(cr float64) string {
	switch cr {
	case 0.125:
		return "1/8"
	case 0.25:
		return "1/4"
	case 0.5:
		return "1/2"
	}
	if cr == float64(int(cr)) {
		return fmt.Sprintf("%d", int(cr))
	}
	return fmt.Sprintf("%g", cr)
}

func commaInt(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	// Insert commas from the right.
	out := []byte(s)
	for i := len(out) - 3; i > 0; i -= 3 {
		out = append(out[:i+1], out[i:]...)
		out[i] = ','
	}
	return string(out)
}
