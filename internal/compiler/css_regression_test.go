package compiler_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/compiler"
)

// TestCSSRegression_DMSidebar tests that DM sidebar CSS is properly applied
func TestCSSRegression_DMSidebar(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="dm-sidebar">
<h5>DM Tip</h5>
<p>This is a DM tip.</p>
</div>
`)

	// Check for expected CSS properties
	checks := []string{
		".dm-sidebar",
		"background: #f8f4ec",
		"border-left: 4px solid #8b0000",
		"DM Only",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: missing '%s' in generated HTML", check)
		}
	}
}

// TestCSSRegression_StatBlockV2 tests stat-block-v2 CSS rendering
func TestCSSRegression_StatBlockV2(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="stat-block-v2">
<h3>Goblin</h3>
<div class="stat-line">
<span class="stat-label">AC</span>
<span class="stat-value">15</span>
</div>
</div>
`)

	checks := []string{
		".stat-block-v2",
		"linear-gradient",
		"border-top: 4px solid #8b0000",
		".stat-line",
		".stat-label",
		".stat-value",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: missing '%s' in generated HTML", check)
		}
	}
}

// TestCSSRegression_ShockPoint tests shock-point CSS with severity variants
func TestCSSRegression_ShockPoint(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		class    string
	}{
		{"mild", "mild", ".shock-point.mild"},
		{"moderate", "moderate", ".shock-point.moderate"},
		{"intense", "intense", ".shock-point.intense"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := generateCSSFixture(t, fmt.Sprintf(`
<div class="shock-point %s">
<span class="severity-badge">%s</span>
<strong>Test</strong>: Description
</div>
`, tt.severity, tt.severity))

			checks := []string{
				".shock-point",
				".severity-badge",
				"border-left: 4px solid",
			}

			for _, check := range checks {
				if !strings.Contains(html, check) {
					t.Errorf("CSS regression: missing '%s' in generated HTML", check)
				}
			}
		})
	}
}

// TestCSSRegression_SessionPrepCard tests session-prep-card CSS
func TestCSSRegression_SessionPrepCard(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="session-prep-card">
<h3>Session Prep</h3>
<div class="prep-item">Item 1</div>
<div class="prep-item">Item 2</div>
</div>
`)

	checks := []string{
		".session-prep-card",
		"border: 2px solid #5a3d2b",
		"border-radius: 6px",
		".prep-item",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: missing '%s' in generated HTML", check)
		}
	}
}

// TestCSSRegression_CharacterWorksheet tests character-worksheet CSS
func TestCSSRegression_CharacterWorksheet(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="character-worksheet">
<div class="worksheet-section">
<h4>Backstory</h4>
<div class="prompt-box">Prompt text</div>
</div>
</div>
`)

	checks := []string{
		".character-worksheet",
		"border: 2px dashed #c9ad6a",
		".worksheet-section",
		".prompt-box",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: missing '%s' in generated HTML", check)
		}
	}
}

