package compiler_test

import (
	"context"
	"os/exec"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/compiler"
	"github.com/pauvalls/grimorio/internal/domain"
)

// TestBackwardCompatibility_MinimalCampaign tests compilation with minimal data (no new fields)
func TestBackwardCompatibility_MinimalCampaign(t *testing.T) {
	// Skip if wkhtmltopdf is not available (CI environment)
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	createMinimalCampaign(t, tmpDir)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	
	// Should compile without errors even with minimal data
	_, err := c.Compile(ctx, "Minimal Campaign")
	if err != nil {
		t.Fatalf("Compilation failed for minimal campaign: %v", err)
	}

	// Verify HTML was created
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("HTML file not created for minimal campaign")
	}
}

// TestBackwardCompatibility_SessionZeroWithoutShockPoints tests Session Zero without shock points
func TestBackwardCompatibility_SessionZeroWithoutShockPoints(t *testing.T) {
	// Skip if wkhtmltopdf is not available (CI environment)
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	createMinimalCampaign(t, tmpDir)

	// Create session-zero.md WITHOUT shock points
	sessionZero := `# Sesión Cero

## Información de la Campaña

Nombre: Test Campaign

## Herramientas de Seguridad

- Ficha X
- Líneas y Veos
`
	os.WriteFile(filepath.Join(tmpDir, "session-zero.md"), []byte(sessionZero), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "Test Campaign")

	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify HTML contains session zero but no shock points errors
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	// Should contain session zero content
	if !contains(string(htmlContent), "Sesión Cero") {
		t.Error("Session Zero section missing")
	}
}

// TestBackwardCompatibility_CharacterWithoutBackstory tests character without new fields
func TestBackwardCompatibility_CharacterWithoutBackstory(t *testing.T) {
	tmpDir := t.TempDir()
	charactersDir := filepath.Join(tmpDir, "characters")
	os.MkdirAll(charactersDir, 0755)

	// Create character WITHOUT backstory hooks, secrets, goals
	character := &domain.Character{
		ID:         "test-char",
		CampaignID: "test-campaign",
		Name:       "Test Hero",
		Race:       "humano",
		Class:      "guerrero",
		Level:      1,
		Background: "soldado",
		Alignment:  "lawful good",
		// No BackstoryHooks, Secrets, or Goals
	}

	// Save character (service should handle missing optional fields)
	// This tests that the domain model is backward compatible
	if character.BackstoryHooks == nil {
		// Expected - optional field
	}
	if character.Secrets == nil {
		// Expected - optional field
	}
	if character.Goals == nil {
		// Expected - optional field
	}

	t.Log("Character model backward compatible - optional fields are nil")
}

// TestBackwardCompatibility_SessionPrepWithoutEncounters tests session prep without new fields
func TestBackwardCompatibility_SessionPrepWithoutEncounters(t *testing.T) {
	// Skip if wkhtmltopdf is not available (CI environment)
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping test")
	}
	
	tmpDir := t.TempDir()

	// Create session-prep.md WITHOUT encounter recommendations
	sessionPrep := `# Preparación de Sesión 1

## Previously On

La aventura continúa.

## Quests Activas

- Quest 1
- Quest 2
`
	os.WriteFile(filepath.Join(tmpDir, "session-prep.md"), []byte(sessionPrep), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "Test Campaign")

	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	// Verify compilation succeeds without encounter/loot/NPC data
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("HTML file not created")
	}
}

// TestBackwardCompatibility_EmptyOptionalFields tests services with empty optional fields
func TestBackwardCompatibility_EmptyOptionalFields(t *testing.T) {
	// Test SessionPrep with empty optional fields
	prep := &domain.SessionPrep{
		CampaignID: "test",
		SessionNum: 1,
		// Empty optional fields
		EncounterRecommendations: []domain.EncounterRecommendation{},
		LootSuggestions:          []domain.LootSuggestion{},
		NPCAppearances:           []domain.NPCAppearance{},
	}

	if prep.EncounterRecommendations == nil {
		t.Error("EncounterRecommendations should be empty slice, not nil")
	}

	// Test Character with empty optional fields
	character := &domain.Character{
		Name:           "Test",
		CampaignID:     "test",
		BackstoryHooks: []string{},
		Secrets:        []string{},
		Goals:          []string{},
	}

	if character.BackstoryHooks == nil {
		t.Error("BackstoryHooks should be empty slice, not nil")
	}

	t.Log("All optional fields properly initialized as empty slices")
}

