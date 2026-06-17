package validators

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Development branch patterns (bilingual ES|EN)
	developmentBranchPattern   = regexp.MustCompile(`(?i)###?\s*(?:Developments?|Desarrollo)`)
	ifThenPattern             = regexp.MustCompile(`(?i)\*\*(?:Si|If)\s+(?:(?:the\s+)?PCs?\s+|los\s+PJs?\s+)?[^:]+:\*\*`)
	consecuenciaPattern       = regexp.MustCompile(`(?i)\*\*(?:(?:Consecuencia|Consequence)\s+(?:inmediata|futura|a\s+largo\s+plazo|immediate|future|long-term)|(?:Immediate|Future|Long-term)\s+(?:consecuencia|consequence)):\*\*`)
	recuperacionPattern       = regexp.MustCompile(`(?i)\*\*(?:Recuperación|Recovery):\*\*`)
	ramasPattern              = regexp.MustCompile(`(?i)(?:rama|branch|opción|path|camino)\s*#?\d*`)
	
	// Solution path patterns
	stealthSolutionPattern    = regexp.MustCompile(`(?i)(sigilo|stealth|oculto|invisibl|sneak)`)
	stealthDCPattern          = regexp.MustCompile(`(?i)(CD|DC)\s*\d+\s*(sigilo|stealth)`)
	socialSolutionPattern     = regexp.MustCompile(`(?i)(persuas|diplom|engañ|deception|intimid|amenaz|negoci)`)
	socialDCPattern           = regexp.MustCompile(`(?i)(CD|DC)\s*\d+\s*(persuas|diplom|engañ|intimid)`)
	combatSolutionPattern     = regexp.MustCompile(`(?i)(combate|atac|pele|fight|attack|enemig|criatura)`)
	
	// Boxed text patterns (bilingual ES|EN, colon optional)
	boxedTextPattern          = regexp.MustCompile(`(?i)>>\s*\*\*(?:Texto para Leer|Read-Aloud\s+Text):?\*\*`)
	secondPersonPattern       = regexp.MustCompile(`(?i)(ves|escuch|sient|percib|not|mir|huel|you\s+see|you\s+hear|you\s+feel|you\s+notice|you\s+observe)`)
	presentTensePattern       = regexp.MustCompile(`(?i)(está|hay|son|encuentr|observ|aparec|\bis\b|\bare\b|stands|appears|stretches|stretching|lies|glows|shimmers|echoes|fills|surrounds|rises|falls)`)
)

// DevelopmentBranch represents a decision branch in an area
type DevelopmentBranch struct {
	Trigger        string
	Immediate      string
	Future         string
	Recovery       string
	HasRecovery    bool
}

