package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAssertExists(t *testing.T) {
	harness := NewTestHarness(t)

	// Create a test file
	testFile := filepath.Join(harness.BaseDir, "test-exists-file.txt")
	if err := os.WriteFile(testFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should pass for existing file
	passed, msg := assertExists(harness, TestAssertion{Type: "exists", Target: "test-exists-file.txt"})
	if !passed {
		t.Errorf("assertExists on existing file should pass, got: %s", msg)
	}

	// Should fail for non-existing file
	passed, msg = assertExists(harness, TestAssertion{Type: "exists", Target: "nonexistent.txt"})
	if passed {
		t.Error("assertExists on non-existing file should fail")
	}
	if msg == "" {
		t.Error("assertExists should return a message on failure")
	}
}

func TestAssertEquals(t *testing.T) {
	harness := NewTestHarness(t)

	// Create a test file with known content
	testFile := filepath.Join(harness.BaseDir, "test-equals-file.txt")
	if err := os.WriteFile(testFile, []byte("expected content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should pass for matching content
	passed, msg := assertEquals(harness, TestAssertion{
		Type:     "equals",
		Target:   "test-equals-file.txt",
		Expected: "expected content",
	})
	if !passed {
		t.Errorf("assertEquals on matching content should pass, got: %s", msg)
	}

	// Should fail for non-matching content
	passed, _ = assertEquals(harness, TestAssertion{
		Type:     "equals",
		Target:   "test-equals-file.txt",
		Expected: "wrong content",
	})
	if passed {
		t.Error("assertEquals on non-matching content should fail")
	}

	// Should fail for missing file
	passed, _ = assertEquals(harness, TestAssertion{
		Type:     "equals",
		Target:   "nonexistent.txt",
		Expected: "anything",
	})
	if passed {
		t.Error("assertEquals on missing file should fail")
	}
}

func TestAssertContains(t *testing.T) {
	harness := NewTestHarness(t)

	// Create a test file with known content
	testFile := filepath.Join(harness.BaseDir, "test-contains-file.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should pass for contained substring
	passed, msg := assertContains(harness, TestAssertion{
		Type:     "contains",
		Target:   "test-contains-file.txt",
		Expected: "world",
	})
	if !passed {
		t.Errorf("assertContains on contained substring should pass, got: %s", msg)
	}

	// Should fail for non-contained substring
	passed, _ = assertContains(harness, TestAssertion{
		Type:     "contains",
		Target:   "test-contains-file.txt",
		Expected: "foo",
	})
	if passed {
		t.Error("assertContains on non-contained substring should fail")
	}

	// Should fail for missing file
	passed, _ = assertContains(harness, TestAssertion{
		Type:     "contains",
		Target:   "nonexistent.txt",
		Expected: "anything",
	})
	if passed {
		t.Error("assertContains on missing file should fail")
	}
}

func TestAssertValidates(t *testing.T) {
	harness := NewTestHarness(t)

	// Placeholder validation should pass
	passed, msg := assertValidates(harness, TestAssertion{Type: "validates", Target: "something"})
	if !passed {
		t.Errorf("assertValidates placeholder should pass, got: %s", msg)
	}
}

func TestAssertFileSystem_UnsupportedType(t *testing.T) {
	harness := NewTestHarness(t)

	passed, msg := assertFileSystem(harness, TestAssertion{Type: "unsupported"})
	if passed {
		t.Error("assertFileSystem with unsupported type should fail")
	}
	if msg == "" {
		t.Error("assertFileSystem should return error message for unsupported type")
	}
}

func TestResolvePath(t *testing.T) {
	harness := NewTestHarness(t)

	// Relative path should be resolved under BaseDir
	resolved := resolvePath(harness, "relative/path")
	if resolved != filepath.Join(harness.BaseDir, "relative/path") {
		t.Errorf("resolvePath(relative) = %q, want %q", resolved, filepath.Join(harness.BaseDir, "relative/path"))
	}

	// Absolute path should remain unchanged
	resolved = resolvePath(harness, "/absolute/path")
	if resolved != "/absolute/path" {
		t.Errorf("resolvePath(absolute) = %q, want %q", resolved, "/absolute/path")
	}
}