// TestCSSRegression_EncounterRecommendation tests encounter-recommendation CSS
func TestCSSRegression_EncounterRecommendation(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="encounter-recommendation">
<span class="cr-badge">CR 1</span>
<span class="encounter-type">combat</span>
<strong>Encounter Name</strong>
</div>
`)

	checks := []string{
		".encounter-recommendation",
		".cr-badge",
		".encounter-type",
		"border-left: 4px solid #5a3d2b",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: missing '%s' in generated HTML", check)
		}
	}
}

// TestCSSRegression_CoverWrapper tests the cover page CSS hardening
// (REQ-3.1, 3.2, 3.3): .cover-wrapper must have BOTH legacy and modern
// page-break properties, use min-height (not fixed height), and the
// .cover-footer must be absolutely positioned so it cannot push overflow.
func TestCSSRegression_CoverWrapper(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Find the .cover-wrapper block
	classIdx := strings.Index(css, ".cover-wrapper {")
	if classIdx == -1 {
		classIdx = strings.Index(css, ".cover-wrapper{")
	}
	if classIdx == -1 {
		t.Fatal("CSS regression: '.cover-wrapper' class not found in CSS")
	}

	closeIdx := strings.Index(css[classIdx:], "}")
	if closeIdx == -1 {
		t.Fatal("CSS regression: could not find closing brace for .cover-wrapper")
	}
	block := css[classIdx : classIdx+closeIdx+1]

	// REQ-3.1: BOTH legacy and modern page-break properties
	if !strings.Contains(block, "page-break-after: always") {
		t.Errorf("CSS regression: .cover-wrapper missing legacy 'page-break-after: always'. Block: %s", block)
	}
	if !strings.Contains(block, "break-after: page") {
		t.Errorf("CSS regression: .cover-wrapper missing modern 'break-after: page'. Block: %s", block)
	}

	// REQ-3.2: min-height (not fixed height) as a safety net
	if !strings.Contains(block, "min-height: 297mm") {
		t.Errorf("CSS regression: .cover-wrapper missing 'min-height: 297mm'. Block: %s", block)
	}
	// Negative check: no bare 'height: 297mm' (would be too rigid).
	// Use a regex with a non-word boundary so it doesn't match 'min-height:'.
	fixedHeightRe := regexp.MustCompile(`(?:^|[^a-z\-])height:\s*297mm`)
	if fixedHeightRe.MatchString(block) {
		t.Errorf("CSS regression: .cover-wrapper still uses fixed 'height: 297mm' (should be min-height). Block: %s", block)
	}

	// position: relative needed for absolute footer positioning
	if !strings.Contains(block, "position: relative") {
		t.Errorf("CSS regression: .cover-wrapper missing 'position: relative' (needed for absolute footer). Block: %s", block)
	}

	// REQ-3.3: cover-footer must be absolutely positioned
	footerIdx := strings.Index(css, ".cover-footer {")
	if footerIdx == -1 {
		footerIdx = strings.Index(css, ".cover-footer{")
	}
	if footerIdx == -1 {
		t.Fatal("CSS regression: '.cover-footer' class not found in CSS")
	}
	footerCloseIdx := strings.Index(css[footerIdx:], "}")
	if footerCloseIdx == -1 {
		t.Fatal("CSS regression: could not find closing brace for .cover-footer")
	}
	footerBlock := css[footerIdx : footerIdx+footerCloseIdx+1]

	if !strings.Contains(footerBlock, "position: absolute") {
		t.Errorf("CSS regression: .cover-footer missing 'position: absolute'. Block: %s", footerBlock)
	}
	if !strings.Contains(footerBlock, "bottom:") {
		t.Errorf("CSS regression: .cover-footer missing 'bottom' declaration. Block: %s", footerBlock)
	}
}

// TestCSSRegression_StatBlockClassic tests the WotC stat-block CSS rules
// (REQ-2.9, 4.7): column-span: all, page-break-inside: avoid, WotC red
// borders, and all new sub-classes (.stat-line, .stat-label, .stat-value,
// .ability-scores, .property-line, .monster-type, .trait, .action,
// .actions-heading).
func TestCSSRegression_StatBlockClassic(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Find the .stat-block class block (not .stat-block-v2)
	classIdx := strings.Index(css, ".stat-block {")
	if classIdx == -1 {
		classIdx = strings.Index(css, ".stat-block{")
	}
	if classIdx == -1 {
		t.Fatal("CSS regression: '.stat-block' class not found in CSS")
	}

	closeIdx := strings.Index(css[classIdx:], "}")
	if closeIdx == -1 {
		t.Fatal("CSS regression: could not find closing brace for .stat-block")
	}
	block := css[classIdx : classIdx+closeIdx+1]

	// Required properties for the faithful WotC stat block
	checks := []struct {
		prop string
		desc string
	}{
		{"column-span: all", "must span both columns"},
		{"page-break-inside: avoid", "must avoid page break inside"},
		{"border-top: 4px solid #8b0000", "WotC red top border"},
		{"border-bottom: 4px solid #8b0000", "WotC red bottom border"},
	}
	for _, c := range checks {
		if !strings.Contains(block, c.prop) {
			t.Errorf("CSS regression: .stat-block missing %q (%s). Block: %s", c.prop, c.desc, block)
		}
	}

	// Sub-classes must exist in the CSS
	subClasses := []string{
		".stat-block .monster-type",
		".stat-block .stat-line",
		".stat-block .stat-label",
		".stat-block .stat-value",
		".stat-block .property-line",
		".stat-block .ability-scores",
		".stat-block .actions-heading",
		".stat-block .trait",
		".stat-block .action",
	}
	for _, sc := range subClasses {
		if !strings.Contains(css, sc) {
			t.Errorf("CSS regression: sub-class %q not found in CSS", sc)
		}
	}
}

// TestCSSRegression_Prologue tests prologue CSS classes
func TestCSSRegression_Prologue(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="prologue">
<h2 id="sec-prologue">Prólogo</h2>
<div class="prologue-part-1">
<h3>Gancho Narrativo</h3>
<div class="read-aloud">Hook text</div>
</div>
<div class="prologue-part-2">
<h3>Trasfondo</h3>
<p>Context text</p>
</div>
</div>
`)

	checks := []string{
		".prologue",
		"column-span: all",
		"page-break-inside: avoid",
		".prologue h2",
		".prologue h3",
		".prologue-part-1",
		".prologue-part-2",
		".prologue-part-3",
		".prologue-part-4",
		".prologue .read-aloud",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: missing '%s' in generated HTML", check)
		}
	}
}

