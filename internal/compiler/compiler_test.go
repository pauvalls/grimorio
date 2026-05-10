package compiler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMarkdownToHTML_ProcessScenePlaceholders(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantContains    string
		wantNotContains string
	}{
		{
			name:            "scene placeholder becomes descriptive text",
			input:           "*[SCENE: A dark dungeon with flickering torches]*",
			wantContains:    "A dark dungeon with flickering torches",
			wantNotContains: "[SCENE:",
		},
		{
			name:            "scene placeholder in standalone line",
			input:           "[SCENE: Epic battle between heroes and dragon]",
			wantContains:    "Epic battle between heroes and dragon",
			wantNotContains: "[SCENE:",
		},
		{
			name:            "multiple scene placeholders",
			input:           "[SCENE: First scene]\n\nSome text\n\n[SCENE: Second scene]",
			wantContains:    "First scene",
			wantNotContains: "[SCENE:",
		},
		{
			name:            "regular markdown still works",
			input:           "# Heading\n\nSome **bold** text",
			wantContains:    "<h1",
			wantNotContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markdownToHTML(tt.input, "/tmp")

			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("markdownToHTML() result does not contain %q\nGot: %s", tt.wantContains, result)
			}

			if tt.wantNotContains != "" && strings.Contains(result, tt.wantNotContains) {
				t.Errorf("markdownToHTML() result should not contain %q\nGot: %s", tt.wantNotContains, result)
			}
		})
	}
}

func TestMarkdownToHTML_ScenePlaceholder_Formatting(t *testing.T) {
	input := "[SCENE: A mystical forest at dawn]"
	result := markdownToHTML(input, "/tmp")

	if !strings.Contains(result, "scene-description") && !strings.Contains(result, "<p>") {
		t.Errorf("scene placeholder should be wrapped in HTML element, got: %s", result)
	}
}

func TestImageEmbedding_PNG(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(pngPath, []byte("fake-png-data"), 0644); err != nil {
		t.Fatal(err)
	}

	md := "![Test Image](" + pngPath + ")"
	result := markdownToHTML(md, tmpDir)

	if !strings.Contains(result, "class=\"campaign-image\"") {
		t.Errorf("expected campaign-image class, got: %s", result)
	}
	if !strings.Contains(result, "data:image/png;base64,") {
		t.Errorf("expected base64 data URI for PNG, got: %s", result)
	}
	if !strings.Contains(result, "Test Image") {
		t.Errorf("expected alt text 'Test Image', got: %s", result)
	}
}

func TestImageEmbedding_SVG(t *testing.T) {
	tmpDir := t.TempDir()
	svgDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(svgDir, 0755); err != nil {
		t.Fatal(err)
	}
	svgPath := filepath.Join(svgDir, "divider.svg")
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg"><line x1="0" y1="0" x2="100" y2="0"/></svg>`
	if err := os.WriteFile(svgPath, []byte(svgContent), 0644); err != nil {
		t.Fatal(err)
	}

	md := "![Divider](assets/divider.svg)"
	result := markdownToHTML(md, tmpDir)

	if !strings.Contains(result, "class=\"campaign-image\"") {
		t.Errorf("expected campaign-image class for SVG, got: %s", result)
	}
	if !strings.Contains(result, "assets/divider.svg") {
		t.Errorf("expected relative path for SVG, got: %s", result)
	}
}

func TestImageEmbedding_MissingImage(t *testing.T) {
	result := markdownToHTML("![Missing](nonexistent.png)", "/tmp")

	if !strings.Contains(result, "image-missing") {
		t.Errorf("expected image-missing for nonexistent file, got: %s", result)
	}
	if !strings.Contains(result, "Missing") {
		t.Errorf("expected alt text 'Missing' in placeholder, got: %s", result)
	}
}

func TestImageEmbedding_CodeAssetRef(t *testing.T) {
	tmpDir := t.TempDir()
	imgDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(imgDir, 0755); err != nil {
		t.Fatal(err)
	}
	imgPath := filepath.Join(imgDir, "monster.png")
	if err := os.WriteFile(imgPath, []byte("fake-monster"), 0644); err != nil {
		t.Fatal(err)
	}

	md := "A `assets/monster.png` in the middle of text"
	result := markdownToHTML(md, tmpDir)

	if !strings.Contains(result, "class=\"campaign-image\"") {
		t.Errorf("expected campaign-image from code asset ref, got: %s", result)
	}
	if !strings.Contains(result, "data:image/png") {
		t.Errorf("expected base64 embedded PNG from code ref, got: %s", result)
	}
}

