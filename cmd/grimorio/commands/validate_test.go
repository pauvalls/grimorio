package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	"github.com/pauvalls/grimorio/internal/services"
	"github.com/urfave/cli/v2"
)

// --- mapCLIScope (pure helper) ---

func TestMapCLIScope(t *testing.T) {
	tests := []struct {
		name           string
		cliScope       string
		wantEngine     domain.ConsistencyScope
		wantPostFilter []string // rule-name prefix; empty = no filter
		wantErr        bool
	}{
		{
			name:       "all maps to Full with no post-filter",
			cliScope:   "all",
			wantEngine: domain.ConsistencyScopeFull,
		},
		{
			name:       "structure maps to LoreOnly with no post-filter",
			cliScope:   "structure",
			wantEngine: domain.ConsistencyScopeLoreOnly,
		},
		{
			name:           "wotc maps to Full and keeps wotc_* rules",
			cliScope:       "wotc",
			wantEngine:     domain.ConsistencyScopeFull,
			wantPostFilter: []string{"wotc_"},
		},
		{
			name:           "references maps to Full and keeps integration rule",
			cliScope:       "references",
			wantEngine:     domain.ConsistencyScopeFull,
			wantPostFilter: []string{"integration"},
		},
		{
			name:       "empty defaults to all",
			cliScope:   "",
			wantEngine: domain.ConsistencyScopeFull,
		},
		{
			name:     "unknown scope returns error",
			cliScope: "garbage",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, filter, err := mapCLIScope(tt.cliScope)
			if (err != nil) != tt.wantErr {
				t.Fatalf("mapCLIScope(%q) error = %v, wantErr = %v", tt.cliScope, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.wantEngine {
				t.Errorf("mapCLIScope(%q) engine = %q, want %q", tt.cliScope, got, tt.wantEngine)
			}
			if !equalStringSlices(filter, tt.wantPostFilter) {
				t.Errorf("mapCLIScope(%q) postFilter = %v, want %v", tt.cliScope, filter, tt.wantPostFilter)
			}
		})
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- runValidate via cli.App with captured stdout ---

// newTestEngine creates an engine backed by in-memory repos with a seeded canon
// document and an empty (but valid) narrative state so a clean campaign
// produces a "good" report.
func newTestEngine(t *testing.T) *services.ValidationEngine {
	t.Helper()
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := services.NewCanonService(canonRepo, stateRepo, nil)
	stateSvc := services.NewNarrativeStateService(stateRepo, canonRepo)

	// Seed a minimal canon so LoadCanon succeeds.
	if err := canonRepo.Save("demo", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		Facts:         []domain.CanonFact{},
		Entities:      []domain.CanonEntity{},
		Rules:         []domain.CanonRule{},
		Timeline:      []domain.CanonTimelineEvent{},
		Relationships: []domain.CanonRelationship{},
	}); err != nil {
		t.Fatalf("seed canon: %v", err)
	}
	// Seed a minimal narrative state.
	if err := stateRepo.Save("demo", &domain.NarrativeState{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
	}); err != nil {
		t.Fatalf("seed state: %v", err)
	}

	return services.NewValidationEngine(canonSvc, stateSvc, nil, "")
}

func newTestApp(engine *services.ValidationEngine) *cli.App {
	return &cli.App{
		Commands: []*cli.Command{
			NewValidateCommand(engine, ""),
		},
		// Override ExitErrHandler so *cli.Exit doesn't call os.Exit and end the test.
		ExitErrHandler: func(_ *cli.Context, _ error) {},
	}
}

// runTestApp runs the test app and captures stdout + stderr. Returns the
// error from app.Run and the captured bytes.
func runTestApp(t *testing.T, app *cli.App, args ...string) (error, string, string) {
	t.Helper()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	err := app.Run(append([]string{"grimorio"}, args...))
	_ = wOut.Close()
	_ = wErr.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	var bufOut, bufErr bytes.Buffer
	_, _ = bufOut.ReadFrom(rOut)
	_, _ = bufErr.ReadFrom(rErr)
	return err, bufOut.String(), bufErr.String()
}

// exitCodeFrom extracts the exit code from a cli.ExitCoder error.
func exitCodeFrom(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(cli.ExitCoder); ok {
		return exitErr.ExitCode()
	}
	return -1
}

func TestRunValidate_RequiresCampaignArg(t *testing.T) {
	engine := newTestEngine(t)
	app := newTestApp(engine)

	err, _, stderr := runTestApp(t, app, "validate")
	if err == nil {
		t.Fatal("expected error for missing campaign arg, got nil")
	}
	if got := exitCodeFrom(err); got != 2 {
		t.Errorf("expected exit code 2 for usage error, got %d (err=%v)", got, err)
	}
	if !strings.Contains(stderr, "campaign name") {
		t.Errorf("expected stderr to contain 'campaign name', got: %q", stderr)
	}
}

func TestRunValidate_CleanCampaign_ExitZero(t *testing.T) {
	engine := newTestEngine(t)
	app := newTestApp(engine)

	err, stdout, _ := runTestApp(t, app, "validate", "demo")
	if err != nil {
		t.Fatalf("expected clean campaign to return nil, got: %v", err)
	}
	if got := exitCodeFrom(err); got != 0 {
		t.Errorf("expected exit code 0, got %d", got)
	}
	if !strings.Contains(stdout, "demo") {
		t.Errorf("expected stdout to contain campaign name 'demo', got: %q", stdout)
	}
	if !strings.Contains(stdout, "All checks passed") {
		t.Errorf("expected stdout to contain 'All checks passed' for clean campaign, got: %q", stdout)
	}
}

func TestRunValidate_UnknownScope_UsageError(t *testing.T) {
	engine := newTestEngine(t)
	app := newTestApp(engine)

	// Note: urfave/cli/v2 only parses flags BEFORE positional args. The
	// real CLI surface (run by main) accepts both, but tests must mirror
	// the library's parsing order. Users will write `grimorio validate --scope=x demo`.
	err, stdout, stderr := runTestApp(t, app, "validate", "--scope=garbage", "demo")
	if err == nil {
		t.Fatalf("expected error for unknown scope, got nil (stdout=%q stderr=%q)", stdout, stderr)
	}
	if got := exitCodeFrom(err); got != 2 {
		t.Errorf("expected exit code 2 for usage error, got %d (err=%v)", got, err)
	}
}

func TestRunValidate_AllScopesAccepted(t *testing.T) {
	engine := newTestEngine(t)

	for _, scope := range []string{"all", "structure", "wotc", "references"} {
		t.Run(scope, func(t *testing.T) {
			app := newTestApp(engine)
			err, _, _ := runTestApp(t, app, "validate", "demo", "--scope="+scope)
			if err != nil {
				t.Errorf("scope=%s should not error on clean campaign, got: %v", scope, err)
			}
		})
	}
}

func TestRunValidate_JSONOutput_RoundTrips(t *testing.T) {
	engine := newTestEngine(t)
	app := newTestApp(engine)

	err, stdout, _ := runTestApp(t, app, "validate", "--json", "demo")
	if err != nil {
		t.Fatalf("--json should not error on clean campaign, got: %v", err)
	}

	if !strings.Contains(stdout, `"campaign_id"`) {
		t.Errorf("expected JSON output to contain campaign_id, got: %s", stdout)
	}
	if !strings.Contains(stdout, `"issues"`) {
		t.Errorf("expected JSON output to contain issues array, got: %s", stdout)
	}

	// round-trip: unmarshal into ConsistencyReport
	var got domain.ConsistencyReport
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("JSON did not round-trip into ConsistencyReport: %v", err)
	}
	if got.CampaignID != "demo" {
		t.Errorf("round-tripped CampaignID = %q, want %q", got.CampaignID, "demo")
	}
}

func TestRunValidate_CriticalIssue_ExitOne(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := services.NewCanonService(canonRepo, stateRepo, nil)
	stateSvc := services.NewNarrativeStateService(stateRepo, canonRepo)

	// Seed canon WITHOUT the npc that will be marked dead.
	if err := canonRepo.Save("demo", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		Facts:         []domain.CanonFact{},
		Entities:      []domain.CanonEntity{},
		Rules:         []domain.CanonRule{},
		Timeline:      []domain.CanonTimelineEvent{},
		Relationships: []domain.CanonRelationship{},
	}); err != nil {
		t.Fatalf("seed canon: %v", err)
	}
	// Seed state with a dead NPC not present in canon → engine reports
	// an "error" severity finding.
	ctx := context.Background()
	if err := stateSvc.Save(ctx, &domain.NarrativeState{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-missing", Name: "MissingNPC", Session: 1, Cause: "test"},
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	engine := services.NewValidationEngine(canonSvc, stateSvc, nil, "")
	app := newTestApp(engine)
	err, _, _ := runTestApp(t, app, "validate", "demo", "--json")

	if got := exitCodeFrom(err); got != 1 {
		t.Errorf("expected exit code 1 for errors, got %d (err=%v)", got, err)
	}
}

func TestRunValidate_DefaultScope_IsAll(t *testing.T) {
	engine := newTestEngine(t)
	app := newTestApp(engine)

	err, stdout, _ := runTestApp(t, app, "validate", "demo")
	if err != nil {
		t.Fatalf("default scope should not error on clean campaign, got: %v", err)
	}
	if !strings.Contains(stdout, "demo") {
		t.Errorf("expected output to contain campaign name 'demo', got: %q", stdout)
	}
}

// TestRunValidate_WotCScope_FiltersNonWotCRules verifies the post-filter
// keeps only wotc_* rules. We seed a dead-NPC error (rule = "npc_alive_check")
// which is NOT a wotc_ rule, then assert the report's issues list is empty
// under --scope=wotc and the exit code is 0.
func TestRunValidate_WotCScope_FiltersNonWotCRules(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := services.NewCanonService(canonRepo, stateRepo, nil)
	stateSvc := services.NewNarrativeStateService(stateRepo, canonRepo)

	if err := canonRepo.Save("demo", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		Facts:         []domain.CanonFact{},
		Entities:      []domain.CanonEntity{},
		Rules:         []domain.CanonRule{},
		Timeline:      []domain.CanonTimelineEvent{},
		Relationships: []domain.CanonRelationship{},
	}); err != nil {
		t.Fatalf("seed canon: %v", err)
	}
	ctx := context.Background()
	if err := stateSvc.Save(ctx, &domain.NarrativeState{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-missing", Name: "MissingNPC", Session: 1, Cause: "test"},
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	engine := services.NewValidationEngine(canonSvc, stateSvc, nil, "")
	app := newTestApp(engine)
	err, stdout, _ := runTestApp(t, app, "validate", "--scope=wotc", "--json", "demo")

	// wotc filter drops the npc_alive_check error → exit 0
	if err != nil {
		t.Fatalf("--scope=wotc should filter out non-wotc rules and pass, got: %v", err)
	}

	// JSON should show no errors after filtering.
	var got domain.ConsistencyReport
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("JSON did not round-trip: %v", err)
	}
	for _, issue := range got.Issues {
		if issue.Rule != "wotc_developments" && issue.Rule != "wotc_multiple_solutions" &&
			issue.Rule != "wotc_character_hooks" && issue.Rule != "wotc_boxed_text" &&
			issue.Rule != "wotc_npc_word_count" {
			t.Errorf("expected only wotc_* rules after filter, found: %s", issue.Rule)
		}
	}
}

// TestRunValidate_ReferencesScope_FiltersNonIntegrationRules verifies the
// post-filter keeps only "integration" rules. We seed a wotc error and a
// npc_alive_check error; the references scope should drop both and exit 0.
func TestRunValidate_ReferencesScope_FiltersNonIntegrationRules(t *testing.T) {
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	canonSvc := services.NewCanonService(canonRepo, stateRepo, nil)
	stateSvc := services.NewNarrativeStateService(stateRepo, canonRepo)

	if err := canonRepo.Save("demo", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		Facts:         []domain.CanonFact{},
		Entities:      []domain.CanonEntity{},
		Rules:         []domain.CanonRule{},
		Timeline:      []domain.CanonTimelineEvent{},
		Relationships: []domain.CanonRelationship{},
	}); err != nil {
		t.Fatalf("seed canon: %v", err)
	}
	ctx := context.Background()
	if err := stateSvc.Save(ctx, &domain.NarrativeState{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "demo",
		DeadNPCs: []domain.NPCDeathRecord{
			{NPCID: "npc-missing", Name: "MissingNPC", Session: 1, Cause: "test"},
		},
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	engine := services.NewValidationEngine(canonSvc, stateSvc, nil, "")
	app := newTestApp(engine)
	err, stdout, _ := runTestApp(t, app, "validate", "--scope=references", "--json", "demo")

	// references filter drops npc_alive_check and wotc_* errors → exit 0
	if err != nil {
		t.Fatalf("--scope=references should filter out non-integration rules, got: %v", err)
	}

	var got domain.ConsistencyReport
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("JSON did not round-trip: %v", err)
	}
	for _, issue := range got.Issues {
		if issue.Rule != "integration" {
			t.Errorf("expected only integration rules after filter, found: %s", issue.Rule)
		}
	}
}

// TestRenderTextReport_CoversAllIssues exercises renderTextReport directly
// to ensure it lists all severity types — keeps the pure formatter honest
// independent of CLI plumbing.
func TestRenderTextReport_CoversAllIssues(t *testing.T) {
	report := &domain.ConsistencyReport{
		CampaignID:    "demo",
		OverallHealth: "poor",
		TotalChecks:   3,
		Errors:        1,
		Warnings:      1,
		Criticals:     1,
		Issues: []domain.CheckResult{
			{Rule: "npc_alive_check", Passed: false, Severity: "error", Message: "Missing NPC"},
			{Rule: "wotc_boxed_text", Passed: false, Severity: "warning", Message: "Too short"},
			{Rule: "integration", Passed: false, Severity: "critical", Message: "Broken cross-ref"},
		},
	}

	var buf bytes.Buffer
	renderTextReport(&buf, report)
	out := buf.String()

	if !strings.Contains(out, "Missing NPC") {
		t.Errorf("expected report to contain 'Missing NPC', got: %q", out)
	}
	if !strings.Contains(out, "Too short") {
		t.Errorf("expected report to contain 'Too short', got: %q", out)
	}
	if !strings.Contains(out, "Broken cross-ref") {
		t.Errorf("expected report to contain 'Broken cross-ref', got: %q", out)
	}
	if !strings.Contains(out, "1 errors, 1 warnings, 1 criticals") {
		t.Errorf("expected summary line, got: %q", out)
	}
}

func TestFilterReport_KeepsMatchingRulesAndRecomputesStats(t *testing.T) {
	report := &domain.ConsistencyReport{
		CampaignID:    "demo",
		OverallHealth: "ignored",
		TotalChecks:   99,
		Errors:        99,
		Issues: []domain.CheckResult{
			{Rule: "wotc_developments", Passed: false, Severity: "warning", Message: "no branches"},
			{Rule: "npc_alive_check", Passed: false, Severity: "error", Message: "dead npc"},
			{Rule: "wotc_boxed_text", Passed: true, Severity: "info", Message: "ok"},
		},
	}

	filterReport(report, []string{"wotc_"})

	if report.TotalChecks != 2 {
		t.Errorf("TotalChecks = %d, want 2", report.TotalChecks)
	}
	if report.Passed != 1 {
		t.Errorf("Passed = %d, want 1", report.Passed)
	}
	if report.Warnings != 1 {
		t.Errorf("Warnings = %d, want 1", report.Warnings)
	}
	if report.Errors != 0 {
		t.Errorf("Errors = %d, want 0 after filter", report.Errors)
	}
	if report.OverallHealth != "fair" {
		t.Errorf("OverallHealth = %q, want 'fair' (only warning present)", report.OverallHealth)
	}
}

func TestFilterReport_EmptyAfterFilter_HealthIsExcellent(t *testing.T) {
	report := &domain.ConsistencyReport{
		CampaignID:    "demo",
		OverallHealth: "poor",
		Issues: []domain.CheckResult{
			{Rule: "npc_alive_check", Passed: false, Severity: "error", Message: "x"},
		},
	}
	filterReport(report, []string{"integration"})
	if report.TotalChecks != 0 {
		t.Errorf("TotalChecks = %d, want 0", report.TotalChecks)
	}
	if report.OverallHealth != "excellent" {
		t.Errorf("OverallHealth = %q, want 'excellent'", report.OverallHealth)
	}
}

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefixes []string
		want     bool
	}{
		{"empty prefixes", "rule", nil, false},
		{"exact prefix match", "wotc_box", []string{"wotc_"}, true},
		{"no match", "integration", []string{"wotc_"}, false},
		{"multiple prefixes, second matches", "integration", []string{"wotc_", "integration"}, true},
		{"prefix longer than name", "wot", []string{"wotc_"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesAny(tt.input, tt.prefixes); got != tt.want {
				t.Errorf("matchesAny(%q, %v) = %v, want %v", tt.input, tt.prefixes, got, tt.want)
			}
		})
	}
}
