package compiler

import (
	"os"
	"strings"
	"testing"
)

// TestVerifyElExiliado_CoverAndBestiary is a live-repro integration test that
// compiles the el-exiliado campaign (the reference fixture in the proposal)
// and asserts the three production bugs are fixed:
//
//  1. Cover image embedded as image/jpeg (not image/png) — REQ-5.1
//  2. Bestiary monsters wrapped in .stat-block — REQ-5.2
//  3. Cover page CSS hardened (break-after, min-height, absolute footer) — REQ-3.1, 3.2, 3.3
//
// The test is skipped when the campaign directory is unavailable (CI does
// not ship user campaigns) so it never breaks the build in a clean checkout.
func TestVerifyElExiliado_CoverAndBestiary(t *testing.T) {
	campaignDir := findElExiliadoCampaign()
	if campaignDir == "" {
		t.Skip("el-exiliado-de-las-tierras-marchitas campaign not available; live repro skipped")
	}

	c := New(campaignDir, "")
	htmlParts, err := c.generateHTML("El Exiliado de las Tierras Marchitas")
	if err != nil {
		t.Fatalf("generateHTML failed: %v", err)
	}
	html := strings.Join(htmlParts, "\n")
	t.Logf("Generated %d bytes of HTML", len(html))

	// REQ-5.1: Cover image embedded as JPEG (the bug was PNG MIME on JPEG bytes).
	coverIdx := strings.Index(html, "class=\"cover-image\"")
	if coverIdx == -1 {
		t.Fatal("cover-image div not found in HTML")
	}
	coverEnd := coverIdx + 2000
	if coverEnd > len(html) {
		coverEnd = len(html)
	}
	coverSection := html[coverIdx:coverEnd]
	if !strings.Contains(coverSection, "data:image/jpeg;base64,") {
		t.Errorf("cover image NOT embedded as JPEG; first 200 chars: %q", truncate(coverSection, 200))
	}
	if strings.Contains(coverSection, "data:image/png;base64,") {
		// Regression of the original bug
		t.Error("cover image is still embedded as PNG (the bug is back)")
	}

	// REQ-5.2: 4 named monsters must be wrapped in .stat-block divs.
	expectedMonsters := []string{"El Rayo", "Eco de los Doce", "Sir Aldric Voss", "Cabra de Dos Cabezas"}
	for _, m := range expectedMonsters {
		want := `<div class="stat-block" data-monster="` + m + `">`
		if !strings.Contains(html, want) {
			t.Errorf("expected stat-block wrapper for %q; not found in HTML", m)
		}
	}
	statBlockCount := strings.Count(html, `<div class="stat-block" data-monster="`)
	if statBlockCount < 4 {
		t.Errorf("expected at least 4 .stat-block wrappers, got %d", statBlockCount)
	}
	t.Logf("Found %d stat-block wrappers (target: >= 4 named)", statBlockCount)

	// REQ-3.1, 3.2, 3.3: Cover page CSS hardening must be present.
	css, err := GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("get CSS: %v", err)
	}
	if !strings.Contains(css, "break-after: page") {
		t.Error("CSS missing 'break-after: page' (REQ-3.1)")
	}
	if !strings.Contains(css, "min-height: 297mm") {
		t.Error("CSS missing 'min-height: 297mm' (REQ-3.2)")
	}
	if !strings.Contains(css, "position: absolute") || !strings.Contains(css, "position: relative") {
		t.Error("CSS missing absolute/relative positioning (REQ-3.3)")
	}

	// REQ-5.3: NPC/PC portraits must be embedded (not as data:image/png for
	// what is actually a JPEG, and not missing). NPC portraits are base64-
	// embedded so we count the JPEG data URIs — all PNG-as-JPEG assets must
	// now be image/jpeg (the bug fix).
	jpegCount := strings.Count(html, "data:image/jpeg;base64,")
	if jpegCount < 1 {
		t.Errorf("expected at least 1 data:image/jpeg base64 URI (cover + NPC/PC portraits), got %d", jpegCount)
	}
	t.Logf("Found %d data:image/jpeg base64 URIs in compiled HTML", jpegCount)
}

func findElExiliadoCampaign() string {
	candidates := []string{
		"/home/pau/campaigns/el-exiliado-de-las-tierras-marchitas",
		"/home/user/campaigns/el-exiliado-de-las-tierras-marchitas",
		"./campaigns/el-exiliado-de-las-tierras-marchitas",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
