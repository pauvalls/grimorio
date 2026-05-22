package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func testInstallShPath() string {
	return filepath.Join(repoRoot(), "scripts", "test-install.sh")
}

// ---------------------------------------------------------------------------
// T014: E2E install test script
// ---------------------------------------------------------------------------

func TestE2E_Exists(t *testing.T) {
	if _, err := os.Stat(testInstallShPath()); os.IsNotExist(err) {
		t.Fatalf("test-install.sh does not exist at %s", testInstallShPath())
	}
}

func TestE2E_IsPOSIXCompliant(t *testing.T) {
	content, err := os.ReadFile(testInstallShPath())
	if err != nil {
		t.Fatalf("failed to read test-install.sh: %v", err)
	}
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		t.Fatal("test-install.sh is empty")
	}
	if !strings.HasPrefix(lines[0], "#!/bin/sh") {
		t.Errorf("test-install.sh must use #!/bin/sh, got: %s", lines[0])
	}
}

func TestE2E_PassesSyntaxCheck(t *testing.T) {
	cmd := exec.Command("sh", "-n", testInstallShPath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("test-install.sh failed sh -n syntax check: %v\nOutput: %s", err, out)
	}
}

func TestE2E_VerifiesInstallCompletion(t *testing.T) {
	content, err := os.ReadFile(testInstallShPath())
	if err != nil {
		t.Fatalf("failed to read test-install.sh: %v", err)
	}
	c := string(content)
	checks := []string{
		"grimorio --version",
		".config/opencode/plugins/grimorio",
		"opencode.json",
	}
	for _, check := range checks {
		if !strings.Contains(c, check) {
			t.Errorf("test-install.sh must verify: %s", check)
		}
	}
}

func TestE2E_VerifiesBinaryInPath(t *testing.T) {
	content, err := os.ReadFile(testInstallShPath())
	if err != nil {
		t.Fatalf("failed to read test-install.sh: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "PATH") && !strings.Contains(c, "which grimorio") && !strings.Contains(c, "command -v grimorio") {
		t.Error("test-install.sh must verify binary is in PATH")
	}
}

func TestE2E_VerifiesAgentsAndSkills(t *testing.T) {
	content, err := os.ReadFile(testInstallShPath())
	if err != nil {
		t.Fatalf("failed to read test-install.sh: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "agents") {
		t.Error("test-install.sh must verify agents directory")
	}
	if !strings.Contains(c, "skills") {
		t.Error("test-install.sh must verify skills directory")
	}
}

func TestE2E_HasCleanup(t *testing.T) {
	content, err := os.ReadFile(testInstallShPath())
	if err != nil {
		t.Fatalf("failed to read test-install.sh: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "rm -rf") && !strings.Contains(c, "cleanup") {
		t.Error("test-install.sh must have cleanup logic")
	}
}

func TestE2E_CanRunInDocker(t *testing.T) {
	content, err := os.ReadFile(testInstallShPath())
	if err != nil {
		t.Fatalf("failed to read test-install.sh: %v", err)
	}
	c := string(content)
	// Should be usable in Docker/CI without interactive prompts
	if strings.Contains(c, "read ") || strings.Contains(c, "read -p") {
		t.Error("test-install.sh should not have interactive prompts for CI/Docker")
	}
}
