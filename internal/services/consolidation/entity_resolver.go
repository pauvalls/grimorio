package consolidation

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// EntityResolver detects similar entity name variants and unifies them.
type EntityResolver struct {
	threshold float64
}

// NewEntityResolver creates a resolver with the given similarity threshold.
// The threshold combines Levenshtein ratio and token-Jaccard; values above
// the threshold trigger a canonical merge suggestion.
func NewEntityResolver(threshold float64) *EntityResolver {
	if threshold <= 0 {
		threshold = 0.85
	}
	return &EntityResolver{threshold: threshold}
}

// Name returns the analyzer name.
func (r *EntityResolver) Name() string {
	return "entity_resolver"
}

// Analyze scans campaign files for entity names and reports collisions.
func (r *EntityResolver) Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error) {
	names := r.extractNames(files)
	if len(names) == 0 {
		return &AnalysisResult{Passed: true, Rule: "entity_name_uniqueness"}, nil
	}

	unique := r.uniqueNames(names)
	clusters := r.clusterNames(unique)

	result := &AnalysisResult{
		Rule:      "entity_name_uniqueness",
		Passed:    len(clusters) == 0,
		Severity:  "warning",
		Locations: []string{},
	}

	for _, cluster := range clusters {
		canonical := r.chooseCanonical(cluster)
		locations := r.clusterLocations(cluster, files)
		score := r.similarity(cluster[0], cluster[1])

		msg := fmt.Sprintf("Entity name collision: %s (suggested canonical: %s, similarity %.2f)",
			strings.Join(cluster, ", "), canonical, score)
		result.Locations = append(result.Locations, locations...)

		if score < 0.6 {
			// Too dissimilar to even ask — keep as issue only.
			result.Issues = append(result.Issues, domainIssue{
				Rule:       result.Rule,
				Severity:   "warning",
				Message:    msg,
				Locations:  locations,
				Suggestion: "Review these names; they may be unrelated.",
			})
		} else if score < r.threshold {
			// Ambiguous: ask user/agent.
			result.Questions = append(result.Questions, domainQuestion{
				ID:      fmt.Sprintf("entity-%s-%s", sanitizeID(cluster[0]), sanitizeID(cluster[1])),
				Rule:    result.Rule,
				Question: fmt.Sprintf("Are '%s' and '%s' the same entity?", cluster[0], cluster[1]),
				Options: []string{cluster[0], cluster[1], "keep_both"},
				Context: map[string]string{
					"canonical_suggestion": canonical,
					"similarity":           fmt.Sprintf("%.2f", score),
				},
			})
		} else {
			// High confidence: propose safe auto-fix.
			result.Fixes = append(result.Fixes, domainFix{
				Rule:      result.Rule,
				Target:    strings.Join(cluster, ", "),
				Before:    strings.Join(cluster, ", "),
				After:     canonical,
				Locations: locations,
			})
		}
	}

	if !result.Passed && result.Message == "" {
		result.Message = fmt.Sprintf("Detected %d potential entity name collision(s)", len(clusters))
	}

	r.deduplicateStrings(&result.Locations)
	return result, nil
}

// CanonicalMap returns a map from variant -> canonical for fixes above threshold.
func (r *EntityResolver) CanonicalMap(files []CampaignFile) map[string]string {
	unique := r.uniqueNames(r.extractNames(files))
	clusters := r.clusterNames(unique)
	m := make(map[string]string)
	for _, cluster := range clusters {
		score := r.similarity(cluster[0], cluster[1])
		if score >= r.threshold {
			canonical := r.chooseCanonical(cluster)
			for _, name := range cluster {
				if name != canonical {
					m[name] = canonical
				}
			}
		}
	}
	return m
}

