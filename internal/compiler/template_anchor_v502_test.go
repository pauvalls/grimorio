package compiler

import (
	"regexp"
	"strings"
	"testing"
)

// appendicesLinkRe matches markdown links to appendices.md anchors,
// e.g. `[Appendix A: Magic Items](appendices.md#appendix-a)`.
var appendicesLinkRe = regexp.MustCompile(`\[[^\]]*\]\(appendices\.md#([a-z0-9-]+)\)`)

// headingAnchorRe matches explicit HTML-style anchor IDs inserted before
// markdown headings, e.g. `<a id="appendix-a"></a>\n## Appendix A: ...`.
var headingAnchorRe = regexp.MustCompile(`<a\s+id="([a-z0-9-]+)"`)

// githubHeadingAnchorRe derives the GitHub-style auto-anchor for a markdown
// heading line, lowercasing, replacing spaces with hyphens, and stripping
// punctuation.
var githubHeadingAnchorRe = regexp.MustCompile(`[^a-z0-9-]+`)

// extractAnchors returns a set of anchor IDs that the appendices template
// exposes. It merges two sources: explicit `<a id="...">` anchors and
// GitHub-style auto-anchors derived from `##` heading text. Used by WU5
// to assert that chapter cross-references resolve.
func extractAnchors(s string) map[string]bool {
	anchors := make(map[string]bool)
	// Explicit anchors
	for _, m := range headingAnchorRe.FindAllStringSubmatch(s, -1) {
		anchors[m[1]] = true
	}
	// GitHub-style auto-anchors from headings (H1-H6)
	for _, line := range strings.Split(s, "\n") {
		trimmed := strings.TrimSpace(line)
		// Match "# ", "## ", ..., "###### " but NOT "#word" (no space).
		if !strings.HasPrefix(trimmed, "#") {
			continue
		}
		// Strip leading #'s and one space
		i := 0
		for i < len(trimmed) && trimmed[i] == '#' {
			i++
		}
		if i >= len(trimmed) || trimmed[i] != ' ' {
			continue
		}
		text := strings.TrimSpace(trimmed[i+1:])
		text = strings.ToLower(text)
		text = githubHeadingAnchorRe.ReplaceAllString(text, "-")
		text = strings.Trim(text, "-")
		if text != "" {
			anchors[text] = true
		}
	}
	return anchors
}

// TestAppendicesAnchorsMatchChapterRefs asserts that every
// `[...](appendices.md#appendix-X)` cross-reference in the chapter template
// resolves to a heading (or explicit anchor) in the appendices template.
func TestAppendicesAnchorsMatchChapterRefs(t *testing.T) {
	anchors := extractAnchors(appendicesTemplate)
	missing := []string{}
	for _, m := range appendicesLinkRe.FindAllStringSubmatch(chapterTemplate, -1) {
		anchor := m[1]
		if !anchors[anchor] {
			missing = append(missing, anchor)
		}
	}
	if len(missing) > 0 {
		t.Errorf("chapter template cross-references anchors not present in appendices: %v\nappendices anchors: %v", missing, keysSorted(anchors))
	}
}

func keysSorted(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// simple insertion sort
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
