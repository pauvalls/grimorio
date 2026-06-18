package consolidation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MapReferenceChecker verifies that every map reference resolves to a file.
type MapReferenceChecker struct {
	campaignDir string
}

// NewMapReferenceChecker creates a new map reference checker.
func NewMapReferenceChecker(campaignDir string) *MapReferenceChecker {
	return &MapReferenceChecker{campaignDir: campaignDir}
}

// Name returns the analyzer name.
func (m *MapReferenceChecker) Name() string {
	return "map_reference_checker"
}

// Analyze detects broken map/asset references.
func (m *MapReferenceChecker) Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error) {
	result := &AnalysisResult{Rule: "map_asset_existence", Passed: true}
	seen := make(map[string]bool)

	reLink := regexp.MustCompile(`\]\(([^)]+\.(?:svg|png|jpg|jpeg|webp))\)`)
	reBare := regexp.MustCompile(`(?i)(assets/[\w\-./]+\.(?:svg|png|jpg|jpeg|webp)|maps/[\w\-./]+\.(?:svg|png|jpg|jpeg|webp))`)

	for _, f := range files {
		refs := m.collectReferences(f.Content, reLink, reBare)
		for _, ref := range refs {
			clean := strings.TrimPrefix(ref, "/")
			if seen[clean] {
				continue
			}
			seen[clean] = true

			fullPath := filepath.Join(m.campaignDir, clean)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				result.Passed = false
				result.Severity = "error"
				result.Issues = append(result.Issues, domainIssue{
					Rule:      result.Rule,
					Severity:  "error",
					Message:   fmt.Sprintf("Missing map/asset reference: %s", clean),
					Locations: []string{f.RelPath},
					Suggestion: "Generate the missing asset or remove the reference.",
				})
			}
		}
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Detected %d missing map/asset reference(s)", len(result.Issues))
	}
	return result, nil
}

func (m *MapReferenceChecker) collectReferences(content string, reLink, reBare *regexp.Regexp) []string {
	var refs []string
	for _, m := range reLink.FindAllStringSubmatch(content, -1) {
		refs = append(refs, m[1])
	}
	for _, m := range reBare.FindAllStringSubmatch(content, -1) {
		refs = append(refs, m[1])
	}
	return uniqueStrings(refs)
}