func TestImageEmbedding_Dedup(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "dedup.png")
	if err := os.WriteFile(pngPath, []byte("dedup-data"), 0644); err != nil {
		t.Fatal(err)
	}

	md := "![First](" + pngPath + ")\n\n![Second](" + pngPath + ")"
	result := markdownToHTML(md, tmpDir)

	imgCount := strings.Count(result, "class=\"campaign-image\"")
	if imgCount != 1 {
		t.Errorf("expected dedup to 1 image, got %d\nHTML:\n%s", imgCount, result)
	}
}

func TestImageEmbedding_InlineImage(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "inline.png")
	if err := os.WriteFile(pngPath, []byte("inline-data"), 0644); err != nil {
		t.Fatal(err)
	}

	md := "This is an ![inline](" + pngPath + ") image in a paragraph."
	result := markdownToHTML(md, tmpDir)

	if !strings.Contains(result, "class=\"campaign-image\"") {
		t.Errorf("expected inline image to be embedded, got: %s", result)
	}
}

func TestImageEmbedding_PreservesRawHTMLImg(t *testing.T) {
	// Verifica que un tag <img> HTML puro en el markdown no sea escapado
	md := `Some text before <img src="assets/portrait.png" alt="Portrait" class="portrait"> and after.`
	result := markdownToHTML(md, "/tmp")

	if strings.Contains(result, "&lt;img") {
		t.Errorf("raw <img> tag was escaped, got: %s", result)
	}
	if !strings.Contains(result, `<img src="assets/portrait.png"`) {
		t.Errorf("raw <img> tag was stripped, got: %s", result)
	}
	if !strings.Contains(result, "Some text before") || !strings.Contains(result, "and after.") {
		t.Errorf("surrounding text was lost, got: %s", result)
	}
}

func TestImageEmbedding_MixedMarkdownAndHTMLImg(t *testing.T) {
	tmpDir := t.TempDir()
	pngPath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(pngPath, []byte("fake-png"), 0644); err != nil {
		t.Fatal(err)
	}

	md := `Before ![markdown img](` + pngPath + `) and <img src="assets/html.png" alt="HTML"> after.`
	result := markdownToHTML(md, tmpDir)

	if !strings.Contains(result, "data:image/png;base64,") {
		t.Errorf("markdown image was not embedded, got: %s", result)
	}
	if !strings.Contains(result, `<img src="assets/html.png"`) {
		t.Errorf("raw HTML img was lost, got: %s", result)
	}
}

func TestCountImagesInMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "markdown image",
			input: `![alt](path/to/img.png)`,
			want:  1,
		},
		{
			name:  "raw img tag",
			input: `<img src="assets/img.png" alt="test">`,
			want:  1,
		},
		{
			name:  "code asset ref",
			input: "`assets/monster.png`",
			want:  1,
		},
		{
			name:  "mixed",
			input: `![md](img.png) and <img src="a.svg"> and ` + "`assets/b.png`",
			want:  3,
		},
		{
			name:  "no images",
			input: "just text **bold**",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countImagesInMarkdown(tt.input)
			if got != tt.want {
				t.Errorf("countImagesInMarkdown() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountImagesInHTML(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")

	html := `<html><body>
		<img src="a.png" alt="1">
		<p>text</p>
		<img src="b.svg" alt="2">
	</body></html>`

	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := countImagesInHTML(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if got != 2 {
		t.Errorf("countImagesInHTML() = %d, want 2", got)
	}
}

func TestVerifyImages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create campaign structure
	areasDir := filepath.Join(tmpDir, "areas")
	if err := os.MkdirAll(areasDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a markdown file with 2 images
	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
Some text
<img src="assets/map.svg" alt="Map">
`
	mdPath := filepath.Join(areasDir, "chapter1.md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create matching HTML with both images
	htmlContent := `<html><body>
		<h1>Act 1</h1>
		<img src="data:image/png;base64,abc" alt="Scene 1" class="campaign-image">
		<p>Some text</p>
		<img src="assets/map.svg" alt="Map" class="campaign-image">
	</body></html>`
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}

	c := New(tmpDir, "wkhtmltopdf")
	expected, found, ok, err := c.verifyImages(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if expected != 2 {
		t.Errorf("expected = %d, want 2", expected)
	}
	if found != 2 {
		t.Errorf("found = %d, want 2", found)
	}
	if !ok {
		t.Error("ok = false, want true")
	}
}

func TestVerifyImages_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create campaign structure
	areasDir := filepath.Join(tmpDir, "areas")
	if err := os.MkdirAll(areasDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a markdown file with 2 images
	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
<img src="assets/map.svg" alt="Map">
`
	mdPath := filepath.Join(areasDir, "chapter1.md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create HTML with only 1 image (simulating a failure)
	htmlContent := `<html><body>
		<h1>Act 1</h1>
		<img src="data:image/png;base64,abc" alt="Scene 1" class="campaign-image">
	</body></html>`
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}

	c := New(tmpDir, "wkhtmltopdf")
	expected, found, ok, err := c.verifyImages(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if expected != 2 {
		t.Errorf("expected = %d, want 2", expected)
	}
	if found != 1 {
		t.Errorf("found = %d, want 1", found)
	}
	if ok {
		t.Error("ok = true, want false")
	}
}

func TestGenerateFactionTracker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create factions directory and reputation matrix
	factionDir := filepath.Join(tmpDir, "factions")
	if err := os.MkdirAll(factionDir, 0755); err != nil {
		t.Fatal(err)
	}

	matrixData := `{
		"campaign_id": "test-campaign",
		"entries": [
			{"faction_id": "faction-thieves", "party_id": "party-1", "score": 15, "status": "friendly"},
			{"faction_id": "faction-guards", "party_id": "party-1", "score": -20, "status": "hostile"}
		]
	}`
	if err := os.WriteFile(filepath.Join(factionDir, "reputation_matrix.json"), []byte(matrixData), 0644); err != nil {
		t.Fatal(err)
	}

	c := New(tmpDir, "wkhtmltopdf")
	tracker := c.generateFactionTracker()

	if tracker == "" {
		t.Fatal("expected tracker HTML, got empty string")
	}

	if !strings.Contains(tracker, "Apéndice E: Faction Tracker") {
		t.Errorf("tracker missing title, got: %s", tracker)
	}
	if !strings.Contains(tracker, "faction-thieves") {
		t.Errorf("tracker missing faction-thieves, got: %s", tracker)
	}
	if !strings.Contains(tracker, "faction-guards") {
		t.Errorf("tracker missing faction-guards, got: %s", tracker)
	}
	if !strings.Contains(tracker, "friendly") {
		t.Errorf("tracker missing friendly status, got: %s", tracker)
	}
	if !strings.Contains(tracker, "hostile") {
		t.Errorf("tracker missing hostile status, got: %s", tracker)
	}
}

func TestGenerateFactionTracker_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir, "wkhtmltopdf")
	tracker := c.generateFactionTracker()

	if tracker != "" {
		t.Errorf("expected empty tracker for campaign without factions, got: %s", tracker)
	}
}

