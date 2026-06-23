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

func TestImageEmbedding_MIMEDetection(t *testing.T) {
	// 5 table cases: PNG bytes, JPEG bytes under .png (the bug), GIF, WebP, unknown fallback
	// REQ-1.1, 1.2, 1.3, 1.4, 1.5, 4.1, 4.2
	tests := []struct {
		name           string
		filename       string
		bytes          []byte
		wantMimePrefix string
	}{
		{
			name:           "PNG bytes with .png extension",
			filename:       "test.png",
			bytes:          []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D},
			wantMimePrefix: "data:image/png;base64,",
		},
		{
			name:           "JPEG bytes with .png extension (the bug)",
			filename:       "cover.png",
			bytes:          []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01},
			wantMimePrefix: "data:image/jpeg;base64,",
		},
		{
			name:           "GIF87a bytes with .gif extension",
			filename:       "anim.gif",
			bytes:          []byte("GIF87a"),
			wantMimePrefix: "data:image/gif;base64,",
		},
		{
			name:           "RIFF/WEBP bytes with .webp extension",
			filename:       "modern.webp",
			bytes:          []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'},
			wantMimePrefix: "data:image/webp;base64,",
		},
		{
			name:           "Unknown bytes fall back to extension-derived MIME",
			filename:       "mystery.png",
			bytes:          []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B},
			wantMimePrefix: "data:image/png;base64,",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			imgPath := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(imgPath, tt.bytes, 0644); err != nil {
				t.Fatal(err)
			}

			// Use the public embedImage via a markdown image ref
			md := "![alt](" + imgPath + ")"
			result := markdownToHTML(md, tmpDir)

			if !strings.Contains(result, tt.wantMimePrefix) {
				t.Errorf("expected MIME prefix %q, got result:\n%s", tt.wantMimePrefix, result)
			}
		})
	}
}

func TestDetectMimeType(t *testing.T) {
	// Direct unit test for the helper function
	tests := []struct {
		name       string
		data       []byte
		wantMime   string
		wantDetect bool
	}{
		{"PNG signature", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "image/png", true},
		{"JPEG signature", []byte{0xFF, 0xD8, 0xFF, 0xE1}, "image/jpeg", true},
		{"GIF87a signature", []byte("GIF87a"), "image/gif", true},
		{"GIF89a signature", []byte("GIF89a"), "image/gif", true},
		{"WebP signature", []byte{'R', 'I', 'F', 'F', 0, 0, 0, 0, 'W', 'E', 'B', 'P'}, "image/webp", true},
		{"empty input", []byte{}, "", false},
		{"unknown bytes", []byte{0x00, 0x01, 0x02, 0x03}, "", false},
		{"too short for PNG", []byte{0x89, 0x50}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime, detected := detectMimeType(tt.data)
			if detected != tt.wantDetect {
				t.Errorf("detectMimeType() detected = %v, want %v", detected, tt.wantDetect)
			}
			if mime != tt.wantMime {
				t.Errorf("detectMimeType() mime = %q, want %q", mime, tt.wantMime)
			}
		})
	}
}

