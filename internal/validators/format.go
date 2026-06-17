package validators

import (
	"regexp"
	"strings"
)

// ValidationError represents a single validation failure
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult contains the outcome of a validation run
type ValidationResult struct {
	Valid     bool              `json:"valid"`
	Errors    []ValidationError `json:"errors,omitempty"`
	Warnings  []string          `json:"warnings,omitempty"`
}

var (
	// NPC format patterns (bilingual ES|EN)
	npcLocationPattern         = regexp.MustCompile(`(?i)\*\*(?:Ubicación|Location):\*\*`)
	npcCombatStatsPattern      = regexp.MustCompile(`(?i)\*\*(?:Estadísticas de Combate|Combat Stats):\*\*`)
	npcQuestInvolvementPattern = regexp.MustCompile(`(?i)\*\*(?:Involucramiento en Quests|Quest Involvement):\*\*`)
	npcAlignmentPattern        = regexp.MustCompile(`(?i)\*\*(?:Alineamiento|Alignment):\*\*`)
	npcSecretPattern           = regexp.MustCompile(`(?i)\*\*(?:Secreto|Secret):\*\*`)

	// Bestiary format patterns (bilingual ES|EN)
	bestiaryRolePattern           = regexp.MustCompile(`(?i)\*\*(?:Rol de combate|Combat Role):\*\*`)
	bestiaryEncounterGroupPattern = regexp.MustCompile(`(?i)\*\*(?:Grupos de encuentro|Encounter Groups):\*\*`)
	bestiarySourcePattern         = regexp.MustCompile(`(?i)\*\*(?:Fuente/Referencia|Source/Reference):\*\*`)
	bestiaryTacticsPattern        = regexp.MustCompile(`(?i)###\s+(?:Tácticas Estructuradas|Structured Tactics)`)

	// Encounter format patterns (bilingual ES|EN)
	encounterTacticalMapPattern      = regexp.MustCompile(`(?i)###\s+(?:Mapa Táctico|Tactical Map)`)
	encounterConditionsPattern       = regexp.MustCompile(`(?i)###\s+(?:Condiciones y Efectos Ambientales|Conditions and Environmental Effects)`)
	encounterRoundByRoundPattern     = regexp.MustCompile(`(?i)###\s+(?:Desarrollo Round-by-Round|Round-by-Round Development)`)
	encounterAlternativeResPattern   = regexp.MustCompile(`(?i)###\s+(?:Resolución Alternativa|Alternative Resolution)`)
	encounterTemplatePattern         = regexp.MustCompile(`(?i)Template:\s*(Ambush|Defense|Assault|Social|Puzzle/Trap|Chase)`)
)

// ValidateNPCFormat checks that NPC markdown contains all required v2 fields
func ValidateNPCFormat(md string) ValidationResult {
	result := ValidationResult{Valid: true}

	if !npcAlignmentPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "alignment", Message: "missing **Alineamiento:** / **Alignment:** field"})
	}
	if !npcLocationPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "location", Message: "missing **Ubicación:** / **Location:** field"})
	}
	if !npcCombatStatsPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "combat_stats", Message: "missing **Estadísticas de Combate:** / **Combat Stats:** field"})
	}
	if !npcQuestInvolvementPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "quest_involvement", Message: "missing **Involucramiento en Quests:** / **Quest Involvement:** field"})
	}
	if !npcSecretPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "secret", Message: "missing **Secreto:** / **Secret:** field"})
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateBestiaryFormat checks that bestiary markdown contains all required v2 fields
func ValidateBestiaryFormat(md string) ValidationResult {
	result := ValidationResult{Valid: true}

	if !bestiaryRolePattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "role", Message: "missing **Rol de combate:** / **Combat Role:** field"})
	}
	if !bestiaryEncounterGroupPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "encounter_groups", Message: "missing **Grupos de encuentro:** / **Encounter Groups:** field"})
	}
	if !bestiarySourcePattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "source", Message: "missing **Fuente/Referencia:** / **Source/Reference:** field"})
	}
	if !bestiaryTacticsPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "tactics", Message: "missing ### Tácticas Estructuradas / ### Structured Tactics section"})
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateEncounterFormat checks that encounter markdown contains all required v2 fields
func ValidateEncounterFormat(md string) ValidationResult {
	result := ValidationResult{Valid: true}

	if !encounterTacticalMapPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "tactical_map", Message: "missing ### Mapa Táctico / ### Tactical Map section"})
	}
	if !encounterConditionsPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "conditions", Message: "missing ### Condiciones y Efectos Ambientales / ### Conditions and Environmental Effects section"})
	}
	if !encounterRoundByRoundPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "round_by_round", Message: "missing ### Desarrollo Round-by-Round / ### Round-by-Round Development section"})
	}
	if !encounterAlternativeResPattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "alternative_resolution", Message: "missing ### Resolución Alternativa / ### Alternative Resolution section"})
	}
	if !encounterTemplatePattern.MatchString(md) {
		result.Errors = append(result.Errors, ValidationError{Field: "template", Message: "missing encounter template (Ambush|Defense|Assault|Social|Puzzle/Trap|Chase)"})
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// CountWords counts words in a string
func CountWords(s string) int {
	fields := strings.Fields(s)
	return len(fields)
}
