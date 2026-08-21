package compiler_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/compiler"
)

// TestCompileWithAllFeatures tests PDF compilation with all new features enabled
func TestCompileWithAllFeatures(t *testing.T) {
	// Skip if no PDF engine is available
	if !compiler.IsPDFEngineAvailable() {
		t.Skip("No PDF engine available, skipping integration test")
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
	c := compiler.New(tmpDir, "")
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

func TestCompile_NestedCalloutTempCampaignPreservesBlockSemantics(t *testing.T) {
	if !compiler.IsPDFEngineAvailable() {
		t.Skip("No PDF engine available, skipping integration test")
	}

	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Nested Callout Test")
	chaptersDir := filepath.Join(tmpDir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		t.Fatal(err)
	}
	chapter := `# Nested Card

> **Read-Aloud:** The archive is silent.
>
> | Sign | Meaning |
> | --- | --- |
> | Dust | A hidden door |
>
> - Search the shelves
> - [Open the door](#door)
>
> ` + "```text" + `
> key <door>
> ` + "```" + `
`
	if err := os.WriteFile(filepath.Join(chaptersDir, "nested.md"), []byte(chapter), 0644); err != nil {
		t.Fatal(err)
	}

	pdfPath, err := compiler.New(tmpDir, "").Compile(t.Context(), "Nested Callout Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if info, err := os.Stat(pdfPath); err != nil || info.Size() == 0 {
		t.Fatalf("compiled PDF is missing or empty: %v", err)
	}
	htmlContent, err := os.ReadFile(filepath.Join(tmpDir, "campaign.html"))
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}
	htmlStr := string(htmlContent)
	for _, want := range []string{"<table>", "<ul>", `<a href="#door">Open the door</a>`, "key &lt;door&gt;"} {
		if !strings.Contains(htmlStr, want) {
			t.Errorf("nested callout HTML missing %q", want)
		}
	}
	if strings.Contains(htmlStr, "| Sign | Meaning |") || strings.Contains(htmlStr, "```text") {
		t.Errorf("nested callout leaked Markdown markers into HTML")
	}
}

// TestCompileWithDMSidebar tests DM sidebar generation
func TestCompileWithDMSidebar(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "DM Sidebar Test")

	// Create chapters directory with DM sidebar content (legacy areas/ dir
	// was removed in v5.0.2 WU7; chapters/ is the only chapter source).
	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)

	chapterContent := `### Área 1: Entrada

<div class="dm-sidebar">
<h5>DM Tip</h5>
<p>Los jugadores pueden encontrar una trampa aquí.</p>
<h5>Secreto</h5>
<p>Hay un pasaje oculto detrás del trono.</p>
</div>

Contenido normal del área.
`
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter1.md"), []byte(chapterContent), 0644)

	ctx := context.Background()
	c := compiler.New(tmpDir, "")
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
	c := compiler.New(tmpDir, "")
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
	c := compiler.New(tmpDir, "")
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

// TestCompileWithChapters tests compilation with new chapters/ structure
func TestCompileWithChapters(t *testing.T) {
	// Skip if no PDF engine is available
	if !compiler.IsPDFEngineAvailable() {
		t.Skip("No PDF engine available, skipping integration test")
	}

	tmpDir := t.TempDir()

	// Create minimal campaign with chapters/ instead of areas/
	createTestCampaignWithChapters(t, tmpDir, "Chapters Test Campaign")

	// Compile
	ctx := context.Background()
	c := compiler.New(tmpDir, "")
	pdfPath, err := c.Compile(ctx, "Chapters Test Campaign")

	if err != nil {
		t.Fatalf("Compile with chapters/ failed: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("PDF file not created at %s", pdfPath)
	}

	// Verify HTML contains chapter content
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlContent, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}

	htmlStr := string(htmlContent)

	// Check for chapter content
	if !strings.Contains(htmlStr, "Área 1") {
		t.Error("HTML missing Área 1 from chapter content")
	}
	if !strings.Contains(htmlStr, "chapter_01") {
		// Chapter files should be included
		t.Log("Note: chapter filename may not appear in HTML directly")
	}
}