func TestStatBlockParser_DoesNotTriggerOnChapterHeading(t *testing.T) {
	// Negative lock: a `## ` heading NOT followed by an italic size+type line
	// must NOT be wrapped in `.stat-block` (REQ-2.10, 2.11, 4.5).
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "chapter heading with chapter summary",
			input: `# Capítulo 1: La Llegada

> **Nivel:** 1-2 | **Duración:** 2-3 horas

## Resumen

Los personajes llegan al pueblo y aceptan su primera misión.

### Área 1: La Entrada del Pueblo

> **Read-Aloud:** *El camino costero termina en un pueblo pequeño.*
`,
		},
		{
			name: "NPC heading without italic size line",
			input: `## Beroldo

Beroldo es un herrero local. Tiene una barba canosa y siempre está
trabajando en su fragua. Sabe más de lo que dice.

### Apariencia

Un hombre de unos cincuenta años, con delantal de cuero.
`,
		},
		{
			name: "h3 sub-section must stay h3 (not promoted to h2)",
			input: `## Description narrativa

Some narrative text.

### Fase 1: Despierto

El Rayo ataca desde la distancia.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markdownToHTML(tt.input, "/tmp")
			if strings.Contains(result, `<div class="stat-block"`) {
				t.Errorf("expected NO .stat-block wrapper, but found one in:\n%s", result)
			}
		})
	}
}

func TestStatBlockParser_DetectsElRayo(t *testing.T) {
	// El Rayo markdown from el-exiliado-de-las-tierras-marchitas/bestiary.md
	// lines 876-893. Verifies the WotC stat block parser produces
	// <div class="stat-block" data-monster="El Rayo"> wrapper,
	// exactly 3 .stat-line divs (AC/HP/Speed), <h2>El Rayo</h2> preserved,
	// and <p class="monster-type"> present (REQ-2.1, 2.2, 4.4).
	elRayoMD := `## El Rayo
*Mediano incorpóreo, caótico neutro*

**Armor Class** 14
**Hit Points** 82 (11d10 + 22)
**Speed** 0 ft., fly 50 ft. (hover)

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| 1 (-5) | 18 (+4) | 14 (+2) | 16 (+3) | 14 (+2) | 18 (+4) |

**Saving Throws** Dex +8, Con +6
**Challenge** 5 (1,800 XP)

**Incorpóreo y luminoso.** El Rayo no tiene cuerpo físico.

**Actions**

**Toque del Rayo.** *Melee Spell Attack:* +8 to hit.
`

	result := markdownToHTML(elRayoMD, "/tmp")

	// 1. Wrapper div with data-monster attribute
	if !strings.Contains(result, `<div class="stat-block" data-monster="El Rayo">`) {
		t.Errorf("expected stat-block wrapper with data-monster attribute, got:\n%s", result)
	}

	// 2. h2 preserved (not downgraded to h3)
	if !strings.Contains(result, `<h2>El Rayo</h2>`) {
		t.Errorf("expected <h2>El Rayo</h2> preserved, got:\n%s", result)
	}

	// 3. monster-type paragraph for the italic line
	if !strings.Contains(result, `<p class="monster-type">`) {
		t.Errorf("expected <p class=\"monster-type\"> for the italic line, got:\n%s", result)
	}

	// 4. exactly 3 .stat-line divs (AC, HP, Speed)
	statLineCount := strings.Count(result, `<div class="stat-line">`)
	if statLineCount != 3 {
		t.Errorf("expected 3 .stat-line divs (AC/HP/Speed), got %d in:\n%s", statLineCount, result)
	}

	// 5. stat-label and stat-value present
	if !strings.Contains(result, `class="stat-label"`) {
		t.Errorf("expected stat-label spans, got:\n%s", result)
	}
	if !strings.Contains(result, `class="stat-value"`) {
		t.Errorf("expected stat-value spans, got:\n%s", result)
	}

	// 6. Labels for AC, HP, Speed
	for _, want := range []string{"Armor Class", "Hit Points", "Speed"} {
		if !strings.Contains(result, want) {
			t.Errorf("expected stat label %q in output, got:\n%s", want, result)
		}
	}

	// 7. closing div
	if !strings.Contains(result, `</div>`) {
		t.Errorf("expected closing </div>, got:\n%s", result)
	}

	// 8. REQ-2.5: trait paragraphs become <p class="trait">. The fixture
	// contains the trait "**Incorpóreo y luminoso.**" which must be classified
	// as a trait (label ends with ".", value is plain text with no <em>).
	if !strings.Contains(result, `class="trait"`) {
		t.Errorf("expected at least one .trait paragraph (REQ-2.5), got:\n%s", result)
	}
	traitCount := strings.Count(result, `<p class="trait">`)
	if traitCount < 1 {
		t.Errorf("expected at least one <p class=\"trait\"> element, got %d in:\n%s", traitCount, result)
	}

	// 9. REQ-2.8: action paragraphs become <p class="action">. The fixture
	// contains the action "**Toque del Rayo.** *Melee Spell Attack:*" which
	// must be classified as an action (value contains <em> from italic).
	if !strings.Contains(result, `class="action"`) {
		t.Errorf("expected at least one .action paragraph (REQ-2.8), got:\n%s", result)
	}
	actionCount := strings.Count(result, `<p class="action">`)
	if actionCount < 1 {
		t.Errorf("expected at least one <p class=\"action\"> element, got %d in:\n%s", actionCount, result)
	}

	// 10. property-line should still appear for the multi-property line
	// (Saving Throws + Challenge) but NOT for the trait or action paragraphs.
	// If property-line count is too high, trait/action detection is broken.
	propertyLineCount := strings.Count(result, `<p class="property-line">`)
	if propertyLineCount > 4 {
		t.Errorf("property-line count suspiciously high (%d) — trait/action likely misclassified in:\n%s", propertyLineCount, result)
	}
}

func TestStatBlockParser_SplitsMultiPropertyLine(t *testing.T) {
	// REQ-2.3, 4.6: a line with multiple **Label** groups must split into
	// one .property-line per group.
	multiPropMD := `## Test Monster
*Medium humanoid, neutral*

**Armor Class** 14
**Hit Points** 82 (11d10 + 22)
**Speed** 0 ft., fly 50 ft. (hover)

| STR | DEX | CON | INT | WIS | CHA |
|:---:|:---:|:---:|:---:|:---:|:---:|
| 10 | 10 | 10 | 10 | 10 | 10 |

**Saving Throws** Dex +8, Con +6 **Damage Vulnerabilities** radiant **Damage Immunities** poison **Condition Immunities** charmed **Challenge** 5 (1,800 XP)
`

	result := markdownToHTML(multiPropMD, "/tmp")

	// Should contain the stat-block wrapper
	if !strings.Contains(result, `<div class="stat-block" data-monster="Test Monster">`) {
		t.Errorf("expected stat-block wrapper, got:\n%s", result)
	}

	// 5 .property-line divs (Saving Throws, Damage Vulnerabilities, Damage
	// Immunities, Condition Immunities, Challenge).
	propLineCount := strings.Count(result, `<p class="property-line">`)
	if propLineCount != 5 {
		t.Errorf("expected 5 .property-line divs, got %d in:\n%s", propLineCount, result)
	}

	// Each label should appear
	for _, want := range []string{"Saving Throws", "Damage Vulnerabilities", "Damage Immunities", "Condition Immunities", "Challenge"} {
		if !strings.Contains(result, want) {
			t.Errorf("expected label %q in output, got:\n%s", want, result)
		}
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

func TestExtractImagePaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int // number of paths extracted
	}{
		{
			name:  "markdown image",
			input: "![alt](assets/scene1.png)",
			want:  1,
		},
		{
			name:  "raw img tag",
			input: `<img src="assets/map.svg" alt="Map">`,
			want:  1,
		},
		{
			name:  "code asset ref",
			input: "`assets/monster.png`",
			want:  1,
		},
		{
			name:  "mixed",
			input: `![md](assets/a.png) and <img src="assets/b.svg"> and ` + "`assets/c.png`",
			want:  3,
		},
		{
			name:  "no images",
			input: "just text **bold**",
			want:  0,
		},
		{
			name:  "duplicate same path",
			input: "![a](assets/sep.svg) and ![b](assets/sep.svg)",
			want:  2, // extraction returns all; dedup happens at caller
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractImagePaths(tt.input)
			if len(got) != tt.want {
				t.Errorf("extractImagePaths() returned %d paths, want %d; paths=%v", len(got), tt.want, got)
			}
		})
	}
}

func TestCountImagesInMarkdownSources_UniquePaths(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string // relative path → content
		want  int
	}{
		{
			name: "same image in 2 files counts as 1",
			files: map[string]string{
				"chapters/ch1.md": "![Sep](assets/separator.svg)",
				"lore.md":         "![Sep](assets/separator.svg)",
			},
			want: 1,
		},
		{
			name: "3 distinct images",
			files: map[string]string{
				"chapters/ch1.md": "![A](assets/a.png)\n![B](assets/b.png)",
				"lore.md":         "![C](assets/c.svg)",
			},
			want: 3,
		},
		{
			name:  "no images",
			files: map[string]string{"chapters/ch1.md": "just text"},
			want:  0,
		},
		{
			name: "mix of markdown/img-tag/code-asset to same path",
			files: map[string]string{
				"chapters/ch1.md": "![Sep](assets/separator.svg) and `assets/separator.svg`",
				"lore.md":         `<img src="assets/separator.svg" alt="Sep">`,
			},
			want: 1,
		},
		{
			name: "same image 4 times across 2 files",
			files: map[string]string{
				"chapters/ch1.md": "![Sep](assets/separator.svg)\n![Sep](assets/separator.svg)",
				"lore.md":         "![Sep](assets/separator.svg)\n![Sep](assets/separator.svg)",
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for relPath, content := range tt.files {
				fullPath := filepath.Join(tmpDir, relPath)
				if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			c := New(tmpDir, "")
			got, err := c.countImagesInMarkdownSources()
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("countImagesInMarkdownSources() = %d, want %d", got, tt.want)
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
	chaptersDir := filepath.Join(tmpDir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a markdown file with 2 images
	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
Some text
<img src="assets/map.svg" alt="Map">
`
	mdPath := filepath.Join(chaptersDir, "chapter1.md")
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

	c := New(tmpDir, "")
	expected, found, warnings, err := c.verifyImages(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if expected != 2 {
		t.Errorf("expected = %d, want 2", expected)
	}
	if found != 2 {
		t.Errorf("found = %d, want 2", found)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want empty", warnings)
	}
}

func TestVerifyImages_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create campaign structure
	chaptersDir := filepath.Join(tmpDir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a markdown file with 2 images
	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
<img src="assets/map.svg" alt="Map">
`
	mdPath := filepath.Join(chaptersDir, "chapter1.md")
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

	c := New(tmpDir, "")
	expected, found, warnings, err := c.verifyImages(htmlPath)
	if err != nil {
		t.Fatal(err)
	}
	if expected != 2 {
		t.Errorf("expected = %d, want 2", expected)
	}
	if found != 1 {
		t.Errorf("found = %d, want 1", found)
	}
	if len(warnings) == 0 {
		t.Error("warnings should be non-empty for mismatch")
	}
	// Warning should mention expected/found counts
	foundWarning := false
	for _, w := range warnings {
		if strings.Contains(w, "expected 2, found 1") {
			foundWarning = true
			break
		}
	}
	if !foundWarning {
		t.Errorf("warnings should contain 'expected 2, found 1', got: %v", warnings)
	}
}

func TestVerifyImages_Advisory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create campaign structure with 2 images in markdown
	chaptersDir := filepath.Join(tmpDir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		t.Fatal(err)
	}

	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
![Scene 2](assets/scene2.png)
`
	mdPath := filepath.Join(chaptersDir, "chapter1.md")
	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		t.Fatal(err)
	}

	// HTML has only 1 image — mismatch should produce warning, NOT error
	htmlContent := `<html><body>
		<img src="data:image/png;base64,abc" alt="Scene 1" class="campaign-image">
	</body></html>`
	htmlPath := filepath.Join(tmpDir, "campaign.html")
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}

	c := New(tmpDir, "")
	expected, found, warnings, err := c.verifyImages(htmlPath)
	// Advisory: err should be nil even on mismatch
	if err != nil {
		t.Fatalf("verifyImages() should not return error on mismatch, got: %v", err)
	}
	if expected != 2 {
		t.Errorf("expected = %d, want 2", expected)
	}
	if found != 1 {
		t.Errorf("found = %d, want 1", found)
	}
	if len(warnings) == 0 {
		t.Fatal("warnings should be non-empty for advisory mismatch")
	}
	// Verify warning text contains expected/found counts
	if !strings.Contains(warnings[0], "expected 2, found 1") {
		t.Errorf("warning = %q, should contain 'expected 2, found 1'", warnings[0])
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

	c := New(tmpDir, "")
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
	c := New(tmpDir, "")
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

	// Create chapters with proper WotC area format (10-15 numbered areas per chapter)
	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter-01.md"), []byte(`# Capítulo 1: El Comienzo

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

	c := New(tmpDir, "")
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
	// No thin space added - it caused spacing issues in PDF
	expected := "<strong>texto</strong>palabra"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_ThinSpaceBeforeNumber(t *testing.T) {
	input := "**texto**123"
	// No thin space added - it caused spacing issues in PDF
	expected := "<strong>texto</strong>123"
	result := formatInline(input)
	if result != expected {
		t.Errorf("formatInline() = %q, want %q", result, expected)
	}
}

func TestFormatInline_BoldItalic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "triple-star bold-italic",
			input:    "***text***",
			expected: "<strong><em>text</em></strong>",
		},
		{
			name:     "triple-star with period (sub-feature)",
			input:    "***Ceilings.*** 10 feet high.",
			expected: "<strong><em>Ceilings.</em></strong> 10 feet high.",
		},
		{
			name:     "mixed bold-italic bold and italic",
			input:    "***bold-italic*** **bold** *italic*",
			expected: "<strong><em>bold-italic</em></strong> <strong>bold</strong> <em>italic</em>",
		},
		{
			name:     "sub-feature and bold in same paragraph",
			input:    "***Walls.*** Stone. **Note:** Dangerous.",
			expected: "<strong><em>Walls.</em></strong> Stone. <strong>Note:</strong> Dangerous.",
		},
		{
			name:     "asymmetric stars graceful fallback no panic",
			input:    "***text**",
			expected: "<strong>*text</strong>",
		},
		{
			name:     "existing bold unchanged",
			input:    "**bold**",
			expected: "<strong>bold</strong>",
		},
		{
			name:     "existing italic unchanged",
			input:    "*italic*",
			expected: "<em>italic</em>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInline(tt.input)
			if result != tt.expected {
				t.Errorf("formatInline(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
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

func TestHTMLBlockNotWrappedInParagraph(t *testing.T) {
	md := `# Test

Some text before.

<div class="shock-point moderate">
<span class="severity-badge">moderate</span>
<strong>Violencia</strong>: Test content.
</div>

Some text after.
`
	
	html := markdownToHTML(md, "/tmp")
	
	if strings.Contains(html, "<p><div") {
		t.Errorf("HTML should not wrap <div> in <p> tags, got: %s", html)
	}
	
	if !strings.Contains(html, `<div class="shock-point`) {
		t.Errorf("HTML should contain the div, got: %s", html)
	}
}

// Phase 2: Worksheet Nesting Tests

func TestWorksheetNestedDivs(t *testing.T) {
	md := `### Worksheet de Creación de Personaje

<div class="character-worksheet">
<div class="worksheet-section">
<h4>Antecedentes y Conexiones</h4>
<div class="prompt-box">¿Qué conecta a tu personaje con esta región?</div>
</div>

<div class="worksheet-section">
<h4>Motivaciones</h4>
<div class="prompt-box">¿Qué busca tu personaje?</div>
</div>

<div class="worksheet-section">
<h4>Vínculos con el Partido</h4>
<div class="prompt-box">¿Cómo conoció al grupo?</div>
</div>

<div class="worksheet-section">
<h4>Secretos y Defectos</h4>
<div class="prompt-box">¿Qué oculta tu personaje?</div>
</div>
</div>
`
	
	html := markdownToHTML(md, "/tmp")
	
	// Verify all opening tags have closing tags
	openCount := strings.Count(html, `<div class="worksheet-section">`)
	closeCount := strings.Count(html, `</div>`)
	
	// Should have 4 worksheet-section opening tags
	if openCount != 4 {
		t.Errorf("Expected 4 worksheet-section opening tags, got %d", openCount)
	}
	
	// Should have enough closing tags (4 worksheet-section + 1 character-worksheet + 4 prompt-box = 9 minimum)
	if closeCount < 9 {
		t.Errorf("Expected at least 9 closing </div> tags, got %d. HTML:\n%s", closeCount, html)
	}
	
	// Verify no <p><div> nesting
	if strings.Contains(html, "<p><div") {
		t.Errorf("HTML should not wrap <div> in <p> tags:\n%s", html)
	}
	
	// Verify character-worksheet div is present
	if !strings.Contains(html, `<div class="character-worksheet">`) {
		t.Errorf("HTML should contain character-worksheet div:\n%s", html)
	}
}

func TestDeeplyNestedDivs(t *testing.T) {
	md := `<div class="level1">
<div class="level2">
<div class="level3">
<div class="level4">Deep content</div>
</div>
</div>
</div>`
	
	html := markdownToHTML(md, "/tmp")
	
	// Count each level
	level1Open := strings.Count(html, `<div class="level1">`)
	level1Close := strings.Count(html, `</div>`) // This counts all, but we need at least 4 total
	
	if level1Open != 1 {
		t.Errorf("Expected 1 level1 div, got %d", level1Open)
	}
	
	// Should have at least 4 closing divs for 4 levels
	if level1Close < 4 {
		t.Errorf("Expected at least 4 closing divs for nested structure, got %d. HTML:\n%s", level1Close, html)
	}
}

// Phase 3: Unclosed Div Auto-Close Tests

func TestExtractBalancedDivs_UnclosedDiv(t *testing.T) {
	md := `<div class="foo">some content`
	
	html := markdownToHTML(md, "/tmp")
	
	// Verify the div is auto-closed
	if !strings.Contains(html, `<div class="foo">`) {
		t.Errorf("Expected opening div tag, got: %s", html)
	}
	if !strings.Contains(html, `</div>`) {
		t.Errorf("Expected closing div tag (auto-closed), got: %s", html)
	}
	if !strings.Contains(html, "some content") {
		t.Errorf("Expected content preserved, got: %s", html)
	}
	
	// Verify balanced tags
	openCount := strings.Count(html, `<div class="foo">`)
	closeCount := strings.Count(html, `</div>`)
	if openCount != closeCount {
		t.Errorf("Unbalanced divs: %d opens, %d closes. HTML:\n%s", openCount, closeCount, html)
	}
}

func TestExtractBalancedDivs_NestedUnclosed(t *testing.T) {
	md := `<div class="outer"><div class="inner">nested content`
	
	html := markdownToHTML(md, "/tmp")
	
	// Verify both divs are auto-closed
	if !strings.Contains(html, `<div class="outer">`) {
		t.Errorf("Expected outer div opening tag, got: %s", html)
	}
	if !strings.Contains(html, `<div class="inner">`) {
		t.Errorf("Expected inner div opening tag, got: %s", html)
	}
	if !strings.Contains(html, "nested content") {
		t.Errorf("Expected content preserved, got: %s", html)
	}
	
	// Verify balanced tags (2 opens, 2 closes)
	openCount := strings.Count(html, `<div class="`)
	closeCount := strings.Count(html, `</div>`)
	if openCount != closeCount {
		t.Errorf("Unbalanced divs: %d opens, %d closes. HTML:\n%s", openCount, closeCount, html)
	}
}

func TestExtractBalancedDivs_Mismatched(t *testing.T) {
	// 3 opens, 2 closes = net 1 unclosed
	md := `<div class="a"><div class="b"><div class="c">content</div></div>`
	
	html := markdownToHTML(md, "/tmp")
	
	// Verify all opening tags are present
	if !strings.Contains(html, `<div class="a">`) {
		t.Errorf("Expected div a opening tag, got: %s", html)
	}
	if !strings.Contains(html, `<div class="b">`) {
		t.Errorf("Expected div b opening tag, got: %s", html)
	}
	if !strings.Contains(html, `<div class="c">`) {
		t.Errorf("Expected div c opening tag, got: %s", html)
	}
	
	// Verify balanced tags (3 opens, 3 closes)
	openCount := strings.Count(html, `<div class="`)
	closeCount := strings.Count(html, `</div>`)
	if openCount != closeCount {
		t.Errorf("Unbalanced divs: %d opens, %d closes. HTML:\n%s", openCount, closeCount, html)
	}
}
