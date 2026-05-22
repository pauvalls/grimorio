package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(b))
}

func installShPath() string {
	return filepath.Join(repoRoot(), "install.sh")
}

func installPs1Path() string {
	return filepath.Join(repoRoot(), "install.ps1")
}

// ---------------------------------------------------------------------------
// T008: install.sh — core download and install
// ---------------------------------------------------------------------------

func TestInstallSh_Exists(t *testing.T) {
	if _, err := os.Stat(installShPath()); os.IsNotExist(err) {
		t.Fatalf("install.sh does not exist at %s", installShPath())
	}
}

func TestInstallSh_IsPOSIXCompliant(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}

	// Must use /bin/sh, not /bin/bash
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		t.Fatal("install.sh is empty")
	}
	if !strings.HasPrefix(lines[0], "#!/bin/sh") {
		t.Errorf("install.sh must use #!/bin/sh for POSIX compliance, got: %s", lines[0])
	}
}

func TestInstallSh_PassesSyntaxCheck(t *testing.T) {
	cmd := exec.Command("sh", "-n", installShPath())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.sh failed sh -n syntax check: %v\nOutput: %s", err, out)
	}
}

func TestInstallSh_HasDetectPlatform(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "detect_platform()") {
		t.Error("install.sh must contain detect_platform() function")
	}
}

func TestInstallSh_HasDownloadRelease(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "download_release()") {
		t.Error("install.sh must contain download_release() function")
	}
}

func TestInstallSh_HasVerifyChecksum(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "verify_checksum()") {
		t.Error("install.sh must contain verify_checksum() function")
	}
}

func TestInstallSh_HasExtractArchive(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "extract_archive()") {
		t.Error("install.sh must contain extract_archive() function")
	}
}

func TestInstallSh_HasInstallBinary(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "install_binary()") {
		t.Error("install.sh must contain install_binary() function")
	}
}

func TestInstallSh_HasWkhtmltopdfCheck(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "check_wkhtmltopdf()") {
		t.Error("install.sh must contain check_wkhtmltopdf() function")
	}
}

func TestInstallSh_NoBashArray(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	// Bash arrays are a common bashism; flag them as warnings
	// But we allow them in comments or strings
	lines := strings.Split(c, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, "=(") && !strings.Contains(trimmed, "$") {
			t.Errorf("line %d may contain bash array syntax: %s", i+1, trimmed)
		}
	}
}

func TestInstallSh_NoJqDependency(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	// Allow jq in comments or error messages, but not as a command dependency
	lines := strings.Split(c, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.Contains(trimmed, "jq ") && !strings.Contains(trimmed, "not found") {
			t.Errorf("line %d uses jq command — install.sh must not depend on jq: %s", i+1, trimmed)
		}
	}
}

func TestInstallSh_NoGitClone(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	if strings.Contains(c, "git clone") {
		t.Error("install.sh must not use git clone — pure binary download only")
	}
}

func TestInstallSh_NoGoBuild(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	if strings.Contains(c, "go build") {
		t.Error("install.sh must not use go build — assumes precompiled binary")
	}
}

func TestInstallSh_UsesCorrectInstallDir(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "${HOME}/.grimorio") && !strings.Contains(c, "$HOME/.grimorio") {
		t.Error("install.sh must use ~/.grimorio as install directory")
	}
}

func TestInstallSh_SymlinkOrCopyToLocalBin(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "${HOME}/.local/bin") && !strings.Contains(c, "$HOME/.local/bin") {
		t.Error("install.sh must reference ~/.local/bin for binary installation")
	}
}

// Test detect_platform logic by sourcing it in a controlled subshell
func TestDetectPlatform_Linux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping on non-Linux host")
	}

	// Extract just the detect_platform function and dependencies
	script := `
uname() {
  if [ "$1" = "-s" ]; then echo "Linux"; elif [ "$1" = "-m" ]; then echo "x86_64"; fi
}
` + extractFunction(installShPath(), "detect_platform") + `
detect_platform
echo "OS=$OS ARCH=$ARCH"
`

	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect_platform failed: %v\nOutput: %s", err, out)
	}
	output := strings.TrimSpace(string(out))
	if output != "OS=linux ARCH=amd64" {
		t.Errorf("expected OS=linux ARCH=amd64, got: %s", output)
	}
}

func TestDetectPlatform_MacOS(t *testing.T) {
	script := `
uname() {
  if [ "$1" = "-s" ]; then echo "Darwin"; elif [ "$1" = "-m" ]; then echo "arm64"; fi
}
` + extractFunction(installShPath(), "detect_platform") + `
detect_platform
echo "OS=$OS ARCH=$ARCH"
`

	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("detect_platform failed: %v\nOutput: %s", err, out)
	}
	output := strings.TrimSpace(string(out))
	if output != "OS=darwin ARCH=arm64" {
		t.Errorf("expected OS=darwin ARCH=arm64, got: %s", output)
	}
}