// ValidateDevelopments checks that an area has proper development branches
func ValidateDevelopments(md string) ValidationResult {
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

	// Check 1: Has Developments section
	if !developmentBranchPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "developments",
			Message: "missing ### Developments section",
		})
		result.Valid = false
		return result
	}
	
	// Extract Developments section (bilingual)
	developmentsSection := extractSectionRegex(md, developmentBranchPattern)
	
	// Normalize for pattern matching
	normalized := strings.ReplaceAll(developmentsSection, "\n", " ")
	
	// Check 2: Minimum 3 branches
	branchCount := countBranches(developmentsSection)
	if branchCount < 3 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "branch_count",
			Message: "Developments must have minimum 3 branches, found " + itoa(branchCount),
		})
	}
	
	// Check 3: Each branch has IF-THEN structure
	ifThenCount := len(ifThenPattern.FindAllString(normalized, -1))
	if ifThenCount < branchCount {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "if_then_structure",
			Message: "not all branches have **Si [condición]**: structure",
		})
	}
	
	// Check 4: Branches have consequences
	consecuenciaCount := len(consecuenciaPattern.FindAllString(normalized, -1))
	if consecuenciaCount < branchCount {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "consequences",
			Message: "not all branches have **Consecuencia:** field",
		})
	}
	
	// Check 5: At least one recovery path
	recuperacionCount := len(recuperacionPattern.FindAllString(normalized, -1))
	if recuperacionCount == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "recovery",
			Message: "no **Recuperación:** paths found - all failures need recovery options",
		})
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateMultipleSolutions checks that obstacles have 3+ solution paths
func ValidateMultipleSolutions(md string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	// Normalize for pattern matching
	normalized := strings.ReplaceAll(md, "\n", " ")
	
	// Check 1: Stealth option with numeric DC
	hasStealth := stealthSolutionPattern.MatchString(normalized)
	hasStealthDC := stealthDCPattern.MatchString(normalized)
	if hasStealth && !hasStealthDC {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "stealth_dc",
			Message: "stealth solution mentioned but no numeric DC provided (e.g., CD 15 Sigilo)",
		})
	}
	
	// Check 2: Social option with numeric DC
	hasSocial := socialSolutionPattern.MatchString(normalized)
	hasSocialDC := socialDCPattern.MatchString(normalized)
	if hasSocial && !hasSocialDC {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "social_dc",
			Message: "social solution mentioned but no numeric DC provided (e.g., CD 14 Persuasión)",
		})
	}
	
	// Check 3: Combat option
	hasCombat := combatSolutionPattern.MatchString(normalized)
	
	// Check 4: At least 2 of 3 solution types present
	solutionTypes := 0
	if hasStealth || hasStealthDC {
		solutionTypes++
	}
	if hasSocial || hasSocialDC {
		solutionTypes++
	}
	if hasCombat {
		solutionTypes++
	}
	
	if solutionTypes < 2 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "solution_variety",
			Message: "obstacles must offer at least 2 different solution types (stealth/social/combat), found " + itoa(solutionTypes),
		})
	}
	
	// Check 5: No relative DCs (alto/bajo)
	if relativeDCPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "numeric_dc",
			Message: "relative DCs (alto/bajo/high/low) are not allowed - use numeric DCs only",
		})
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateBoxedText checks that boxed text follows WotC standards
func ValidateBoxedText(md string) ValidationResult {
	result := ValidationResult{Valid: true}

	// Check 0: Mixed language detection for boxed text labels
	esBoxed := esBoxedText.MatchString(md)
	enBoxed := enBoxedText.MatchString(md)
	if esBoxed && enBoxed {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "mixed_language",
			Message: "mixed language detected: both 'Texto para Leer' and 'Read-Aloud Text' in same chapter",
		})
		result.Valid = false
		return result
	}

	// Check 1: Has boxed text
	if !boxedTextPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "boxed_text",
			Message: "missing >> **Texto para Leer** / >> **Read-Aloud Text** boxed text",
		})
		result.Valid = false
		return result
	}
	
	// Extract boxed text sections
	boxedTextSections := extractBoxedText(md)
	
	for i, text := range boxedTextSections {
		wordCount := CountWords(text)
		
		// Check 2: Word count 50-400 (WotC calibrated)
		if wordCount < 50 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_length",
				Message: "boxed text #" + itoa(i+1) + " has " + itoa(wordCount) + " words, minimum is 50",
			})
		}
		if wordCount > 400 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_length",
				Message: "boxed text #" + itoa(i+1) + " has " + itoa(wordCount) + " words, maximum is 400",
			})
		}
		
		// Normalize for pattern matching
		normalized := strings.ReplaceAll(text, "\n", " ")
		
		// Check 3: Second person voice
		if !secondPersonPattern.MatchString(normalized) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_voice",
				Message: "boxed text #" + itoa(i+1) + " should use second person (ves, escuchas, sientes)",
			})
		}
		
		// Check 4: Present tense
		if !presentTensePattern.MatchString(normalized) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_tense",
				Message: "boxed text #" + itoa(i+1) + " should use present tense (está, hay, son)",
			})
		}
		
		// Check 5: No spoilers/mechanics
		if containsMechanics(text) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_spoilers",
				Message: "boxed text #" + itoa(i+1) + " contains mechanics or spoilers (DCs, stats, monster names)",
			})
		}
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateCharacterHooks checks that areas have character hooks
func ValidateCharacterHooks(md string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	// Normalize for pattern matching
	normalized := strings.ReplaceAll(md, "\n", " ")
	
	// Check 1: Has character hooks section
	hookPattern := regexp.MustCompile(`(?i)(character\s+hook|gancho\s+de\s+personaje|hook\s+por\s+background)`)
	if !hookPattern.MatchString(normalized) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "character_hooks",
			Message: "missing character hooks section",
		})
		return result
	}
	
	// Check 2: Minimum 2 hooks
	hookCount := len(hookPattern.FindAllString(normalized, -1))
	if hookCount < 2 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "hook_count",
			Message: "areas must have minimum 2 character hooks, found " + itoa(hookCount),
		})
	}
	
	// Check 3: Hooks tied to background/class/race/faction
	backgroundHookPattern := regexp.MustCompile(`(?i)(background|clase|raza|facción|origen)`)
	if !backgroundHookPattern.MatchString(normalized) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "hook_association",
			Message: "character hooks must be tied to background, class, race, or faction",
		})
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// Helper functions

func extractSection(md, sectionName string) string {
	// Find section heading
	pattern := regexp.MustCompile(`(?i)###?\s*` + regexp.QuoteMeta(sectionName))
	loc := pattern.FindStringIndex(md)
	if loc == nil {
		return ""
	}
	
	// Extract from heading to next ### or end
	start := loc[0]
	nextSection := regexp.MustCompile(`(?m)^#{3,}`).FindStringIndex(md[start+1:])
	if nextSection == nil {
		return md[start:]
	}
	
	return md[start : start+1+nextSection[0]]
}

// extractSectionRegex extracts a section using a pre-compiled regex pattern
func extractSectionRegex(md string, pattern *regexp.Regexp) string {
	loc := pattern.FindStringIndex(md)
	if loc == nil {
		return ""
	}
	
	// Extract from heading to next ### or end
	start := loc[0]
	nextSection := regexp.MustCompile(`(?m)^#{3,}`).FindStringIndex(md[start+1:])
	if nextSection == nil {
		return md[start:]
	}
	
	return md[start : start+1+nextSection[0]]
}

