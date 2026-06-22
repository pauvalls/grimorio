package monster

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pauvalls/grimorio/internal/monster/rules/parser"
	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

// MonsterCRDriftAnalyzer is a consolidation.Analyzer that runs the
// monster CR validator on every bestiary block in a campaign and
// surfaces CR drift as Findings.
//
// The analyzer is bestiary-scoped: it only inspects files whose
// path contains "/bestiary/" (or whose name is "bestiary.md"). Other
// files are ignored to keep the report focused on the canonical
// stat-block source.
//
// The analyzer is strictly advisory — it never returns an error
// that blocks a save. A malformed block is reported as a critical
// Finding (with the parse error embedded in the message), not as a
// non-nil error from Analyze().
type MonsterCRDriftAnalyzer struct {
	validator *MonsterValidator
}

// NewMonsterCRDriftAnalyzer creates a new analyzer.
func NewMonsterCRDriftAnalyzer() *MonsterCRDriftAnalyzer {
	return &MonsterCRDriftAnalyzer{
		validator: NewMonsterValidator(),
	}
}

// Name returns the analyzer's stable identifier.
func (a *MonsterCRDriftAnalyzer) Name() string {
	return "monster_cr_drift"
}

// Analyze reads each bestiary file, splits it into per-monster
// blocks, and runs the validator on each. Findings are emitted as
// consolidation.domainIssue records with severity mapped from the
// monster service Severity:
//
//	OK    -> info
//	Minor -> warning
//	Major -> critical
//
// Malformed blocks (parse errors) are reported as critical findings
// rather than as errors so the analyzer never blocks a save.
func (a *MonsterCRDriftAnalyzer) Analyze(ctx context.Context, files []consolidation.CampaignFile) (*consolidation.AnalysisResult, error) {
	result := &consolidation.AnalysisResult{
		Rule:      a.Name(),
		Passed:    true,
		Severity:  "info",
		Locations: []string{},
	}
	if len(files) == 0 {
		result.Message = "no bestiary files to analyze"
		return result, nil
	}

	var allIssues []consolidation.DomainIssue
	var allLocations []string

	for _, f := range files {
		if !isBestiaryFile(f) {
			continue
		}
		allLocations = append(allLocations, f.RelPath)
		blocks := splitMonsters(f.Content)
		for _, blk := range blocks {
			monsterName := extractBlockName(blk)
			location := f.RelPath
			m, perr := parser.ParseStatBlock(blk)
			if perr != nil {
				allIssues = append(allIssues, consolidation.NewIssue(
					a.Name(),
					"critical",
					fmt.Sprintf("monster %q: parse error — %s", monsterName, perr.Error()),
					[]string{location},
					"Fix the markdown so the stat block parses cleanly.",
				))
				continue
			}
			vr := a.validator.Validate(m)
			sev := mapMonsterSeverity(vr.Severity)
			if sev == "info" {
				// OK monsters are reported as info findings so
				// the audit shows coverage; omit would also be
				// acceptable per spec.
				allIssues = append(allIssues, consolidation.NewIssue(
					a.Name(),
					"info",
					fmt.Sprintf("monster %q: OK (official CR %v, calculated %v)", vr.Monster.Name, vr.OfficialCR, vr.CalculatedCR),
					[]string{location},
					"",
				))
				continue
			}
			suggestion := strings.Join(vr.Suggestions, "; ")
			allIssues = append(allIssues, consolidation.NewIssue(
				a.Name(),
				sev,
				fmt.Sprintf("monster %q CR drift: official %v, calculated %v (delta %.2f, %s)", vr.Monster.Name, vr.OfficialCR, vr.CalculatedCR, vr.Delta, vr.Severity),
				[]string{location},
				suggestion,
			))
		}
	}

	result.Issues = allIssues
	result.Locations = uniqueStrings(allLocations)
	if len(allIssues) > 0 {
		result.Passed = false
		// Overall severity is the worst of the issues.
		result.Severity = worstSeverity(allIssues)
	}
	result.Message = fmt.Sprintf("analyzed %d bestiary file(s), %d finding(s)", len(result.Locations), len(result.Issues))
	return result, nil
}

// isBestiaryFile reports whether the given CampaignFile should be
// analyzed. Bestiary files live under a "bestiary" directory or are
// named bestiary.md.
func isBestiaryFile(f consolidation.CampaignFile) bool {
	rel := strings.ToLower(f.RelPath)
	if strings.Contains(rel, "bestiary") {
		return true
	}
	return filepath.Base(rel) == "bestiary.md"
}

// extractBlockName returns the H2 heading of a markdown stat block.
func extractBlockName(blk string) string {
	for _, l := range strings.Split(blk, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "## ") && !strings.HasPrefix(t, "### ") {
			return strings.TrimSpace(strings.TrimPrefix(t, "## "))
		}
	}
	return ""
}

// mapMonsterSeverity translates a monster service Severity into the
// consolidation severity string. The mapping is:
//
//	OK    -> "info"
//	Minor -> "warning"
//	Major -> "critical"
func mapMonsterSeverity(s Severity) string {
	switch s {
	case SeverityOK:
		return "info"
	case SeverityMinor:
		return "warning"
	case SeverityMajor:
		return "critical"
	}
	return "warning"
}

// worstSeverity returns the highest severity in the issue list.
func worstSeverity(issues []consolidation.DomainIssue) string {
	worst := "info"
	for _, i := range issues {
		switch i.Severity {
		case "critical":
			return "critical"
		case "error":
			if worst != "critical" {
				worst = "error"
			}
		case "warning":
			if worst != "critical" && worst != "error" {
				worst = "warning"
			}
		}
	}
	return worst
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