// ---------------------------------------------------------------------------
// T009: install.sh — plugin setup and config merge
// ---------------------------------------------------------------------------

func TestInstallSh_HasSetupPlugins(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "setup_plugins()") {
		t.Error("install.sh must contain setup_plugins() function")
	}
}

func TestInstallSh_HasCreateMCPJson(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "create_mcp_json()") {
		t.Error("install.sh must contain create_mcp_json() function")
	}
}

func TestInstallSh_HasMergeOpencodeConfig(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "merge_opencode_config()") {
		t.Error("install.sh must contain merge_opencode_config() function")
	}
}

func TestInstallSh_HasPathUpdate(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "update_path()") {
		t.Error("install.sh must contain update_path() function")
	}
}

func TestInstallSh_HasUpdateFlag(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), "--update") {
		t.Error("install.sh must support --update flag")
	}
}

func TestInstallSh_CopiesAgentsAndSkills(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "agents") {
		t.Error("install.sh must reference agents directory")
	}
	if !strings.Contains(c, "skills") {
		t.Error("install.sh must reference skills directory")
	}
}

func TestInstallSh_PluginDirIsOpencode(t *testing.T) {
	content, err := os.ReadFile(installShPath())
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}
	if !strings.Contains(string(content), ".config/opencode/plugins/grimorio") {
		t.Error("install.sh must reference ~/.config/opencode/plugins/grimorio as plugin directory")
	}
}

// ---------------------------------------------------------------------------
// T010: install.ps1 — Windows installer
// ---------------------------------------------------------------------------

func TestInstallPs1_Exists(t *testing.T) {
	if _, err := os.Stat(installPs1Path()); os.IsNotExist(err) {
		t.Fatalf("install.ps1 does not exist at %s", installPs1Path())
	}
}

func TestInstallPs1_HasArchitectureDetection(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	c := string(content)
	if !strings.Contains(c, "PROCESSOR_ARCHITECTURE") {
		t.Error("install.ps1 must detect architecture via $env:PROCESSOR_ARCHITECTURE")
	}
}

func TestInstallPs1_HasDownloadFunction(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "Download-Release") && !strings.Contains(string(content), "Invoke-WebRequest") {
		t.Error("install.ps1 must have download functionality")
	}
}

func TestInstallPs1_HasExpandArchive(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "Expand-Archive") {
		t.Error("install.ps1 must use Expand-Archive for extraction")
	}
}

func TestInstallPs1_HasWkhtmltopdfCheck(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "wkhtmltopdf") {
		t.Error("install.ps1 must check for wkhtmltopdf")
	}
}

func TestInstallPs1_HasUpdateSwitch(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "-Update") {
		t.Error("install.ps1 must support -Update switch")
	}
}

func TestInstallPs1_HasExecutionPolicyHandling(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "ExecutionPolicy") {
		t.Error("install.ps1 must handle execution policy")
	}
}

func TestInstallPs1_HasPathUpdate(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "Path") && !strings.Contains(string(content), "Environment") {
		t.Error("install.ps1 must update PATH")
	}
}

func TestInstallPs1_UsesLocalAppData(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "LOCALAPPDATA") {
		t.Error("install.ps1 must use $env:LOCALAPPDATA for installation directory")
	}
}

func TestInstallPs1_HasPluginDir(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), ".config\\opencode\\plugins\\grimorio") && !strings.Contains(string(content), ".config/opencode/plugins/grimorio") {
		t.Error("install.ps1 must reference opencode plugin directory")
	}
}

func TestInstallPs1_HasMCPJson(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), ".mcp.json") {
		t.Error("install.ps1 must create .mcp.json")
	}
}

func TestInstallPs1_HasConvertFromJson(t *testing.T) {
	content, err := os.ReadFile(installPs1Path())
	if err != nil {
		t.Fatalf("failed to read install.ps1: %v", err)
	}
	if !strings.Contains(string(content), "ConvertFrom-Json") || !strings.Contains(string(content), "ConvertTo-Json") {
		t.Error("install.ps1 must use ConvertFrom-Json / ConvertTo-Json for opencode.json")
	}
}

// ---------------------------------------------------------------------------
// Triangulation: Behavioral tests for shell functions
// ---------------------------------------------------------------------------

