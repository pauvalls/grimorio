package services

import (
	"context"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestCampaignHealthCheck_RunHealthCheck(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock repositories
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	npcRepo := repository.NewMemoryNPCRepository()

	svc := NewCampaignHealthCheck(canonRepo, stateRepo, questRepo, factionRepo, npcRepo, tmpDir)

	t.Run("empty campaign returns excellent health", func(t *testing.T) {
		campaignID := "test-empty"

		// Save empty canon
		canon := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Entities:      []domain.CanonEntity{},
		}
		if err := canonRepo.Save(campaignID, canon); err != nil {
			t.Fatalf("failed to save canon: %v", err)
		}

		// Save empty state
		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 1,
			SessionLog:     []domain.SessionRecord{},
		}
		if err := stateRepo.Save(campaignID, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		report, err := svc.RunHealthCheck(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("RunHealthCheck failed: %v", err)
		}

		if report.OverallHealth != domain.HealthStatusExcellent {
			t.Errorf("expected excellent health, got %v", report.OverallHealth)
		}
		if len(report.Findings) != 0 {
			t.Errorf("expected no findings, got %d", len(report.Findings))
		}
	})

	t.Run("detects stale quest", func(t *testing.T) {
		campaignID := "test-stale"

		// Save canon
		canon := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Entities:      []domain.CanonEntity{},
		}
		if err := canonRepo.Save(campaignID, canon); err != nil {
			t.Fatalf("failed to save canon: %v", err)
		}

		// Save state with stale quest (active for >10 sessions)
		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 15,
			ActiveQuests: []domain.QuestState{
				{
					ID:     "quest-1",
					Name:   "Stale Quest",
					Status: "active",
				},
			},
			SessionLog: []domain.SessionRecord{
				{SessionNum: 1, Date: time.Now().AddDate(0, 0, -14)},
			},
		}
		if err := stateRepo.Save(campaignID, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		report, err := svc.RunHealthCheck(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("RunHealthCheck failed: %v", err)
		}

		if report.OverallHealth != domain.HealthStatusFair {
			t.Errorf("expected fair health, got %v", report.OverallHealth)
		}

		found := false
		for _, f := range report.Findings {
			if f.Rule == "stale_quest" {
				found = true
				if f.Severity != domain.SeverityWarning {
					t.Errorf("expected warning severity, got %v", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected stale_quest finding")
		}
	})
}

func TestCampaignHealthCheck_checkFactionContradictions(t *testing.T) {
	tmpDir := t.TempDir()

	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	npcRepo := repository.NewMemoryNPCRepository()

	svc := NewCampaignHealthCheck(canonRepo, stateRepo, questRepo, factionRepo, npcRepo, tmpDir)

	t.Run("detects faction contradiction", func(t *testing.T) {
		campaignID := "test-faction"

		// Save canon with friendly faction
		canon := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Entities: []domain.CanonEntity{
				{
					ID:          "faction-1",
					Type:        domain.EntityTypeFaction,
					Name:        "Friendly Faction",
					CanonState:  domain.EntityStateAlive,
					Properties:  map[string]any{"attitude": "ally"},
				},
			},
		}
		if err := canonRepo.Save(campaignID, canon); err != nil {
			t.Fatalf("failed to save canon: %v", err)
		}

		// Save state
		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 1,
		}
		if err := stateRepo.Save(campaignID, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		// Save hostile reputation
		matrix := &domain.FactionReputationMatrix{
			CampaignID: campaignID,
			Entries: []domain.ReputationEntry{
				{
					FactionID: "faction-1",
					PartyID:   "party-1",
					Score:     -90, // Hostile
					Status:    "hostile",
				},
			},
		}
		if err := factionRepo.Save(campaignID, matrix); err != nil {
			t.Fatalf("failed to save faction reputation: %v", err)
		}

		report, err := svc.RunHealthCheck(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("RunHealthCheck failed: %v", err)
		}

		found := false
		for _, f := range report.Findings {
			if f.Rule == "faction_contradiction" {
				found = true
				if f.Severity != domain.SeverityCritical {
					t.Errorf("expected critical severity, got %v", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected faction_contradiction finding")
		}
	})
}

func TestCampaignHealthCheck_checkDeadNPCMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	npcRepo := repository.NewMemoryNPCRepository()

	svc := NewCampaignHealthCheck(canonRepo, stateRepo, questRepo, factionRepo, npcRepo, tmpDir)

	t.Run("detects dead NPC mismatch", func(t *testing.T) {
		campaignID := "test-dead-npc"

		// Save canon with alive NPC
		canon := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Entities: []domain.CanonEntity{
				{
					ID:         "npc-1",
					Type:       domain.EntityTypeNPC,
					Name:       "Victim NPC",
					CanonState: domain.EntityStateAlive,
				},
			},
		}
		if err := canonRepo.Save(campaignID, canon); err != nil {
			t.Fatalf("failed to save canon: %v", err)
		}

		// Save state with dead NPC
		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 5,
			DeadNPCs: []domain.NPCDeathRecord{
				{
					NPCID:   "npc-1",
					Name:    "Victim NPC",
					Session: 3,
				},
			},
		}
		if err := stateRepo.Save(campaignID, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		report, err := svc.RunHealthCheck(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("RunHealthCheck failed: %v", err)
		}

		found := false
		for _, f := range report.Findings {
			if f.Rule == "dead_npc_mismatch" {
				found = true
				if f.Severity != domain.SeverityCritical {
					t.Errorf("expected critical severity, got %v", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected dead_npc_mismatch finding")
		}
	})
}

func TestCampaignHealthCheck_checkMcGuffinDrift(t *testing.T) {
	tmpDir := t.TempDir()

	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	npcRepo := repository.NewMemoryNPCRepository()

	svc := NewCampaignHealthCheck(canonRepo, stateRepo, questRepo, factionRepo, npcRepo, tmpDir)

	t.Run("detects McGuffin drift", func(t *testing.T) {
		campaignID := "test-mcguffin"

		// Save canon with McGuffin expected location
		canon := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Entities: []domain.CanonEntity{
				{
					ID:         "item-1",
					Type:       domain.EntityTypeItem,
					Name:       "Magic McGuffin",
					CanonState: domain.EntityStateAlive,
					Properties: map[string]any{
						"expected_location": "temple",
					},
				},
			},
		}
		if err := canonRepo.Save(campaignID, canon); err != nil {
			t.Fatalf("failed to save canon: %v", err)
		}

		// Save state with McGuffin in wrong location
		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 3,
			KeyItems: []domain.KeyItem{
				{
					ID:         "item-1",
					Name:       "Magic McGuffin",
					Holder:     "player-inventory",
					IsMcGuffin: true,
				},
			},
		}
		if err := stateRepo.Save(campaignID, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		report, err := svc.RunHealthCheck(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("RunHealthCheck failed: %v", err)
		}

		found := false
		for _, f := range report.Findings {
			if f.Rule == "mcguffin_drift" {
				found = true
				if f.Severity != domain.SeverityCritical {
					t.Errorf("expected critical severity, got %v", f.Severity)
				}
			}
		}
		if !found {
			t.Error("expected mcguffin_drift finding")
		}
	})
}

func TestCampaignHealthCheck_GetHealthReport(t *testing.T) {
	tmpDir := t.TempDir()

	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	npcRepo := repository.NewMemoryNPCRepository()

	svc := NewCampaignHealthCheck(canonRepo, stateRepo, questRepo, factionRepo, npcRepo, tmpDir)

	t.Run("returns error when no report exists", func(t *testing.T) {
		_, err := svc.GetHealthReport(context.Background(), "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent report")
		}
	})

	t.Run("returns saved report", func(t *testing.T) {
		campaignID := "test-report"

		// Save canon and state
		canon := &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    campaignID,
			Entities:      []domain.CanonEntity{},
		}
		if err := canonRepo.Save(campaignID, canon); err != nil {
			t.Fatalf("failed to save canon: %v", err)
		}

		state := &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 1,
		}
		if err := stateRepo.Save(campaignID, state); err != nil {
			t.Fatalf("failed to save state: %v", err)
		}

		// Run health check to create report
		_, err := svc.RunHealthCheck(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("RunHealthCheck failed: %v", err)
		}

		// Retrieve report
		report, err := svc.GetHealthReport(context.Background(), campaignID)
		if err != nil {
			t.Fatalf("GetHealthReport failed: %v", err)
		}

		if report.CampaignID != campaignID {
			t.Errorf("expected campaign ID %s, got %s", campaignID, report.CampaignID)
		}
	})
}

func TestSortFindings(t *testing.T) {
	findings := []HealthFinding{
		{Rule: "rule-b", Severity: domain.SeverityWarning, EntityID: "2"},
		{Rule: "rule-a", Severity: domain.SeverityCritical, EntityID: "1"},
		{Rule: "rule-a", Severity: domain.SeverityCritical, EntityID: "3"},
		{Rule: "rule-c", Severity: domain.SeverityInfo, EntityID: "4"},
	}

	sortFindings(findings)

	// Should be sorted: Critical first, then by rule, then by entity ID
	expected := []string{"1", "3", "2", "4"}
	for i, f := range findings {
		if f.EntityID != expected[i] {
			t.Errorf("position %d: expected entity %s, got %s", i, expected[i], f.EntityID)
		}
	}
}

func TestCalculateSummary(t *testing.T) {
	findings := []HealthFinding{
		{Severity: domain.SeverityCritical},
		{Severity: domain.SeverityCritical},
		{Severity: domain.SeverityWarning},
		{Severity: domain.SeverityWarning},
		{Severity: domain.SeverityWarning},
		{Severity: domain.SeverityInfo},
	}

	summary := calculateSummary(findings)

	if summary.TotalFindings != 6 {
		t.Errorf("expected 6 total findings, got %d", summary.TotalFindings)
	}
	if summary.CriticalCount != 2 {
		t.Errorf("expected 2 critical, got %d", summary.CriticalCount)
	}
	if summary.WarningCount != 3 {
		t.Errorf("expected 3 warnings, got %d", summary.WarningCount)
	}
	if summary.InfoCount != 1 {
		t.Errorf("expected 1 info, got %d", summary.InfoCount)
	}
}

func TestCalculateOverallHealth(t *testing.T) {
	tests := []struct {
		name     string
		summary  HealthSummary
		expected domain.HealthStatus
	}{
		{
			name:     "critical present",
			summary:  HealthSummary{CriticalCount: 1},
			expected: domain.HealthStatusCritical,
		},
		{
			name:     "warning only",
			summary:  HealthSummary{WarningCount: 1},
			expected: domain.HealthStatusFair,
		},
		{
			name:     "info only",
			summary:  HealthSummary{InfoCount: 1},
			expected: domain.HealthStatusGood,
		},
		{
			name:     "no findings",
			summary:  HealthSummary{},
			expected: domain.HealthStatusExcellent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateOverallHealth(tt.summary)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
