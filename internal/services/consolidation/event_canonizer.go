package consolidation

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// EventCanonizer ensures key events occur in a single place and moment.
type EventCanonizer struct{}

// NewEventCanonizer creates a new event canonizer.
func NewEventCanonizer() *EventCanonizer {
	return &EventCanonizer{}
}

// Name returns the analyzer name.
func (e *EventCanonizer) Name() string {
	return "event_canonizer"
}

// Analyze detects key events placed in multiple locations.
func (e *EventCanonizer) Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error) {
	re := regexp.MustCompile(`(?i)(Murder of|Theft of|Signing of|Fall of|Rise of|Battle of|Death of|Assassination of)\s+([A-Z][A-Za-z\s'-]+)`)
	events := make(map[string]map[string]string)

	for _, f := range files {
		for _, line := range strings.Split(f.Content, "\n") {
			for _, m := range re.FindAllStringSubmatch(line, -1) {
				eventKey := normalizeKey(m[2])
				eventLabel := strings.TrimSpace(m[2])
				if events[eventKey] == nil {
					events[eventKey] = make(map[string]string)
				}
				events[eventKey][f.RelPath] = eventLabel
			}
		}
	}

	result := &AnalysisResult{Rule: "event_canonical_location", Passed: true}
	for event, locs := range events {
		if len(locs) <= 1 {
			continue
		}
		result.Passed = false
		result.Severity = "error"
		var locations []string
		var label string
		for loc, l := range locs {
			locations = append(locations, loc)
			if label == "" {
				label = l
			}
		}
		result.Questions = append(result.Questions, domainQuestion{
			ID:       fmt.Sprintf("event-%s", sanitizeID(event)),
			Rule:     result.Rule,
			Question: fmt.Sprintf("Which act owns the event '%s'?", label),
			Options:  locations,
			Context: map[string]string{
				"event":     event,
				"locations": strings.Join(locations, ", "),
			},
		})
		result.Issues = append(result.Issues, DomainIssue{
			Rule:       result.Rule,
			Severity:   "error",
			Message:    fmt.Sprintf("Event '%s' appears in multiple acts/files", label),
			Locations:  uniqueStrings(locations),
			Suggestion: "Choose one canonical location for this event.",
		})
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Detected %d event with duplicate placement", len(result.Issues))
	}
	return result, nil
}
