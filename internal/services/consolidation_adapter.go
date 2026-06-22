package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/services/consolidation"
	"github.com/pauvalls/grimorio/internal/services/monster"
)

// ConsolidationAdapter bridges the internal/services/consolidation engine
// with the legacy ValidationEngine and CampaignHealthCheck shapes.
//
// It owns a CampaignConsolidator and exposes:
//   - Files: load markdown files for a campaign (or nil if the directory is
//     missing, so cross-file checks skip cleanly for early-stage campaigns).
//   - RunAnalyzer: run a single analyzer and translate its AnalysisResult
//     into []CheckResult for ValidationEngine.
//   - RunAnalyzerHealthFindings: same but translated to []HealthFinding
//     for CampaignHealthCheck.
//   - Consolidate/Detect/ResolveAmbiguity/RegenerateIndex/VerifyFreshness:
//     pass-through to the underlying CampaignConsolidator, used by the
//     MCP consolidation tools.
type ConsolidationAdapter struct {
	consolidator *consolidation.CampaignConsolidator
	baseDir      string
}

// NewConsolidationAdapter creates an adapter bound to the given base
// directory (the root under which campaign subdirectories live).
func NewConsolidationAdapter(baseDir string) *ConsolidationAdapter {
	return &ConsolidationAdapter{
		consolidator: consolidation.NewCampaignConsolidator(baseDir),
		baseDir:      baseDir,
	}
}

// NewConsolidationAdapterWithReader allows tests to inject a CampaignReader.
func NewConsolidationAdapterWithReader(baseDir string, reader consolidation.CampaignReader) *ConsolidationAdapter {
	return &ConsolidationAdapter{
		consolidator: consolidation.NewCampaignConsolidatorWithReader(baseDir, reader),
		baseDir:      baseDir,
	}
}

// Files returns the campaign markdown files. Returns (nil, nil) when the
// campaign directory does not exist (or any other reader error) so callers
// can skip cross-file checks without surfacing "campaign not generated yet"
// as a validation failure.
func (a *ConsolidationAdapter) Files(ctx context.Context, campaignID string) ([]consolidation.CampaignFile, error) {
	files, err := a.consolidator.ReadFiles(ctx, campaignID)
	if err != nil {
		// Treat reader errors as "no files" — early-stage campaigns
		// (e.g. before create_campaign runs) shouldn't fail validation.
		return nil, nil
	}
	return files, nil
}

