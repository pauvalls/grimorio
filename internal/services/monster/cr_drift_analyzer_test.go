package monster

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

// makeTestBestiaryFile builds a CampaignFile with a synthetic bestiary
// body. It is the unit fixture for the analyzer.
func makeTestBestiaryFile(content string) consolidation.CampaignFile {
	return consolidation.CampaignFile{
		Path:    "bestiary/bestiary.md",
		RelPath: "bestiary/bestiary.md",
		Content: content,
		ModTime: time.Now(),
	}
}

// goodGoblinBlock is a canonical WotC stat block at CR 1/4.
const goodGoblinBlock = `## Goblin

*Small humanoid, neutral evil*

**Armor Class** 15 (Leather Armor, Shield)
**Hit Points** 7 (2d6)
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

// minorDriftBlock is a goblin with HP slightly out of CR 1/4 band
// (CR 1/4 covers 36-49 HP; this one has 80 → CR 1 defensive
// → delta 0.75 → Minor drift).
const minorDriftBlock = `## Bandit

*Medium humanoid, any alignment*

**Armor Class** 15 (Studded Leather)
**Hit Points** 80 (11d8 + 11)
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

// majorDriftBlock is a 999 HP / 99 AC monster declared as CR 1/4.
const majorDriftBlock = `## Outlier

*Medium humanoid, unaligned*

**Armor Class** 99
**Hit Points** 999
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`

func TestMonsterCRDriftAnalyzer_Name(t *testing.T) {
	t.Parallel()
	a := NewMonsterCRDriftAnalyzer()
	if got := a.Name(); got != "monster_cr_drift" {
		t.Errorf("Name() = %q, want monster_cr_drift", got)
	}
}

func TestMonsterCRDriftAnalyzer_OKMinorMajor(t *testing.T) {
	t.Parallel()
	// Build a bestiary with 3 monsters: OK, Minor, Major.
	content := goodGoblinBlock + "\n" + minorDriftBlock + "\n" + majorDriftBlock + "\n"
	files := []consolidation.CampaignFile{makeTestBestiaryFile(content)}

	a := NewMonsterCRDriftAnalyzer()
	res, err := a.Analyze(context.Background(), files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if res == nil {
		t.Fatal("Analyze returned nil result")
	}
	if res.Passed {
		t.Errorf("AnalysisResult.Passed = true, want false (drift detected)")
	}
	if res.Rule != "monster_cr_drift" {
		t.Errorf("Rule = %q, want monster_cr_drift", res.Rule)
	}

	// Count by severity — the analyzer must emit at least:
	//   1 critical (Major) + 1 warning (Minor) for the 2 drifters.
	// The OK goblin may or may not be reported; if reported, it must
	// be tagged "info".
	var warnings, criticals, infos int
	for _, issue := range res.Issues {
		switch issue.Severity {
		case "warning":
			warnings++
		case "critical":
			criticals++
		case "info":
			infos++
		}
	}
	if criticals < 1 {
		t.Errorf("criticals = %d, want >= 1 (major drift must surface as critical)", criticals)
	}
	if warnings < 1 {
		t.Errorf("warnings = %d, want >= 1 (minor drift must surface as warning)", warnings)
	}
	if warnings+criticals+infos == 0 {
		t.Errorf("expected at least one finding, got none")
	}

	// Locations should reference the bestiary file.
	if len(res.Locations) == 0 {
		t.Errorf("expected at least one location, got none")
	}
	for _, l := range res.Locations {
		if !strings.Contains(l, "bestiary") {
			t.Errorf("location = %q, want contains 'bestiary'", l)
		}
	}
}

func TestMonsterCRDriftAnalyzer_MalformedMonsterReportsAsCritical(t *testing.T) {
	t.Parallel()
	// A bestiary with one good goblin and one block that cannot be
	// parsed (non-numeric AC).
	bad := `## Bad Monster

**Armor Class** not-a-number
**Hit Points** 10
**Speed** 30 ft.
**Challenge** 1/4 (50 XP)
`
	content := goodGoblinBlock + "\n" + bad
	files := []consolidation.CampaignFile{makeTestBestiaryFile(content)}

	a := NewMonsterCRDriftAnalyzer()
	res, err := a.Analyze(context.Background(), files)
	if err != nil {
		t.Fatalf("Analyze returned error for malformed block, want advisory finding: %v", err)
	}
	if res == nil {
		t.Fatal("Analyze returned nil result")
	}
	// Must not block: res.Issues must contain a critical finding
	// for the malformed monster, but the analyzer must NOT bubble
	// up the parse error.
	if res.Passed {
		t.Errorf("AnalysisResult.Passed = true, want false (malformed is drift)")
	}
	foundCritical := false
	for _, issue := range res.Issues {
		if issue.Severity == "critical" && (strings.Contains(issue.Message, "Bad Monster") || strings.Contains(issue.Message, "parse") || strings.Contains(issue.Message, "malformed")) {
			foundCritical = true
			break
		}
	}
	if !foundCritical {
		t.Errorf("expected a critical finding referencing the malformed monster, got %+v", res.Issues)
	}
}

func TestMonsterCRDriftAnalyzer_NoFilesIsClean(t *testing.T) {
	t.Parallel()
	a := NewMonsterCRDriftAnalyzer()
	res, err := a.Analyze(context.Background(), nil)
	if err != nil {
		t.Fatalf("Analyze(nil) returned error: %v", err)
	}
	if res == nil {
		t.Fatal("Analyze(nil) returned nil")
	}
	if !res.Passed {
		t.Errorf("Analyze(nil).Passed = false, want true (no files = no drift)")
	}
}

func TestMonsterCRDriftAnalyzer_IgnoresNonBestiaryFiles(t *testing.T) {
	t.Parallel()
	// A non-bestiary file with stat-block-looking content should
	// NOT be analyzed (the analyzer is bestiary-scoped).
	files := []consolidation.CampaignFile{{
		Path:    "chapters/chapter-1.md",
		RelPath: "chapters/chapter-1.md",
		Content: goodGoblinBlock,
		ModTime: time.Now(),
	}}
	a := NewMonsterCRDriftAnalyzer()
	res, err := a.Analyze(context.Background(), files)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if !res.Passed {
		t.Errorf("Analyze(non-bestiary) = not passed; want passed (analyzer is bestiary-scoped)")
	}
	if len(res.Issues) != 0 {
		t.Errorf("Issues = %d, want 0 (chapter file should be ignored)", len(res.Issues))
	}
}

func TestMonsterCRDriftAnalyzer_Performance(t *testing.T) {
	t.Parallel()
	// 10 monsters must complete in < 1s. The analyzer is
	// called inline during validation; a slow analyzer would
	// block every chapter save.
	var b strings.Builder
	b.WriteString("# Bestiario\n\n")
	for i := 0; i < 10; i++ {
		fmt.Fprintf(&b, "## Monster %d\n\n*Medium humanoid*\n\n", i)
		b.WriteString("**Armor Class** 13\n")
		b.WriteString("**Hit Points** 30\n")
		b.WriteString("**Speed** 30 ft.\n")
		b.WriteString("**Challenge** 1/2 (100 XP)\n\n")
	}
	files := []consolidation.CampaignFile{makeTestBestiaryFile(b.String())}

	a := NewMonsterCRDriftAnalyzer()
	start := time.Now()
	res, err := a.Analyze(context.Background(), files)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if res == nil {
		t.Fatal("Analyze returned nil")
	}
	if elapsed > time.Second {
		t.Errorf("Analyze took %v for 10 monsters, want < 1s", elapsed)
	}
}
