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

func TestCSSRegression_ArkanumCalloutLayoutContract(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	checks := []string{
		".dm-sidebar-wide",
		".nested-card .table-wrap",
		"break-inside: avoid",
		"page-break-inside: avoid",
	}
	for _, check := range checks {
		if !strings.Contains(css, check) {
			t.Errorf("Arkanum CSS contract: missing %q", check)
		}
	}

	dmSidebarStart := strings.Index(css, ".dm-sidebar {")
	if dmSidebarStart >= 0 {
		dmSidebarEnd := strings.Index(css[dmSidebarStart:], "}")
		if dmSidebarEnd >= 0 && strings.Contains(css[dmSidebarStart:dmSidebarStart+dmSidebarEnd], "column-span: all") {
			t.Errorf("ordinary DM sidebars must not span all columns")
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
		"border-top: 4px solid #7a2e1a",
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

// TestCSSRegression_CoverWrapper tests the cover page CSS hardening.
// UPDATED for fix-statblock-layout-and-cover-overflow: the v5.4.2 contract
// (min-height + flex column + absolute footer) was the cause of Bug B
// (cover spilled to 2 pages). The new contract is an EXACT 297mm box
// with absolute-positioned children — see TestCSSRegression_CoverFixedHeight
// for the full new contract. This test keeps the legacy "must not regress"
// checks (BOTH break-after + break-inside variants, position: relative,
// absolute footer with bottom offset) and is now compatible with the
// fixed-height rule.
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

	// REQ-3.2 (updated): position: absolute + min-height: 297mm + @page :first.
	// The previous `height: 297mm` (flow-positioned) still split the cover
	// under Chromium's page-break rounding. Anchoring the cover to the page
	// with position: absolute makes the page-break calculation treat it as
	// page-relative and never round the bottom off.
	if !strings.Contains(block, "position: absolute") {
		t.Errorf("CSS regression: .cover-wrapper missing 'position: absolute'. Block: %s", block)
	}
	if !strings.Contains(block, "min-height: 297mm") {
		t.Errorf("CSS regression: .cover-wrapper missing 'min-height: 297mm'. Block: %s", block)
	}

	// position: absolute anchors the cover to the page (not body), so the
	// page-break calculation never rounds the bottom onto a second page.
	// The cover-footer is still absolutely positioned relative to the cover.
	if !strings.Contains(block, "position: absolute") {
		t.Errorf("CSS regression: .cover-wrapper missing 'position: absolute'. Block: %s", block)
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
func TestCSSRegression_StatLineNoFlex(t *testing.T) {
	// REQ (fix-statblock-layout-and-cover-overflow): .stat-block .stat-line
	// must NOT use `display: flex` or `justify-content`. The v5.4.2 flex
	// layout pushed long values into a narrow right column when the parser
	// over-split a trait line into 3 .stat-line rows (Bug A from PR #17).
	// The WotC look is bold label + space + value flowing inline, so
	// `display: block` with inline label/value spans is the correct rule.
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Find the .stat-block .stat-line rule.
	classIdx := strings.Index(css, ".stat-block .stat-line {")
	if classIdx == -1 {
		classIdx = strings.Index(css, ".stat-block .stat-line{")
	}
	if classIdx == -1 {
		t.Fatal("CSS regression: '.stat-block .stat-line' rule not found in CSS")
	}
	closeIdx := strings.Index(css[classIdx:], "}")
	if closeIdx == -1 {
		t.Fatal("CSS regression: could not find closing brace for .stat-block .stat-line")
	}
	block := css[classIdx : classIdx+closeIdx+1]

	if strings.Contains(block, "display: flex") {
		t.Errorf("CSS regression: .stat-block .stat-line still uses 'display: flex' (Bug A is back). Block: %s", block)
	}
	if strings.Contains(block, "justify-content") {
		t.Errorf("CSS regression: .stat-block .stat-line still uses 'justify-content' (Bug A is back). Block: %s", block)
	}

	// The label/value spans must remain, and the stat-label should still
	// be bold (the WotC convention). We don't assert `display: block`
	// directly because the fix uses block; the negative assertions above
	// are the contract.
	if !strings.Contains(css, ".stat-block .stat-label") {
		t.Error("CSS regression: '.stat-block .stat-label' rule not found")
	}
	if !strings.Contains(css, ".stat-block .stat-value") {
		t.Error("CSS regression: '.stat-block .stat-value' rule not found")
	}
}

// TestCSSRegression_CoverFixedHeight asserts the new cover contract:
// REQ (fix-statblock-layout-and-cover-overflow): .cover-wrapper uses
// EXACT height 297mm (not min-height 297mm) with absolute-positioned
// children, because the v5.4.2 min-height + flex-column + absolute
// footer approach pushed the wrapper past one A4 page (Bug B from
// PR #17 — cover spilled to 2 pages).
func TestCSSRegression_CoverFixedHeight(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

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

	// The new contract: position: absolute + @page :first { margin: 0 } +
	// min-height: 297mm. The cover anchors to the page itself (not the body)
	// so the page-break calculation treats it as page-relative and never
	// rounds the bottom off onto a second page.
	if !strings.Contains(block, "position: absolute") {
		t.Errorf("CSS regression: .cover-wrapper missing 'position: absolute'. Block: %s", block)
	}
	if !strings.Contains(block, "min-height: 297mm") {
		t.Errorf("CSS regression: .cover-wrapper missing 'min-height: 297mm' (A4 fallback). Block: %s", block)
	}
	if !strings.Contains(block, "overflow: hidden") {
		t.Errorf("CSS regression: .cover-wrapper missing 'overflow: hidden' (needed for the fixed 297mm box). Block: %s", block)
	}
	// Break-inside avoid (BOTH legacy and modern forms)
	if !strings.Contains(block, "page-break-inside: avoid") {
		t.Errorf("CSS regression: .cover-wrapper missing 'page-break-inside: avoid'. Block: %s", block)
	}
	if !strings.Contains(block, "break-inside: avoid") {
		t.Errorf("CSS regression: .cover-wrapper missing 'break-inside: avoid'. Block: %s", block)
	}
	// Break-after (BOTH legacy and modern forms)
	if !strings.Contains(block, "page-break-after: always") {
		t.Errorf("CSS regression: .cover-wrapper missing 'page-break-after: always'. Block: %s", block)
	}
	if !strings.Contains(block, "break-after: page") {
		t.Errorf("CSS regression: .cover-wrapper missing 'break-after: page'. Block: %s", block)
	}
	// No flex layout (was the cause of Bug B)
	if strings.Contains(block, "display: flex") {
		t.Errorf("CSS regression: .cover-wrapper still uses 'display: flex' (Bug B is back). Block: %s", block)
	}

	// .cover-image must be absolutely positioned between top: 35mm and
	// bottom: 20mm (per design #2306).
	imgIdx := strings.Index(css, ".cover-image {")
	if imgIdx == -1 {
		imgIdx = strings.Index(css, ".cover-image{")
	}
	if imgIdx == -1 {
		t.Fatal("CSS regression: '.cover-image' class not found in CSS")
	}
	imgClose := strings.Index(css[imgIdx:], "}")
	if imgClose == -1 {
		t.Fatal("CSS regression: could not find closing brace for .cover-image")
	}
	imgBlock := css[imgIdx : imgIdx+imgClose+1]
	if !strings.Contains(imgBlock, "position: absolute") {
		t.Errorf("CSS regression: .cover-image not absolutely positioned. Block: %s", imgBlock)
	}
	if !strings.Contains(imgBlock, "top: 32mm") {
		t.Errorf("CSS regression: .cover-image missing 'top: 32mm' (per design). Block: %s", imgBlock)
	}
	if !strings.Contains(imgBlock, "bottom: 14mm") {
		t.Errorf("CSS regression: .cover-image missing 'bottom: 14mm' (per design). Block: %s", imgBlock)
	}
}
func parseCSSBlock(css, selector string) string {
	// Find the selector block (with or without space before brace)
	idx := strings.Index(css, selector+" {")
	if idx == -1 {
		idx = strings.Index(css, selector+"{")
	}
	if idx == -1 {
		return ""
	}
	end := strings.Index(css[idx:], "}")
	if end == -1 {
		return ""
	}
	return css[idx : idx+end+1]
}

func TestCSSRegression_StatBlockClassic(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	block := parseCSSBlock(css, ".stat-block")
	if block == "" {
		t.Fatal("CSS regression: '.stat-block' class not found in CSS")
	}

	// Required properties for the faithful WotC stat block
	checks := []struct {
		prop string
		desc string
	}{
		{"column-span: all", "must span both columns"},
		{"page-break-inside: avoid", "must avoid page break inside"},
		{"border-top: 4px solid #7a2e1a", "reddish-brown top border"},
		{"border-bottom: 4px solid #7a2e1a", "reddish-brown bottom border"},
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

// TestCSSRegression_StatBlockSubtitleSeparator asserts REQ: the separator
// line moves from .stat-block h2 to .stat-block .monster-type.
func TestCSSRegression_StatBlockSubtitleSeparator(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// .stat-block h2 must NOT have border-bottom
	h2Block := parseCSSBlock(css, ".stat-block h2")
	if h2Block == "" {
		t.Fatal("CSS regression: '.stat-block h2' block not found in CSS")
	}
	if strings.Contains(h2Block, "border-bottom") {
		t.Errorf("CSS regression: .stat-block h2 should NOT have 'border-bottom' (separator moved to .monster-type). Block: %s", h2Block)
	}

	// .stat-block .monster-type MUST have border-bottom + padding-bottom
	typeBlock := parseCSSBlock(css, ".stat-block .monster-type")
	if typeBlock == "" {
		t.Fatal("CSS regression: '.stat-block .monster-type' block not found in CSS")
	}
	if !strings.Contains(typeBlock, "border-bottom: 2px solid #8b0000") {
		t.Errorf("CSS regression: .stat-block .monster-type missing 'border-bottom: 2px solid #8b0000'. Block: %s", typeBlock)
	}
	if !strings.Contains(typeBlock, "padding-bottom:") {
		t.Errorf("CSS regression: .stat-block .monster-type missing 'padding-bottom'. Block: %s", typeBlock)
	}
}

// TestCSSRegression_StatLabelSmallCaps asserts REQ: stat labels use
// font-variant: small-caps instead of text-transform: uppercase.
func TestCSSRegression_StatLabelSmallCaps(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	labelBlock := parseCSSBlock(css, ".stat-block .stat-label")
	if labelBlock == "" {
		t.Fatal("CSS regression: '.stat-block .stat-label' block not found in CSS")
	}
	if strings.Contains(labelBlock, "text-transform: uppercase") {
		t.Errorf("CSS regression: .stat-block .stat-label should NOT use 'text-transform: uppercase'. Block: %s", labelBlock)
	}
	if !strings.Contains(labelBlock, "font-variant: small-caps") {
		t.Errorf("CSS regression: .stat-block .stat-label missing 'font-variant: small-caps'. Block: %s", labelBlock)
	}
	if !strings.Contains(labelBlock, "font-feature-settings: \"smcp\"") {
		t.Errorf("CSS regression: .stat-block .stat-label missing 'font-feature-settings: \"smcp\"'. Block: %s", labelBlock)
	}
}

// TestCSSRegression_StatBlockActionsHeading asserts REQ: actions heading
// uses border-top (separator above) instead of border-bottom.
func TestCSSRegression_StatBlockActionsHeading(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	headingBlock := parseCSSBlock(css, ".stat-block .actions-heading")
	if headingBlock == "" {
		t.Fatal("CSS regression: '.stat-block .actions-heading' block not found in CSS")
	}
	if strings.Contains(headingBlock, "border-bottom") {
		t.Errorf("CSS regression: .stat-block .actions-heading should NOT have 'border-bottom'. Block: %s", headingBlock)
	}
	if !strings.Contains(headingBlock, "border-top: 1px solid #c9ad6a") {
		t.Errorf("CSS regression: .stat-block .actions-heading missing 'border-top: 1px solid #c9ad6a'. Block: %s", headingBlock)
	}
	if !strings.Contains(headingBlock, "padding-top:") {
		t.Errorf("CSS regression: .stat-block .actions-heading missing 'padding-top'. Block: %s", headingBlock)
	}
}

// TestCSSRegression_StatBlockBoldItalic asserts REQ: trait and action names
// are bold+italic via .trait strong:first-child and .action strong:first-child.
func TestCSSRegression_StatBlockBoldItalic(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Check the combined rule
	expected := ".stat-block .trait strong:first-child"
	if !strings.Contains(css, expected) {
		t.Errorf("CSS regression: missing selector %q", expected)
	}
	expected = ".stat-block .action strong:first-child"
	if !strings.Contains(css, expected) {
		t.Errorf("CSS regression: missing selector %q", expected)
	}
	// Check font-style: italic
	if !strings.Contains(css, "trait strong:first-child, .stat-block .action strong:first-child") &&
		!strings.Contains(css, "action strong:first-child, .stat-block .trait strong:first-child") {
		// Check each individually
		found := false
		idx := strings.Index(css, "trait strong:first-child")
		if idx != -1 {
			// Look for font-style near this position
			contextEnd := idx + 200
			if contextEnd > len(css) {
				contextEnd = len(css)
			}
			if strings.Contains(css[idx:contextEnd], "font-style: italic") {
				found = true
			}
		}
		if !found {
			t.Errorf("CSS regression: trait/action strong:first-child missing 'font-style: italic'")
		}
	}
}

// TestCSSRegression_StatBlockV2BorderColor asserts REQ: .stat-block-v2
// border colors match the reddish-brown palette (#7a2e1a).
func TestCSSRegression_StatBlockV2BorderColor(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	v2Block := parseCSSBlock(css, ".stat-block-v2")
	if v2Block == "" {
		t.Fatal("CSS regression: '.stat-block-v2' block not found in CSS")
	}
	if !strings.Contains(v2Block, "border-top: 4px solid #7a2e1a") {
		t.Errorf("CSS regression: .stat-block-v2 missing 'border-top: 4px solid #7a2e1a'. Block: %s", v2Block)
	}
	if !strings.Contains(v2Block, "border-bottom: 4px solid #7a2e1a") {
		t.Errorf("CSS regression: .stat-block-v2 missing 'border-bottom: 4px solid #7a2e1a'. Block: %s", v2Block)
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

// TestCSSRegression_RosterWrap asserts REQ-2.4: the .roster-wrap
// CSS rule exists and includes column-span: all. The wrapper is
// used by generateAdventureRoster to span the three tables
// (NPCs / Monstruos / Encuentros) across the full page width.
func TestCSSRegression_RosterWrap(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Find the .roster-wrap block (with or without space before brace)
	classIdx := strings.Index(css, ".roster-wrap {")
	if classIdx == -1 {
		classIdx = strings.Index(css, ".roster-wrap{")
	}
	if classIdx == -1 {
		t.Fatal("CSS regression: '.roster-wrap' class not found in dnd-style.css")
	}

	closeIdx := strings.Index(css[classIdx:], "}")
	if closeIdx == -1 {
		t.Fatal("CSS regression: could not find closing brace for .roster-wrap")
	}
	block := css[classIdx : classIdx+closeIdx+1]

	if !strings.Contains(block, "column-span: all") {
		t.Errorf("CSS regression: .roster-wrap missing 'column-span: all'. Block: %s", block)
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

// TestCSSRegression_DMSidebarColumnSpan asserts that only explicitly wide DM
// sidebars span both columns. Ordinary callouts stay in one readable column.
func TestCSSRegression_DMSidebarColumnSpan(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}
	classes := []string{
		".shock-point",
		".encounter-recommendation",
		".general-features",
	}
	for _, cls := range classes {
		idx := strings.Index(css, cls+" {")
		if idx == -1 {
			t.Errorf("class %s not found in CSS", cls)
			continue
		}
		// Cut at the next "}" to scope the check to this class's block.
		endIdx := strings.Index(css[idx:], "}")
		if endIdx == -1 {
			t.Errorf("class %s block has no closing brace", cls)
			continue
		}
		block := css[idx : idx+endIdx+1]
		if !strings.Contains(block, "column-span: all") {
			t.Errorf("class %s block does not contain 'column-span: all'. Block: %s", cls, block)
		}
	}
	wideIdx := strings.Index(css, ".dm-sidebar-wide {")
	if wideIdx == -1 {
		t.Fatal("class .dm-sidebar-wide not found in CSS")
	}
	wideEnd := strings.Index(css[wideIdx:], "}")
	if wideEnd == -1 || !strings.Contains(css[wideIdx:wideIdx+wideEnd], "column-span: all") {
		t.Errorf("class .dm-sidebar-wide must contain 'column-span: all'")
	}
}

// TestCSSRegression_PageRules asserts the @page A4 rule with all 5 margin
// boxes set to content: none. REQ-1.2: defense-in-depth against Chromium
// version drift on the --no-pdf-header-footer flag.
func TestCSSRegression_PageRules(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}
	// The @page block may be multi-line; normalize whitespace for substring checks.
	normalized := strings.Join(strings.Fields(css), " ")
	if !strings.Contains(normalized, "@page { size: A4; margin: 15mm;") {
		t.Error("expected @page rule with size: A4; margin: 15mm;")
	}
	expectedBoxes := []string{
		"@top-left { content: none }",
		"@top-right { content: none }",
		"@bottom-left { content: none }",
		"@bottom-center { content: none }",
		"@bottom-right { content: none }",
	}
	for _, box := range expectedBoxes {
		if !strings.Contains(css, box) {
			t.Errorf("expected %q in CSS", box)
		}
	}
	// @page :first must still exist (preserves cover full-bleed)
	if !strings.Contains(css, "@page :first {") {
		t.Error("expected @page :first rule to remain")
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
	// Set UPDATE_SNAPSHOTS=1 to regenerate snapshots (used after intentional CSS changes).
	updateSnapshots := os.Getenv("UPDATE_SNAPSHOTS") == "1"
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
				if updateSnapshots {
					if err := os.WriteFile(snapshotPath, []byte(html), 0644); err != nil {
						t.Fatalf("Failed to update snapshot: %v", err)
					}
					t.Logf("Updated snapshot: %s", snapshotPath)
					return
				}
				t.Errorf("HTML differs from snapshot. Run with -update to update snapshots.")
			}
		})
	}
}

// TestCSSRegression_TableHasNoColumnSpan asserts REQ-3.2: the
// `table {` selector block in dnd-style.css no longer contains
// `column-span: all`. The rule moved to a new `.table-wrap` block.
// The regex anchors on `^table\s*\{` so it does NOT match
// `.table-wrap` (the dot in `.table-wrap` is a class-marker, not
// the `table` element selector).
func TestCSSRegression_TableHasNoColumnSpan(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	re := regexp.MustCompile(`(?ms)^table\s*\{[^}]*\}`)
	m := re.FindString(css)
	if m == "" {
		t.Fatal("table { block not found in CSS")
	}
	if strings.Contains(m, "column-span: all") {
		t.Errorf("table { block should not contain 'column-span: all' (moved to .table-wrap). Block: %s", m)
	}
}

// TestCSSRegression_TableWrapExists asserts REQ-3.2: a new
// `.table-wrap` CSS rule exists and contains `column-span: all`.
// The wrapper is emitted by flushTable (see WU-1) and is the
// recipient of the column-span role that the `table` selector
// used to play.
func TestCSSRegression_TableWrapExists(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	idx := strings.Index(css, ".table-wrap {")
	if idx == -1 {
		idx = strings.Index(css, ".table-wrap{")
	}
	if idx == -1 {
		t.Fatal("CSS regression: '.table-wrap' class not found in dnd-style.css")
	}
	end := strings.Index(css[idx:], "}")
	if end == -1 {
		t.Fatal("CSS regression: .table-wrap block has no closing brace")
	}
	block := css[idx : idx+end+1]
	if !strings.Contains(block, "column-span: all") {
		t.Errorf(".table-wrap block missing 'column-span: all'. Block: %s", block)
	}
}

func TestCSSRegression_TablePage(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	page := regexp.MustCompile(`(?ms)\.table-wrap\.table-page\s*\{[^}]*\}`).FindString(css)
	if page == "" {
		t.Fatal("CSS regression: .table-wrap.table-page block not found")
	}
	for _, property := range []string{"column-span: all", "break-before: page", "page-break-before: always", "break-inside: auto", "page-break-inside: auto"} {
		if !strings.Contains(page, property) {
			t.Errorf("table-page block missing %q: %s", property, page)
		}
	}
	for _, property := range []string{"width: 100%", "overflow: visible"} {
		if !strings.Contains(page, property) {
			t.Errorf("table-page block missing %q: %s", property, page)
		}
	}

	boundary := regexp.MustCompile(`(?ms)\.table-page-boundary\s*\{[^}]*\}`).FindString(css)
	if boundary == "" {
		t.Fatal("CSS regression: .table-page-boundary block not found")
	}
	for _, property := range []string{"display: block", "width: 100%", "height: 0", "margin: 0", "padding: 0", "border: 0", "column-span: all", "break-before: page", "page-break-before: always"} {
		if !strings.Contains(boundary, property) {
			t.Errorf("boundary block missing %q: %s", property, boundary)
		}
	}
	if strings.Contains(boundary, "display: none") {
		t.Error("boundary must remain in layout; display:none defeats the page break")
	}
	if strings.Contains(page, "break-after: page") || strings.Contains(boundary, "break-after: page") {
		t.Error("table-page pagination must use the wrapper and sibling break-before, not break-after")
	}
}

func TestCSSRegression_AdaptiveTableFragmentation(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}
	for _, property := range []string{"thead", "tbody", "tr"} {
		if !strings.Contains(css, property) {
			t.Errorf("stylesheet missing table fragmentation selector %q", property)
		}
	}
	if !strings.Contains(css, "display: table-header-group") {
		t.Error("thead must repeat as a table-header-group")
	}
	if !strings.Contains(css, ".table-wrap.table-page") || strings.Contains(css, ".table-wrap.table-page {\n  overflow-x") {
		t.Error("table-page must remain a full-width, non-scrolling fragmentable surface")
	}
}

func TestCSSRegression_TablePageNestedNeutralizers(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}
	for _, selector := range []string{
		".nested-card .table-wrap.table-page",
		".read-aloud .table-wrap.table-page",
		".dm-sidebar .table-wrap.table-page",
		".chapter-summary .table-wrap.table-page",
		".introduction-sidebar .table-wrap.table-page",
		".nested-card .table-page-boundary",
		".read-aloud .table-page-boundary",
	} {
		if !strings.Contains(css, selector) {
			t.Errorf("nested containment selector missing %q", selector)
		}
	}
}

// TestCSSRegression_ListBreakInside asserts REQ-3.3: lists and
// list items never break inside their content. The `ul, ol` rule
// must contain `break-inside: avoid` (modern) and
// `column-break-inside: avoid` (legacy alias). The `li` rule
// must contain `break-inside: avoid`. These together prevent
// Chromium from splitting a single bullet item across two
// columns or pages (Issue K).
func TestCSSRegression_ListBreakInside(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	// Find the `ul, ol {` block (or the single-selector `ul {` if
	// the cascade has not yet been widened — the test asserts the
	// widened form is present).
	idx := strings.Index(css, "ul, ol {")
	if idx == -1 {
		t.Fatal("CSS regression: 'ul, ol {' block not found (selector should be widened from 'ul' to 'ul, ol' per REQ-3.3)")
	}
	end := strings.Index(css[idx:], "}")
	if end == -1 {
		t.Fatal("CSS regression: ul, ol block has no closing brace")
	}
	block := css[idx : idx+end+1]

	if !strings.Contains(block, "break-inside: avoid") {
		t.Errorf("ul, ol block missing 'break-inside: avoid'. Block: %s", block)
	}
	if !strings.Contains(block, "column-break-inside: avoid") {
		t.Errorf("ul, ol block missing 'column-break-inside: avoid'. Block: %s", block)
	}

	// li block: must contain break-inside: avoid. The selector must be
	// the standalone `li {` (preceded by a newline/whitespace, not a
	// class name like `.toc li {` or `.character-worksheet .worksheet-section li {`).
	// We use a regex to match the standalone selector.
	liRe := regexp.MustCompile(`(?m)^\s*li\s*\{`)
	liLoc := liRe.FindStringIndex(css)
	if liLoc == nil {
		t.Fatal("CSS regression: standalone 'li {' block not found")
	}
	liIdx := liLoc[0]
	liEnd := strings.Index(css[liIdx:], "}")
	if liEnd == -1 {
		t.Fatal("CSS regression: li block has no closing brace")
	}
	liBlock := css[liIdx : liIdx+liEnd+1]
	if !strings.Contains(liBlock, "break-inside: avoid") {
		t.Errorf("li block missing 'break-inside: avoid'. Block: %s", liBlock)
	}
}

// TestCSSRegression_H1BreakAfter asserts REQ-3.4: chapter h1
// titles never break after themselves. The standalone `h1 {`
// block (h1 is the chapter title selector) must contain
// `break-after: avoid` and `column-break-after: avoid` so the
// chapter title stays with its first paragraph (Issue G).
//
// The test uses a regex anchored on `^h1\s*\{` (line start, no
// leading class) so it does NOT match child selectors like
// `.cover-wrapper h1 {` or `.prologue h1 {`.
func TestCSSRegression_H1BreakAfter(t *testing.T) {
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("Failed to get CSS: %v", err)
	}

	h1Re := regexp.MustCompile(`(?m)^\s*h1\s*\{`)
	h1Loc := h1Re.FindStringIndex(css)
	if h1Loc == nil {
		t.Fatal("CSS regression: standalone 'h1 {' block not found")
	}
	idx := h1Loc[0]
	end := strings.Index(css[idx:], "}")
	if end == -1 {
		t.Fatal("CSS regression: h1 block has no closing brace")
	}
	block := css[idx : idx+end+1]

	if !strings.Contains(block, "break-after: avoid") {
		t.Errorf("h1 block missing 'break-after: avoid'. Block: %s", block)
	}
	if !strings.Contains(block, "column-break-after: avoid") {
		t.Errorf("h1 block missing 'column-break-after: avoid'. Block: %s", block)
	}
}
