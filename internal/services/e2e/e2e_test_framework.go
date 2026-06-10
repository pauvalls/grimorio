package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// E2ETestSuite represents a complete E2E test suite.
type E2ETestSuite struct {
	SuiteID   string     `json:"suite_id"`
	Name      string     `json:"name"`
	Tests     []E2ETest  `json:"tests"`
	Fixtures  []TestFixture `json:"fixtures"`
}

// E2ETest represents a single E2E test.
type E2ETest struct {
	TestID          string        `json:"test_id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Steps           []TestStep    `json:"steps"`
	Assertions      []TestAssertion `json:"assertions"`
	FixturesRequired []string     `json:"fixtures_required"`
}

// TestStep represents a step in an E2E test.
type TestStep struct {
	Action         string         `json:"action"`
	Tool           string         `json:"tool"` // MCP tool name
	Params         map[string]any `json:"params"`
	ExpectedStatus string         `json:"expected_status"`
}

// TestAssertion represents an assertion in an E2E test.
type TestAssertion struct {
	Type     string `json:"type"` // exists, equals, contains, validates, references
	Target   string `json:"target"`
	Expected any    `json:"expected"`
}

// TestFixture represents a test fixture.
type TestFixture struct {
	FixtureID       string `json:"fixture_id"`
	Name            string `json:"name"`
	Data            any    `json:"data"`
	CleanupRequired bool   `json:"cleanup_required"`
}

// TestResult represents the result of a test execution.
type TestResult struct {
	TestID     string        `json:"test_id"`
	Name       string        `json:"name"`
	Status     string        `json:"status"` // pass, fail, skip
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
	Assertions []AssertionResult `json:"assertions"`
}

// AssertionResult represents the result of an assertion.
type AssertionResult struct {
	Type     string `json:"type"`
	Passed   bool   `json:"passed"`
	Message  string `json:"message"`
}

// TestResults represents the results of a test suite execution.
type TestResults struct {
	SuiteID      string        `json:"suite_id"`
	Name         string        `json:"name"`
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	Duration     time.Duration `json:"duration"`
	Results      []TestResult  `json:"results"`
}

// E2ETestRunner runs E2E test suites.
type E2ETestRunner struct {
	ctx     context.Context
	harness *TestHarness
}

// NewE2ETestRunner creates a new E2ETestRunner.
func NewE2ETestRunner(ctx context.Context) *E2ETestRunner {
	return &E2ETestRunner{ctx: ctx}
}

// NewE2ETestRunnerWithHarness creates a runner with a specific harness.
func NewE2ETestRunnerWithHarness(ctx context.Context, harness *TestHarness) *E2ETestRunner {
	return &E2ETestRunner{ctx: ctx, harness: harness}
}

// RunSuite executes all tests in a suite.
func (r *E2ETestRunner) RunSuite(suite *E2ETestSuite) (*TestResults, error) {
	results := &TestResults{
		SuiteID:     suite.SuiteID,
		Name:        suite.Name,
		TotalTests:  len(suite.Tests),
		Results:     []TestResult{},
	}

	startTime := time.Now()

	for _, test := range suite.Tests {
		result := r.RunTest(&test, suite.Fixtures)
		results.Results = append(results.Results, result)
		
		if result.Status == "pass" {
			results.PassedTests++
		} else {
			results.FailedTests++
		}
	}

	results.Duration = time.Since(startTime)
	return results, nil
}

// RunTest executes a single test.
func (r *E2ETestRunner) RunTest(test *E2ETest, fixtures []TestFixture) TestResult {
	result := TestResult{
		TestID:     test.TestID,
		Name:       test.Name,
		Status:     "pass",
		Assertions: []AssertionResult{},
	}

	startTime := time.Now()

	// Execute steps
	for _, step := range test.Steps {
		if r.harness == nil {
			result.Status = "fail"
			result.Error = fmt.Sprintf("no harness available for step %s", step.Action)
			result.Duration = time.Since(startTime)
			return result
		}
		if err := executeStep(r.ctx, r.harness, step); err != nil {
			result.Status = "fail"
			result.Error = fmt.Sprintf("step %s failed: %v", step.Action, err)
			result.Duration = time.Since(startTime)
			return result
		}
	}

	// Run assertions
	for _, assertion := range test.Assertions {
		passed, message := assertFileSystem(r.harness, assertion)
		assertionResult := AssertionResult{
			Type:    assertion.Type,
			Passed:  passed,
			Message: message,
		}
		result.Assertions = append(result.Assertions, assertionResult)

		if !assertionResult.Passed {
			result.Status = "fail"
			result.Error = fmt.Sprintf("Assertion failed: %s - %s", assertion.Type, message)
		}
	}

	result.Duration = time.Since(startTime)
	return result
}

// GenerateTestFixtures creates fixtures for a campaign type.
func (r *E2ETestRunner) GenerateTestFixtures(campaignType string) ([]TestFixture, error) {
	fixtures := []TestFixture{
		{
			FixtureID:       "campaign_" + campaignType,
			Name:            "Test Campaign",
			Data:            map[string]string{"type": campaignType},
			CleanupRequired: true,
		},
	}
	return fixtures, nil
}

// CleanupFixture cleans up a test fixture.
// It deletes campaign directories with the "test_" prefix under the harness base dir.
// Directories without the "test_" prefix are skipped for safety.
func (r *E2ETestRunner) CleanupFixture(fixture TestFixture) error {
	if !fixture.CleanupRequired {
		return nil
	}
	if r.harness == nil {
		return fmt.Errorf("no harness available for cleanup")
	}

	baseDir := r.harness.BaseDir
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base dir %s: %w", baseDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "test-") {
			continue // Safety: only delete test- prefixed directories
		}
		targetPath := filepath.Join(baseDir, name)
		if err := os.RemoveAll(targetPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", targetPath, err)
		}
	}
	return nil
}
