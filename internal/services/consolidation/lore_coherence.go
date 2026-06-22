package consolidation

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// LoreCoherence extracts dates, treaties, events, and primordial entities and
// detects contradictions across campaign files.
type LoreCoherence struct{}

// NewLoreCoherence creates a new lore coherence analyzer.
func NewLoreCoherence() *LoreCoherence {
	return &LoreCoherence{}
}

// Name returns the analyzer name.
func (l *LoreCoherence) Name() string {
	return "lore_coherence"
}

// Analyze detects timeline contradictions.
func (l *LoreCoherence) Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error) {
	treaties := make(map[string]map[int][]string)        // normalized name -> year -> relpaths
	treatyDisplay := make(map[string]string)              // normalized name -> first display form
	events := make(map[string][]string)
	eventDisplay := make(map[string]string)
	primordial := make(map[string][]string)
	primordialDisplay := make(map[string]string)

	treatyRe := regexp.MustCompile(`(?i)Treaty of\s+([A-Z][A-Za-z'-]+)`)
	dateRe := regexp.MustCompile(`\b(\d{3,4})\b`)
	primordialRe := regexp.MustCompile(`(?i)(?:primordial|ancient|old one|elder|first)\s+([A-Z][A-Za-z'-]+)`)
	eventRe := regexp.MustCompile(`(?i)^#{1,3}\s+(Murder of|Theft of|Signing of|Fall of|Rise of|Battle of|Death of)\s+([A-Z][A-Za-z\s'-]+)`)

	for _, f := range files {
		lines := strings.Split(f.Content, "\n")
		for _, line := range lines {
			if match := treatyRe.FindStringSubmatch(line); match != nil {
				treaty := normalizeKey(match[1])
				if _, ok := treatyDisplay[treaty]; !ok {
					treatyDisplay[treaty] = match[1]
				}
				year := l.extractYear(line, dateRe)
				if year != 0 {
					if treaties[treaty] == nil {
						treaties[treaty] = make(map[int][]string)
					}
					treaties[treaty][year] = append(treaties[treaty][year], f.RelPath)
				}
			}
			if match := eventRe.FindStringSubmatch(line); match != nil {
				event := normalizeKey(match[2])
				if _, ok := eventDisplay[event]; !ok {
					eventDisplay[event] = strings.TrimSpace(match[2])
				}
				events[event] = append(events[event], f.RelPath)
			}
			if matches := primordialRe.FindAllStringSubmatch(line, -1); matches != nil {
				for _, m := range matches {
					entity := normalizeKey(m[1])
					if _, ok := primordialDisplay[entity]; !ok {
						primordialDisplay[entity] = m[1]
					}
					primordial[entity] = append(primordial[entity], f.RelPath)
				}
			}
		}
	}

	result := &AnalysisResult{Rule: "timeline_consistency", Passed: true}

	for treaty, years := range treaties {
		if len(years) <= 1 {
			continue
		}
		var ys []int
		for y := range years {
			ys = append(ys, y)
		}
		if len(ys) > 1 {
			result.Passed = false
			result.Severity = "error"
			result.Issues = append(result.Issues, DomainIssue{
				Rule:      result.Rule,
				Severity:  "error",
				Message:   fmt.Sprintf("Treaty of %s has contradictory dates: %v", treatyDisplay[treaty], ys),
				Locations: flattenYearLocations(years),
				Suggestion: "Choose a single canonical date; lore files take precedence over acts.",
			})
		}
	}

	for event, locs := range events {
		if len(locs) > 1 {
			result.Passed = false
			if result.Severity == "" {
				result.Severity = "warning"
			}
			result.Issues = append(result.Issues, DomainIssue{
				Rule:      "event_placement",
				Severity:  "warning",
				Message:   fmt.Sprintf("Event '%s' is described in multiple locations", eventDisplay[event]),
				Locations: uniqueStrings(locs),
				Suggestion: "Ensure key events occur in a single place and moment.",
			})
		}
	}

	for entity, locs := range primordial {
		if len(locs) > 1 {
			result.Passed = false
			if result.Severity == "" {
				result.Severity = "warning"
			}
			result.Issues = append(result.Issues, DomainIssue{
				Rule:      "primordial_entity",
				Severity:  "warning",
				Message:   fmt.Sprintf("Primordial entity '%s' referenced in multiple contexts", primordialDisplay[entity]),
				Locations: uniqueStrings(locs),
			})
		}
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Detected %d lore coherence issue(s)", len(result.Issues))
	}
	return result, nil
}

func (l *LoreCoherence) extractYear(line string, re *regexp.Regexp) int {
	matches := re.FindAllStringSubmatch(line, -1)
	for _, m := range matches {
		if len(m) > 1 {
			if y, err := strconv.Atoi(m[1]); err == nil && y >= 100 && y <= 9999 {
				return y
			}
		}
	}
	return 0
}

func normalizeKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.TrimRight(s, ":#*-_")
	return s
}

func flattenYearLocations(years map[int][]string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, locs := range years {
		for _, loc := range locs {
			if !seen[loc] {
				seen[loc] = true
				out = append(out, loc)
			}
		}
	}
	return out
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
