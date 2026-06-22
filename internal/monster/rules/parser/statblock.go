// Package parser converts markdown stat blocks (WotC 2014 MM and 2025 MM
// formats) into a *rules.Monster value.
//
// The parser is intentionally lenient: "None" entries are skipped, unknown
// fields are ignored, and partial blocks still produce a *Monster with as
// much data as could be extracted. The only fatal errors are syntactic
// ones (e.g. a non-numeric AC value, missing required header). All errors
// are returned as *ParseError, which carries the source line number for
// caller diagnostics.
package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pauvalls/grimorio/internal/monster/rules"
)

// ParseError carries a line-annotated parse error.
type ParseError struct {
	Line int
	Msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d: %s", e.Line, e.Msg)
}

// line is a (text, 1-based-line-number) tuple used during parsing.
type line struct {
	Text string
	N    int
}

// ParseStatBlock parses a markdown stat block into a *rules.Monster.
// Supports both 2014 MM (section names like "Armor Class", "Hit Points")
// and 2025 MM ("Combat Highlights" section + Initiative as "+X (+Y)").
func ParseStatBlock(markdown string) (*rules.Monster, error) {
	if strings.TrimSpace(markdown) == "" {
		return nil, &ParseError{Line: 1, Msg: "empty input"}
	}
	ls := splitLines(markdown)

	m := &rules.Monster{
		Speed:     map[rules.SpeedKind]int{},
		Skills:    map[string]int{},
		Senses:    rules.Senses{Special: []string{}},
		Languages: []string{},
		Tags:      []string{},
	}

	// 1. Find the H2 header (name).
	name, _, err := findName(ls)
	if err != nil {
		// Promote Line=0 to Line=1 for callers.
		if pe, ok := err.(*ParseError); ok && pe.Line == 0 {
			pe.Line = 1
		}
		return nil, err
	}
	m.Name = name

	// 2. Look for a size/type/alignment italic line under the name.
	parseSizeTypeAlignment(ls, m)

	// 3. Iterate field lines.
	for _, l := range ls {
		if err := parseFieldLine(l, m); err != nil {
			return nil, err
		}
	}

	// 4. Parse Traits/Actions/Bonus Actions/Reactions/Legendary bodies.
	parseSectionedBodies(ls, m)

	// 5. Compute missing Initiative score from DEX mod if 2025.
	if m.InitScore == 0 && m.Initiative != 0 {
		m.InitScore = 10 + m.Initiative
	}

	return m, nil
}

func splitLines(md string) []line {
	var out []line
	for i, raw := range strings.Split(md, "\n") {
		out = append(out, line{Text: raw, N: i + 1})
	}
	return out
}

// findName locates the first H2 header (## Name) and returns the name.
func findName(ls []line) (string, int, error) {
	for _, l := range ls {
		t := strings.TrimSpace(l.Text)
		if strings.HasPrefix(t, "## ") {
			name := strings.TrimSpace(strings.TrimPrefix(t, "## "))
			if name == "" {
				return "", l.N, &ParseError{Line: l.N, Msg: "empty H2 header"}
			}
			return name, l.N, nil
		}
	}
	return "", 0, &ParseError{Line: 0, Msg: "no H2 header found (expected '## Monster Name')"}
}

