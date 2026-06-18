package consolidation

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// StatBlockResolver detects multiple stat blocks/CR values for the same boss.
type StatBlockResolver struct{}

// NewStatBlockResolver creates a new stat-block resolver.
func NewStatBlockResolver() *StatBlockResolver {
	return &StatBlockResolver{}
}

// Name returns the analyzer name.
func (s *StatBlockResolver) Name() string {
	return "stat_block_resolver"
}

// bossStat holds a single parsed stat block reference.
type bossStat struct {
	Name     string
	CR       int
	Source   string
	Location string
}

// Analyze parses stat blocks and reports CR conflicts.
func (s *StatBlockResolver) Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error) {
	var stats []bossStat
	reHeading := regexp.MustCompile(`^#{1,3}\s+([A-Z][A-Za-z\s'-]+)$`)
	reCR := regexp.MustCompile(`(?i)CR\s*:?\s*(\d+)`)

	for _, f := range files {
		lines := strings.Split(f.Content, "\n")
		var current string
		for _, line := range lines {
			if m := reHeading.FindStringSubmatch(line); m != nil {
				current = strings.TrimSpace(m[1])
				continue
			}
			if current == "" {
				continue
			}
			if m := reCR.FindStringSubmatch(line); m != nil {
				cr, _ := strconv.Atoi(m[1])
				source := "act"
				if strings.Contains(strings.ToLower(f.RelPath), "bestiary") {
					source = "bestiary"
				}
				stats = append(stats, bossStat{
					Name:     current,
					CR:       cr,
					Source:   source,
					Location: f.RelPath,
				})
				current = "" // one CR per heading block
			}
		}
	}

	byName := make(map[string][]bossStat)
	for _, st := range stats {
		key := normalizeKey(st.Name)
		byName[key] = append(byName[key], st)
	}

	result := &AnalysisResult{Rule: "stat_block_consistency", Passed: true}
	for name, entries := range byName {
		if len(entries) <= 1 {
			continue
		}
		canonicalCR := entries[0].CR
		for _, e := range entries {
			if e.Source == "bestiary" {
				canonicalCR = e.CR
				break
			}
		}

		var conflicting []string
		var locations []string
		for _, e := range entries {
			locations = append(locations, e.Location)
			if e.CR != canonicalCR {
				conflicting = append(conflicting, fmt.Sprintf("%s CR %d (%s)", e.Name, e.CR, e.Source))
			}
		}

		if len(conflicting) > 0 {
			result.Passed = false
			result.Severity = "error"
			result.Issues = append(result.Issues, domainIssue{
				Rule:      result.Rule,
				Severity:  "error",
				Message:   fmt.Sprintf("Boss '%s' has conflicting CR values; bestiary CR %d is canonical. Conflicts: %s", name, canonicalCR, strings.Join(conflicting, "; ")),
				Locations: uniqueStrings(locations),
				Suggestion: "Update non-bestiary stat blocks to match the bestiary CR.",
			})
		}
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Detected %d stat-block conflict(s)", len(result.Issues))
	}
	return result, nil
}
