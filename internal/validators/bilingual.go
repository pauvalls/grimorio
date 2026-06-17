package validators

import (
	"fmt"
	"regexp"
)

// Spanish-only markers (not shared with English)
var (
	esBoxedText     = regexp.MustCompile(`(?i)>>\s*\*\*Texto para Leer:?\*\*`)
	esIfThen        = regexp.MustCompile(`(?i)\*\*Si\s+[^:]+:\*\*`)
	esConsecuencia  = regexp.MustCompile(`(?i)\*\*Consecuencia\s+(?:inmediata|futura|a\s+largo\s+plazo):\*\*`)
	esRecuperacion  = regexp.MustCompile(`(?i)\*\*Recuperación:\*\*`)
	esNPCAlignment  = regexp.MustCompile(`(?i)\*\*Alineamiento:\*\*`)
	esNPCLocation   = regexp.MustCompile(`(?i)\*\*Ubicación:\*\*`)
	esNPCSecret     = regexp.MustCompile(`(?i)\*\*Secreto:\*\*`)
	esNPCCombat     = regexp.MustCompile(`(?i)\*\*Estadísticas de Combate:\*\*`)
	esNPCQuest      = regexp.MustCompile(`(?i)\*\*Involucramiento en Quests:\*\*`)
	esEncounter     = regexp.MustCompile(`(?i)(?:Encuentro)\s+\d+`)
	esArea          = regexp.MustCompile(`(?i)(?:Área)\s+\d+`)
)

// English-only markers (not shared with Spanish)
var (
	enBoxedText     = regexp.MustCompile(`(?i)>>\s*\*\*Read-Aloud\s+Text:?\*\*`)
	enIfThen        = regexp.MustCompile(`(?i)\*\*If\s+(the\s+PCs?\s+)?[^:]+:\*\*`)
	enConsecuencia  = regexp.MustCompile(`(?i)\*\*(?:Immediate|Future|Long-term)\s+consequence:\*\*`)
	enRecuperacion  = regexp.MustCompile(`(?i)\*\*Recovery:\*\*`)
	enNPCAlignment  = regexp.MustCompile(`(?i)\*\*Alignment:\*\*`)
	enNPCLocation   = regexp.MustCompile(`(?i)\*\*Location:\*\*`)
	enNPCSecret     = regexp.MustCompile(`(?i)\*\*Secret:\*\*`)
	enNPCCombat     = regexp.MustCompile(`(?i)\*\*Combat\s+Stats:\*\*`)
	enNPCQuest      = regexp.MustCompile(`(?i)\*\*Quest\s+Involvement:\*\*`)
	enEncounter     = regexp.MustCompile(`(?i)(?:Encounter)\s+\d+`)
	enArea          = regexp.MustCompile(`(?i)(?:Area)\s+\d+`)
)

// DetectLanguage detects the language of a markdown chapter by counting
// Spanish-only and English-only markers. Returns "es", "en", or error if mixed.
func DetectLanguage(md string) (string, error) {
	esCount := 0
	enCount := 0

	// Count Spanish markers
	esPatterns := []*regexp.Regexp{
		esBoxedText, esIfThen, esConsecuencia, esRecuperacion,
		esNPCAlignment, esNPCLocation, esNPCSecret, esNPCCombat, esNPCQuest,
		esEncounter, esArea,
	}
	for _, p := range esPatterns {
		esCount += len(p.FindAllString(md, -1))
	}

	// Count English markers
	enPatterns := []*regexp.Regexp{
		enBoxedText, enIfThen, enConsecuencia, enRecuperacion,
		enNPCAlignment, enNPCLocation, enNPCSecret, enNPCCombat, enNPCQuest,
		enEncounter, enArea,
	}
	for _, p := range enPatterns {
		enCount += len(p.FindAllString(md, -1))
	}

	if esCount > 0 && enCount > 0 {
		return "mixed", fmt.Errorf("mixed language detected: Spanish markers (%d) and English markers (%d) in same chapter", esCount, enCount)
	}

	if esCount > 0 {
		return "es", nil
	}
	if enCount > 0 {
		return "en", nil
	}

	// No markers found — default to "es" for backward compatibility
	return "es", nil
}

// BilingualPattern creates a compiled regex from paired ES|EN alternatives.
func BilingualPattern(es, en string) *regexp.Regexp {
	return regexp.MustCompile(`(?i)(?:` + es + `|` + en + `)`)
}
