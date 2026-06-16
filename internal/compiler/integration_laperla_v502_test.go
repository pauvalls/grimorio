package compiler

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// laPerlaCampaignPath is the on-disk path to the reference campaign used by
// the v5.0.2 integration test. Override with GRIMORIO_LAPERLA_PATH if your
// copy lives elsewhere. Gated on GRIMORIO_TEST_CAMPAIGNS=1 to keep CI fast.
const laPerlaCampaignPath = "/home/pau/campaigns/la-perla-de-umberlee"

// TestIntegrationLaPerlaV502 compiles the la-perla-de-umberlee campaign
// and asserts that the rendered HTML contains the v5.0.2 CSS fixes
// (column-span, table-layout, overflow-wrap) and that generateHTML
// returns no error with non-empty output.
//
// Skipped unless both:
//   - GRIMORIO_TEST_CAMPAIGNS=1 is set
//   - the campaign directory exists on disk
func TestIntegrationLaPerlaV502(t *testing.T) {
	if os.Getenv("GRIMORIO_TEST_CAMPAIGNS") != "1" {
		t.Skip("set GRIMORIO_TEST_CAMPAIGNS=1 to run integration test against la-perla-de-umberlee")
	}

	campaignPath := laPerlaCampaignPath
	if override := os.Getenv("GRIMORIO_LAPERLA_PATH"); override != "" {
		campaignPath = override
	}
	if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
		t.Skipf("campaign dir not found: %s", campaignPath)
	}

	// Copy the campaign to a temp dir so we don't pollute the reference
	// with a fresh campaign.html / campaign.pdf.
	tmpDir := t.TempDir()
	if err := copyDirForTest(t, campaignPath, tmpDir); err != nil {
		t.Fatalf("copy campaign to temp: %v", err)
	}

	c := New(tmpDir, "")
	parts, err := c.generateHTML("Integration Test")
	if err != nil {
		t.Fatalf("generateHTML error: %v", err)
	}
	if len(parts) == 0 {
		t.Fatal("generateHTML returned no html parts")
	}
	html := strings.Join(parts, "\n")
	if len(html) == 0 {
		t.Fatal("compiled HTML is empty")
	}

	// WU1 CSS fixes must be present in the compiled HTML.
	required := []string{
		"column-span: all",
		"table-layout: fixed",
		"overflow-wrap: break-word",
	}
	for _, want := range required {
		if !strings.Contains(html, want) {
			t.Errorf("compiled HTML missing %q (CSS regression)", want)
		}
	}
}

// copyDirForTest recursively copies a directory tree. We use io.Copy +
// filepath.Walk instead of shelling out so the test stays pure-Go.
func copyDirForTest(t *testing.T, src, dst string) error {
	t.Helper()
	return filepath.Walk(src, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		// Skip the previously compiled artifacts.
		base := filepath.Base(path)
		if base == "campaign.html" || base == "campaign.pdf" {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
