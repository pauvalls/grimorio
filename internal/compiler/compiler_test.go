package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMarkdownToHTML_ProcessScenePlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantContains string
		wantNotContains string
	}{
		{
			name:     "scene placeholder becomes descriptive text",
			input:    "*[SCENE: A dark dungeon with flickering torches]*",
			wantContains: "A dark dungeon with flickering torches",
			wantNotContains: "[SCENE:",
		},
		{
			name:     "scene placeholder in standalone line",
			input:    "[SCENE: Epic battle between heroes and dragon]",
			wantContains: "Epic battle between heroes and dragon",
			wantNotContains: "[SCENE:",
		},
		{
			name:     "multiple scene placeholders",
			input:    "[SCENE: First scene]\n\nSome text\n\n[SCENE: Second scene]",
			wantContains: "First scene",
			wantNotContains: "[SCENE:",
		},
		{
			name:     "regular markdown still works",
			input:    "# Heading\n\nSome **bold** text",
			wantContains: "<h1",
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
	os.MkdirAll(svgDir, 0755)
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
	os.MkdirAll(imgDir, 0755)
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
	actsDir := filepath.Join(tmpDir, "acts")
	os.MkdirAll(actsDir, 0755)
	
	// Create a markdown file with 2 images
	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
Some text
<img src="assets/map.svg" alt="Map">
`
	mdPath := filepath.Join(actsDir, "act1.md")
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
	actsDir := filepath.Join(tmpDir, "acts")
	os.MkdirAll(actsDir, 0755)
	
	// Create a markdown file with 2 images
	mdContent := `# Act 1
![Scene 1](assets/scene1.png)
<img src="assets/map.svg" alt="Map">
`
	mdPath := filepath.Join(actsDir, "act1.md")
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
