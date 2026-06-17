package validators

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// Area heading pattern: ### Área X: Name or ### Area X: Name (supports numbered and lettered areas)
	areaHeadingPattern = regexp.MustCompile(`(?m)^#{3,4}\s+(?:[Áa]rea|Area)\s+(\d+|[A-Z]\d*)(?:\s*:\s*(.+))?$`)

	// Relative DC pattern (invalid): "DC alto", "DC bajo", "DC high", "DC low"
	relativeDCPattern = regexp.MustCompile(`(?i)DC\s+(alto|bajo|high|low|medio|moderado|dif[íi]cil)`)

	// Connection patterns (support both numbered and lettered area refs)
	connectionForwardPattern  = regexp.MustCompile(`(?i)→\s*(?:[Áa]rea|Area)\s+(\d+|[A-Z]\d*)`)
	connectionBackwardPattern = regexp.MustCompile(`(?i)←\s*(?:[Áa]rea|Area)\s+(\d+|[A-Z]\d*)`)

	// Treasure with XP pattern (handles **XP:** 50 XP or XP: 50)
	treasureXPPattern = regexp.MustCompile(`(?i)XP\W*\d+`)

	// Creature/NPC pattern (bilingual)
	creaturePattern = regexp.MustCompile(`(?i)\*\*(?:Criaturas|Creatures):\*\*`)
	interactivePattern = regexp.MustCompile(`(?i)(criaturas?|npcs?|tesoro|trampa|secreto|pista|puzzle|creatures?|treasure|trap|secret|clue)`)
)

// Area represents a parsed area block
type Area struct {
	Number      int
	LetterID    string // For lettered areas (A1, B2, etc.)
	Name        string
	Content     string
	Connections []int
}

// AreaID returns a string identifier for the area (number or letter)
func (a Area) AreaID() string {
	if a.LetterID != "" {
		return a.LetterID
	}
	return itoa(a.Number)
}

// ValidateAreaMarkdown validates a full act markdown for area format compliance
func ValidateAreaMarkdown(md string) ValidationResult {
	result := ValidationResult{Valid: true}

	// Check 0: Mixed language detection
	_, langErr := DetectLanguage(md)
	if langErr != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "mixed_language",
			Message: langErr.Error(),
		})
		result.Valid = false
		return result
	}

	areas := parseAreas(md)

	// Check 1: Area count (7-15 for standard chapter, WotC calibrated)
	if len(areas) < 7 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "area_count",
			Message: "chapter has " + itoa(len(areas)) + " areas, minimum is 7",
		})
	}
	if len(areas) > 15 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "area_count",
			Message: "chapter has " + itoa(len(areas)) + " areas, maximum is 15",
		})
	}

	// Build bidirectional connection map (using string keys for lettered areas)
	connectionMap := make(map[string]map[string]bool)
	for _, area := range areas {
		connectionMap[area.AreaID()] = make(map[string]bool)
	}

	for _, area := range areas {
		// Extract forward connections (→ Área X)
		fwdMatches := connectionForwardPattern.FindAllStringSubmatch(area.Content, -1)
		for _, m := range fwdMatches {
			connectionMap[area.AreaID()][m[1]] = true
		}

		// Extract backward connections (← Área X)
		bwdMatches := connectionBackwardPattern.FindAllStringSubmatch(area.Content, -1)
		for _, m := range bwdMatches {
			connectionMap[area.AreaID()][m[1]] = true
		}
	}

	// Check each area
	for _, area := range areas {
		wordCount := CountWords(area.Content)

		// Check 2: Word count per area (150-600, WotC calibrated)
		if wordCount < 150 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "word_count",
				Message: "Area " + area.AreaID() + " has " + itoa(wordCount) + " words, minimum is 150",
			})
		}
		if wordCount > 600 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "word_count",
				Message: "Area " + area.AreaID() + " has " + itoa(wordCount) + " words, maximum is 600",
			})
		}

		// Check 3: Numeric DCs only
		if relativeDCPattern.MatchString(area.Content) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "numeric_dc",
				Message: "Area " + area.AreaID() + " contains non-numeric DC (alto/bajo/high/low)",
			})
		}

		// Check 4: Treasure with XP if creatures present
		hasCreatures := creaturePattern.MatchString(area.Content)
		hasTreasureXP := treasureXPPattern.MatchString(area.Content)
		if hasCreatures && !hasTreasureXP {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "treasure",
				Message: "Area " + area.AreaID() + " has creatures but no XP treasure",
			})
		}

		// Check 5: At least one interactive element
		hasInteractive := interactivePattern.MatchString(area.Content)
		if !hasInteractive {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "interactive_element",
				Message: "Area " + area.AreaID() + " has no interactive elements (creatures, NPCs, traps, treasure, clues)",
			})
		}
	}

	// Check 6: Bidirectional connections
	for _, area := range areas {
		for targetID := range connectionMap[area.AreaID()] {
			if targetID == area.AreaID() {
				continue // self-connection is allowed
			}
			if _, exists := connectionMap[targetID]; !exists {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "bidirectional",
					Message: "Area " + area.AreaID() + " connects to Area " + targetID + " but Area " + targetID + " does not exist",
				})
				continue
			}
			if !connectionMap[targetID][area.AreaID()] {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "bidirectional",
					Message: "Area " + area.AreaID() + " → Area " + targetID + " but reverse connection is missing",
				})
			}
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateChapterWordCount validates total chapter word count (3000-16000 WotC calibrated)
func ValidateChapterWordCount(md string) ValidationResult {
	result := ValidationResult{Valid: true}
	wordCount := CountWords(md)

	if wordCount < 3000 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "chapter_word_count",
			Message: "Chapter word count " + itoa(wordCount) + " below minimum 3000",
		})
	}
	if wordCount > 16000 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "chapter_word_count",
			Message: "Chapter word count " + itoa(wordCount) + " exceeds maximum 16000",
		})
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// parseAreas extracts individual areas from act markdown (supports numbered and lettered)
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

			name := ""
			if len(matches) > 2 && matches[2] != "" {
				name = strings.TrimSpace(matches[2])
			}

			areaID := matches[1]
			num, err := strconv.Atoi(areaID)
			if err != nil {
				// Lettered area (A1, B2, etc.)
				currentArea = &Area{
					Number:   0,
					LetterID: areaID,
					Name:     name,
				}
			} else {
				currentArea = &Area{
					Number: num,
					Name:   name,
				}
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