// TestCSSRegression_PrologueDefaultStyles tests that prologue styles include expected CSS properties
func TestCSSRegression_PrologueDefaultStyles(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Find the .prologue class block
	classIdx := strings.Index(css, ".prologue {")
	if classIdx == -1 {
		classIdx = strings.Index(css, ".prologue{")
	}
	if classIdx == -1 {
		t.Fatal("CSS regression: '.prologue' class not found in CSS")
	}

	closeIdx := strings.Index(css[classIdx:], "}")
	if closeIdx == -1 {
		t.Fatal("CSS regression: could not find closing brace for .prologue")
	}
	block := css[classIdx : classIdx+closeIdx+1]

	if !strings.Contains(block, "column-span: all") {
		t.Errorf("CSS regression: .prologue missing 'column-span: all'. Block: %s", block)
	}
	if !strings.Contains(block, "page-break-inside: avoid") {
		t.Errorf("CSS regression: .prologue missing 'page-break-inside: avoid'. Block: %s", block)
	}
}

// TestCSSRegression_PageBreakAvoid tests that all div-based classes have page-break-inside: avoid
func TestCSSRegression_PageBreakAvoid(t *testing.T) {
	// Get the CSS directly
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// All div-based classes must have page-break-inside: avoid
	classesRequiringPageBreak := []string{
		".cover-wrapper",
		".toc",
		".stat-block",
		".read-aloud",
		".handout-section",
		".map-description",
		".campaign-image",
		".code-block",
		".dm-sidebar",
		".stat-block-v2",
		".session-prep-card",
		".shock-point",
		".character-worksheet",
		".character-worksheet .worksheet-section",
		".character-worksheet .prompt-box",
		".encounter-recommendation",
		".flowchart",
		".scene-description",
		".prologue",
	}

	for _, class := range classesRequiringPageBreak {
		// Find the class definition in CSS
		classIdx := strings.Index(css, class+" {")
		if classIdx == -1 {
			// Try with space before brace
			classIdx = strings.Index(css, class+"{")
		}
		if classIdx == -1 {
			t.Errorf("CSS regression: class '%s' not found in CSS", class)
			continue
		}

		// Find the closing brace for this class
		closeIdx := strings.Index(css[classIdx:], "}")
		if closeIdx == -1 {
			t.Errorf("CSS regression: could not find closing brace for class '%s'", class)
			continue
		}

		classBlock := css[classIdx : classIdx+closeIdx+1]

		// Check for page-break-inside: avoid
		if !strings.Contains(classBlock, "page-break-inside: avoid") {
			t.Errorf("CSS regression: class '%s' missing 'page-break-inside: avoid'. Block: %s", class, classBlock)
		}
	}
}

