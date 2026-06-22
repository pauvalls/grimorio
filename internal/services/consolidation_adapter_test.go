package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services/consolidation"
)

// writeFile is a tiny helper that fails the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// fakeReader lets tests inject in-memory CampaignFile lists without
// touching the filesystem.
type fakeReader struct {
	files []consolidation.CampaignFile
}

func (r *fakeReader) ReadCampaign(_ context.Context, _ string) ([]consolidation.CampaignFile, error) {
	return r.files, nil
}

func newTestEngineWithFiles(t *testing.T, files []consolidation.CampaignFile, baseDir string) (*ValidationEngine, *CanonService, *NarrativeStateService) {
	t.Helper()
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonService := NewCanonService(canonRepo, stateRepo, repository.NewMemoryCheckpointRepository())
	stateService := NewNarrativeStateService(stateRepo, canonRepo)
	engine := NewValidationEngine(canonService, stateService, repository.NewMemoryFactionRepository(), baseDir)
	engine.consolidator = NewConsolidationAdapterWithReader(baseDir, &fakeReader{files: files})
	return engine, canonService, stateService
}

// seedCanon stores an empty canon document so CheckConsistency's
// LoadCanon call does not fail. Returns the canon service so tests can
// mutate further.
func seedCanon(t *testing.T, cs *CanonService, campaignID string) {
	t.Helper()
	if err := cs.SaveCanon(context.Background(), &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    campaignID,
		Entities:      []domain.CanonEntity{},
		Timeline:      []domain.CanonTimelineEvent{},
		Rules:         []domain.CanonRule{},
		Facts:         []domain.CanonFact{},
		Relationships: []domain.CanonRelationship{},
	}); err != nil {
		t.Fatalf("seedCanon: %v", err)
	}
}

func TestConsolidationAdapter_FilesReturnsNilForMissingDir(t *testing.T) {
	adapter := NewConsolidationAdapter(t.TempDir())
	files, err := adapter.Files(context.Background(), "no-such-campaign")
	if err != nil {
		t.Fatalf("Files: %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files for missing dir, got %d", len(files))
	}
}

func TestConsolidationAdapter_RunAnalyzerSkipsEmptyCampaigns(t *testing.T) {
	adapter := NewConsolidationAdapter(t.TempDir())
	res, err := adapter.RunAnalyzer(context.Background(), "absent", consolidation.NewLoreCoherence())
	if err != nil {
		t.Fatalf("RunAnalyzer: %v", err)
	}
	if res != nil {
		t.Errorf("expected nil result for empty campaign, got %+v", res)
	}
}

func TestConsolidationAdapter_RunAnalyzerOnFilesystem(t *testing.T) {
	// Build a real campaign directory on disk to exercise the integration.
	base := t.TempDir()
	campaign := "fixture"
	campDir := filepath.Join(base, campaign)
	// Note: one treaty per line, so the analyzer can pair each name with its year.
	content := "Treaty of Ashford 1247 sealed the border.\n" +
		"Treaty of Ashford 1251 broke the pact.\n"
	writeFile(t, filepath.Join(campDir, "lore.md"), content)

	adapter := NewConsolidationAdapter(base)
	res, err := adapter.RunAnalyzer(context.Background(), campaign, consolidation.NewLoreCoherence())
	if err != nil {
		t.Fatalf("RunAnalyzer: %v", err)
	}
	if res == nil {
		t.Fatal("expected non-nil result")
	}
	if res.Passed {
		t.Errorf("expected lore issue to fail, got passed=true")
	}
}

func TestConsolidationAdapter_RunAllAnalyzers(t *testing.T) {
	adapter := NewConsolidationAdapter(t.TempDir())
	results, err := adapter.RunAllAnalyzers(context.Background(), "absent")
	if err != nil {
		t.Fatalf("RunAllAnalyzers: %v", err)
	}
	// Empty campaign → empty result map (no analyzers run, no failures).
	if len(results) != 0 {
		t.Errorf("expected 0 analyzers for empty campaign, got %d", len(results))
	}
}

func TestConsolidationAdapter_RunAllAnalyzersWithFiles(t *testing.T) {
	files := []consolidation.CampaignFile{
		{RelPath: "lore.md", Content: "Treaty of Ashford 1247.\nTreaty of Ashford 1251.\n"},
	}
	adapter := NewConsolidationAdapterWithReader(t.TempDir(), &fakeReader{files: files})
	results, err := adapter.RunAllAnalyzers(context.Background(), "any")
	if err != nil {
		t.Fatalf("RunAllAnalyzers: %v", err)
	}
	if len(results) != 7 {
		t.Errorf("expected 7 analyzers, got %d", len(results))
	}
}

