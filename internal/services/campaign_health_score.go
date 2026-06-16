package services

import (
	"context"
	"fmt"

	"github.com/pauvalls/grimorio/internal/domain"
)

// CampaignHealthScore aggregates health checks and consistency reports into 0-100 scores.
type CampaignHealthScore struct {
	healthCheck       *CampaignHealthCheck
	validationEngine  *ValidationEngine
}

// NewCampaignHealthScore creates a new health score aggregator.
func NewCampaignHealthScore(healthCheck *CampaignHealthCheck, validationEngine *ValidationEngine) *CampaignHealthScore {
	return &CampaignHealthScore{
		healthCheck:      healthCheck,
		validationEngine: validationEngine,
	}
}

// Compute runs health checks and consistency validation, then returns 0-100 scores.
func (s *CampaignHealthScore) Compute(ctx context.Context, campaignID string) (*domain.CampaignHealthReport, error) {
	if s.healthCheck == nil || s.validationEngine == nil {
		return nil, fmt.Errorf("health score dependencies not initialized")
	}

	health, err := s.healthCheck.RunHealthCheck(ctx, campaignID)
	if err != nil {
		return nil, fmt.Errorf("health check failed: %w", err)
	}

	consistency, err := s.validationEngine.CheckConsistency(ctx, campaignID, domain.ConsistencyScopeFull)
	if err != nil {
		return nil, fmt.Errorf("consistency check failed: %w", err)
	}

	return calculateScores(health, consistency), nil
}

// calculateScores derives 0-100 scores from raw reports.
func calculateScores(health *HealthReport, consistency *domain.ConsistencyReport) *domain.CampaignHealthReport {
	// Empty campaign guard
	if health.Summary.TotalFindings == 0 && consistency.TotalChecks == 0 {
		return &domain.CampaignHealthReport{
			OverallHealth:      0,
			CanonCompleteness:  0,
			NarrativeCoherence: 0,
			FactionBalance:     0,
			WotCCompliance:     0,
			HookCoverage:       0,
			Status:             "unknown",
		}
	}

	// Consistency score: 100 - (criticals*25 + errors*10 + warnings*2)
	consistencyScore := 100 - (consistency.Criticals*25 + consistency.Errors*10 + consistency.Warnings*2)
	consistencyScore = clampScore(consistencyScore)

	// Narrative coherence from health findings
	coherenceScore := 100 - (health.Summary.CriticalCount*25 + health.Summary.WarningCount*10 + health.Summary.InfoCount*2)
	coherenceScore = clampScore(coherenceScore)

	// Canon completeness: percentage of passed checks vs total
	canonScore := 0
	if consistency.TotalChecks > 0 {
		canonScore = (consistency.Passed * 100) / consistency.TotalChecks
	}

	// WotC compliance: penalize failures
	wotcFailures := countWotCFailures(consistency)
	wotcScore := 100 - (wotcFailures * 5)
	wotcScore = clampScore(wotcScore)

	// Faction balance and hook coverage default to coherence when no data
	factionScore := coherenceScore
	hookScore := coherenceScore

	// Weighted overall: consistency 30%, coherence 25%, canon 20%, wotc 15%, faction 5%, hook 5%
	overall := (consistencyScore*30 + coherenceScore*25 + canonScore*20 + wotcScore*15 + factionScore*5 + hookScore*5) / 100

	status := "healthy"
	switch {
	case overall < 40:
		status = "critical"
	case overall < 60:
		status = "poor"
	case overall < 70:
		status = "fair"
	case overall < 85:
		status = "good"
	}

	return &domain.CampaignHealthReport{
		OverallHealth:      overall,
		CanonCompleteness:  canonScore,
		NarrativeCoherence: coherenceScore,
		FactionBalance:     factionScore,
		WotCCompliance:     wotcScore,
		HookCoverage:       hookScore,
		Status:             status,
	}
}

// countWotCFailures counts consistency issues related to WotC format rules.
func countWotCFailures(r *domain.ConsistencyReport) int {
	count := 0
	for _, issue := range r.Issues {
		if !issue.Passed && isWotCRule(issue.Rule) {
			count++
		}
	}
	return count
}

func isWotCRule(rule string) bool {
	switch rule {
	case "wotc_developments", "wotc_multiple_solutions", "wotc_character_hooks",
		"wotc_boxed_text", "wotc_npc_word_count":
		return true
	}
	return false
}

// clampScore restricts a score to 0-100.
func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
