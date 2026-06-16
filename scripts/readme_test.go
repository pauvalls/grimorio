package scripts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readmePath() string {
	return filepath.Join(repoRoot(), "README.md")
}

// ---------------------------------------------------------------------------
// T015: README tests
// ---------------------------------------------------------------------------

func TestREADME_HasLinuxMacOSInstall(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "curl -sSL") || !strings.Contains(c, "install.sh") {
		t.Error("README must contain Linux/macOS curl | sh install command")
	}
}

func TestREADME_HasWindowsInstall(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "irm") || !strings.Contains(c, "install.ps1") || !strings.Contains(c, "iex") {
		t.Error("README must contain Windows irm | iex install command")
	}
}

func TestREADME_HasDeveloperSection(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "git clone") || !strings.Contains(c, "make install") {
		t.Error("README must contain developer git clone + make install instructions")
	}
}

func TestREADME_HasUpdateSection(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "grimorio update") {
		t.Error("README must mention grimorio update")
	}
	if !strings.Contains(c, "--update") {
		t.Error("README must mention install.sh --update")
	}
}

func TestREADME_HasTroubleshooting(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "Troubleshooting") && !strings.Contains(c, "Solución de Problemas") {
		t.Error("README must have Troubleshooting section")
	}
}

func TestREADME_HasRequirementsWkhtmltopdf(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "wkhtmltopdf") {
		t.Error("README must mention wkhtmltopdf in requirements")
	}
}

func TestREADME_HasSpanishAndEnglish(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "## 🇬🇧 English") {
		t.Error("README must have English section")
	}
	if !strings.Contains(c, "## 🇪🇸 Español") {
		t.Error("README must have Spanish section")
	}
}

func TestREADME_InstallCommandsAreCopyPasteReady(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	// Linux/macOS command should be in a code block and be copy-paste ready
	if !strings.Contains(c, "curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh") {
		t.Error("README must have copy-paste ready Linux/macOS install command")
	}
	// Windows command should be in a code block
	if !strings.Contains(c, "irm https://raw.githubusercontent.com/pauvalls/grimorio/main/install.ps1 | iex") {
		t.Error("README must have copy-paste ready Windows install command")
	}
}

// TestREADME_DropsExperimentalWindowsLabel ensures we no longer mark
// Windows as experimental now that the pdftotext path is gated and CI
// validates Windows builds.
func TestREADME_DropsExperimentalWindowsLabel(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)
	if strings.Contains(c, "Windows support is experimental") {
		t.Error("README must not call Windows support 'experimental' (English section)")
	}
	if strings.Contains(c, "Soporte de Windows es experimental") {
		t.Error("README must not call Windows support 'experimental' (Spanish section)")
	}
}

// TestREADME_DocLinksResolve walks every relative docs/ link in README.md
// and asserts the file exists. Stubs were added for mcp-tools.md and
// architecture.md so the table-of-contents rows resolve cleanly.
func TestREADME_DocLinksResolve(t *testing.T) {
	content, err := os.ReadFile(readmePath())
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	c := string(content)

	// Extract every relative docs/* and examples/* link target.
	var targets []string
	for _, line := range strings.Split(c, "\n") {
		for _, needle := range []string{"](docs/", "](examples/"} {
			idx := 0
			for {
				i := strings.Index(line[idx:], needle)
				if i < 0 {
					break
				}
				// needle is "](" + path-prefix; the path begins right after.
				pathStart := idx + i + len("](")
				rest := line[pathStart:]
				j := strings.Index(rest, ")")
				if j > 0 {
					targets = append(targets, rest[:j])
				}
				idx = pathStart
			}
		}
	}

	if len(targets) == 0 {
		t.Fatal("no docs/ or examples/ links found in README.md — did the parser break?")
	}

	root := repoRoot()
	for _, path := range targets {
		full := filepath.Join(root, path)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("README link target does not exist: %s", path)
		}
	}
}
