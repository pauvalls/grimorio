package compiler_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/compiler"
)

// TestCompileWithAllFeatures tests PDF compilation with all new features enabled
func TestCompileWithAllFeatures(t *testing.T) {
	// Skip if wkhtmltopdf is not available
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping integration test")
	}

	tmpDir := t.TempDir()

	// Create minimal campaign structure
	createTestCampaign(t, tmpDir, "Test Campaign")

	// Add session-zero.md with shock points
	sessionZero := `# Sesión Cero

## Puntos de Shock

<div class="shock-point moderate">
<span class="severity-badge">moderate</span>
<strong>Violencia</strong>: Combate fantástico
</div>
`
	_ = os.WriteFile(filepath.Join(tmpDir, "session-zero.md"), []byte(sessionZero), 0644)

	// Add session-prep.md
	sessionPrep := `# Preparación de Sesión 1

## Previously On
La aventura comienza.

## Recomendaciones de Encuentros

<div class="encounter-recommendation">
<span class="cr-badge">CR 1</span>
<span class="encounter-type">combat</span>
<strong>Combate</strong>
</div>
`
	_ = os.WriteFile(filepath.Join(tmpDir, "session-prep.md"), []byte(sessionPrep), 0644)

	// Create characters directory with a character sheet
	charactersDir := filepath.Join(tmpDir, "characters")
	_ = os.MkdirAll(charactersDir, 0755)

	characterSheet := `# Hoja de Personaje: Test Hero

**Humano Guerrero Nivel 1**

## Estadísticas

| STR | DEX | CON | INT | WIS | CHA |
|-----|-----|-----|-----|-----|-----|
| 15  | 12  | 14  | 10  | 10  | 8   |

## Backstory

<div class="character-worksheet">
<div class="worksheet-section">
<h4>Ganchos de Historia</h4>
<div class="prompt-box">Veterano de guerra</div>
</div>
</div>
`
	_ = os.WriteFile(filepath.Join(charactersDir, "test-hero.md"), []byte(characterSheet), 0644)

	// Compile
	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	pdfPath, err := c.Compile(ctx, "Test Campaign")

	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("PDF file not created at %s", pdfPath)
	}

	// Verify HTML was created and contains expected sections
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	htmlStr := string(htmlContent)

	// Check for new CSS classes
	expectedClasses := []string{
		"shock-point",
		"severity-badge",
		"encounter-recommendation",
		"cr-badge",
		"character-worksheet",
		"worksheet-section",
	}

	for _, class := range expectedClasses {
		if !strings.Contains(htmlStr, class) {
			t.Errorf("HTML missing expected CSS class: %s", class)
		}
	}

	// Check for session zero section
	if !strings.Contains(htmlStr, "Sesión Cero") {
		t.Error("HTML missing Session Zero section")
	}

	// Check for character sheet section
	if !strings.Contains(htmlStr, "Test Hero") {
		t.Error("HTML missing character sheet")
	}
}

// TestCompileWithDMSidebar tests DM sidebar generation
func TestCompileWithDMSidebar(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "DM Sidebar Test")

	// Create areas directory with DM sidebar content
	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)

	areaContent := `### Área 1: Entrada

<div class="dm-sidebar">
<h5>DM Tip</h5>
<p>Los jugadores pueden encontrar una trampa aquí.</p>
<h5>Secreto</h5>
<p>Hay un pasaje oculto detrás del trono.</p>
</div>

Contenido normal del área.
`
	_ = os.WriteFile(filepath.Join(areasDir, "act1.md"), []byte(areaContent), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "DM Sidebar Test")

	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Verify HTML contains DM sidebar
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	if !strings.Contains(string(htmlContent), "dm-sidebar") {
		t.Error("HTML missing DM sidebar class")
	}

	if !strings.Contains(string(htmlContent), "DM Tip") {
		t.Error("HTML missing DM tip content")
	}
}

// TestCompileWithStatBlockV2 tests enhanced stat block rendering
func TestCompileWithStatBlockV2(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Stat Block V2 Test")

	// Create bestiary with v2 stat block
	bestiaryDir := filepath.Join(tmpDir, "bestiary")
	_ = os.MkdirAll(bestiaryDir, 0755)

	statBlock := `### Goblin

<div class="stat-block-v2">
<h3>Goblin</h3>
<div class="stat-line">
<span class="stat-label">AC</span>
<span class="stat-value">15</span>
</div>
<div class="stat-line">
<span class="stat-label">HP</span>
<span class="stat-value">7</span>
</div>
</div>
`
	_ = os.WriteFile(filepath.Join(bestiaryDir, "goblin.md"), []byte(statBlock), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "Stat Block V2 Test")

	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Verify HTML contains stat-block-v2
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	if !strings.Contains(string(htmlContent), "stat-block-v2") {
		t.Error("HTML missing stat-block-v2 class")
	}

	if !strings.Contains(string(htmlContent), "stat-line") {
		t.Error("HTML missing stat-line class")
	}
}

// TestCompileBackwardCompatibility tests that campaigns without new features still compile
func TestCompileBackwardCompatibility(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Backward Compatibility Test")

	// Minimal campaign with no new features
	ctx := context.Background()
	c := compiler.New(tmpDir, "wkhtmltopdf")
	_, err := c.Compile(ctx, "Backward Compatibility Test")

	if err != nil {
		t.Fatalf("Compile failed for minimal campaign: %v", err)
	}

	// Verify PDF was created
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Error("HTML file not created for minimal campaign")
	}
}

// Helper function to create minimal test campaign
func createTestCampaign(t *testing.T, dir, name string) {
	t.Helper()

	// Create required directories
	dirs := []string{"areas", "npcs", "bestiary", "encounters", "maps", "assets"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	// Create introduction.md
	intro := `# ` + name + `

Esta es una campaña de prueba.
`
	_ = os.WriteFile(filepath.Join(dir, "introduction.md"), []byte(intro), 0644)

	// Create lore.md
	lore := `# Lore y Ambientación

Este es el lore de la campaña de prueba.
`
	_ = os.WriteFile(filepath.Join(dir, "lore.md"), []byte(lore), 0644)

	// Create appendices.md
	appendices := `# Appendices

Apéndices de la campaña.
`
	_ = os.WriteFile(filepath.Join(dir, "appendices.md"), []byte(appendices), 0644)
}