// parseSizeTypeAlignment parses the italic line below the name.
func parseSizeTypeAlignment(ls []line, m *rules.Monster) {
	for _, l := range ls {
		t := strings.TrimSpace(l.Text)
		if !strings.HasPrefix(t, "*") || !strings.HasSuffix(t, "*") {
			continue
		}
		// Skip action italic lines like "*Hit:* 5 (1d6+2) slashing damage."
		inner := strings.Trim(t, "*")
		if strings.Contains(inner, "Attack:") || strings.Contains(inner, "Hit:") || strings.Contains(inner, "Miss:") {
			continue
		}
		parts := strings.SplitN(inner, ",", 2)
		if len(parts) != 2 {
			continue
		}
		first := strings.TrimSpace(parts[0])
		firstWords := strings.Fields(first)
		if len(firstWords) < 2 {
			continue
		}
		m.Size = rules.Size(firstWords[0])
		typePart := strings.TrimSpace(first[len(firstWords[0]):])
		typeWords := strings.Fields(typePart)
		if len(typeWords) == 0 {
			continue
		}
		m.Type = rules.CreatureType(strings.ToLower(typeWords[0]))
		if idx := strings.Index(typePart, "("); idx != -1 {
			end := strings.Index(typePart, ")")
			if end > idx {
				tagStr := typePart[idx+1 : end]
				for _, tag := range strings.Split(tagStr, ",") {
					m.Tags = append(m.Tags, strings.TrimSpace(strings.ToLower(tag)))
				}
			}
		}
		m.Alignment = rules.Alignment(strings.ToLower(strings.TrimSpace(parts[1])))
		return
	}
}

var (
	acRE             = regexp.MustCompile(`(?i)^\*\*\s*Armor\s+Class\s*\*\*[:\s]*\s*(\d+)\s*(?:\(([^)]*)\))?`)
	hpRE             = regexp.MustCompile(`(?i)^\*\*\s*Hit\s+Points\s*\*\*[:\s]*\s*(\d+)\s*(?:\(([^)]*)\))?`)
	speedRE          = regexp.MustCompile(`(?i)^\*\*\s*Speed\s*\*\*[:\s]*\s*(.+?)\s*$`)
	init2025RE       = regexp.MustCompile(`(?i)^\*\*\s*Initiative\s*\*\*[:\s]*\s*([+-]?\d+)\s*\(([+-]?\d+)\)`)
	init2014RE       = regexp.MustCompile(`(?i)^\*\*\s*Initiative\s*\*\*[:\s]*\s*([+-]?\d+)\s*$`)
	abilityCellRE    = regexp.MustCompile(`(\d+)\s*\(([+-]?\d+)\)`)
	crRE             = regexp.MustCompile(`(?i)^\*\*\s*Challenge\s*\*\*[:\s]*\s*(\S+)\s*(?:\((\d[\d,]*)\s*XP\))?`)
	sensesRE         = regexp.MustCompile(`(?i)^\*\*\s*Senses\s*\*\*[:\s]*\s*(.+?)\s*$`)
	langsRE          = regexp.MustCompile(`(?i)^\*\*\s*Languages\s*\*\*[:\s]*\s*(.+?)\s*$`)
	savesRE          = regexp.MustCompile(`(?i)^\*\*\s*Saving\s+Throws\s*\*\*[:\s]*\s*(.+?)\s*$`)
	skillsRE         = regexp.MustCompile(`(?i)^\*\*\s*Skills\s*\*\*[:\s]*\s*(.+?)\s*$`)
	dmgVulnRE        = regexp.MustCompile(`(?i)^\*\*\s*Damage\s+Vulnerabilities\s*\*\*[:\s]*\s*(.+?)\s*$`)
	dmgResRE         = regexp.MustCompile(`(?i)^\*\*\s*Damage\s+Resistances\s*\*\*[:\s]*\s*(.+?)\s*$`)
	dmgImmRE         = regexp.MustCompile(`(?i)^\*\*\s*Damage\s+Immunities\s*\*\*[:\s]*\s*(.+?)\s*$`)
	condImmRE        = regexp.MustCompile(`(?i)^\*\*\s*Condition\s+Immunities\s*\*\*[:\s]*\s*(.+?)\s*$`)
	legendaryIntroRE = regexp.MustCompile(`(?i)can\s+take\s+(\d+)\s+legendary\s+actions?`)
)

