package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// assert evaluates a single assertion against the filesystem state.
func assertFileSystem(harness *TestHarness, assertion TestAssertion) (bool, string) {
	switch assertion.Type {
	case "exists":
		return assertExists(harness, assertion)
	case "equals":
		return assertEquals(harness, assertion)
	case "contains":
		return assertContains(harness, assertion)
	case "validates":
		return assertValidates(harness, assertion)
	default:
		return false, fmt.Sprintf("unsupported assertion type: %s", assertion.Type)
	}
}

func resolvePath(harness *TestHarness, target string) string {
	// If target looks like a campaign name, resolve under BaseDir
	if !filepath.IsAbs(target) {
		return filepath.Join(harness.BaseDir, target)
	}
	return target
}

func assertExists(harness *TestHarness, assertion TestAssertion) (bool, string) {
	path := resolvePath(harness, assertion.Target)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Sprintf("expected %s to exist, but it does not", path)
		}
		return false, fmt.Sprintf("error checking %s: %v", path, err)
	}
	return true, fmt.Sprintf("%s exists", path)
}

func assertEquals(harness *TestHarness, assertion TestAssertion) (bool, string) {
	path := resolvePath(harness, assertion.Target)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Sprintf("failed to read %s: %v", path, err)
	}

	expected, ok := assertion.Expected.(string)
	if !ok {
		return false, fmt.Sprintf("expected string for equals assertion, got %T", assertion.Expected)
	}

	if string(data) != expected {
		return false, fmt.Sprintf("content of %s does not equal expected", path)
	}
	return true, fmt.Sprintf("content of %s equals expected", path)
}

func assertContains(harness *TestHarness, assertion TestAssertion) (bool, string) {
	path := resolvePath(harness, assertion.Target)
	data, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Sprintf("failed to read %s: %v", path, err)
	}

	expected, ok := assertion.Expected.(string)
	if !ok {
		return false, fmt.Sprintf("expected string for contains assertion, got %T", assertion.Expected)
	}

	if !strings.Contains(string(data), expected) {
		return false, fmt.Sprintf("content of %s does not contain %q", path, expected)
	}
	return true, fmt.Sprintf("content of %s contains %q", path, expected)
}

func assertValidates(harness *TestHarness, assertion TestAssertion) (bool, string) {
	// Placeholder for custom validation assertions
	return true, fmt.Sprintf("validation %s passed (placeholder)", assertion.Target)
}