// RunAnalyzer loads the campaign files and runs a single analyzer, returning
// its AnalysisResult. Returns (nil, nil) when the campaign has no files.
func (a *ConsolidationAdapter) RunAnalyzer(ctx context.Context, campaignID string, analyzer consolidation.Analyzer) (*consolidation.AnalysisResult, error) {
	files, err := a.Files(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if files == nil {
		return nil, nil
	}
	return analyzer.Analyze(ctx, files)
}

// RunAllAnalyzers runs every analyzer in the campaign and returns the
// per-analyzer results keyed by name. Returns an empty map when the
// campaign has no markdown files yet.
func (a *ConsolidationAdapter) RunAllAnalyzers(ctx context.Context, campaignID string) (map[string]*consolidation.AnalysisResult, error) {
	files, err := a.Files(ctx, campaignID)
	if err != nil {
		return nil, err
	}
	if files == nil {
		return map[string]*consolidation.AnalysisResult{}, nil
	}

	analyzers := []consolidation.Analyzer{
		consolidation.NewEntityResolver(0.85),
		consolidation.NewLoreCoherence(),
		consolidation.NewStatBlockResolver(),
		consolidation.NewEventCanonizer(),
		consolidation.NewFileConsolidator(),
		consolidation.NewMapReferenceChecker(filepath.Join(a.baseDir, campaignID)),
		monster.NewMonsterCRDriftAnalyzer(),
	}
	results := make(map[string]*consolidation.AnalysisResult, len(analyzers))
	for _, an := range analyzers {
		res, err := an.Analyze(ctx, files)
		if err != nil {
			results[an.Name()] = &consolidation.AnalysisResult{
				Rule:      an.Name(),
				Passed:    false,
				Severity:  "error",
				Message:   fmt.Sprintf("analyzer %s failed: %v", an.Name(), err),
				Locations: []string{},
			}
			continue
		}
		results[an.Name()] = res
	}
	return results, nil
}

// AnalysisToCheckResults translates an AnalysisResult into the legacy
// CheckResult shape. Each issue becomes a CheckResult with Passed=false
// and the issue's severity. The overall check is also emitted as a
// "passed" CheckResult when the analyzer reports no problems.
//
// If result is nil (e.g. no files), it returns a single passing CheckResult
// so the caller still emits at least one check record for the rule.
func AnalysisToCheckResults(rule string, result *consolidation.AnalysisResult) []domain.CheckResult {
	if result == nil {
		return []domain.CheckResult{{
			Rule:     rule,
			Passed:   true,
			Severity: "info",
			Message:  "no markdown files to analyze",
		}}
	}
	out := make([]domain.CheckResult, 0, len(result.Issues)+1)
	out = append(out, domain.CheckResult{
		Rule:     rule,
		Passed:   result.Passed,
		Severity: normalizeSeverity(result.Severity),
		Message:  result.Message,
		Location: joinLocations(result.Locations),
	})
	for _, issue := range result.Issues {
		out = append(out, domain.CheckResult{
			Rule:     issue.Rule,
			Passed:   false,
			Severity: normalizeSeverity(issue.Severity),
			Message:  issue.Message,
			Location: joinLocations(issue.Locations),
		})
	}
	return out
}

// AnalysisToHealthFindings translates an AnalysisResult into HealthFinding
// shape for the CampaignHealthCheck.
func AnalysisToHealthFindings(rule string, result *consolidation.AnalysisResult) []HealthFinding {
	if result == nil {
		return nil
	}
	out := make([]HealthFinding, 0, len(result.Issues))
	for _, issue := range result.Issues {
		out = append(out, HealthFinding{
			Rule:           rule,
			Severity:       severityToDomain(issue.Severity),
			Message:        issue.Message,
			Recommendation: issue.Suggestion,
			Metadata: map[string]any{
				"locations": issue.Locations,
			},
		})
	}
	return out
}

// Consolidate delegates to the underlying consolidator.
func (a *ConsolidationAdapter) Consolidate(ctx context.Context, campaignID string, opts domain.ConsolidationOptions) (*domain.ConsolidationReport, error) {
	return a.consolidator.Consolidate(ctx, campaignID, opts)
}

// Detect delegates to the underlying consolidator.
func (a *ConsolidationAdapter) Detect(ctx context.Context, campaignID string) (*domain.ConsolidationReport, error) {
	return a.consolidator.Detect(ctx, campaignID)
}

// ResolveAmbiguity delegates to the underlying consolidator.
func (a *ConsolidationAdapter) ResolveAmbiguity(ctx context.Context, campaignID, qid, decision string) error {
	return a.consolidator.ResolveAmbiguity(ctx, campaignID, qid, decision)
}

// RegenerateIndex delegates to the underlying consolidator.
func (a *ConsolidationAdapter) RegenerateIndex(ctx context.Context, campaignID string) error {
	return a.consolidator.RegenerateIndex(ctx, campaignID)
}

// VerifyFreshness delegates to the underlying consolidator.
func (a *ConsolidationAdapter) VerifyFreshness(ctx context.Context, campaignID string) (*domain.FreshnessReport, error) {
	return a.consolidator.VerifyFreshness(ctx, campaignID)
}

// Consolidator exposes the underlying CampaignConsolidator (used by MCP handlers).
func (a *ConsolidationAdapter) Consolidator() *consolidation.CampaignConsolidator {
	return a.consolidator
}

// normalizeSeverity coerces analyzer severities into the three values
// ValidationEngine expects: critical, error, warning. Empty becomes warning.
func normalizeSeverity(s string) string {
	switch s {
	case "critical":
		return "critical"
	case "error":
		return "error"
	case "warning":
		return "warning"
	case "info":
		return "warning"
	}
	return "warning"
}

// severityToDomain maps a string severity into the domain.Severity enum.
func severityToDomain(s string) domain.Severity {
	switch s {
	case "critical":
		return domain.SeverityCritical
	case "error":
		// There is no "error" severity in the health enum; treat as critical.
		return domain.SeverityCritical
	case "warning":
		return domain.SeverityWarning
	}
	return domain.SeverityWarning
}

// joinLocations collapses a locations slice into a comma-separated string
// for the CheckResult.Location field.
func joinLocations(locs []string) string {
	if len(locs) == 0 {
		return ""
	}
	out := ""
	for i, l := range locs {
		if i > 0 {
			out += ", "
		}
		out += l
	}
	return out
}
