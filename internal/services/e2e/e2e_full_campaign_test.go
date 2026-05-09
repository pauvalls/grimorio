package e2e

import (
	"context"
	"testing"
)

// TestFullCampaignGeneration tests end-to-end campaign generation.
func TestFullCampaignGeneration(t *testing.T) {
	ctx := context.Background()
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_full_campaign",
		Name:    "Full Campaign Generation",
		Tests: []E2ETest{
			{
				TestID:      "generate_campaign_from_brief",
				Name:        "Generate Campaign from Brief",
				Description: "Test campaign generation from initial brief",
				Steps: []TestStep{
					{
						Action:         "generate_campaign",
						Tool:           "grimorio_generate_campaign",
						Params:         map[string]any{"brief": "test campaign"},
						ExpectedStatus: "success",
					},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "campaign", Expected: true},
					{Type: "exists", Target: "acts", Expected: true},
				},
			},
			{
				TestID:      "generate_all_chapters",
				Name:        "Generate All Chapters",
				Description: "Test generation of all 3 acts",
				Steps: []TestStep{
					{Action: "generate_act", Tool: "grimorio_generate_act", Params: map[string]any{"act": 1}},
					{Action: "generate_act", Tool: "grimorio_generate_act", Params: map[string]any{"act": 2}},
					{Action: "generate_act", Tool: "grimorio_generate_act", Params: map[string]any{"act": 3}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "act_1", Expected: true},
					{Type: "exists", Target: "act_2", Expected: true},
					{Type: "exists", Target: "act_3", Expected: true},
				},
			},
			{
				TestID:      "generate_all_npcs",
				Name:        "Generate All NPCs",
				Description: "Test NPC generation for campaign",
				Steps: []TestStep{
					{Action: "generate_npcs", Tool: "grimorio_generate_npcs"},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "npcs", Expected: true},
					{Type: "contains", Target: "npcs", Expected: "allies"},
					{Type: "contains", Target: "npcs", Expected: "enemies"},
				},
			},
			{
				TestID:      "compile_pdf",
				Name:        "Compile PDF",
				Description: "Test PDF compilation",
				Steps: []TestStep{
					{Action: "compile_pdf", Tool: "grimorio_compile_pdf"},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "pdf_file", Expected: true},
					{Type: "validates", Target: "pdf_file", Expected: "valid_pdf"},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_xp_progression",
		Name:    "XP Progression Continuity",
		Tests: []E2ETest{
			{
				TestID:      "generate_xp_tables",
				Name:        "Generate XP Tables for All Chapters",
				Description: "Test XP table generation",
				Steps: []TestStep{
					{Action: "generate_xp_table", Tool: "grimorio_generate_xp_table", Params: map[string]any{"chapter": 1}},
					{Action: "generate_xp_table", Tool: "grimorio_generate_xp_table", Params: map[string]any{"chapter": 2}},
					{Action: "generate_xp_table", Tool: "grimorio_generate_xp_table", Params: map[string]any{"chapter": 3}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "xp_tables", Expected: true},
				},
			},
			{
				TestID:      "verify_no_level_gaps",
				Name:        "Verify No Level Gaps",
				Description: "Test continuous level progression",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "validates", Target: "level_progression", Expected: "continuous"},
				},
			},
			{
				TestID:      "verify_phb_thresholds",
				Name:        "Verify PHB Thresholds",
				Description: "Test cumulative XP matches PHB",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "equals", Target: "xp_thresholds", Expected: "PHB_standard"},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_tactics_environment",
		Name:    "Tactics Environment Reference",
		Tests: []E2ETest{
			{
				TestID:      "generate_tactics",
				Name:        "Generate Tactics for Encounters",
				Description: "Test tactics generation",
				Steps: []TestStep{
					{Action: "generate_tactics", Tool: "grimorio_generate_tactics"},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "tactics", Expected: true},
				},
			},
			{
				TestID:      "verify_area_features",
				Name:        "Verify Tactics Reference Area Features",
				Description: "Test tactics use actual area features",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "contains", Target: "environmental_tactics", Expected: "area_features"},
				},
			},
			{
				TestID:      "verify_intelligence_tier",
				Name:        "Verify Intelligence Tier Matches Monster INT",
				Description: "Test INT score to tier mapping",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "validates", Target: "intelligence_tier", Expected: "matches_INT"},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_quest_approaches",
		Name:    "Quest Approach Diversity",
		Tests: []E2ETest{
			{
				TestID:      "generate_quest_chain",
				Name:        "Generate Quest Chain Across Chapter",
				Description: "Test quest chain generation",
				Steps: []TestStep{
					{Action: "generate_quest", Tool: "grimorio_generate_quest"},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "quest_chain", Expected: true},
				},
			},
			{
				TestID:      "verify_three_approaches",
				Name:        "Verify 3+ Distinct Approaches",
				Description: "Test approach diversity",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "contains", Target: "approaches", Expected: "combat"},
					{Type: "contains", Target: "approaches", Expected: "social"},
					{Type: "contains", Target: "approaches", Expected: "stealth"},
				},
			},
			{
				TestID:      "verify_failure_states",
				Name:        "Verify Failure State Consequences",
				Description: "Test failure consequences propagate",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "failure_states", Expected: true},
					{Type: "contains", Target: "failure_states", Expected: "soft"},
					{Type: "contains", Target: "failure_states", Expected: "hard"},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_campaign_consistency",
		Name:    "Campaign Consistency Validation",
		Tests: []E2ETest{
			{
				TestID:      "validate_npc_references",
				Name:        "Validate NPC References Exist",
				Description: "Test all NPC references are valid",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "validates", Target: "npc_references", Expected: "all_exist"},
				},
			},
			{
				TestID:      "validate_quest_chains",
				Name:        "Validate Quest Chains Complete",
				Description: "Test quest chains have no gaps",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "validates", Target: "quest_chains", Expected: "complete"},
				},
			},
			{
				TestID:      "validate_xp_progression",
				Name:        "Validate XP Progression No Gaps",
				Description: "Test XP has no gaps",
				Steps:       []TestStep{},
				Assertions: []TestAssertion{
					{Type: "validates", Target: "xp_progression", Expected: "no_gaps"},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_handout_generation",
		Name:    "Handout Generation",
		Tests: []E2ETest{
			{
				TestID:      "generate_letter",
				Name:        "Generate Letter Handout",
				Description: "Test letter generation",
				Steps: []TestStep{
					{Action: "generate_handout", Tool: "grimorio_generate_handout", Params: map[string]any{"type": "letter"}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "handout", Expected: true},
					{Type: "contains", Target: "content", Expected: "sender"},
					{Type: "contains", Target: "content", Expected: "recipient"},
				},
			},
			{
				TestID:      "generate_clue",
				Name:        "Generate Clue Handout",
				Description: "Test clue generation",
				Steps: []TestStep{
					{Action: "generate_handout", Tool: "grimorio_generate_handout", Params: map[string]any{"type": "clue"}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "handout", Expected: true},
					{Type: "contains", Target: "reveal_conditions", Expected: true},
				},
			},
			{
				TestID:      "export_handout",
				Name:        "Export Handout in Multiple Formats",
				Description: "Test handout export",
				Steps: []TestStep{
					{Action: "export_handout", Tool: "grimorio_export_handout", Params: map[string]any{"format": "pdf"}},
					{Action: "export_handout", Tool: "grimorio_export_handout", Params: map[string]any{"format": "text"}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "exported_file", Expected: true},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_random_tables",
		Name:    "Random Table Generation",
		Tests: []E2ETest{
			{
				TestID:      "generate_encounter_table",
				Name:        "Generate Encounter Table",
				Description: "Test encounter table generation",
				Steps: []TestStep{
					{Action: "generate_random_table", Tool: "grimorio_generate_random_tables", Params: map[string]any{"type": "encounter"}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "table", Expected: true},
					{Type: "contains", Target: "entries", Expected: "10+"},
				},
			},
			{
				TestID:      "generate_rumor_table",
				Name:        "Generate Rumor Table",
				Description: "Test rumor table generation",
				Steps: []TestStep{
					{Action: "generate_random_table", Tool: "grimorio_generate_random_tables", Params: map[string]any{"type": "rumor"}},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "table", Expected: true},
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
	runner := NewE2ETestRunner(ctx)

	suite := &E2ETestSuite{
		SuiteID: "e2e_canon_validation",
		Name:    "Canon Validation",
		Tests: []E2ETest{
			{
				TestID:      "validate_proposal",
				Name:        "Validate Content Proposal",
				Description: "Test proposal validation",
				Steps: []TestStep{
					{Action: "validate_proposal", Tool: "grimorio_validate_canon"},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "validation_result", Expected: true},
				},
			},
			{
				TestID:      "check_consistency",
				Name:        "Check Campaign Consistency",
				Description: "Test consistency checking",
				Steps: []TestStep{
					{Action: "check_consistency", Tool: "grimorio_check_consistency"},
				},
				Assertions: []TestAssertion{
					{Type: "exists", Target: "consistency_report", Expected: true},
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

// Helper test to verify test framework works
func TestE2ETestRunner_Basic(t *testing.T) {
	ctx := context.Background()
	runner := NewE2ETestRunner(ctx)

	test := &E2ETest{
		TestID:      "basic_test",
		Name:        "Basic Test",
		Description: "Verify test runner works",
		Steps:       []TestStep{},
		Assertions: []TestAssertion{
			{Type: "exists", Target: "test", Expected: true},
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
