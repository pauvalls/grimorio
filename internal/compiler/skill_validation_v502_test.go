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

// loadSkill reads the grimorio-chapters SKILL.md from disk, failing the test
// if the file is missing or unreadable. Cached via t.TempDir is not used
// because the file lives outside the temp dir.
func loadSkill(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(skillPath)
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
