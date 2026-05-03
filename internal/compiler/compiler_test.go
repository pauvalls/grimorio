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
