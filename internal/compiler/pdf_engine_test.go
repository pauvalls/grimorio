package compiler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestDetectPDFEngine_Priority tests that engine detection prefers Chromium over wkhtmltopdf.
func TestDetectPDFEngine_Priority(t *testing.T) {
	// We can't easily mock exec.LookPath, but we can verify the function returns
	// a non-empty engine name.
	engine := detectPDFEngine()
	if engine == "" {
		t.Error("detectPDFEngine() should return a non-empty engine name")
	}

	// Verify the detected engine is in the supported list
	found := false
	for _, e := range pdfEnginePriority {
		if e == engine {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("detectPDFEngine() returned unsupported engine: %s", engine)
	}
}

// TestIsPDFEngineAvailable tests the availability checker.
func TestIsPDFEngineAvailable(t *testing.T) {
	// This is a simple smoke test - we just verify it doesn't panic
	// and returns a bool value.
	_ = IsPDFEngineAvailable()
}

// TestSupportedEngines tests that we expose the engine list.
func TestSupportedEngines(t *testing.T) {
	engines := SupportedEngines()
	if len(engines) == 0 {
		t.Error("SupportedEngines() should return non-empty list")
	}

	// Should contain chromium and wkhtmltopdf
	hasChromium := false
	hasWkhtmltopdf := false
	for _, e := range engines {
		if e == "chromium" {
			hasChromium = true
		}
		if e == "wkhtmltopdf" {
			hasWkhtmltopdf = true
		}
	}
	if !hasChromium {
		t.Error("SupportedEngines() should include 'chromium'")
	}
	if !hasWkhtmltopdf {
		t.Error("SupportedEngines() should include 'wkhtmltopdf'")
	}
}

// TestIsChromiumEngine tests the Chromium engine detection.
func TestIsChromiumEngine(t *testing.T) {
	tests := []struct {
		engine string
		want   bool
	}{
		{"chromium", true},
		{"chrome", true},
		{"google-chrome", true},
		{"google-chrome-stable", true},
		{"chromium-browser", true},
		{"msedge", true},
		{"wkhtmltopdf", false},
		{"weasyprint", false},
		{"", false},
		{"firefox", false},
	}

	for _, tt := range tests {
		t.Run(tt.engine, func(t *testing.T) {
			got := isChromiumEngine(tt.engine)
			if got != tt.want {
				t.Errorf("isChromiumEngine(%q) = %v, want %v", tt.engine, got, tt.want)
			}
		})
	}
}

// TestNew_AutoDetection tests that New() auto-detects an engine when empty.
func TestNew_AutoDetection(t *testing.T) {
	c := New("/tmp/campaign", "")
	if c.PDFEngine == "" {
		t.Error("New() with empty engine should auto-detect a PDF engine")
	}
	if c.CompilerVersion != 2 {
		t.Errorf("New() CompilerVersion = %d, want 2", c.CompilerVersion)
	}
}

// TestNew_ExplicitEngine tests that New() respects an explicitly provided engine.
func TestNew_ExplicitEngine(t *testing.T) {
	c := New("/tmp/campaign", "custom-engine")
	if c.PDFEngine != "custom-engine" {
		t.Errorf("New() with explicit engine = %s, want custom-engine", c.PDFEngine)
	}
}

// TestHtmlToPDF_Chromium tests the Chromium PDF generation path.
// This test is skipped if Chromium is not available.
func TestHtmlToPDF_Chromium(t *testing.T) {
	chromiumBin := ""
	for _, bin := range []string{"chromium", "chrome", "google-chrome", "google-chrome-stable"} {
		if _, err := exec.LookPath(bin); err == nil {
			chromiumBin = bin
			break
		}
	}

	if chromiumBin == "" {
		t.Skip("No Chromium/Chrome browser found in PATH, skipping integration test")
	}

	// Verify Chromium can actually run in headless mode (CI environments may have it installed but not functional)
	testCtx, testCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer testCancel()
	cmd := exec.CommandContext(testCtx, chromiumBin, "--headless", "--disable-gpu", "--version")
	if err := cmd.Run(); err != nil {
		t.Skip("Chromium cannot run in headless mode (likely missing dependencies in CI), skipping integration test")
	}

	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")
	pdfPath := filepath.Join(tmpDir, "test.pdf")

	// Create a simple HTML file
	htmlContent := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Test</title></head>
<body><h1>Hello PDF</h1><p>This is a test.</p></body>
</html>`
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}

	c := New(tmpDir, chromiumBin)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.htmlToPDF(ctx, htmlPath, pdfPath)
	if err != nil {
		t.Fatalf("htmlToPDF with Chromium failed: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("PDF file not created at %s", pdfPath)
	}

	// Verify PDF has some content (not empty)
	info, err := os.Stat(pdfPath)
	if err != nil {
		t.Fatalf("Failed to stat PDF: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Generated PDF is empty")
	}

	// Verify it's a valid PDF (starts with %PDF)
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("Failed to read PDF: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Errorf("Generated file is not a valid PDF (does not start with %%PDF)")
	}
}

// TestHtmlToPDF_Wkhtmltopdf tests the wkhtmltopdf generation path.
// This test is skipped if wkhtmltopdf is not available.
func TestHtmlToPDF_Wkhtmltopdf(t *testing.T) {
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		t.Skip("wkhtmltopdf not installed, skipping integration test")
	}

	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")
	pdfPath := filepath.Join(tmpDir, "test.pdf")

	// Create a simple HTML file
	htmlContent := `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Test</title></head>
<body><h1>Hello PDF</h1><p>This is a test.</p></body>
</html>`
	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0644); err != nil {
		t.Fatal(err)
	}

	c := New(tmpDir, "wkhtmltopdf")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := c.htmlToPDF(ctx, htmlPath, pdfPath)
	if err != nil {
		t.Fatalf("htmlToPDF with wkhtmltopdf failed: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("PDF file not created at %s", pdfPath)
	}

	// Verify PDF has some content
	info, err := os.Stat(pdfPath)
	if err != nil {
		t.Fatalf("Failed to stat PDF: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Generated PDF is empty")
	}

	// Verify it's a valid PDF
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("Failed to read PDF: %v", err)
	}
	if !strings.HasPrefix(string(data), "%PDF") {
		t.Errorf("Generated file is not a valid PDF")
	}
}

// TestHtmlToPDF_ContextTimeout tests that context cancellation works.
func TestHtmlToPDF_ContextTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	htmlPath := filepath.Join(tmpDir, "test.html")
	pdfPath := filepath.Join(tmpDir, "test.pdf")

	if err := os.WriteFile(htmlPath, []byte("<html><body>Test</body></html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// Use "sleep" as a fake engine that will hang
	c := New(tmpDir, "sleep")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Give time for the context to expire
	time.Sleep(5 * time.Millisecond)

	err := c.htmlToPDF(ctx, htmlPath, pdfPath)
	if err == nil {
		t.Fatal("Expected error due to context timeout, got nil")
	}
}

// TestCompile_SuccessWithAnyEngine tests compilation works with auto-detected engine.
// This test is skipped if no PDF engine is available.
func TestCompile_SuccessWithAnyEngine(t *testing.T) {
	if !IsPDFEngineAvailable() {
		t.Skip("No PDF engine available in PATH, skipping integration test")
	}

	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)

	// Use auto-detected engine
	c := New(tmpDir, "")
	ctx := context.Background()

	pdfPath, err := c.Compile(ctx, "Test Campaign")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfPath); os.IsNotExist(err) {
		t.Errorf("PDF file not created at %s", pdfPath)
	}
}
