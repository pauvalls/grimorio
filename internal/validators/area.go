package validators

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Area heading pattern: ### Área X: Name or ### Area X: Name
	areaHeadingPattern = regexp.MustCompile(`(?m)^#{3,4}\s+[Áa]rea\s+(\d+):\s*(.+)$`)

	// DC pattern: must be numeric like "DC 14" or "DC15"
	dcPattern = regexp.MustCompile(`(?i)DC\s*\d+`)

	// Relative DC pattern (invalid): "DC alto", "DC bajo", "DC high", "DC low"
	relativeDCPattern = regexp.MustCompile(`(?i)DC\s+(alto|bajo|high|low|medio|moderado|dif[íi]cil)`)

	// Connection patterns
	connectionForwardPattern  = regexp.MustCompile(`(?i)→\s*[Áa]rea\s+(\d+)`)
	connectionBackwardPattern = regexp.MustCompile(`(?i)←\s*[Áa]rea\s+(\d+)`)

	// Treasure with XP pattern (handles **XP:** 50 XP or XP: 50)
	treasureXPPattern = regexp.MustCompile(`(?i)XP\W*\d+`)

	// Creature/NPC pattern
	creaturePattern = regexp.MustCompile(`(?i)\*\*Criaturas:\*\*`)
	npcPattern      = regexp.MustCompile(`(?i)\*\*NPCs?:\*\*`)
	trapPattern     = regexp.MustCompile(`(?i)\*\*Secretos y Trampas:\*\*`)
	interactivePattern = regexp.MustCompile(`(?i)(criaturas?|npcs?|tesoro|trampa|secreto|pista|puzzle)`)
)

// Area represents a parsed area block
type Area struct {
	Number      int
	Name        string
	Content     string
	Connections []int
}

// ValidateAreaMarkdown validates a full act markdown for area format compliance
func ValidateAreaMarkdown(md string) ValidationResult {
	result := ValidationResult{Valid: true}

	areas := parseAreas(md)

	// Check 1: Area count (10-15 for standard act)
	if len(areas) < 10 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "area_count",
			Message: "act has " + itoa(len(areas)) + " areas, minimum is 10",
		})
	}
	if len(areas) > 15 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "area_count",
			Message: "act has " + itoa(len(areas)) + " areas, maximum is 15",
		})
	}

	// Build bidirectional connection map
	connectionMap := make(map[int]map[int]bool)
	for _, area := range areas {
		connectionMap[area.Number] = make(map[int]bool)
	}

	for _, area := range areas {
		// Extract forward connections (→ Área X)
		fwdMatches := connectionForwardPattern.FindAllStringSubmatch(area.Content, -1)
		for _, m := range fwdMatches {
			if num, err := strconv.Atoi(m[1]); err == nil {
				connectionMap[area.Number][num] = true
			}
		}

		// Extract backward connections (← Área X)
		bwdMatches := connectionBackwardPattern.FindAllStringSubmatch(area.Content, -1)
		for _, m := range bwdMatches {
			if num, err := strconv.Atoi(m[1]); err == nil {
				connectionMap[area.Number][num] = true
			}
		}
	}

	// Check each area
	for _, area := range areas {
		wordCount := CountWords(area.Content)

		// Check 2: Word count per area (150-200)
		if wordCount < 150 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "word_count",
				Message: "Área " + itoa(area.Number) + " has " + itoa(wordCount) + " words, minimum is 150",
			})
		}
		if wordCount > 200 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "word_count",
				Message: "Área " + itoa(area.Number) + " has " + itoa(wordCount) + " words, maximum is 200",
			})
		}

		// Check 3: Numeric DCs only
		if relativeDCPattern.MatchString(area.Content) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "numeric_dc",
				Message: "Área " + itoa(area.Number) + " contains non-numeric DC (alto/bajo/high/low)",
			})
		}

		// Check 4: Treasure with XP if creatures present
		hasCreatures := creaturePattern.MatchString(area.Content)
		hasTreasureXP := treasureXPPattern.MatchString(area.Content)
		if hasCreatures && !hasTreasureXP {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "treasure",
				Message: "Área " + itoa(area.Number) + " has creatures but no XP treasure",
			})
		}

		// Check 5: At least one interactive element
		hasInteractive := interactivePattern.MatchString(area.Content)
		if !hasInteractive {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "interactive_element",
				Message: "Área " + itoa(area.Number) + " has no interactive elements (creatures, NPCs, traps, treasure, clues)",
			})
		}
	}

	// Check 6: Bidirectional connections
	for _, area := range areas {
		for targetNum := range connectionMap[area.Number] {
			// Check reverse connection exists
			if targetNum == area.Number {
				continue // self-connection is allowed
			}
			if _, exists := connectionMap[targetNum]; !exists {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "bidirectional",
					Message: "Área " + itoa(area.Number) + " connects to Área " + itoa(targetNum) + " but Área " + itoa(targetNum) + " does not exist",
				})
				continue
			}
			if !connectionMap[targetNum][area.Number] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "bidirectional",
					Message: "Área " + itoa(area.Number) + " → Área " + itoa(targetNum) + " but reverse connection is missing",
				})
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// parseAreas extracts individual areas from act markdown
func parseAreas(md string) []Area {
	var areas []Area
	lines := strings.Split(md, "\n")

	var currentArea *Area
	var contentLines []string

	for _, line := range lines {
		if matches := areaHeadingPattern.FindStringSubmatch(line); matches != nil {
			// Save previous area
			if currentArea != nil {
				currentArea.Content = strings.Join(contentLines, "\n")
				areas = append(areas, *currentArea)
			}

			num, _ := strconv.Atoi(matches[1])
			currentArea = &Area{
				Number: num,
				Name:   strings.TrimSpace(matches[2]),
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