func (r *EntityResolver) extractNames(files []CampaignFile) []string {
	var names []string
	headingRe := regexp.MustCompile(`^#{1,3}\s+(.+)$`)
	boldRe := regexp.MustCompile(`\*\*([A-Z][A-Za-z\s'-]+?)\*\*`)

	for _, f := range files {
		lines := strings.Split(f.Content, "\n")
		for _, line := range lines {
			if m := headingRe.FindStringSubmatch(line); m != nil {
				name := strings.TrimSpace(m[1])
				if n := cleanName(name); n != "" {
					names = append(names, n)
				}
			}
		}
		for _, m := range boldRe.FindAllStringSubmatch(f.Content, -1) {
			if n := cleanName(m[1]); n != "" {
				names = append(names, n)
			}
		}
	}
	return names
}

func (r *EntityResolver) uniqueNames(names []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, n := range names {
		key := strings.ToLower(n)
		if !seen[key] {
			seen[key] = true
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out
}

func (r *EntityResolver) clusterNames(names []string) [][]string {
	var clusters [][]string
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			score := r.similarity(names[i], names[j])
			if score >= 0.6 {
				clusters = append(clusters, []string{names[i], names[j]})
			}
		}
	}
	return clusters
}

func (r *EntityResolver) similarity(a, b string) float64 {
	lev := levenshteinRatio(a, b)
	jac := tokenJaccard(a, b)
	if lev >= r.threshold || jac >= 0.8 {
		if lev > jac {
			return lev
		}
		return jac
	}
	return (lev + jac) / 2
}

func (r *EntityResolver) chooseCanonical(cluster []string) string {
	// Prefer the longest name (most specific) or the one that appears first alphabetically as tie-breaker.
	canonical := cluster[0]
	for _, name := range cluster[1:] {
		if len(name) > len(canonical) || (len(name) == len(canonical) && name < canonical) {
			canonical = name
		}
	}
	return canonical
}

func (r *EntityResolver) clusterLocations(cluster []string, files []CampaignFile) []string {
	var locs []string
	for _, f := range files {
		lower := strings.ToLower(f.Content)
		for _, name := range cluster {
			if strings.Contains(lower, strings.ToLower(name)) {
				locs = append(locs, f.RelPath)
				break
			}
		}
	}
	return locs
}

func (r *EntityResolver) deduplicateStrings(slice *[]string) {
	seen := make(map[string]bool)
	var out []string
	for _, s := range *slice {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	*slice = out
}

func cleanName(name string) string {
	name = strings.TrimSpace(name)
	if len(name) < 3 {
		return ""
	}
	// Drop trailing punctuation / markdown.
	name = strings.TrimRight(name, ":#*-_")
	if len(name) < 3 {
		return ""
	}
	return name
}

func sanitizeID(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	return strings.ToLower(re.ReplaceAllString(s, "-"))
}

func levenshteinRatio(a, b string) float64 {
	aRunes := []rune(strings.ToLower(a))
	bRunes := []rune(strings.ToLower(b))
	if len(aRunes) == 0 && len(bRunes) == 0 {
		return 1.0
	}
	dist := levenshteinDistance(aRunes, bRunes)
	maxLen := len(aRunes)
	if len(bRunes) > maxLen {
		maxLen = len(bRunes)
	}
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(dist)/float64(maxLen)
}

func levenshteinDistance(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			curr[j] = min(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func tokenJaccard(a, b string) float64 {
	tokensA := tokenSet(a)
	tokensB := tokenSet(b)
	if len(tokensA) == 0 && len(tokensB) == 0 {
		return 1.0
	}
	intersection := 0
	for tok := range tokensA {
		if tokensB[tok] {
			intersection++
		}
	}
	union := len(tokensA) + len(tokensB) - intersection
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

func tokenSet(s string) map[string]bool {
	set := make(map[string]bool)
	for _, f := range strings.Fields(strings.ToLower(s)) {
		f = strings.TrimFunc(f, func(r rune) bool { return !('a' <= r && r <= 'z' || '0' <= r && r <= '9') })
		if len(f) >= 2 {
			set[f] = true
		}
	}
	return set
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
