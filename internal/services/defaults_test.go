package services

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// serviceProhibitedSpanish are Spanish prose tokens that MUST NOT
// appear in service default strings (encounter recommendations,
// loot tables, session context, etc.). Service-layer error messages
// and short identifiers (e.g. "combat", "social", "exploration")
// stay in English; the D&D 5e glossary table in the DM agent
// remains bilingual by spec.
var serviceProhibitedSpanish = []string{
	"Criaturas apropiadas",
	"NPC Importante",
	"Encuentro de Exploración",
	"Aparece en la sesión",
	"Puede ser encontrado",
	"Posible encuentro casual",
	"Recompensa estándar",
	"Poción de Curación",
	"Pergamino de Hechizo",
	"Arma Mágica",
	"Armadura Mágica",
	"Anillo de Protección",
	"Objeto Maravilloso",
	"Artefacto Menor",
	"Poción de Velocidad",
	"Poción de Resistencia",
	"Antídoto",
	"Varita Mágica",
	"Amuleto de Protección",
	"Encuentro genérico",
	"Interacción social con",
	"Oportunidad para obtener",
	"Descubrimiento de ruinas",
	"Exploración de cueva",
	"Navegación por territorio",
	"Investigación de fenómeno",
	"Encuentro enfocado",
	"Momento para que el partido",
	"Encuentro que combina",
	"Situación compleja",
	"Ajustar según la narrativa",
	"Interacción con NPCs",
	"Descubrimiento de ubicación",
	"Recompensa estándar",
	"Ítem consumible",
	"Suministros o consumibles",
	"Objeto mágico",
	"Objeto mágico apropiado",
	"Objeto mágico raro",
	"Prólogo",
	"Resúmenes de Sesiones",
	"Sesión ",
	"(sesión ",
	"## Prólogo",
}

// TestServicesNoSpanishDefaults verifies that the listed service files
// do not contain Spanish default strings. The list of banned tokens
// covers the prose strings that surface to DMs and players in
// encounter recommendations, loot tables, and session context. Short
// type identifiers (e.g., "combat", "social") and D&D 5e glossary
// terms are exempt.
func TestServicesNoSpanishDefaults(t *testing.T) {
	targets := []string{
		"session_generator.go",
		"area_service.go",
		"validation_engine.go",
		"dm_context_service.go",
	}

	for _, filename := range targets {
		t.Run(filename, func(t *testing.T) {
			data := readServiceFile(t, filename)
			content := string(data)
			for _, banned := range serviceProhibitedSpanish {
				assert.False(t, strings.Contains(content, banned),
					"service file %s still contains Spanish default %q — i18n-english-default regression", filename, banned)
			}
		})
	}
}

func readServiceFile(t *testing.T, filename string) []byte {
	t.Helper()
	candidates := []string{
		"./" + filename,
		"../" + filename,
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			return data
		}
	}
	t.Fatalf("could not read %s from any of %v", filename, candidates)
	return nil
}
