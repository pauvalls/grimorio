package domain

// CampaignHealthReport holds six 0-100 scores for campaign health.
type CampaignHealthReport struct {
	OverallHealth      int    `json:"overall_health"`
	CanonCompleteness  int    `json:"canon_completeness"`
	NarrativeCoherence int    `json:"narrative_coherence"`
	FactionBalance     int    `json:"faction_balance"`
	WotCCompliance     int    `json:"wotc_compliance"`
	HookCoverage       int    `json:"hook_coverage"`
	Status             string `json:"status"`
}