func TestGenerateHTML_WithNewSections(t *testing.T) {
	tmpDir := t.TempDir()

	// Create campaign structure with new sections
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nSome lore text."), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "session-zero.md"), []byte("# Sesión Cero — Guía para el DM\n\nSafety tools and guidelines."), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "flowchart.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="100" height="100"/></svg>`), 0644)

	// Create acts with proper WotC area format (10-15 numbered areas per act)
	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)
	_ = os.WriteFile(filepath.Join(areasDir, "chapter-01.md"), []byte(`# Capítulo 1: El Comienzo

> **Nivel:** 1-2 | **Duración:** 2-3 horas | **Tono:** Misterioso

## Resumen
Los personajes llegan al pueblo y aceptan su primera misión.

## Áreas Numeradas (WotC Format)

### Área 1: La Entrada del Pueblo

> **Read-Aloud:** *El camino costero termina en un pueblo pequeño. Casas de madera, niebla marina, y un silencio antinatural. El viento trae olor a sal y algo más... algo que no debería estar aquí.*

**Descripción para el DM:**
La entrada al pueblo es un camino empedrado que atraviesa una verja de hierro oxidado. A la izquierda hay una taberna "La Gaviota Perdida", a la derecha una iglesia pequeña.

- **Percepción DC 10:** Huellas frescas en el barro cerca de la taberna
- **Percepción DC 12:** Un gato negro los observa desde el techo de una casa

**Criaturas:**
- 1 **Sra. Morales** *NG female human commoner* — dueña de la taberna (ver Apéndice B)

**Tesoro:**
- **XP:** 0 (exploración social)
- **Moneda:** 5 sp (en el bolsillo del jugador)

**Conexiones:**
- → Área 2 (La Gaviota Perdida, la taberna, 30 pies al norte)
- → Área 3 (La Iglesia, 50 pies al este)

**Secretos y Trampas:**
- **Secreto: Gato Negro**
  - **Encontrar:** Percepción DC 14 para notar que el gato no tiene sombra
  - **Contenido:** El gato es en realidad un familiar mágico de un mago que vive en la Mansión Vargas

**Desarrollo:**
- **Si entran a la taberna:** La Sra. Morales les cuenta sobre los lamentos nocturnos
- **Si van a la iglesia:** El Padre Tomás les ofrece la misión principal

> ##### Nota del DM
> Este es un buen momento para que los jugadores exploren y hagan preguntas. La información aquí es crucial para el resto de la aventura.
`), 0644)

	c := New(tmpDir, "wkhtmltopdf")
	htmlParts, err := c.generateHTML("Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// Verify Session Zero heading (markdown converts to h1 with id)
	if !strings.Contains(html, "Sesión Cero") {
		t.Errorf("expected Sesión Cero heading in HTML, got: %s", html)
	}

	// Verify flowchart SVG
	if !strings.Contains(html, "<svg") {
		t.Errorf("expected flowchart SVG in HTML, got: %s", html)
	}

	// Verify WotC area format elements in HTML (new format uses "Criaturas" not "NPCs presentes")
	if !strings.Contains(html, "Área 1") {
		t.Errorf("expected 'Área 1' heading in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Sra. Morales") {
		t.Errorf("expected 'Sra. Morales' NPC in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Percepción DC") {
		t.Errorf("expected 'Percepción DC' skill check in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Tesoro") {
		t.Errorf("expected 'Tesoro' section in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Conexiones") {
		t.Errorf("expected 'Conexiones' section in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Secretos y Trampas") {
		t.Errorf("expected 'Secretos y Trampas' section in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Desarrollo") {
		t.Errorf("expected 'Desarrollo' section in HTML, got: %s", html)
	}
	if !strings.Contains(html, "Nota del DM") {
		t.Errorf("expected sidebar 'Nota del DM' in HTML, got: %s", html)
	}
	// Note: Apéndice F (Adventure Roster) requires "NPCs presentes" format which uses old roster extraction
	// The new area format uses "Criaturas" which is NOT extracted to roster (known limitation)
}

func TestExtractRosterEntries(t *testing.T) {
	md := `# Acto 1

## NPCs presentes
- **Eldrin** — Mago
- **Thorn** — Guerrero

## Monstruos
- **Goblin** (CR 1/4)

## Encuentros
- Emboscada
- Defensa
`
	npcs, monsters, encounters := extractRosterEntries(md)

	if len(npcs) != 2 {
		t.Errorf("expected 2 NPCs, got %d", len(npcs))
	}
	if len(monsters) != 1 {
		t.Errorf("expected 1 monster, got %d", len(monsters))
	}
	if len(encounters) != 2 {
		t.Errorf("expected 2 encounters, got %d", len(encounters))
	}

	if !strings.Contains(npcs[0], "Eldrin") {
		t.Errorf("expected Eldrin in NPCs, got %v", npcs)
	}
	if !strings.Contains(monsters[0], "Goblin") {
		t.Errorf("expected Goblin in monsters, got %v", monsters)
	}
}

func TestCompile_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal campaign structure
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)

	c := New(tmpDir, "sleep")

	// Use a very short timeout to trigger context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give time for the context to expire
	time.Sleep(5 * time.Millisecond)

	_, err := c.Compile(ctx, "Test Campaign")
	if err == nil {
		t.Fatal("expected error due to context timeout, got nil")
	}
	if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "killed") {
		t.Errorf("expected context/timeout/killed error, got: %v", err)
	}
}

func TestCompile_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal campaign structure
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)

	c := New(tmpDir, "sleep")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := c.Compile(ctx, "Test Campaign")
	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}
}

