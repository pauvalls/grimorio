package compiler

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

// executeChapterTemplate parses the embedded chapterTemplate and executes it
// against the given data, returning the output string. Used by all WU4 tests.
func executeChapterTemplate(t *testing.T, data any) string {
	t.Helper()
	tmpl, err := template.New("chapter").Parse(chapterTemplate)
	if err != nil {
		t.Fatalf("parse chapterTemplate: %v", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		t.Fatalf("execute chapterTemplate: %v", err)
	}
	return buf.String()
}

// TestChapterTemplateEnglishStruct asserts that the chapter template renders
// the English front-matter when fed English keys (Title, GameMode, Duration,
// LevelRange, Tone). Required substrings: `# Chapter 1:`, `**Game Mode:**`,
// `**Duration:**`, `**Level:**`, `**Tone:**`.
//
// We use a map so missing fields (e.g. the Spanish body keys) evaluate to
// the zero value without panicking — this lets the same template render an
// English-only campaign where the body uses Spanish-named fields that are
// just empty.
func TestChapterTemplateEnglishStruct(t *testing.T) {
	data := map[string]any{
		"ChapterNumber": 1,
		"Title":         "The Drowned Vault",
		"GameMode":      "dungeon_lineal",
		"Duration":      "1 session",
		"LevelRange":    "3-5",
		"Tone":          "mystery",
	}
	out := executeChapterTemplate(t, data)

	required := []string{
		"# Chapter 1:",
		"**Game Mode:**",
		"**Duration:**",
		"**Level:**",
		"**Tone:**",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("English front-matter missing %q in output:\n%s", s, out)
		}
	}
}

// TestChapterTemplateSpanishStruct asserts that the chapter template renders
// the Spanish front-matter when fed Spanish keys (Título, ModoDeJuego, etc.).
func TestChapterTemplateSpanishStruct(t *testing.T) {
	data := map[string]any{
		"ChapterNumber": 1,
		"Título":        "La Bóveda Ahogada",
		"ModoDeJuego":   "dungeon_lineal",
		"Duración":      "1 sesión",
		"RangoDeNivel":  "3-5",
		"Tono":          "misterio",
	}
	out := executeChapterTemplate(t, data)

	required := []string{
		"# Capítulo 1:",
		"**Modo de Juego:**",
		"**Duración:**",
		"**Rango de Nivel:**",
		"**Tono:**",
	}
	for _, s := range required {
		if !strings.Contains(out, s) {
			t.Errorf("Spanish front-matter missing %q in output:\n%s", s, out)
		}
	}
}

// TestChapterTemplateSevenPlaceholders asserts that the embedded
// chapterTemplate references the seven WotC block hooks. The check is
// permissive: it accepts the placeholder as a top-level field reference
// (`{{.X}}`, `{{.X ...}}`, `{{.X}}`, `{{range .X}}`, `{{with .X}}`, etc.).
func TestChapterTemplateSevenPlaceholders(t *testing.T) {
	placeholders := []string{
		".RoleplayCues",
		".DMSidebars",
		".RandomEncounterTables",
		".Traps",
		".FactionTracker",
		".WhatsNext",
		".XPPPJ",
	}
	missing := []string{}
	for _, p := range placeholders {
		if !strings.Contains(chapterTemplate, p) {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		t.Errorf("chapterTemplate missing placeholders: %v", missing)
	}
}

// TestChapterTemplateGeneralFeaturesRendersSection asserts that when data
// includes a non-nil GeneralFeatures with Content, the template renders a
// "General Features" section with a <div class="general-features"> wrapper
// before the Areas section.
func TestChapterTemplateGeneralFeaturesRendersSection(t *testing.T) {
	data := map[string]any{
		"ChapterNumber": 1,
		"Title":         "The Drowned Vault",
		"GameMode":      "dungeon_lineal",
		"Duration":      "1 session",
		"LevelRange":    "3-5",
		"Tone":          "mystery",
		"GeneralFeatures": map[string]any{
			"Content": "***Ceilings.*** 30 feet high, vaulted stone.\n***Light.*** Dim torchlight.",
		},
	}
	out := executeChapterTemplate(t, data)

	if !strings.Contains(out, "## General Features") {
		t.Error("expected '## General Features' heading in output")
	}
	if !strings.Contains(out, `<div class="general-features">`) {
		t.Error("expected '<div class=\"general-features\">' wrapper in output")
	}
	if !strings.Contains(out, "***Ceilings.***") {
		t.Error("expected GeneralFeatures content in output")
	}
	// Verify General Features appears before Areas
	gfIdx := strings.Index(out, "## General Features")
	areasIdx := strings.Index(out, "## Areas")
	if gfIdx >= 0 && areasIdx >= 0 && gfIdx >= areasIdx {
		t.Error("General Features section must appear before Areas section")
	}
}

// TestChapterTemplateGeneralFeaturesNilOmitsSection asserts that when
// GeneralFeatures is nil (or absent), the template does NOT render any
// "General Features" section (backward compatible).
func TestChapterTemplateGeneralFeaturesNilOmitsSection(t *testing.T) {
	data := map[string]any{
		"ChapterNumber": 1,
		"Title":         "The Drowned Vault",
		"GameMode":      "dungeon_lineal",
		"Duration":      "1 session",
		"LevelRange":    "3-5",
		"Tone":          "mystery",
	}
	out := executeChapterTemplate(t, data)

	if strings.Contains(out, "## General Features") {
		t.Error("nil GeneralFeatures should NOT produce '## General Features' section")
	}
	if strings.Contains(out, "general-features") {
		t.Error("nil GeneralFeatures should NOT produce 'general-features' div")
	}
}
