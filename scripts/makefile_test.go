package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func makefilePath() string {
	return filepath.Join(repoRoot(), "Makefile")
}

// ---------------------------------------------------------------------------
// T011: Makefile tests
// ---------------------------------------------------------------------------

func TestMakefile_Exists(t *testing.T) {
	if _, err := os.Stat(makefilePath()); os.IsNotExist(err) {
		t.Fatalf("Makefile does not exist at %s", makefilePath())
	}
}

func TestMakefile_InstallDelegatesToInstallSh(t *testing.T) {
	content, err := os.ReadFile(makefilePath())
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "install.sh") {
		t.Error("Makefile install target must reference install.sh")
	}
}

func TestMakefile_UpdateDelegatesToGrimorioUpdate(t *testing.T) {
	content, err := os.ReadFile(makefilePath())
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "grimorio update") {
		t.Error("Makefile update target must reference 'grimorio update'")
	}
}

func TestMakefile_BuildStillExists(t *testing.T) {
	content, err := os.ReadFile(makefilePath())
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	if !strings.Contains(string(content), "build:") {
		t.Error("Makefile must still have build target")
	}
}

func TestMakefile_PreservesExistingTargets(t *testing.T) {
	content, err := os.ReadFile(makefilePath())
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	c := string(content)
	required := []string{"test:", "lint:", "coverage:", "bench:", "docker:", "clean:", "release:", "changelog:"}
	for _, target := range required {
		if !strings.Contains(c, target) {
			t.Errorf("Makefile must preserve %s target", target)
		}
	}
}

func TestMakefile_InstallHasSourceFallback(t *testing.T) {
	content, err := os.ReadFile(makefilePath())
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	c := string(content)
	// The install target should have a fallback for developers (go build)
	// We check for the presence of both install.sh and go build in the install target area
	installStart := strings.Index(c, "install:")
	if installStart == -1 {
		t.Fatal("Makefile missing install target")
	}
	// Find the next target or end of file to scope the install target
	nextTarget := strings.Index(c[installStart+1:], "\n\n")
	installBlock := c[installStart:]
	if nextTarget != -1 {
		installBlock = c[installStart : installStart+1+nextTarget]
	}
	if !strings.Contains(installBlock, "install.sh") {
		t.Error("Makefile install target must reference install.sh")
	}
	if !strings.Contains(installBlock, "go build") {
		t.Error("Makefile install target must have fallback go build for developers")
	}
}

func TestMakefile_InstallDryRun(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}
	cmd := exec.Command("make", "-n", "install")
	cmd.Dir = repoRoot()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("make -n install failed: %v\nOutput: %s", err, out)
	}
	output := string(out)
	// Should reference install.sh or show the fallback build path
	if !strings.Contains(output, "install.sh") && !strings.Contains(output, "go build") {
		t.Errorf("make -n install should reference install.sh or go build, got:\n%s", output)
	}
}
