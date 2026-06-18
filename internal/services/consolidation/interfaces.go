package consolidation

import (
	"context"
	"time"
)

// CampaignFile represents a single markdown file loaded from a campaign.
type CampaignFile struct {
	Path    string
	RelPath string
	Content string
	ModTime time.Time
}

// CampaignReader loads campaign files for analysis.
type CampaignReader interface {
	ReadCampaign(ctx context.Context, campaignID string) ([]CampaignFile, error)
}

// FixApplicator applies a safe fix to campaign files.
type FixApplicator interface {
	ApplyFix(ctx context.Context, campaignID string, fix domainFix, files []CampaignFile) ([]CampaignFile, error)
}

// Analyzer is implemented by every consolidation analyzer.
type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, files []CampaignFile) (*AnalysisResult, error)
}

// AnalysisResult collects findings from a single analyzer.
type AnalysisResult struct {
	Passed    bool
	Rule      string
	Severity  string
	Message   string
	Locations []string
	Fixes     []domainFix
	Issues    []domainIssue
	Questions []domainQuestion
}

// domainFix is the internal fix representation before conversion to domain types.
type domainFix struct {
	Rule      string
	Target    string
	Before    string
	After     string
	Locations []string
}

// domainIssue is the internal issue representation.
type domainIssue struct {
	Rule       string
	Severity   string
	Message    string
	Locations  []string
	Suggestion string
}

// domainQuestion is the internal ambiguity representation.
type domainQuestion struct {
	ID      string
	Rule    string
	Question string
	Options []string
	Context map[string]string
}