// parseFieldLine inspects a single line and, if it matches a known
// field pattern, populates the monster. Returns an error only when a
// present field is malformed (e.g. AC value is not a number). Returns
// silently for unrecognized lines.
func parseFieldLine(l line, m *rules.Monster) error {
	t := l.Text

	// Initiative (2025: "+4 (+14) (Dexterity)").
	if matches := init2025RE.FindStringSubmatch(t); matches != nil {
		mod, _ := strconv.Atoi(matches[1])
		score, _ := strconv.Atoi(matches[2])
		m.Initiative = mod
		m.InitScore = score
		return nil
	}

	// Armor Class.
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(t)), "**armor class**") {
		matches := acRE.FindStringSubmatch(t)
		if matches == nil {
			return &ParseError{Line: l.N, Msg: "malformed Armor Class line: expected a number"}
		}
		ac, _ := strconv.Atoi(matches[1])
		m.AC = ac
		if len(matches) > 2 && matches[2] != "" {
			m.ACSource = matches[2]
		}
		return nil
	}

	// Hit Points.
	if matches := hpRE.FindStringSubmatch(t); matches != nil {
		hp, _ := strconv.Atoi(matches[1])
		m.HP = hp
		if len(matches) > 2 && matches[2] != "" {
			m.HPDice = matches[2]
		}
		return nil
	}

	// Speed.
	if matches := speedRE.FindStringSubmatch(t); matches != nil {
		parseSpeed(matches[1], m)
		return nil
	}

	// Challenge.
	if matches := crRE.FindStringSubmatch(t); matches != nil {
		cr, err := rules.ParseCR(matches[1])
		if err != nil {
			return &ParseError{Line: l.N, Msg: "malformed Challenge line: " + err.Error()}
		}
		m.CR = cr
		if len(matches) > 2 && matches[2] != "" {
			xpStr := strings.ReplaceAll(matches[2], ",", "")
			if xp, err := strconv.Atoi(xpStr); err == nil {
				m.XP = xp
			}
		}
		return nil
	}

	// Senses.
	if matches := sensesRE.FindStringSubmatch(t); matches != nil {
		val := strings.TrimSpace(matches[1])
		if !isNone(val) {
			parseSenses(val, m)
		}
		return nil
	}

	// Languages.
	if matches := langsRE.FindStringSubmatch(t); matches != nil {
		val := strings.TrimSpace(matches[1])
		if !isNone(val) {
			for _, lang := range strings.Split(val, ",") {
				m.Languages = append(m.Languages, strings.TrimSpace(lang))
			}
		}
		return nil
	}

	// Saving Throws.
	if matches := savesRE.FindStringSubmatch(t); matches != nil {
		for _, s := range strings.Split(matches[1], ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			parts := strings.Fields(s)
			if len(parts) > 0 {
				m.Saves = append(m.Saves, parts[0])
			}
		}
		return nil
	}

	// Skills.
	if matches := skillsRE.FindStringSubmatch(t); matches != nil {
		for _, s := range strings.Split(matches[1], ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			parts := strings.Fields(s)
			if len(parts) == 0 {
				continue
			}
			bonus := 0
			if len(parts) >= 2 {
				b, _ := strconv.Atoi(strings.TrimPrefix(parts[1], "+"))
				bonus = b
			}
			m.Skills[parts[0]] = bonus
		}
		return nil
	}

	// Damage Vulnerabilities / Resistances / Immunities.
	if matches := dmgVulnRE.FindStringSubmatch(t); matches != nil {
		m.DamageVulnerabilities = parseDamageTypes(matches[1])
		return nil
	}
	if matches := dmgResRE.FindStringSubmatch(t); matches != nil {
		m.DamageResistances = parseDamageTypes(matches[1])
		return nil
	}
	if matches := dmgImmRE.FindStringSubmatch(t); matches != nil {
		m.DamageImmunities = parseDamageTypes(matches[1])
		return nil
	}

	// Condition Immunities.
	if matches := condImmRE.FindStringSubmatch(t); matches != nil {
		for _, c := range strings.Split(matches[1], ",") {
			c = strings.TrimSpace(c)
			if c == "" || isNone(c) {
				continue
			}
			m.ConditionImmunities = append(m.ConditionImmunities, rules.Condition(strings.ToLower(c)))
		}
		return nil
	}

	// Ability row (table cell with "8 (-1)" pattern).
	if cells := abilityCellRE.FindAllStringSubmatch(t, -1); len(cells) >= 6 {
		if v, err := strconv.Atoi(cells[0][1]); err == nil {
			m.Abilities.STR = v
		}
		if v, err := strconv.Atoi(cells[1][1]); err == nil {
			m.Abilities.DEX = v
		}
		if v, err := strconv.Atoi(cells[2][1]); err == nil {
			m.Abilities.CON = v
		}
		if v, err := strconv.Atoi(cells[3][1]); err == nil {
			m.Abilities.INT = v
		}
		if v, err := strconv.Atoi(cells[4][1]); err == nil {
			m.Abilities.WIS = v
		}
		if v, err := strconv.Atoi(cells[5][1]); err == nil {
			m.Abilities.CHA = v
		}
		return nil
	}

	// Initiative (2014 fallback).
	if matches := init2014RE.FindStringSubmatch(t); matches != nil {
		mod, _ := strconv.Atoi(matches[1])
		m.Initiative = mod
		return nil
	}

	// Legendary actions intro line.
	if matches := legendaryIntroRE.FindStringSubmatch(t); matches != nil {
		if m.Legendary == nil {
			m.Legendary = &rules.LegendaryGroup{}
		}
		uses, _ := strconv.Atoi(matches[1])
		m.Legendary.Uses = uses
		return nil
	}

	return nil
}

