package compiler_test

import (
	"regexp"
	"testing"

	"github.com/pauvalls/grimorio/internal/compiler"
)

// assertCSSRuleContains asserts that the CSS block whose selector list contains
// `selector` (allowing comma-separated lists like `table, tr, td, th`) also
// contains a `prop:` declaration. Used by the v5.0.2 PDF-render CSS regression
// tests.
func assertCSSRuleContains(t *testing.T, css, selector, prop string) {
	t.Helper()
	// Allow the target selector to appear at the start of the selector list
	// or after a comma: e.g. `table {`, `table, tr, td, th {`.
	sel := regexp.QuoteMeta(selector)
	re := regexp.MustCompile(`(?:^|[\s,])` + sel + `(?:[\s,]|$)[^{]*\{[^}]*` + regexp.QuoteMeta(prop) + `\s*:`)
	if !re.MatchString(css) {
		t.Errorf("CSS block containing %q must contain %q; regex did not match", selector, prop)
	}
}

// loadCSS returns the embedded dnd-style CSS, failing the test on error.
func loadCSS(t *testing.T) string {
	t.Helper()
	css, err := compiler.GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("GetTemplate(dnd-style) error: %v", err)
	}
	if css == "" {
		t.Fatal("dnd-style CSS is empty")
	}
	return css
}

// TestCoverWrapperColumnSpan asserts that the .cover-wrapper CSS block contains
// the `column-span: all` declaration, so the cover breaks out of the two-column
// body flow in PDF output.
func TestCoverWrapperColumnSpan(t *testing.T) {
	assertCSSRuleContains(t, loadCSS(t), ".cover-wrapper", "column-span")
}

// TestTableLayoutFixed asserts the `table` rule declares `table-layout: fixed`
// so wide tables fit within the page width and don't overflow.
func TestTableLayoutFixed(t *testing.T) {
	assertCSSRuleContains(t, loadCSS(t), "table", "table-layout")
}

// TestTdThOverflowWrap asserts the shared `td` / `th` block contains
// `overflow-wrap: break-word` so long cell content wraps instead of overflowing.
func TestTdThOverflowWrap(t *testing.T) {
	css := loadCSS(t)
	// The shared selector at L235 is `table, tr, td, th` — assert both
	// `td` and `th` blocks contain the wrap property.
	assertCSSRuleContains(t, css, "td", "overflow-wrap")
	assertCSSRuleContains(t, css, "th", "overflow-wrap")
}