// TestCompile_DuplicatedImagesAcrossFiles tests that compilation succeeds when
// the same image is referenced multiple times across different markdown files.
// This is the exact scenario from mareas-oscuras-v2 with separator.svg.
func TestCompile_DuplicatedImagesAcrossFiles(t *testing.T) {
	if !compiler.IsPDFEngineAvailable() {
		t.Skip("No PDF engine available, skipping integration test")
	}

	tmpDir := t.TempDir()

	// Create campaign structure
	dirs := []string{"chapters", "npcs", "bestiary", "encounters", "maps", "assets"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	// Create a minimal SVG file
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="10"><rect width="100" height="10"/></svg>`
	if err := os.WriteFile(filepath.Join(tmpDir, "assets", "separator.svg"), []byte(svgContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create lore.md referencing separator.svg twice
	lore := "# Lore\n\n![Separador](assets/separator.svg)\n\nSome lore text.\n\n![Separador](assets/separator.svg)\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte(lore), 0644); err != nil {
		t.Fatal(err)
	}

	// Create chapter1.md referencing separator.svg twice
	chapter := "# Chapter 1\n\n![Separador](assets/separator.svg)\n\nChapter content.\n\n![Separador](assets/separator.svg)\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "chapters", "chapter1.md"), []byte(chapter), 0644); err != nil {
		t.Fatal(err)
	}

	c := compiler.New(tmpDir, "")
	pdfPath, err := c.Compile(context.Background(), "Test Duplicated Images")
	if err != nil {
		t.Fatalf("Compile() should succeed with duplicated images, got error: %v", err)
	}
	if pdfPath == "" {
		t.Fatal("Compile() should return non-empty pdfPath")
	}
	if !strings.HasSuffix(pdfPath, "campaign.pdf") {
		t.Errorf("pdfPath = %q, should end with campaign.pdf", pdfPath)
	}

	// Verify PDF file was actually created
	if _, err := os.Stat(pdfPath); err != nil {
		t.Errorf("PDF file should exist at %q: %v", pdfPath, err)
	}
}

// Helper function to create minimal test campaign with chapters/ structure
func createTestCampaignWithChapters(t *testing.T, dir, name string) {
	t.Helper()

	// Create required directories (no areas/)
	dirs := []string{"chapters", "npcs", "bestiary", "encounters", "maps", "assets"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", d, err)
		}
	}

	// Create introduction.md
	intro := `# ` + name + `

Esta es una campaña de prueba con capítulos.
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

	// Create a chapter file
	chapterContent := `# Capítulo 1: El Comienzo

## Apertura Narrativa

Los héroes llegan al pueblo.

## Áreas

### Área 1: Entrada del Pueblo

> Los jugadores ven el pueblo desde la colina.

La entrada está custodiada por guardias.

### Área 2: La Taberna

> Humo y risas salen de la taberna.

Un lugar acogedor para descansar.

## Consecuencias y Transición

El pueblo queda a salvo.
`
	_ = os.WriteFile(filepath.Join(dir, "chapters", "chapter_01.md"), []byte(chapterContent), 0644)
}

// Helper function to create minimal test campaign
func createTestCampaign(t *testing.T, dir, name string) {
	t.Helper()

	// Create required directories (grimorio-areas removed in v5.0.2 WU7)
	dirs := []string{"chapters", "npcs", "bestiary", "encounters", "maps", "assets"}
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

func TestCompile_BlockquoteClassification(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "BQ Test")

	intro := `# Introduction

<!-- introduction-sidebar -->
> ##### Sidebar Name
> Optional rules detail.

> *Read this aloud.*
`
	_ = os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte(intro), 0644)

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	chapter := `# Chapter 1

> - Level 1-2
> - 2-3 hours

### Area 1: Entrance

> **Read-Aloud Text:** The door creaks open.

> ##### DM Sidebar: Traps
> The floor is trapped.
`
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(chapter), 0644)

	c := compiler.New(tmpDir, "")
	_, err := c.Compile(context.Background(), "BQ Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML: %v", err)
	}
	htmlStr := string(htmlBytes)

	checks := []struct {
		class   string
		content string
	}{
		{"chapter-summary", "Level 1-2"},
		{"read-aloud", "The door creaks open"},
		{"dm-sidebar", "The floor is trapped"},
		{"introduction-sidebar", "Optional rules detail"},
	}

	for _, check := range checks {
		if !strings.Contains(htmlStr, `class="`+check.class+`"`) {
			t.Errorf("HTML missing %s block", check.class)
		}
		if !strings.Contains(htmlStr, check.content) {
			t.Errorf("HTML missing content %q for %s", check.content, check.class)
		}
	}

	// Prefix labels should be stripped.
	if strings.Contains(htmlStr, "Read-Aloud Text:") {
		t.Error("Read-Aloud Text prefix should be stripped")
	}
	if strings.Contains(htmlStr, "DM Sidebar:") {
		t.Error("DM Sidebar prefix should be stripped")
	}
}