// parseSpeed splits a speed string like "30 ft., fly 60 ft., swim 30 ft."
// into the monster's Speed map.
func parseSpeed(s string, m *rules.Monster) {
	cleaned := strings.ReplaceAll(s, "ft.", "")
	cleaned = strings.ReplaceAll(cleaned, "ft", "")
	for _, part := range strings.Split(cleaned, ",") {
		part = strings.TrimSpace(part)
		if part == "" || isNone(part) {
			continue
		}
		if idx := strings.Index(part, "("); idx != -1 {
			part = strings.TrimSpace(part[:idx])
		}
		fields := strings.Fields(part)
		switch len(fields) {
		case 1:
			if v, err := strconv.Atoi(fields[0]); err == nil {
				m.Speed[rules.SpeedWalk] = v
			}
		case 2:
			kind := strings.ToLower(fields[0])
			v, err := strconv.Atoi(fields[1])
			if err != nil {
				continue
			}
			switch kind {
			case "walk":
				m.Speed[rules.SpeedWalk] = v
			case "fly":
				m.Speed[rules.SpeedFly] = v
			case "swim":
				m.Speed[rules.SpeedSwim] = v
			case "burrow":
				m.Speed[rules.SpeedBurrow] = v
			case "climb":
				m.Speed[rules.SpeedClimb] = v
			}
		}
	}
}

// parseSenses splits "darkvision 60 ft., passive Perception 9" into the
// senses struct.
func parseSenses(s string, m *rules.Monster) {
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lower := strings.ToLower(part)
		if strings.HasPrefix(lower, "passive perception") {
			fields := strings.Fields(part)
			if len(fields) >= 3 {
				if v, err := strconv.Atoi(fields[2]); err == nil {
					m.Senses.PassivePerception = v
				}
			}
			continue
		}
		m.Senses.Special = append(m.Senses.Special, part)
	}
}

// parseDamageTypes splits a damage list into a slice of DamageType.
// Unknown types are dropped.
func parseDamageTypes(s string) []rules.DamageType {
	var out []rules.DamageType
	for _, part := range strings.Split(s, ",") {
		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}
		first := strings.ToLower(fields[0])
		if first == "and" || first == "from" || isNone(first) {
			continue
		}
		dt := rules.DamageType(first)
		if isKnownDamageType(dt) {
			out = append(out, dt)
		}
	}
	return out
}

