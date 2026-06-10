package services

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
)

func TestCalculateScores_Healthy(t *testing.T) {
	health := &HealthReport{
		Summary: HealthSummary{
			CriticalCount: 0,
			WarningCount:  1,
			InfoCount:     2,
		},
	}
	consistency := &domain.ConsistencyReport{
		Issues: []domain.CheckResult{
			{Passed: true},
			{Passed: false, Severity: "warning"},
		},
		TotalChecks: 10,
		Passed:      8,
		Warnings:    1,
		Errors:      0,
		Criticals:   0,
	}

	scores := calculateScores(health, consistency)

	if scores.OverallHealth < 70 {
		t.Errorf("expected OverallHealth >= 70, got %d", scores.OverallHealth)
	}
	if scores.Status != "healthy" {
		t.Errorf("expected status 'healthy', got %q", scores.Status)
	}
}

func TestCalculateScores_EmptyCampaign(t *testing.T) {
	health := &HealthReport{
		Summary: HealthSummary{
			CriticalCount: 0,
			WarningCount:  0,
			InfoCount:     0,
		},
	}
	consistency := &domain.ConsistencyReport{
		Issues:      []domain.CheckResult{},
		TotalChecks: 0,
		Passed:      0,
		Warnings:    0,
		Errors:      0,
		Criticals:   0,
	}

	scores := calculateScores(health, consistency)

	if scores.OverallHealth != 0 {
		t.Errorf("expected OverallHealth 0 for empty campaign, got %d", scores.OverallHealth)
	}
	if scores.Status != "unknown" {
		t.Errorf("expected status 'unknown' for empty campaign, got %q", scores.Status)
	}
}

func TestCalculateScores_CriticalIssues(t *testing.T) {
	health := &HealthReport{
		Summary: HealthSummary{
			CriticalCount: 3,
			WarningCount:  5,
			InfoCount:     0,
		},
	}
	consistency := &domain.ConsistencyReport{
		Issues: []domain.CheckResult{
			{Passed: false, Severity: "critical"},
			{Passed: false, Severity: "critical"},
			{Passed: false, Severity: "critical"},
			{Passed: false, Severity: "error"},
			{Passed: false, Severity: "error"},
		},
		TotalChecks: 5,
		Passed:      0,
		Warnings:    0,
		Errors:      2,
		Criticals:   3,
	}

	scores := calculateScores(health, consistency)

	if scores.OverallHealth >= 70 {
		t.Errorf("expected OverallHealth < 70 with criticals, got %d", scores.OverallHealth)
	}
	if scores.Status != "critical" {
		t.Errorf("expected status 'critical', got %q", scores.Status)
	}
}

func TestClampScore(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{-10, 0},
		{0, 0},
		{50, 50},
		{100, 100},
		{150, 100},
	}

	for _, tt := range tests {
		got := clampScore(tt.input)
		if got != tt.expected {
			t.Errorf("clampScore(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