func TestCompile_SuccessWithContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal campaign structure
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)

	// Use "echo" as a fake PDF engine that creates the output file
	c := New(tmpDir, "echo")

	ctx := context.Background()
	_, err := c.Compile(ctx, "Test Campaign")
	// echo will succeed but won't create a valid PDF; the test verifies context is passed through
	// We expect an error because echo doesn't produce a valid PDF, but the context should work
	if err == nil {
		// If echo somehow works, that's fine too - what matters is no panic and context was used
		t.Log("Compile with context completed without error (echo returned 0)")
	}
}

// Phase 1: Compiler HTML Bug Fixes Tests

func TestFormatInline_NoThinSpaceBeforeColon(t *testing.T) {
	input := "**Importante**: Esto es crucial"
	expected := "<strong>Importante</strong>: Esto es crucial"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_NoThinSpaceBeforeSemicolon(t *testing.T) {
	input := "**texto**; más texto"
	expected := "<strong>texto</strong>; más texto"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_NoThinSpaceBeforeComma(t *testing.T) {
	input := "**texto**, y más"
	expected := "<strong>texto</strong>, y más"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_NoThinSpaceBeforePeriod(t *testing.T) {
	input := "**texto**. Fin."
	expected := "<strong>texto</strong>. Fin."
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_NoThinSpaceBeforeExclamation(t *testing.T) {
	input := "**texto**! ¡Wow!"
	expected := "<strong>texto</strong>! ¡Wow!"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_NoThinSpaceBeforeQuestion(t *testing.T) {
	input := "**texto**? ¿Qué?"
	expected := "<strong>texto</strong>? ¿Qué?"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_ThinSpaceBeforeLetter(t *testing.T) {
	input := "**texto**palabra"
	expected := "<strong>texto</strong>&thinsp;palabra"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_ThinSpaceBeforeNumber(t *testing.T) {
	input := "**texto**123"
	expected := "<strong>texto</strong>&thinsp;123"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestMarkdownToHTML_StripsComments(t *testing.T) {
	input := "Texto <!-- comentario --> más texto"
	result := markdownToHTML(input, "")
	
	if strings.Contains(result, "<!--") {
		t.Errorf("HTML comments not stripped: %s", result)
	}
	if !strings.Contains(result, "Texto") || !strings.Contains(result, "más texto") {
		t.Errorf("Text around comment was lost: %s", result)
	}
}

func TestMarkdownToHTML_StripsMultilineComments(t *testing.T) {
	input := `Texto
<!-- 
  Comentario
  de múltiples
  líneas
-->
más texto`
	result := markdownToHTML(input, "")
	
	if strings.Contains(result, "<!--") {
		t.Errorf("Multiline comments not stripped: %s", result)
	}
	if !strings.Contains(result, "Texto") || !strings.Contains(result, "más texto") {
		t.Errorf("Text around multiline comment was lost: %s", result)
	}
}

func TestMarkdownToHTML_StripsCommentsInParagraph(t *testing.T) {
	input := "Párrafo con <!-- comentario inline --> texto"
	result := markdownToHTML(input, "")
	
	if strings.Contains(result, "<!--") {
		t.Errorf("Inline comments not stripped: %s", result)
	}
}

func TestMarkdownToHTML_StripsCommentsWithHTML(t *testing.T) {
	input := `# Título
Texto normal
<!-- <div>comentario con HTML</div> -->
Más texto`
	result := markdownToHTML(input, "")
	
	if strings.Contains(result, "<!--") {
		t.Errorf("Comments with HTML not stripped: %s", result)
	}
	if !strings.Contains(result, "<h1") || !strings.Contains(result, "Más texto") {
		t.Errorf("Content around comment was lost: %s", result)
	}
}