func countBranches(text string) int {
	// Normalize text for matching (remove newlines)
	normalized := strings.ReplaceAll(text, "\n", " ")
	
	count := len(ifThenPattern.FindAllString(normalized, -1))
	if count == 0 {
		// Fallback: count numbered items
		count = len(ramasPattern.FindAllString(text, -1))
	}
	if count == 0 {
		// Fallback: count bullet points under Developments
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "-") || strings.HasPrefix(strings.TrimSpace(line), "*") {
				count++
			}
		}
	}
	return count
}

func extractBoxedText(md string) []string {
	var texts []string
	// Match >> followed by content until next >>, # heading, or end
	// Note: Using [\s\S]*? instead of (?s:.) for Go regex compatibility
	pattern := regexp.MustCompile(`>>([\s\S]*?)(?:\n>>|\n#|$)`)
	matches := pattern.FindAllString(md, -1)
	for _, m := range matches {
		texts = append(texts, strings.TrimPrefix(m, ">>"))
	}
	return texts
}

func containsMechanics(text string) bool {
	// Check for DCs, stats, monster names
	mechanicsPattern := regexp.MustCompile(`(?i)(CD\s*\d+|DC\s*\d+|AC|HP|CR|XP|proficienc|saving throw|skill bonus)`)
	return mechanicsPattern.MatchString(text)
}

// ValidateNPCStatLinks checks that NPCs reference stat blocks in bestiary
func ValidateNPCStatLinks(npcsMD, bestiaryMD string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	// Extract NPC names from npcs.md
	npcNamePattern := regexp.MustCompile(`(?i)###\s+([A-Za-z][A-Za-z\s']+?)\s*\n`)
	npcMatches := npcNamePattern.FindAllStringSubmatch(npcsMD, -1)
	
	var npcNames []string
	for _, m := range npcMatches {
		name := strings.TrimSpace(m[1])
		if len(name) > 2 {
			npcNames = append(npcNames, name)
		}
	}
	
	// Check each NPC has a stat block reference
	for _, npcName := range npcNames {
		// Look for "Ver bestiary.md: [Name]" or "[Name]" in bestiary
		refPattern := regexp.MustCompile(`(?i)(ver\s+bestiary\.md:\s*` + regexp.QuoteMeta(npcName) + `|` + regexp.QuoteMeta(npcName) + `)`)
		
		hasRef := refPattern.MatchString(npcsMD) || strings.Contains(bestiaryMD, npcName)
		
		if !hasRef {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "npc_stat_link",
				Message: fmt.Sprintf("NPC '%s' has no stat block reference in bestiary.md", npcName),
			})
		}
	}
	
	// Check for stat blocks in bestiary without corresponding NPC
	bestiaryNPCPattern := regexp.MustCompile(`(?i)###\s+([A-Za-z][A-Za-z\s']+?)\s*\n.*?\*[A-Z][a-z]+\s+(male|female|neutral)\s+`)
	bestiaryMatches := bestiaryNPCPattern.FindAllStringSubmatch(bestiaryMD, -1)
	
	for _, m := range bestiaryMatches {
		name := strings.TrimSpace(m[1])
		found := false
		for _, npcName := range npcNames {
			if strings.EqualFold(name, npcName) {
				found = true
				break
			}
		}
		
		if !found {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Stat block '%s' in bestiary has no corresponding NPC in npcs.md", name))
		}
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateNPCWordCount checks that NPCs meet WotC word count standards
func ValidateNPCWordCount(npcsMD string) ValidationResult {
	result := ValidationResult{Valid: true}
	
	// Split by NPC sections
	sections := regexp.MustCompile(`(?m)^###\s+`).Split(npcsMD, -1)
	
	for i, section := range sections {
		if i == 0 {
			continue // Skip content before first NPC
		}
		
		lines := strings.Split(section, "\n")
		npcName := strings.TrimSpace(lines[0])
		wordCount := CountWords(section)
		
		// Determine if major or minor NPC
		isMajor := strings.Contains(strings.ToLower(section), "apariencia física") ||
			strings.Contains(strings.ToLower(section), "personalidad") ||
			strings.Contains(strings.ToLower(section), "secretos")
		
		if isMajor {
			if wordCount < 500 {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "npc_word_count",
					Message: fmt.Sprintf("Major NPC '%s' has %d words, minimum is 500", npcName, wordCount),
				})
			}
			if wordCount > 800 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Major NPC '%s' has %d words, recommended maximum is 800", npcName, wordCount))
			}
		} else {
			if wordCount < 200 {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "npc_word_count",
					Message: fmt.Sprintf("Minor NPC '%s' has %d words, minimum is 200", npcName, wordCount),
				})
			}
		}
	}
	
	result.Valid = len(result.Errors) == 0
	return result
}
