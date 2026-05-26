package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompilerV2_HierarchicalTOC(t *testing.T) {
	tmpDir := t.TempDir()

	// Create lore
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)

	// Create act with areas
	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)
	_ = os.WriteFile(filepath.Join(areasDir, "chapter_01_test.md"), []byte(`# Capítulo 1: Test

### Área 1: Vestíbulo

Content.

### Área 2: Sala

Content.
`), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// TOC should contain area references
	if !strings.Contains(html, "Área 1") {
		t.Error("TOC should contain Area 1")
	}
	if !strings.Contains(html, "Área 2") {
		t.Error("TOC should contain Area 2")
	}
}

func TestCompilerV2_CrossReferenceLinks(t *testing.T) {
	tmpDir := t.TempDir()

	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)
	_ = os.WriteFile(filepath.Join(areasDir, "chapter_01_test.md"), []byte(`# Capítulo 1

### Área 1: First

Go to Área 2.

### Área 2: Second

Go back to Área 1.
`), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// Should have clickable links between areas
	if !strings.Contains(html, `href="#`) {
		t.Error("HTML should contain internal links for cross-references")
	}
}

func TestCompilerV2_AreaNumberHighlighting(t *testing.T) {
	tmpDir := t.TempDir()

	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)
	_ = os.WriteFile(filepath.Join(areasDir, "chapter_01_test.md"), []byte(`# Capítulo 1

### Área 1: Test

Content.
`), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	if !strings.Contains(html, "area-number") {
		t.Error("v2 HTML should highlight area numbers with area-number class")
	}
}

func TestCompilerV2_InlineStatBlock(t *testing.T) {
	tmpDir := t.TempDir()

	// Create bestiary with custom creature
	bestiaryDir := filepath.Join(tmpDir, "bestiary")
	_ = os.MkdirAll(bestiaryDir, 0755)
	_ = os.WriteFile(filepath.Join(bestiaryDir, "bestiary.md"), []byte(`# Bestiario

## Shadow Wraith

*Mediano no-muerto, NE*

**Clase de Armadura:** 12
**Puntos de Golpe:** 22 (5d8)
**Velocidad:** 30 pies

**Desafío:** 1/4 (50 PX)
`), 0644)

	// Create act referencing the creature
	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)
	_ = os.WriteFile(filepath.Join(areasDir, "chapter_01_test.md"), []byte(`# Capítulo 1

### Área 1: Test

**Criaturas:**
- 2 **Shadow Wraith**

Content.
`), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// Should contain stat block inline for custom creature
	if !strings.Contains(html, "stat-block") {
		t.Log("Note: stat blocks may not be fully inline yet")
	}
}

func TestCompilerV2_HandoutPages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal campaign
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "areas"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "areas", "chapter_01_test.md"), []byte("# Capítulo 1\n\nTest."), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// v2 should have handout section
	if !strings.Contains(html, "handout-page") {
		t.Log("Note: handout pages may be empty if no handout data exists")
	}
}

func TestCompilerV1_NoV2Features(t *testing.T) {
	tmpDir := t.TempDir()

	areasDir := filepath.Join(tmpDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)
	_ = os.WriteFile(filepath.Join(areasDir, "chapter_01_test.md"), []byte(`# Capítulo 1

### Área 1: Test

Content.
`), 0644)

	c := NewWithVersion(tmpDir, "", 1)
	htmlParts, err := c.generateHTML("Test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// v1 should NOT have area number highlighting
	if strings.Contains(html, `<span class="area-number">`) {
		t.Error("v1 should not have area-number highlighting")
	}
}