func TestArchiveName_LinuxAMD64(t *testing.T) {
	script := `
` + extractFunction(installShPath(), "archive_name") + `
OS=linux
ARCH=amd64
archive_name
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("archive_name failed: %v\nOutput: %s", err, out)
	}
	output := strings.TrimSpace(string(out))
	if output != "grimorio_Linux_x86_64" {
		t.Errorf("expected grimorio_Linux_x86_64, got: %s", output)
	}
}

func TestArchiveName_DarwinARM64(t *testing.T) {
	script := `
` + extractFunction(installShPath(), "archive_name") + `
OS=darwin
ARCH=arm64
archive_name
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("archive_name failed: %v\nOutput: %s", err, out)
	}
	output := strings.TrimSpace(string(out))
	if output != "grimorio_Darwin_arm64" {
		t.Errorf("expected grimorio_Darwin_arm64, got: %s", output)
	}
}

func TestArchiveExt_Linux(t *testing.T) {
	script := `
` + extractFunction(installShPath(), "archive_ext") + `
OS=linux
archive_ext
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("archive_ext failed: %v\nOutput: %s", err, out)
	}
	output := strings.TrimSpace(string(out))
	if output != "tar.gz" {
		t.Errorf("expected tar.gz for Linux, got: %s", output)
	}
}

func TestArchiveExt_Darwin(t *testing.T) {
	script := `
` + extractFunction(installShPath(), "archive_ext") + `
OS=darwin
archive_ext
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("archive_ext failed: %v\nOutput: %s", err, out)
	}
	output := strings.TrimSpace(string(out))
	if output != "tar.gz" {
		t.Errorf("expected tar.gz for Darwin, got: %s", output)
	}
}

func TestCreateMCPJson(t *testing.T) {
	tmpDir := t.TempDir()
	script := `
` + extractFunction(installShPath(), "create_mcp_json") + `
OS=linux
plugin_dir="` + tmpDir + `"
create_mcp_json "$plugin_dir"
cat "$plugin_dir/.mcp.json"
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create_mcp_json failed: %v\nOutput: %s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "grimorio") {
		t.Error(".mcp.json must contain grimorio key")
	}
	if !strings.Contains(output, "command") {
		t.Error(".mcp.json must contain command field")
	}
}

func TestMergeOpencodeConfig_CreateNew(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")
	os.MkdirAll(configDir, 0755)

	script := `
` + extractFunction(installShPath(), "merge_opencode_config") + `
` + extractFunction(installShPath(), "log") + `
` + extractFunction(installShPath(), "warn") + `
` + extractFunction(installShPath(), "success") + `
` + extractFunction(installShPath(), "command_exists") + `
` + extractFunction(installShPath(), "create_mcp_json") + `
HOME="` + tmpDir + `"
OPENCODE_PLUGIN_DIR="` + tmpDir + `/.config/opencode/plugins/grimorio"
merge_opencode_config
cat "$HOME/.config/opencode/opencode.json"
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("merge_opencode_config failed: %v\nOutput: %s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "grimorio") {
		t.Error("opencode.json must contain grimorio config")
	}
	if !strings.Contains(output, "mcp") {
		t.Error("opencode.json must contain mcp section")
	}
}

func TestMergeOpencodeConfig_PreservesExisting(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "opencode")
	os.MkdirAll(configDir, 0755)

	existingConfig := `{"custom_key": "custom_value", "mcp": {"other": {"enabled": true}}}`
	configFile := filepath.Join(configDir, "opencode.json")
	os.WriteFile(configFile, []byte(existingConfig), 0644)

	script := `
` + extractFunction(installShPath(), "merge_opencode_config") + `
` + extractFunction(installShPath(), "log") + `
` + extractFunction(installShPath(), "warn") + `
` + extractFunction(installShPath(), "success") + `
` + extractFunction(installShPath(), "command_exists") + `
` + extractFunction(installShPath(), "create_mcp_json") + `
HOME="` + tmpDir + `"
OPENCODE_PLUGIN_DIR="` + tmpDir + `/.config/opencode/plugins/grimorio"
merge_opencode_config
cat "$HOME/.config/opencode/opencode.json"
`
	cmd := exec.Command("sh", "-c", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("merge_opencode_config failed: %v\nOutput: %s", err, out)
	}
	output := string(out)
	if !strings.Contains(output, "custom_key") {
		t.Error("opencode.json must preserve existing custom_key")
	}
	if !strings.Contains(output, "grimorio") {
		t.Error("opencode.json must contain grimorio config after merge")
	}
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// extractFunction pulls a shell function definition from a file by name.
// It returns everything from "funcname() {" to the matching closing brace.
func extractFunction(path, name string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	start := -1
	for i, line := range lines {
		if strings.Contains(line, name+"() {") || strings.Contains(line, name+" () {") {
			start = i
			break
		}
	}
	if start == -1 {
		return ""
	}

	// Track brace depth to find the matching closing brace
	depth := 0
	var result []string
	for i := start; i < len(lines); i++ {
		line := lines[i]
		// Count braces (rough heuristic)
		for _, ch := range line {
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
			}
		}
		result = append(result, line)
		if depth == 0 && i > start {
			break
		}
	}
	return strings.Join(result, "\n")
}
