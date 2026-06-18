package domain

import "time"

// ConsolidationReport is the output of the campaign consistency consolidation engine.
type ConsolidationReport struct {
	CampaignID      string                 `json:"campaign_id"`
	ChecksRun       []ConsolidationCheck   `json:"checks_run"`
	FixesApplied    []ConsolidationFix     `json:"fixes_applied"`
	RemainingIssues []ConsolidationIssue   `json:"remaining_issues"`
	NeedsHuman      []AmbiguityQuestion    `json:"needs_human"`
	Freshness       *FreshnessReport       `json:"freshness,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ConsolidationCheck records the result of a single consolidation rule.
type ConsolidationCheck struct {
	Rule      string   `json:"rule"`
	Passed    bool     `json:"passed"`
	Severity  string   `json:"severity"`
	Message   string   `json:"message"`
	Locations []string `json:"locations,omitempty"`
}

// ConsolidationFix records an automatic fix applied by the engine.
type ConsolidationFix struct {
	Rule      string   `json:"rule"`
	Target    string   `json:"target"`
	Before    string   `json:"before"`
	After     string   `json:"after"`
	Locations []string `json:"locations,omitempty"`
}

// ConsolidationIssue records a problem that could not be fixed automatically.
type ConsolidationIssue struct {
	Rule        string   `json:"rule"`
	Severity    string   `json:"severity"`
	Message     string   `json:"message"`
	Locations   []string `json:"locations,omitempty"`
	Suggestion  string   `json:"suggestion,omitempty"`
}

// AmbiguityQuestion records a decision that requires human or agent input.
type AmbiguityQuestion struct {
	ID       string            `json:"id"`
	Rule     string            `json:"rule"`
	Question string            `json:"question"`
	Options  []string          `json:"options"`
	Context  map[string]string `json:"context,omitempty"`
}

// FreshnessReport describes whether generated artifacts are stale.
type FreshnessReport struct {
	CampaignID      string    `json:"campaign_id"`
	SourcesNewest   time.Time `json:"sources_newest"`
	CampaignMDTime  time.Time `json:"campaign_md_time"`
	IndexMDTime     time.Time `json:"index_md_time"`
	CampaignMDStale bool      `json:"campaign_md_stale"`
	IndexStale      bool      `json:"index_stale"`
	Message         string    `json:"message"`
}

// ConsolidationOptions controls consolidation behavior.
type ConsolidationOptions struct {
	AutoFix              bool    `json:"auto_fix"`
	EntitySimilarityThreshold float64 `json:"entity_similarity_threshold"`
	BackupDir            string  `json:"backup_dir,omitempty"`
}
