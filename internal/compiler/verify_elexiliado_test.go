package compiler

import (
	"fmt"
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

	// REQ (fix-statblock-layout-and-cover-overflow): El Rayo's hero image
	// must be hoisted INSIDE the .stat-block div (Bug C from PR #17).
	// Grep the El Rayo stat-block for an <img> tag; the image is embedded
	// as a base64 data URI by the existing processImages flow, so the
	// source filename is not present in the compiled HTML.
	elRayoOpenIdx := strings.Index(html, `<div class="stat-block" data-monster="El Rayo">`)
	if elRayoOpenIdx != -1 {
		// Walk the div nesting to find the matching </div>.
		depth := 0
		elRayoEnd := -1
		for j := elRayoOpenIdx; j < len(html); {
			switch {
			case strings.HasPrefix(html[j:], `<div`):
				depth++
				j += len("<div")
			case strings.HasPrefix(html[j:], `</div>`):
				depth--
				if depth == 0 {
					elRayoEnd = j + len("</div>")
					j = elRayoEnd
					break
				}
				j += len("</div>")
			default:
				j++
			}
			if elRayoEnd != -1 {
				break
			}
		}
		if elRayoEnd == -1 {
			t.Error("could not find matching </div> for El Rayo stat-block")
		} else {
			elRayoBlock := html[elRayoOpenIdx:elRayoEnd]
			if !strings.Contains(elRayoBlock, `<img`) {
				t.Error("El Rayo stat-block does NOT contain the hoisted hero image (Bug C is back)")
			} else {
				t.Logf("El Rayo stat-block contains the hoisted hero image (Bug C fixed)")
			}
		}
	}

	// REQ-3.1, 3.2, 3.3: Cover page CSS hardening must be present.
	// UPDATED: the contract is `height: calc(297mm - 20px)` (A4 minus body
	// vertical padding) + `max-height: calc(297mm - 20px)` + `overflow: hidden`.
	// The previous exact `height: 297mm` + negative margin spilled the cover
	// to 2 pages under Chromium's column-span: all layout.
	css, err := GetTemplate("dnd-style")
	if err != nil {
		t.Fatalf("get CSS: %v", err)
	}
	if !strings.Contains(css, "break-after: page") {
		t.Error("CSS missing 'break-after: page' (REQ-3.1)")
	}
	if !strings.Contains(css, "height: calc(297mm - 20px)") {
		t.Error("CSS missing 'height: calc(297mm - 20px)' (REQ-3.2 contract)")
	}
	if !strings.Contains(css, "max-height: calc(297mm - 20px)") {
		t.Error("CSS missing 'max-height: calc(297mm - 20px)' (REQ-3.2 contract)")
	}
	if strings.Contains(css, "min-height: 297mm") {
		t.Error("CSS still uses 'min-height: 297mm' (Bug B is back)")
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

	// REQ-2.5 / REQ-2.8 (live repro): trait and action paragraphs must be
	// classified with the right CSS class in the el-exiliado compile, not
	// dumped into .property-line. We assert a sane minimum and log the
	// counts so the report can be cited.
	traitCount := strings.Count(html, `<p class="trait">`)
	actionCount := strings.Count(html, `<p class="action">`)
	propertyLineCount := strings.Count(html, `<p class="property-line">`)
	if traitCount < 1 {
		t.Errorf("expected at least 1 <p class=\"trait\"> in el-exiliado compile (REQ-2.5), got %d", traitCount)
	}
	if actionCount < 1 {
		t.Errorf("expected at least 1 <p class=\"action\"> in el-exiliado compile (REQ-2.8), got %d", actionCount)
	}
	t.Logf("El-exiliado live repro: trait=%d action=%d property-line=%d stat-block=%d jpeg-uris=%d",
		traitCount, actionCount, propertyLineCount, statBlockCount, jpegCount)

	// Persist the live-repro counts to a sidecar file so a follow-up shell
	// pipeline (or `make` target) can grep them without re-running the
	// compiler. Useful for the verify/apply reports.
	reportPath := campaignDir + "/.live-repro-counts.txt"
	report := fmt.Sprintf("trait=%d action=%d property-line=%d stat-block=%d jpeg-uris=%d html-bytes=%d\n",
		traitCount, actionCount, propertyLineCount, statBlockCount, jpegCount, len(html))
	if err := os.WriteFile(reportPath, []byte(report), 0o644); err != nil {
		t.Logf("could not write live-repro report: %v", err)
	}

	// Also dump the full live-repro HTML to disk so the user can grep
	// (the binary itself does not expose a `compile` CLI subcommand — the
	// only compile path is the in-process `generateHTML` we just called).
	livePath := campaignDir + "/campaign.live-repro.html"
	if err := os.WriteFile(livePath, []byte(html), 0o644); err != nil {
		t.Logf("could not write live-repro HTML: %v", err)
	}
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
