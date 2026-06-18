package services

import (
	"context"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

// newTestHealthCheckWithFiles wires a CampaignHealthCheck with a fake reader
// so tests can inject markdown files without touching the filesystem.
func newTestHealthCheckWithFiles(t *testing.T, files []consolidation.CampaignFile) (*CampaignHealthCheck, *domain.CanonDocument) {
	t.Helper()
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	npcRepo := repository.NewMemoryNPCRepository()

	canon := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "fixture",
		Entities:      []domain.CanonEntity{},
		Timeline:      []domain.CanonTimelineEvent{},
		Rules:         []domain.CanonRule{},
		Facts:         []domain.CanonFact{},
		Relationships: []domain.CanonRelationship{},
	}
	if err := canonRepo.Save("fixture", canon); err != nil {
		t.Fatalf("save canon: %v", err)
	}
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "fixture",
		CurrentSession: 1,
		SessionLog:     []domain.SessionRecord{},
	}
	if err := stateRepo.Save("fixture", state); err != nil {
		t.Fatalf("save state: %v", err)
	}

	svc := NewCampaignHealthCheck(canonRepo, stateRepo, questRepo, factionRepo, npcRepo, t.TempDir())
	svc.consolidator = NewConsolidationAdapterWithReader(t.TempDir(), &fakeReader{files: files})
	return svc, canon
}

func findingByRule(findings []HealthFinding, rule string) *HealthFinding {
	for i := range findings {
		if findings[i].Rule == rule {
			return &findings[i]
		}
	}
	return nil
}

func TestCampaignHealthCheck_Consolidation_DetectAllDrift(t *testing.T) {
	files := []consolidation.CampaignFile{
		{
			RelPath: "lore.md",
			Content: "Treaty of Ashford 1247.\nTreaty of Ashford 1251.\n",
		},
		{
			RelPath: "bestiary/boss.md",
			Content: "# Boss\n\nCR 5\n",
		},
		{
			RelPath: "acts/act1.md",
			Content: "# Boss\n\nCR 9\n",
		},
		{
			RelPath: "npcs/v.md",
			Content: "# Velara\n",
		},
		{
			RelPath: "npcs/v2.md",
			Content: "# Velarra\n",
		},
		{
			RelPath: "areas/cellar.md",
			Content: "# Cellar\n\nSee [map](assets/maps/cellar.svg).\n",
		},
		{
			RelPath: "areas/dup1.md",
			Content: "Same content.\n",
		},
		{
			RelPath: "areas/dup2.md",
			Content: "Same content.\n",
		},
	}
	svc, _ := newTestHealthCheckWithFiles(t, files)
	report, err := svc.RunHealthCheck(context.Background(), "fixture")
	if err != nil {
		t.Fatalf("RunHealthCheck: %v", err)
	}

	// Each consolidation rule should emit at least one finding.
	for _, rule := range []string{
		"duplicate_file",
		"missing_map",
		"lore_contradiction",
		"entity_name_collision",
		"boss_stat_block_drift",
	} {
		if findingByRule(report.Findings, rule) == nil {
			t.Errorf("expected finding for rule %q, got %d findings", rule, len(report.Findings))
		}
	}
}

func TestCampaignHealthCheck_Consolidation_SeverityMapping(t *testing.T) {
	files := []consolidation.CampaignFile{
		{
			RelPath: "lore.md",
			Content: "Treaty of Ashford 1247.\nTreaty of Ashford 1251.\n",
		},
		{
			RelPath: "bestiary/boss.md",
			Content: "# Boss\n\nCR 5\n",
		},
		{
			RelPath: "acts/act1.md",
			Content: "# Boss\n\nCR 9\n",
		},
		{
			RelPath: "areas/cellar.md",
			Content: "# Cellar\n\nSee [map](assets/maps/missing.svg).\n",
		},
		{
			RelPath: "areas/dup1.md",
			Content: "Same content.\n",
		},
		{
			RelPath: "areas/dup2.md",
			Content: "Same content.\n",
		},
	}
	svc, _ := newTestHealthCheckWithFiles(t, files)
	report, err := svc.RunHealthCheck(context.Background(), "fixture")
	if err != nil {
		t.Fatalf("RunHealthCheck: %v", err)
	}

	// Lore contradictions and missing maps should be critical.
	if f := findingByRule(report.Findings, "lore_contradiction"); f == nil || f.Severity != domain.SeverityCritical {
		t.Errorf("expected critical lore_contradiction, got %+v", f)
	}
	if f := findingByRule(report.Findings, "missing_map"); f == nil || f.Severity != domain.SeverityCritical {
		t.Errorf("expected critical missing_map, got %+v", f)
	}
	if f := findingByRule(report.Findings, "boss_stat_block_drift"); f == nil || f.Severity != domain.SeverityCritical {
		t.Errorf("expected critical boss_stat_block_drift, got %+v", f)
	}
	// Duplicates should be warnings.
	if f := findingByRule(report.Findings, "duplicate_file"); f == nil || f.Severity != domain.SeverityWarning {
		t.Errorf("expected warning duplicate_file, got %+v", f)
	}
}

func TestCampaignHealthCheck_Consolidation_EmptyCampaign(t *testing.T) {
	// No markdown files at all → consolidator returns no files → no
	// consolidation findings, but the existing checks still run on the
	// seeded canon (which is empty → no findings either).
	svc, _ := newTestHealthCheckWithFiles(t, nil)
	report, err := svc.RunHealthCheck(context.Background(), "fixture")
	if err != nil {
		t.Fatalf("RunHealthCheck: %v", err)
	}
	for _, f := range report.Findings {
		switch f.Rule {
		case "duplicate_file", "stale_generated_file", "missing_map",
			"lore_contradiction", "entity_name_collision", "boss_stat_block_drift":
			t.Errorf("unexpected consolidation finding for empty campaign: %+v", f)
		}
	}
}

func TestAnalysisToHealthFindings_NilResult(t *testing.T) {
	if got := AnalysisToHealthFindings("rule_x", nil); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
