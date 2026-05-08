package validators

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Bestiary entry header: ## Creature Name
	bestiaryHeaderPattern = regexp.MustCompile(`(?m)^#{2,3}\s+(.+)$`)

	// NPC header: ## NPC Name
	npcHeaderPattern = regexp.MustCompile(`(?m)^#{2,3}\s+(.+)$`)

	// XP in bestiary: (50 PX) or (100 XP)
	bestiaryXPPattern = regexp.MustCompile(`(?i)\((\d+)\s*(?:PX|XP)\)`)

	// Creature quantity in act: "2 **Goblin**" or "1 **Orc**"
	creatureQuantityPattern = regexp.MustCompile(`(?im)^\s*[-*]?\s*(\d+)\s+\*\*([^*]+)\*\*`)

	// NPC reference in act: "**NPC Name**" outside of Criaturas section
	npcRefInActPattern = regexp.MustCompile(`(?i)\*\*([^*]{3,50})\*\*`)
)

// ValidateIntegration checks cross-references between act, bestiary, and npcs
func ValidateIntegration(act, bestiary, npcs string) ValidationResult {
	result := ValidationResult{Valid: true}

	// Extract creatures from bestiary
	bestiaryCreatures := extractBestiaryCreatures(bestiary)

	// Extract NPCs from npcs file
	npcList := extractNPCs(npcs)

	// Parse act areas (use more permissive pattern for integration tests)
	areas := parseAreasIntegration(act)

	// Check cross-references per area
	for _, area := range areas {
		// Check creatures exist in bestiary
		creatureMatches := creatureQuantityPattern.FindAllStringSubmatch(area.Content, -1)
		for _, m := range creatureMatches {
			creatureName := strings.TrimSpace(m[2])
			if !contains(bestiaryCreatures, creatureName) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "creature_reference",
					Message: "Área " + itoa(area.Number) + " references creature '" + creatureName + "' not found in bestiary",
				})
			}
		}

		// Check NPCs exist in npcs file
		// Look for NPC references in explicit NPC sections only
		lines := strings.Split(area.Content, "\n")
		inNPCSection := false
		for _, line := range lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "**npcs:**") || strings.Contains(lower, "**npc:**") || strings.Contains(lower, "**personajes:**") {
				inNPCSection = true
				continue
			}
			if inNPCSection && (strings.HasPrefix(strings.TrimSpace(line), "## ") || strings.HasPrefix(strings.TrimSpace(line), "### ") || strings.HasPrefix(strings.TrimSpace(line), "**Criaturas:**") || strings.HasPrefix(strings.TrimSpace(line), "**Tesoro:**") || strings.HasPrefix(strings.TrimSpace(line), "**Conexiones:**")) {
				inNPCSection = false
				continue
			}

			if inNPCSection {
				matches := npcRefInActPattern.FindAllStringSubmatch(line, -1)
				for _, m := range matches {
					npcName := strings.TrimSpace(m[1])
					// Skip section headers and common terms
					if isHeaderOrCommonTerm(npcName) {
						continue
					}
					if len(npcName) > 2 && len(npcs) > 0 && !contains(npcList, npcName) {
						result.Errors = append(result.Errors, ValidationError{
							Field:   "npc_reference",
							Message: "Área " + itoa(area.Number) + " references NPC '" + npcName + "' not found in npcs file",
						})
					}
				}
			}
		}
	}

	// Check treasure consistency per area
	for _, area := range areas {
		hasCreatures := creaturePattern.MatchString(area.Content)
		hasTreasureXP := treasureXPPattern.MatchString(area.Content)
		if hasCreatures && !hasTreasureXP {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "treasure",
				Message: "Área " + itoa(area.Number) + " has creatures but no XP treasure",
			})
		}
	}

	// Check XP budget
	totalXP := CalculateTotalXP(act, bestiary)
	partySize := 5
	xpPerPC := totalXP / partySize

	// For level 1 act, XP per PC should be 300-400
	// Flag if significantly over (unbalanced act)
	if xpPerPC > 500 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "xp_budget",
			Message: "XP per PC (" + itoa(xpPerPC) + ") exceeds recommended threshold for act level",
		})
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// CalculateTotalXP computes total XP from creatures in the act based on bestiary
func CalculateTotalXP(act, bestiary string) int {
	total := 0

	// Build creature XP map from bestiary
	xpMap := make(map[string]int)
	lines := strings.Split(bestiary, "\n")
	var currentCreature string
	for _, line := range lines {
		if m := bestiaryHeaderPattern.FindStringSubmatch(line); m != nil {
			currentCreature = strings.TrimSpace(m[1])
		}
		if m := bestiaryXPPattern.FindStringSubmatch(line); m != nil {
			xp, _ := strconv.Atoi(m[1])
			if currentCreature != "" {
				xpMap[currentCreature] = xp
			}
		}
	}

	// Sum XP from act creature references
	areas := parseAreas(act)
	for _, area := range areas {
		matches := creatureQuantityPattern.FindAllStringSubmatch(area.Content, -1)
		for _, m := range matches {
			qty, _ := strconv.Atoi(m[1])
			creatureName := strings.TrimSpace(m[2])
			if xp, ok := xpMap[creatureName]; ok {
				total += qty * xp
			}
		}
	}

	return total
}

func extractBestiaryCreatures(bestiary string) []string {
	var creatures []string
	matches := bestiaryHeaderPattern.FindAllStringSubmatch(bestiary, -1)
	for _, m := range matches {
		creatures = append(creatures, strings.TrimSpace(m[1]))
	}
	return creatures
}

func extractNPCs(npcs string) []string {
	var list []string
	matches := npcHeaderPattern.FindAllStringSubmatch(npcs, -1)
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		// Skip section headers like "NPCs Principales"
		if !strings.Contains(strings.ToLower(name), "npcs") && !strings.Contains(strings.ToLower(name), "principales") {
			list = append(list, name)
		}
	}
	return list
}

func contains(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

// parseAreasIntegration extracts areas with more permissive matching (for integration tests)
func parseAreasIntegration(md string) []Area {
	var areas []Area
	lines := strings.Split(md, "\n")

	var currentArea *Area
	var contentLines []string

	// More permissive: ### Área N or ### Area N, with or without colon
	integrationHeadingPattern := regexp.MustCompile(`(?m)^#{3,4}\s+[Áa]rea\s+(\d+)`)

	for _, line := range lines {
		if matches := integrationHeadingPattern.FindStringSubmatch(line); matches != nil {
			if currentArea != nil {
				currentArea.Content = strings.Join(contentLines, "\n")
				areas = append(areas, *currentArea)
			}

			num, _ := strconv.Atoi(matches[1])
			currentArea = &Area{
				Number: num,
				Name:   "",
			}
			contentLines = nil
		} else if currentArea != nil {
			contentLines = append(contentLines, line)
		}
	}

	if currentArea != nil {
		currentArea.Content = strings.Join(contentLines, "\n")
		areas = append(areas, *currentArea)
	}

	return areas
}

func isHeaderOrCommonTerm(name string) bool {
	nameLower := strings.ToLower(name)
	terms := []string{
		"read-aloud", "descripción", "descripcion", "criaturas", "tesoro", "tesoros",
		"conexiones", "secretos", "trampas", "desarrollo", "npcs", "npc",
		"detectar", "mecanismo", "consecuencia", "desactivar", "encontrar",
		"contenido", "objetos", "moneda", "xp", "gp", "sp", "area",
	}
	for _, term := range terms {
		if nameLower == term || strings.Contains(nameLower, ":") {
			return true
		}
	}
	return false
}
