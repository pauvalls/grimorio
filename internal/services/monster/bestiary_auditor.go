package monster

import (
	"context"
	"log/slog"

	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

// crAnalyzer is the small subset of consolidation.Analyzer that
// BestiaryAuditor needs. It exists so tests can inject a failing
// analyzer without depending on the full consolidation package.
type crAnalyzer interface {
	Analyze(ctx context.Context, files []consolidation.CampaignFile) (*consolidation.AnalysisResult, error)
}

// BestiaryAuditor is a thin wrapper around the consolidation
// MonsterCRDriftAnalyzer that runs the audit inline at save time
// (CampaignService.SaveBestiary and CampaignService.FinalizeChapter).
//
// It is strictly advisory: it NEVER returns an error to the caller.
// The audit method:
//
//  1. Catches every analyzer error and logs it at ERROR.
//
//  2. Catches every panic and logs it at DEBUG (so the caller is
//     never crashed by a malformed block).
//
//  3. Maps the analyzer's severity bucket to a slog level:
//
//     OK     (info)    -> Debug
//     Minor  (warning) -> Info
//     Major  (critical) -> Warn
//
// A nil Analyzer and a nil Logger are tolerated: the auditor
// short-circuits to a no-op (or falls back to slog.Default()).
type BestiaryAuditor struct {
	// Analyzer is the consolidation analyzer that runs the CR
	// drift check. May be nil — the auditor will be a no-op.
	Analyzer crAnalyzer

	// Logger receives the audit findings. May be nil — the
	// auditor falls back to slog.Default().
	Logger *slog.Logger
}

// NewBestiaryAuditor returns a BestiaryAuditor with the given
// analyzer and logger. Nil analyzer defaults to
// NewMonsterCRDriftAnalyzer(); nil logger defaults to
// slog.Default().
func NewBestiaryAuditor(analyzer *MonsterCRDriftAnalyzer, logger *slog.Logger) *BestiaryAuditor {
	var iface crAnalyzer
	if analyzer == nil {
		iface = NewMonsterCRDriftAnalyzer()
	} else {
		iface = analyzer
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &BestiaryAuditor{Analyzer: iface, Logger: logger}
}

// AuditBestiarySave is called from CampaignService.SaveBestiary after
// the bestiary markdown is persisted. It runs the CR drift analyzer
// on the bestiary content and logs the findings. It NEVER returns
// an error — the audit is strictly advisory.
func (b *BestiaryAuditor) AuditBestiarySave(ctx context.Context, content string, campaignID string) {
	if b == nil {
		return
	}
	b.audit(ctx, content, campaignID, "bestiary/bestiary.md")
}

// AuditChapterFinalize is called from CampaignService.FinalizeChapter
// after the chapter is persisted. It runs the CR drift analyzer on
// the chapter content (which may reference monster stat blocks) and
// logs the findings. It NEVER returns an error.
//
// The location hint embeds "bestiary" so the consolidation
// analyzer's bestiary-scope filter accepts the synthetic
// CampaignFile. The path is metadata only; what matters is the
// content.
func (b *BestiaryAuditor) AuditChapterFinalize(ctx context.Context, chapterContent string, campaignID string) {
	if b == nil {
		return
	}
	b.audit(ctx, chapterContent, campaignID, "bestiary/chapter-scan")
}

// audit is the shared body of AuditBestiarySave and
// AuditChapterFinalize. It runs the analyzer on a single
// consolidation.CampaignFile, buckets the findings, and logs the
// summary at the appropriate level. It catches all errors and
// panics so the caller is never affected.
func (b *BestiaryAuditor) audit(ctx context.Context, content string, campaignID string, location string) {
	if b.Analyzer == nil {
		return
	}
	logger := b.Logger
	if logger == nil {
		logger = slog.Default()
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Debug("cr audit recovered from panic",
				"campaign", campaignID,
				"location", location,
				"panic", r)
		}
	}()

	file := consolidation.CampaignFile{
		Path:    location,
		RelPath: location,
		Content: content,
	}

	result, err := b.Analyzer.Analyze(ctx, []consolidation.CampaignFile{file})
	if err != nil {
		logger.Error("cr audit analyzer failed",
			"campaign", campaignID,
			"location", location,
			"err", err)
		return
	}

	ok, minor, major := 0, 0, 0
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "info":
			ok++
		case "warning":
			minor++
		case "critical":
			major++
		}
	}

	// Map severity to log level. The mapping is the inverse of
	// cr_drift_analyzer.go's mapMonsterSeverity: the analyzer
	// emits the same string buckets the auditor logs at.
	//
	//   - major > 0  -> WARN  (the monster drifts more than 1 CR
	//                            band; the DM should review)
	//   - minor > 0  -> INFO  (the monster drifts 1 CR band; the
	//                            DM may want to tighten the stats)
	//   - otherwise  -> DEBUG (clean bestiary; no action needed)
	if major > 0 {
		logger.Warn("cr audit found major drift",
			"campaign", campaignID,
			"location", location,
			"major", major,
			"minor", minor,
			"ok", ok,
		)
		return
	}
	if minor > 0 {
		logger.Info("cr audit found minor drift",
			"campaign", campaignID,
			"location", location,
			"major", major,
			"minor", minor,
			"ok", ok,
		)
		return
	}
	logger.Debug("cr audit clean",
		"campaign", campaignID,
		"location", location,
		"findings", len(result.Issues),
	)
}
