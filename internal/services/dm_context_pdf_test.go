package services

import (
	"os"
	"strings"
	"testing"
)

// TestExtractPDFText_WindowsGuard verifies that extractPDFText returns an
// explicit, actionable error when running on Windows. The runtime check is
// gated by an env var override so the test can exercise the Windows branch
// on any host (CI runs Linux, dev machines may run macOS).
//
// Set GRIMORIO_FAKE_WINDOWS=1 to force the Windows path; otherwise the test
// asserts the Linux / non-Windows branch is taken (which is what the CI
// environment will exercise).
func TestExtractPDFText_WindowsGuard(t *testing.T) {
	svc := &DMContextService{baseDir: t.TempDir()}

	if os.Getenv("GRIMORIO_FAKE_WINDOWS") == "1" {
		// Forced Windows path: must return an explicit error mentioning
		// poppler-utils AND WSL.
		_, err := svc.extractPDFText("/nonexistent/file.pdf")
		if err == nil {
			t.Fatal("expected error on Windows, got nil")
		}
		msg := err.Error()
		if !strings.Contains(msg, "poppler-utils") {
			t.Errorf("expected error to mention 'poppler-utils', got: %q", msg)
		}
		if !strings.Contains(msg, "WSL") {
			t.Errorf("expected error to mention 'WSL', got: %q", msg)
		}
		return
	}

	// Non-Windows (Linux/macOS) path: the function proceeds past the OS
	// guard. With no pdftotext installed, it returns a different error
	// (about the binary not being available) — but the error must NOT be
	// the Windows-only message.
	_, err := svc.extractPDFText("/nonexistent/file.pdf")
	if err == nil {
		t.Skip("pdftotext available; skipping non-Windows path check")
	}
	if strings.Contains(err.Error(), "WSL") {
		t.Errorf("non-Windows path leaked Windows-only error: %q", err.Error())
	}
}

// TestIsWindows_HonorsFakeOverride asserts the helper that powers the
// guard honours the test override and reads runtime.GOOS otherwise.
func TestIsWindows_HonorsFakeOverride(t *testing.T) {
	// Save & restore the package-level override.
	orig := fakeWindowsOverride
	defer func() { fakeWindowsOverride = orig }()

	fakeWindowsOverride = true
	if !isWindows() {
		t.Error("expected isWindows() = true when fakeWindowsOverride = true")
	}
	fakeWindowsOverride = false
	if isWindows() {
		t.Error("expected isWindows() = false when fakeWindowsOverride = false and runtime.GOOS != windows")
	}
}