var knownDamageTypes = map[rules.DamageType]bool{
	rules.DmgAcid: true, rules.DmgBludgeon: true, rules.DmgCold: true,
	rules.DmgFire: true, rules.DmgForce: true, rules.DmgLightning: true,
	rules.DmgNecrotic: true, rules.DmgPierce: true, rules.DmgPoison: true,
	rules.DmgPsychic: true, rules.DmgRadiant: true, rules.DmgSlash: true, rules.DmgThunder: true,
}

func isKnownDamageType(dt rules.DamageType) bool {
	return knownDamageTypes[dt]
}

func isNone(s string) bool {
	s = strings.TrimSpace(s)
	return strings.EqualFold(s, "None") || strings.EqualFold(s, "—") || s == "-"
}

// parseSectionedBodies extracts Traits, Actions, Bonus Actions, Reactions,
// and Legendary actions from H3/H2 section bodies.
//
// The format is "### Traits\n\n**Name.** description\n\n**Name2.** ..."
func parseSectionedBodies(ls []line, m *rules.Monster) {
	type section struct {
		name   string
		target *[]rules.Action
		reac   *[]rules.Reaction
		trait  *[]rules.Trait
		legend *[]rules.Action
	}
	// Map section header (lowercased) → target slice.
	sections := []section{
		{"traits", nil, nil, &m.Traits, nil},
		{"actions", &m.Actions, nil, nil, nil},
		{"bonus actions", &m.BonusActions, nil, nil, nil},
		{"reactions", nil, &m.Reactions, nil, nil},
		{"legendary actions", nil, nil, nil, nil}, // handled below
	}
	_ = sections[4].legend

	// Simple state machine: walk lines; on a section header, switch context.
	currentHeader := ""
	pendingName := ""
	pendingDesc := strings.Builder{}
	flushEntry := func() {
		if pendingName == "" {
			return
		}
		desc := strings.TrimSpace(pendingDesc.String())
		switch currentHeader {
		case "traits":
			m.Traits = append(m.Traits, rules.Trait{Name: pendingName, Description: desc})
		case "actions":
			m.Actions = append(m.Actions, rules.Action{Name: pendingName, Description: desc})
		case "bonus actions":
			m.BonusActions = append(m.BonusActions, rules.Action{Name: pendingName, Description: desc})
		case "reactions":
			m.Reactions = append(m.Reactions, rules.Reaction{Action: rules.Action{Name: pendingName, Description: desc}})
		case "legendary actions":
			if m.Legendary == nil {
				m.Legendary = &rules.LegendaryGroup{}
			}
			m.Legendary.Actions = append(m.Legendary.Actions, rules.Action{Name: pendingName, Description: desc})
		}
		pendingName = ""
		pendingDesc.Reset()
	}

	traitRE := regexp.MustCompile(`^\*\*\s*([^*]+?)\.\*\*\s+(.+)$`)
	for _, l := range ls {
		t := strings.TrimSpace(l.Text)
		// H2/H3/H4 section header.
		if strings.HasPrefix(t, "#") {
			flushEntry()
			header := strings.TrimSpace(strings.TrimLeft(t, "#"))
			currentHeader = strings.ToLower(strings.TrimSpace(header))
			continue
		}
		// Match "**Name.** description" as the start of a new entry.
		if matches := traitRE.FindStringSubmatch(t); matches != nil {
			if currentHeader != "" {
				flushEntry()
				pendingName = strings.TrimSpace(matches[1])
				pendingDesc.WriteString(strings.TrimSpace(matches[2]))
			}
			continue
		}
		// Continuation line: append to current description.
		if currentHeader != "" && pendingName != "" && t != "" {
			pendingDesc.WriteString(" ")
			pendingDesc.WriteString(t)
		}
	}
	flushEntry()
}
