package validators

import (
	"regexp"
	"strings"
)

var (
	// Development branch patterns
	developmentBranchPattern   = regexp.MustCompile(`(?i)###?\s*Developments?`)
	ifThenPattern             = regexp.MustCompile(`(?i)\*\*Si\s+[^:]+:\*\*`)
	consecuenciaPattern       = regexp.MustCompile(`(?i)\*\*Consecuencia\s+(inmediata|futura|a\s+largo\s+plazo):\*\*`)
	recuperacionPattern       = regexp.MustCompile(`(?i)\*\*Recuperación:\*\*`)
	ramasPattern              = regexp.MustCompile(`(?i)(?:rama|branch|opción|path|camino)\s*#?\d*`)
	
	// Solution path patterns
	stealthSolutionPattern    = regexp.MustCompile(`(?i)(sigilo|stealth|oculto|invisibl|sneak)`)
	stealthDCPattern          = regexp.MustCompile(`(?i)(CD|DC)\s*\d+\s*(sigilo|stealth)`)
	socialSolutionPattern     = regexp.MustCompile(`(?i)(persuas|diplom|engañ|deception|intimid|amenaz|negoci)`)
	socialDCPattern           = regexp.MustCompile(`(?i)(CD|DC)\s*\d+\s*(persuas|diplom|engañ|intimid)`)
	combatSolutionPattern     = regexp.MustCompile(`(?i)(combate|atac|pele|fight|attack|enemig|criatura)`)
	
	// Boxed text patterns
	boxedTextPattern          = regexp.MustCompile(`(?i)>>\s*\*\*Texto para Leer\*\*`)
	secondPersonPattern       = regexp.MustCompile(`(?i)(ves|escuch|sient|percib|not|mir|huel)`)
	presentTensePattern       = regexp.MustCompile(`(?i)(está|hay|son|encuentr|observ|aparec)`)
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
	
	// Check 1: Has Developments section
	if !developmentBranchPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "developments",
			Message: "missing ### Developments section",
		})
		return result
	}
	
	// Extract Developments section
	developmentsSection := extractSection(md, "Developments")
	
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
	
	// Check 1: Has boxed text
	if !boxedTextPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "boxed_text",
			Message: "missing >> **Texto para Leer** boxed text",
		})
		return result
	}
	
	// Extract boxed text sections
	boxedTextSections := extractBoxedText(md)
	
	for i, text := range boxedTextSections {
		wordCount := CountWords(text)
		
		// Check 2: Word count 100-600
		if wordCount < 100 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_length",
				Message: "boxed text #" + itoa(i+1) + " has " + itoa(wordCount) + " words, minimum is 100",
			})
		}
		if wordCount > 600 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "boxed_text_length",
				Message: "boxed text #" + itoa(i+1) + " has " + itoa(wordCount) + " words, maximum is 600",
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
	pattern := regexp.MustCompile(`>>(?s:.+?)(?=\n>>|\n#|\z)`)
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