func TestCompile_MarkdownLinks(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Link Test")

	appendices := `# Appendices

<a id="appendix-a-magic-items"></a>
## Appendix A: Magic Items

*Magic items.*
`
	_ = os.WriteFile(filepath.Join(tmpDir, "appendices.md"), []byte(appendices), 0644)

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	chapter := `# Chapter 1

### Area 1: Entrance

See [Background](#adventure-background) and [Appendix A](appendices.md#appendix-a-magic-items).
`
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(chapter), 0644)

	c := compiler.New(tmpDir, "")
	_, err := c.Compile(context.Background(), "Link Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlBytes, _ := os.ReadFile(htmlPath)
	htmlStr := string(htmlBytes)

	if !strings.Contains(htmlStr, `<a href="#adventure-background">Background</a>`) {
		t.Errorf("internal link not converted, got: %s", htmlStr)
	}
	if !strings.Contains(htmlStr, `<a href="#appendix-a-magic-items">Appendix A</a>`) {
		t.Errorf("cross-file link not rewritten, got: %s", htmlStr)
	}
}

func TestCompile_WorksheetSuppression(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Worksheet Test")

	sessionZero := `# Session Zero

<div class="character-worksheet">
<div class="worksheet-section">
<h4>Prompt</h4>
<div class="prompt-box">Question?</div>
</div>
</div>
`
	_ = os.WriteFile(filepath.Join(tmpDir, "session-zero.md"), []byte(sessionZero), 0644)

	c := compiler.New(tmpDir, "")
	_, err := c.Compile(context.Background(), "Worksheet Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlBytes, _ := os.ReadFile(htmlPath)
	htmlStr := string(htmlBytes)

	if strings.Contains(htmlStr, `class="character-worksheet"`) {
		t.Error("character-worksheet block should be suppressed in PDF output")
	}
}

func TestCompile_ImageCoverage(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Image Test")

	_ = os.MkdirAll(filepath.Join(tmpDir, "assets"), 0755)
	for _, name := range []string{"sz.png", "intro.png", "setting.png", "appendix.png"} {
		_ = os.WriteFile(filepath.Join(tmpDir, "assets", name), []byte("fake"), 0644)
	}

	_ = os.WriteFile(filepath.Join(tmpDir, "session-zero.md"), []byte("# SZ\n\n![SZ](assets/sz.png)"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte("# Intro\n\n![Intro](assets/intro.png)"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "setting-guide.md"), []byte("# Setting\n\n![Setting](assets/setting.png)"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "appendices.md"), []byte("# Appendices\n\n![Appendix](assets/appendix.png)"), 0644)

	c := compiler.New(tmpDir, "")
	_, err := c.Compile(context.Background(), "Image Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlBytes, _ := os.ReadFile(htmlPath)
	htmlStr := string(htmlBytes)

	for _, alt := range []string{"SZ", "Intro", "Setting", "Appendix"} {
		if !strings.Contains(htmlStr, `alt="`+alt+`"`) {
			t.Errorf("image %q missing from HTML", alt)
		}
	}
}

func TestCompile_CrossReferenceAnchors(t *testing.T) {
	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "Anchor Test")

	appendices := `# Appendices

<a id="appendix-a-magic-items"></a>
## Appendix A: Magic Items

*Magic items.*
`
	_ = os.WriteFile(filepath.Join(tmpDir, "appendices.md"), []byte(appendices), 0644)

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	chapter := `# Chapter 1

### Area 5: The Crypt

Content.
`
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(chapter), 0644)

	c := compiler.New(tmpDir, "")
	_, err := c.Compile(context.Background(), "Anchor Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	htmlPath := filepath.Join(tmpDir, "campaign.html")
	htmlBytes, _ := os.ReadFile(htmlPath)
	htmlStr := string(htmlBytes)

	if !strings.Contains(htmlStr, `id="area-5"`) {
		t.Error("area heading missing stable area-5 id")
	}
	if !strings.Contains(htmlStr, `id="appendix-a-magic-items"`) {
		t.Error("explicit appendix anchor missing")
	}
}

func TestCompile_PDFEngineAgnostic(t *testing.T) {
	if !compiler.IsPDFEngineAvailable() {
		t.Skip("No PDF engine available, skipping engine smoke test")
	}

	tmpDir := t.TempDir()
	createTestCampaign(t, tmpDir, "PDF Smoke Test")

	c := compiler.New(tmpDir, "")
	pdfPath, err := c.Compile(context.Background(), "PDF Smoke Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	info, err := os.Stat(pdfPath)
	if err != nil {
		t.Fatalf("PDF file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("PDF file is empty")
	}
}

// TestCompile_TableIntegrity asserts REQ-3.5 (end-to-end): a 10-row
// 3-column table in a fixture campaign renders with all 10 rows
// visible in the compiled PDF, verified via `pdftotext -layout`.
// The test is gated on `compiler.IsPDFEngineAvailable()` so it
// skips cleanly when Chromium is not in PATH. The test is also
// skipped if `pdftotext` (poppler-utils) is not installed.
func TestCompile_TableIntegrity(t *testing.T) {
	if !compiler.IsPDFEngineAvailable() {
		t.Skip("Chromium not in PATH, skipping table integrity test")
	}

	dir := t.TempDir()

	// Minimal campaign structure (chapters/ is the canonical source per
	// v5.0.2 WU7 removal of legacy areas/).
	dirs := []string{"chapters", "npcs", "bestiary", "encounters", "maps", "assets"}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "introduction.md"), []byte("# Table Test\n\nIntro.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "lore.md"), []byte("# Lore\n\nLore.\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "appendices.md"), []byte("# Appendices\n\nAppx.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// 10-row 3-column table in a chapter file. The first column of
	// each row is a unique marker ("Row N Col1" for N=1..10) so we
	// can verify all 10 rows survived the page-break logic.
	chapter := "# Chapter 1\n\n## Table Test\n\n| Col1 | Col2 | Col3 |\n|------|------|------|\n"
	for i := 1; i <= 10; i++ {
		chapter += fmt.Sprintf("| Row %d Col1 | Row %d Col2 | Row %d Col3 |\n", i, i, i)
	}
	chapter += "\nEnd of chapter.\n"
	if err := os.WriteFile(filepath.Join(dir, "chapters", "tables.md"), []byte(chapter), 0644); err != nil {
		t.Fatal(err)
	}

	// Compile end-to-end.
	c := compiler.New(dir, "")
	pdfPath, err := c.Compile(context.Background(), "Table Test")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}
	if _, err := os.Stat(pdfPath); err != nil {
		t.Fatalf("PDF not created at %s: %v", pdfPath, err)
	}

	// Run pdftotext on the PDF. If pdftotext is not in PATH, skip.
	pdftotext, err := exec.LookPath("pdftotext")
	if err != nil {
		t.Skip("pdftotext not in PATH, skipping text extraction verification")
	}
	out, err := exec.Command(pdftotext, "-layout", pdfPath, "-").Output()
	if err != nil {
		t.Skipf("pdftotext failed (cannot verify row markers): %v", err)
	}
	text := string(out)

	// Each row's first column is a unique marker "Row N Col1" for N=1..10.
	// All 10 must appear in the extracted PDF text.
	seen := make(map[string]bool)
	for i := 1; i <= 10; i++ {
		marker := fmt.Sprintf("Row %d Col1", i)
		if strings.Contains(text, marker) {
			seen[marker] = true
		}
	}
	if len(seen) < 10 {
		missing := []string{}
		for i := 1; i <= 10; i++ {
			marker := fmt.Sprintf("Row %d Col1", i)
			if !seen[marker] {
				missing = append(missing, marker)
			}
		}
		t.Errorf("PDF missing %d of 10 row markers: %v", 10-len(seen), missing)
	}
}
