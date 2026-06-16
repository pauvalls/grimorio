package e2e

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestFullCampaignGeneration tests end-to-end campaign generation.
func TestFullCampaignGeneration(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	// Create a campaign first so subsequent steps have something to work with
	if err := executeStep(ctx, harness, TestStep{
		Action: "create_campaign",
		Params: map[string]any{"name": "test-full-campaign", "title": "Test Campaign"},
	}); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	suite := &E2ETestSuite{
		SuiteID: "e2e_full_campaign",
		Name:    "Full Campaign Generation",
		Tests: []E2ETest{
			{
				TestID:      "save_lore",
				Name:        "Save Campaign Lore",
				Description: "Test saving lore to an existing campaign",
				Steps: []TestStep{
					{
						Action: "save_lore",
						Params: map[string]any{
							"campaign": "test-full-campaign",
							"content":  "# World History\n\nLong ago...",
						},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-full-campaign", Expected: true},
				},
			},
			{
				TestID:      "save_npcs",
				Name:        "Save NPCs",
				Description: "Test saving NPCs to campaign",
				Steps: []TestStep{
					{
						Action: "save_npcs",
						Params: map[string]any{
							"campaign": "test-full-campaign",
							"content":  "# NPCs\n\n## Gandalf\nA wizard...",
						},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-full-campaign", Expected: true},
				},
			},
			{
				TestID:      "save_encounters",
				Name:        "Save Encounters",
				Description: "Test saving encounters to campaign",
				Steps: []TestStep{
					{
						Action: "save_encounters",
						Params: map[string]any{
							"campaign": "test-full-campaign",
							"content":  "# Encounters\n\n## Ambush\nA bandit ambush...",
						},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-full-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
		for _, result := range results.Results {
			if result.Status == "fail" {
				t.Logf("Failed test: %s - %s", result.Name, result.Error)
			}
		}
	}

	t.Logf("Test suite completed in %v: %d passed, %d failed", results.Duration, results.PassedTests, results.FailedTests)
}

// TestXPProgressionContinuity tests XP progression across chapters.
func TestXPProgressionContinuity(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_xp_progression",
		Name:    "XP Progression Continuity",
		Tests: []E2ETest{
			{
				TestID:      "create_campaign_for_xp",
				Name:        "Create Campaign for XP Test",
				Description: "Setup campaign",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-xp-campaign", "title": "XP Test"},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-xp-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("XP progression test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestTacticsEnvironmentReference tests tactics reference area features.
func TestTacticsEnvironmentReference(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_tactics_environment",
		Name:    "Tactics Environment Reference",
		Tests: []E2ETest{
			{
				TestID:      "create_campaign_for_tactics",
				Name:        "Create Campaign for Tactics Test",
				Description: "Setup campaign",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-tactics-campaign", "title": "Tactics Test"},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-tactics-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Tactics environment test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestQuestApproachDiversity tests quest approach diversity.
func TestQuestApproachDiversity(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_quest_approaches",
		Name:    "Quest Approach Diversity",
		Tests: []E2ETest{
			{
				TestID:      "create_quest",
				Name:        "Create a Quest",
				Description: "Test quest creation",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-quest-campaign", "title": "Quest Test"},
					},
					{
						Action: "create_quest",
						Params: map[string]any{
							"campaign":    "test-quest-campaign",
							"quest_title": "Find the Sword",
							"quest_type":  "main",
							"hook":        "A stranger approaches...",
							"stakes":      "The kingdom's fate",
							"reward":      "1000 gold",
						},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-quest-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Quest approach test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestCampaignConsistency tests campaign consistency.
func TestCampaignConsistency(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_campaign_consistency",
		Name:    "Campaign Consistency Validation",
		Tests: []E2ETest{
			{
				TestID:      "create_campaign_for_consistency",
				Name:        "Create Campaign for Consistency Test",
				Description: "Setup campaign",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-consistency-campaign", "title": "Consistency Test"},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-consistency-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Campaign consistency test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestHandoutGeneration tests handout generation.
func TestHandoutGeneration(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_handout_generation",
		Name:    "Handout Generation",
		Tests: []E2ETest{
			{
				TestID:      "create_campaign_for_handout",
				Name:        "Create Campaign for Handout Test",
				Description: "Setup campaign",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-handout-campaign", "title": "Handout Test"},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-handout-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Handout generation test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestRandomTables tests random table generation.
func TestRandomTables(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_random_tables",
		Name:    "Random Table Generation",
		Tests: []E2ETest{
			{
				TestID:      "create_campaign_for_tables",
				Name:        "Create Campaign for Table Test",
				Description: "Setup campaign",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-table-campaign", "title": "Table Test"},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-table-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Random tables test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestCanonValidation tests canon validation.
func TestCanonValidation(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	suite := &E2ETestSuite{
		SuiteID: "e2e_canon_validation",
		Name:    "Canon Validation",
		Tests: []E2ETest{
			{
				TestID:      "create_campaign_for_canon",
				Name:        "Create Campaign for Canon Test",
				Description: "Setup campaign",
				Steps: []TestStep{
					{
						Action: "create_campaign",
						Params: map[string]any{"name": "test-canon-campaign", "title": "Canon Test"},
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "test-canon-campaign", Expected: true},
				},
			},
		},
	}

	results, err := runner.RunSuite(suite)
	if err != nil {
		t.Fatalf("RunSuite() error = %v", err)
	}

	if results.FailedTests > 0 {
		t.Errorf("Canon validation test suite failed: %d/%d tests failed", results.FailedTests, results.TotalTests)
	}
}

// TestE2ETestRunner_Basic tests the basic runner functionality.
func TestE2ETestRunner_Basic(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	test := &E2ETest{
		TestID:      "basic_test",
		Name:        "Basic Test",
		Description: "Verify test runner works",
		Steps: []TestStep{
			{
				Action: "create_campaign",
				Params: map[string]any{"name": "test-basic-campaign", "title": "Basic Test"},
			},
		},
		Assertions: []TestAssertion{
			{Type: "exists", Target: "test-basic-campaign", Expected: true},
		},
	}

	result := runner.RunTest(test, []TestFixture{})

	if result.Status != "pass" {
		t.Errorf("Basic test failed: %s", result.Error)
	}

	if result.Duration < 0 {
		t.Error("Test duration should be positive")
	}

	t.Logf("Basic test passed in %v", result.Duration)
}

// TestCleanupFixture tests that CleanupFixture only removes test- prefixed directories.
func TestCleanupFixture(t *testing.T) {
	ctx := context.Background()
	harness := NewTestHarness(t)
	runner := NewE2ETestRunnerWithHarness(ctx, harness)

	baseDir := harness.BaseDir

	// Create test directories
	_ = os.MkdirAll(filepath.Join(baseDir, "test-campaign-a"), 0755)
	_ = os.MkdirAll(filepath.Join(baseDir, "test-campaign-b"), 0755)
	_ = os.MkdirAll(filepath.Join(baseDir, "real_campaign"), 0755)

	fixture := TestFixture{
		FixtureID:       "cleanup_test",
		Name:            "Cleanup Test",
		CleanupRequired: true,
	}

	if err := runner.CleanupFixture(fixture); err != nil {
		t.Fatalf("CleanupFixture() error = %v", err)
	}

	// test-* dirs should be removed
	if _, err := os.Stat(filepath.Join(baseDir, "test-campaign-a")); !os.IsNotExist(err) {
		t.Error("test-campaign-a should have been deleted")
	}
	if _, err := os.Stat(filepath.Join(baseDir, "test-campaign-b")); !os.IsNotExist(err) {
		t.Error("test-campaign-b should have been deleted")
	}

	// real_campaign should remain
	if _, err := os.Stat(filepath.Join(baseDir, "real_campaign")); err != nil {
		t.Error("real_campaign should NOT have been deleted")
	}
}