func TestAnalysisToCheckResults_NilResult(t *testing.T) {
	checks := AnalysisToCheckResults("rule_x", nil)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if !checks[0].Passed {
		t.Errorf("expected passing check for nil result, got %+v", checks[0])
	}
}

func TestAnalysisToCheckResults_PropagatesSeverityAndMessage(t *testing.T) {
	res := &consolidation.AnalysisResult{
		Rule:      "rule_x",
		Passed:    false,
		Severity:  "critical",
		Message:   "boom",
		Locations: []string{"a.md", "b.md"},
	}
	checks := AnalysisToCheckResults("rule_x", res)
	if len(checks) != 1 {
		t.Fatalf("expected 1 check, got %d", len(checks))
	}
	if checks[0].Severity != "critical" {
		t.Errorf("severity = %q, want critical", checks[0].Severity)
	}
	if checks[0].Message != "boom" {
		t.Errorf("message = %q, want boom", checks[0].Message)
	}
	if checks[0].Location != "a.md, b.md" {
		t.Errorf("location = %q, want 'a.md, b.md'", checks[0].Location)
	}
}

// TestValidationEngine_ConsolidationChecks_DetectAllDrift is the headline
// integration test: it injects a markdown campaign with several classes of
// drift and asserts each new validation method emits the right CheckResult.
func TestValidationEngine_ConsolidationChecks_DetectAllDrift(t *testing.T) {
	files := []consolidation.CampaignFile{
		{
			RelPath: "lore.md",
			Content: "Treaty of Ashford 1247 sealed the border.\n" +
				"Treaty of Ashford 1251 broke the pact.\n",
		},
		{
			RelPath: "bestiary/araoxs.md",
			Content: "# Araxos\n\nCR 7\n",
		},
		{
			RelPath: "acts/act1.md",
			Content: "# Araxos\n\nCR 9\n",
		},
		{
			RelPath: "npcs/velaplata.md",
			Content: "# Velaplata the Bold\n",
		},
		{
			RelPath: "npcs/velaplanta.md",
			Content: "# Velaplanta the Bold\n",
		},
		{
			RelPath: "areas/dungeon.md",
			Content: "# Dungeon\n\nThe map ([cellar](assets/maps/cellar.svg)) leads underground.\n",
		},
		{
			RelPath: "areas/duplicate1.md",
			Content: "# Duplicate\n\nSame content here.\n",
		},
		{
			RelPath: "areas/duplicate2.md",
			Content: "# Duplicate\n\nSame content here.\n",
		},
	}
	engine, canon, _ := newTestEngineWithFiles(t, files, t.TempDir())
	seedCanon(t, canon, "drift")
	report, err := engine.CheckConsistency(context.Background(), "drift", domain.ConsistencyScopeFull)
	if err != nil {
		t.Fatalf("CheckConsistency: %v", err)
	}

	rulesSeen := map[string]int{}
	for _, issue := range report.Issues {
		rulesSeen[issue.Rule]++
	}

	// Each consolidation check should have produced at least one issue.
	for _, want := range []string{
		"consolidation_lore_coherence",
		"consolidation_stat_block_consistency",
		"consolidation_entity_uniqueness",
		"consolidation_map_assets_exist",
		"consolidation_no_duplicate_files",
	} {
		if rulesSeen[want] == 0 {
			t.Errorf("expected at least one issue for rule %q (got rules: %v)", want, rulesSeen)
		}
	}
}

// TestValidationEngine_ConsolidationChecks_NotRunForNonFullScope verifies
// the new checks are scoped to ConsistencyScopeFull.
func TestValidationEngine_ConsolidationChecks_NotRunForNonFullScope(t *testing.T) {
	files := []consolidation.CampaignFile{
		{RelPath: "lore.md", Content: "Treaty of Ashford 1247. Treaty of Ashford 1251.\n"},
	}
	engine, canon, _ := newTestEngineWithFiles(t, files, t.TempDir())
	seedCanon(t, canon, "drift")
	report, err := engine.CheckConsistency(context.Background(), "drift", domain.ConsistencyScopeLoreOnly)
	if err != nil {
		t.Fatalf("CheckConsistency: %v", err)
	}
	for _, issue := range report.Issues {
		if issue.Rule == "consolidation_lore_coherence" {
			t.Errorf("consolidation check should not run for non-full scope")
		}
	}
}