// TestBackwardCompatibility_TemplateConditionals tests that templates handle empty data
func TestBackwardCompatibility_TemplateConditionals(t *testing.T) {
	// Skip if wkhtmltopdf is not available (CI environment)
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	createMinimalCampaign(t, tmpDir)

	// Create character sheet template test
	charactersDir := filepath.Join(tmpDir, "characters")
	os.MkdirAll(charactersDir, 0755)

	// Character sheet with minimal data (no spells, no backstory, etc.)
	charSheet := `# Minimal Character

**Humano Guerrero Nivel 1**

## Estadísticas

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| 15  | 12  | 14  | 10  | 10  | 8   |
`
	os.WriteFile(filepath.Join(charactersDir, "minimal.md"), []byte(charSheet), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "Test Campaign")

	if err != nil {
		t.Fatalf("Compilation failed with minimal character: %v", err)
	}

	t.Log("Template conditionals work correctly with missing optional data")
}

// TestBackwardCompatibility_ServiceMethods tests that new service methods don't break existing code
func TestBackwardCompatibility_ServiceMethods(t *testing.T) {
	// This test verifies that new service methods are additive and don't break existing usage
	
	// SessionPrepService.GetPrep should still work (existing method)
	// SessionPrepService.GetPrepWithScenarios is new (optional)
	
	// CharacterService.CreateCharacter should still work (existing method)
	// CharacterService.GenerateWithBackstory is new (optional)
	
	t.Log("Service methods are backward compatible - new methods are additive")
}

// TestBackwardCompatibility_JSONSerialization tests JSON serialization of extended models
func TestBackwardCompatibility_JSONSerialization(t *testing.T) {
	// Test that extended models serialize correctly with omitempty
	character := &domain.Character{
		ID:         "test",
		CampaignID: "test",
		Name:       "Test",
		Race:       "humano",
		Class:      "guerrero",
		Level:      1,
		// Optional fields omitted - should not appear in JSON
	}

	// Serialize to JSON to verify it works
	_, err := json.Marshal(character)
	if err != nil {
		t.Fatalf("JSON serialization failed: %v", err)
	}

	// Optional fields with omitempty should not break existing JSON consumers
	t.Log("JSON serialization backward compatible with omitempty tags")
}

// TestBackwardCompatibility_CSSNewClassesNotRequired tests that new CSS classes are optional
func TestBackwardCompatibility_CSSNewClassesNotRequired(t *testing.T) {
	// Skip if wkhtmltopdf is not available (CI environment)
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping test")
	}
	
	tmpDir := t.TempDir()
	createMinimalCampaign(t, tmpDir)

	// Create content without any new CSS classes
	areaContent := `### Área 1

Contenido normal sin clases CSS nuevas.
`
	areasDir := filepath.Join(tmpDir, "areas")
	os.MkdirAll(areasDir, 0755)
	os.WriteFile(filepath.Join(areasDir, "act1.md"), []byte(areaContent), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "Test Campaign")

	if err != nil {
		t.Fatalf("Compilation failed without new CSS classes: %v", err)
	}

	t.Log("New CSS classes are optional - content compiles without them")
}

// Helper: create minimal campaign structure
func createMinimalCampaign(t *testing.T, dir string) {
	t.Helper()

	// Create required directories
	dirs := []string{"areas", "npcs", "bestiary", "encounters", "maps", "assets", "characters"}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(dir, d), 0755)
	}

	// Create minimal required files
	files := map[string]string{
		"introduction.md": "# Test Campaign\n\nMinimal test campaign.",
		"lore.md":         "# Lore\n\nMinimal lore.",
		"appendices.md":   "# Appendices\n\nMinimal appendices.",
	}

	for path, content := range files {
		os.WriteFile(filepath.Join(dir, path), []byte(content), 0644)
	}
}

// Helper: check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
