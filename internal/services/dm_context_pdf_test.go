package services

import (
	"os"
	"runtime"
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
//
// On Windows, isWindows() always returns true via runtime.GOOS, so the
// non-Windows branch is unreachable. The test is skipped there unless the
// env var is set, in which case it exercises the Windows path explicitly.
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

	// On Windows the runtime.GOOS check forces the Windows path, so the
	// non-Windows branch is never taken. Skip rather than fail.
	if runtime.GOOS == "windows" {
		t.Skip("non-Windows path test does not apply on Windows (set GRIMORIO_FAKE_WINDOWS=1 to force)")
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
//
// On Windows the second assertion is unreachable because runtime.GOOS is
// always "windows"; the test is gated accordingly.
func TestIsWindows_HonorsFakeOverride(t *testing.T) {
	// Save & restore the package-level override.
	orig := fakeWindowsOverride
	defer func() { fakeWindowsOverride = orig }()

	fakeWindowsOverride = true
	if !isWindows() {
		t.Error("expected isWindows() = true when fakeWindowsOverride = true")
	}

	// The "override is false, function should be false" assertion only
	// makes sense when we are NOT actually running on Windows. On Windows
	// runtime.GOOS makes the function true regardless of the override.
	if runtime.GOOS == "windows" {
		t.Skip("override-false assertion does not apply on Windows (runtime.GOOS is always windows)")
	}

	fakeWindowsOverride = false
	if isWindows() {
		t.Error("expected isWindows() = false when fakeWindowsOverride = false and runtime.GOOS != windows")
	}
}

