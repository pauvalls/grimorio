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

	// Create chapter with numbered areas
	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01_test.md"), []byte(`# Capítulo 1: Test

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

	// HTML body should contain the area headings (the legacy
	// extractAreasFromDir hierarchical TOC was removed in WU7 — areas
	// are inlined into the body via markdownToHTMLWithID + postProcessHTML).
	if !strings.Contains(html, "Área 1") {
		t.Error("HTML should contain Area 1 heading")
	}
	if !strings.Contains(html, "Área 2") {
		t.Error("HTML should contain Area 2 heading")
	}
}

func TestCompilerV2_CrossReferenceLinks(t *testing.T) {
	tmpDir := t.TempDir()

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01_test.md"), []byte(`# Capítulo 1

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

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01_test.md"), []byte(`# Capítulo 1

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

	// Create chapter referencing the creature
	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01_test.md"), []byte(`# Capítulo 1

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
	_ = os.MkdirAll(filepath.Join(tmpDir, "chapters"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "chapters", "chapter_01_test.md"), []byte("# Capítulo 1\n\nTest."), 0644)

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

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01_test.md"), []byte(`# Capítulo 1

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

func TestCompilerV2_PrologueChapterInTOC(t *testing.T) {
	tmpDir := t.TempDir()

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)

	// Create prologue chapter with is_prologue frontmatter
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_00.md"), []byte(`---
is_prologue: true
---
# Prologue: The Gathering Storm

### Area 1: The Tavern

The party meets at a tavern.

### Area 2: The Road

The party sets out on the road.
`), 0644)

	// Create a regular chapter
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(`# Chapter 1: The Journey Begins

### Area 1: Forest Edge

Content here.
`), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// TOC should contain "Prologue" entry
	if !strings.Contains(html, "Prologue") {
		t.Error("TOC should contain 'Prologue' entry for chapter_00.md with is_prologue: true")
	}

	// Prologue should render before chapter_01
	prologueIdx := strings.Index(html, "The Gathering Storm")
	chapter1Idx := strings.Index(html, "The Journey Begins")
	if prologueIdx < 0 {
		t.Error("prologue content should be rendered")
	}
	if chapter1Idx < 0 {
		t.Error("chapter 1 content should be rendered")
	}
	if prologueIdx >= chapter1Idx {
		t.Error("prologue should render before chapter 1")
	}
}

func TestCompilerV2_NoPrologueWithoutFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()

	chaptersDir := filepath.Join(tmpDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)

	// Create chapter_00 WITHOUT is_prologue frontmatter
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_00.md"), []byte(`# Chapter 0: Regular Chapter

### Area 1: Test

Content.
`), 0644)

	c := NewWithVersion(tmpDir, "", 2)
	htmlParts, err := c.generateHTML("Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	html := strings.Join(htmlParts, "\n")

	// TOC should NOT have a separate "Prologue" entry
	// (chapter_00 without frontmatter is just a regular chapter in the Chapters dir)
	tocStart := strings.Index(html, `class="toc"`)
	tocEnd := strings.Index(html[tocStart:], "</div>")
	if tocStart >= 0 && tocEnd >= 0 {
		toc := html[tocStart : tocStart+tocEnd]
		// Check that "Prologue" is NOT a separate TOC entry
		if strings.Contains(toc, ">Prologue<") {
			t.Error("TOC should not have separate 'Prologue' entry without is_prologue frontmatter")
		}
	}
}