// generateCSSFixture generates HTML with the embedded CSS for testing
func generateCSSFixture(t *testing.T, bodyContent string) string {
	t.Helper()

	// Get the CSS from the compiler package
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		// CSS is embedded, this shouldn't happen
		css = ""
	}

	// Build minimal HTML document
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<style>%s</style>
</head>
<body>
%s
</body>
</html>`, css, bodyContent)

	return html
}

// TestCSS_NoConflictsWithExistingClasses tests that new CSS doesn't override existing classes
func TestCSS_NoConflictsWithExistingClasses(t *testing.T) {
	html := generateCSSFixture(t, `
<div class="stat-block">Old stat block</div>
<div class="read-aloud">Read aloud text</div>
<div class="toc">Table of contents</div>
`)

	// Verify existing classes still work
	checks := []string{
		".stat-block",
		".read-aloud",
		".toc",
	}

	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("CSS regression: existing class '%s' not found", check)
		}
	}
}

// TestCSS_SnapshotComparison compares generated HTML against saved snapshots
func TestCSS_SnapshotComparison(t *testing.T) {
	snapshotDir := "testdata/css-snapshots"
	
	// Create snapshot directory if it doesn't exist
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		_ = os.MkdirAll(snapshotDir, 0755)
	}

	tests := []struct {
		name    string
		content string
	}{
		{"dm-sidebar", `<div class="dm-sidebar"><h5>DM Tip</h5><p>Tip content</p></div>`},
		{"stat-block-v2", `<div class="stat-block-v2"><h3>Monster</h3><div class="stat-line"><span class="stat-label">AC</span><span class="stat-value">15</span></div></div>`},
		{"shock-point-mild", `<div class="shock-point mild"><span class="severity-badge">mild</span>Content</div>`},
		{"shock-point-moderate", `<div class="shock-point moderate"><span class="severity-badge">moderate</span>Content</div>`},
		{"shock-point-intense", `<div class="shock-point intense"><span class="severity-badge">intense</span>Content</div>`},
		{"session-prep-card", `<div class="session-prep-card"><h3>Prep</h3><div class="prep-item">Item</div></div>`},
		{"character-worksheet", `<div class="character-worksheet"><div class="worksheet-section"><h4>Section</h4><div class="prompt-box">Prompt</div></div></div>`},
		{"encounter-recommendation", `<div class="encounter-recommendation"><span class="cr-badge">CR 1</span><span class="encounter-type">combat</span>Name</div>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html := generateCSSFixture(t, tt.content)
			snapshotPath := filepath.Join(snapshotDir, tt.name+".html")

			// If snapshot doesn't exist, create it
			if _, err := os.Stat(snapshotPath); os.IsNotExist(err) {
				if err := os.WriteFile(snapshotPath, []byte(html), 0644); err != nil {
				t.Fatalf("Failed to write snapshot: %v", err)
			}
				t.Logf("Created snapshot: %s", snapshotPath)
				return
			}

			// Compare against snapshot
			expected, err := os.ReadFile(snapshotPath)
			if err != nil {
				t.Fatalf("Failed to read snapshot: %v", err)
			}

			if string(expected) != html {
				t.Errorf("HTML differs from snapshot. Run with -update to update snapshots.")
			}
		})
	}
}
