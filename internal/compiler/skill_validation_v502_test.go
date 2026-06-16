package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skillPath is the path to the grimorio-chapters SKILL.md, relative to the
// internal/compiler package directory (where these tests run).
const skillPath = "../../skills/grimorio-chapters/SKILL.md"

// areasSkillPath is the legacy grimorio-areas/SKILL.md (deprecated).
const areasSkillPath = "../../skills/grimorio-areas/SKILL.md"

// loadSkill reads the grimorio-chapters SKILL.md from disk, failing the test
// if the file is missing or unreadable. Cached via t.TempDir is not used
// because the file lives outside the temp dir.
func loadSkill(t *testing.T) string {
	t.Helper()
	return loadSkillFile(t, skillPath)
}

// loadSkillFile reads any SKILL.md from disk given a package-relative path.
func loadSkillFile(t *testing.T, relPath string) string {
	t.Helper()
	abs, err := filepath.Abs(relPath)
	if err != nil {
		t.Fatalf("abs skill path: %v", err)
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		t.Fatalf("read skill file %s: %v", abs, err)
	}
	return string(data)
}

// TestSkillLineCount asserts that the rewritten skill file is at least 400
// lines long — the spec's structural minimum for the WotC rule set.
func TestSkillLineCount(t *testing.T) {
	content := loadSkill(t)
	lines := strings.Count(content, "\n") + 1
	if lines < 400 {
		t.Errorf("SKILL.md has %d lines, want >= 400", lines)
	}
}

// TestSkillSevenRuleHeaders asserts that the seven new WotC rule headers are
// all present in the rewritten skill file. The headers (case-insensitive):
// DM Sidebar, Roleplay, Random Encounter, XP (per-encounter), Trap, Faction,
// What's Next.
func TestSkillSevenRuleHeaders(t *testing.T) {
	content := strings.ToLower(loadSkill(t))
	headers := []string{
		"dm sidebar",
		"roleplay",
		"random encounter",
		"xp",
		"trap",
		"faction",
		"what's next",
	}
	missing := []string{}
	for _, h := range headers {
		if !strings.Contains(content, h) {
			missing = append(missing, h)
		}
	}
	if len(missing) > 0 {
		t.Errorf("SKILL.md missing rule headers: %v", missing)
	}
}

// TestSkillSevenFencedExamples asserts that there are at least 7 fenced
// markdown code blocks within the rewritten skill file (one per WotC rule).
func TestSkillSevenFencedExamples(t *testing.T) {
	content := loadSkill(t)
	fences := strings.Count(content, "```")
	// Each fenced block has an opening ``` and a closing ```, so count/2.
	if fences < 14 {
		t.Errorf("SKILL.md has %d backticks, want >= 14 (>= 7 fenced examples)", fences)
	}
}

// TestAreasSkillDeprecationBanner asserts that the legacy grimorio-areas
// skill file has a deprecation banner in the first 10 lines that points
// the reader to the v2 grimorio-chapters skill.
func TestAreasSkillDeprecationBanner(t *testing.T) {
	content := loadSkillFile(t, areasSkillPath)
	lines := strings.SplitN(content, "\n", 11)
	if len(lines) < 10 {
		t.Fatalf("areas skill has fewer than 10 lines (%d); cannot validate banner", len(lines))
	}
	header := strings.ToLower(strings.Join(lines[:10], "\n"))
	if !strings.Contains(header, "deprecated") {
		t.Errorf("first 10 lines of areas SKILL.md must contain 'deprecated'; got:\n%s", lines[0])
	}
	if !strings.Contains(header, "skills/grimorio-chapters/skill.md") {
		t.Errorf("first 10 lines of areas SKILL.md must point to skills/grimorio-chapters/SKILL.md")
	}
	// File must still parse as valid markdown — at least 1 markdown heading
	if !strings.Contains(content, "# ") {
		t.Errorf("areas SKILL.md must contain at least one markdown heading")
	}
}
