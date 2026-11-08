package scripts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// powershellPath returns a usable PowerShell interpreter, if any. Windows ships
// powershell.exe (5.1) and pwsh (7+); Linux and macOS may have pwsh.
func powershellPath() string {
	for _, name := range []string{"pwsh", "powershell"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// installPs1Driver is the script executed against a temporary HOME. It loads
// install.ps1 without running MAIN, points the install paths at the temp
// directory and asserts the OpenCode V2 contract.
const installPs1Driver = `
$ErrorActionPreference = 'Stop'

$root = $env:GRIMORIO_TEST_ROOT
$target = $env:GRIMORIO_TEST_SCRIPT
$failures = 0

function Check {
    param([string]$Name, [bool]$Condition, [string]$Detail = '')
    if ($Condition) { Write-Output "OK $Name" }
    else { Write-Output "FAIL $Name $Detail"; $script:failures++ }
}

function Get-NoBom {
    param([string]$Path)
    $b = [System.IO.File]::ReadAllBytes($Path)
    if ($b.Length -lt 3) { return $true }
    return -not ($b[0] -eq 0xEF -and $b[1] -eq 0xBB -and $b[2] -eq 0xBF)
}

$env:USERPROFILE = Join-Path $root 'home'
$env:LOCALAPPDATA = Join-Path $root 'appdata'
New-Item -ItemType Directory -Path $env:USERPROFILE -Force | Out-Null
New-Item -ItemType Directory -Path $env:LOCALAPPDATA -Force | Out-Null

$raw = Get-Content $target -Raw
$mainIndex = $raw.IndexOf('# MAIN')
if ($mainIndex -lt 0) { throw 'MAIN block not found in install.ps1' }
$harness = Join-Path $root 'harness.ps1'
[System.IO.File]::WriteAllText($harness, $raw.Substring(0, $mainIndex), (New-Object System.Text.UTF8Encoding($false)))
. $harness

# On Unix the script's Windows-style paths would become literal filenames.
if ($IsLinux -or $IsMacOS) {
    $InstallDir = Join-Path $env:LOCALAPPDATA 'Grimorio'
    $OpenCodeDir = Join-Path $env:USERPROFILE '.config/opencode'
    $SkillsDir = Join-Path $OpenCodeDir 'skills'
    $AgentsDir = Join-Path $OpenCodeDir 'agents'
    $CommandsDir = Join-Path $OpenCodeDir 'commands'
}
else {
    $InstallDir = Join-Path $env:LOCALAPPDATA 'Grimorio'
    $OpenCodeDir = Join-Path $env:USERPROFILE '.config\opencode'
    $SkillsDir = Join-Path $OpenCodeDir 'skills'
    $AgentsDir = Join-Path $OpenCodeDir 'agents'
    $CommandsDir = Join-Path $OpenCodeDir 'commands'
}

# --- fresh config ---
Merge-OpenCodeConfig
$configPath = Join-Path $OpenCodeDir 'opencode.json'
$commandPath = Join-Path $CommandsDir 'grimorio.md'

Check 'config-created' (Test-Path $configPath)
Check 'config-no-bom' (Get-NoBom $configPath)
Check 'command-created' (Test-Path $commandPath)
Check 'command-no-bom' (Get-NoBom $commandPath)

$cfg = Get-Content $configPath -Raw | ConvertFrom-Json
Check 'v2-server-entry' ($null -ne $cfg.mcp.servers.grimorio)
Check 'no-legacy-mcp-entry' ($null -eq $cfg.mcp.grimorio)
Check 'disabled-false' ($cfg.mcp.servers.grimorio.disabled -eq $false)
Check 'environment-present' ($null -ne $cfg.mcp.servers.grimorio.environment)
Check 'command-is-array' ($cfg.mcp.servers.grimorio.command -is [array])

$cmdText = Get-Content $commandPath -Raw
Check 'command-selects-agent' ($cmdText.Contains('agent: grimorio-architect'))

# --- legacy migration on an existing config ---
$legacyDir = Join-Path $root 'legacy/.config/opencode'
New-Item -ItemType Directory -Path $legacyDir -Force | Out-Null
$legacy = @'
{
  "agent": { "keep-me": { "description": "x" } },
  "mcp": {
    "servers": { "codegraph": { "type": "local", "command": ["codegraph"] } },
    "grimorio": { "command": ["/old"], "type": "local", "enabled": true }
  },
  "command": { "grimorio": { "template": "old" }, "other": { "template": "keep" } },
  "deep": { "a": { "b": { "c": { "d": { "e": "kept" } } } } }
}
'@
[System.IO.File]::WriteAllText((Join-Path $legacyDir 'opencode.json'), $legacy, (New-Object System.Text.UTF8Encoding($false)))

$OpenCodeDir = $legacyDir
$CommandsDir = Join-Path $legacyDir 'commands'
Merge-OpenCodeConfig
$legacyCfg = Get-Content (Join-Path $legacyDir 'opencode.json') -Raw | ConvertFrom-Json
Check 'preserves-unrelated-agent' ($null -ne $legacyCfg.agent.'keep-me')
Check 'preserves-unrelated-server' ($null -ne $legacyCfg.mcp.servers.codegraph)
Check 'preserves-unrelated-command' ($null -ne $legacyCfg.command.other)
Check 'drops-legacy-mcp' ($null -eq $legacyCfg.mcp.grimorio)
Check 'drops-legacy-command' ($null -eq $legacyCfg.command.grimorio)
Check 'keeps-deep-nesting' ($legacyCfg.deep.a.b.c.d.e -eq 'kept')
Check 'writes-backup' ((Get-ChildItem $legacyDir -Filter 'opencode.json.backup.*').Count -ge 1)

# --- invalid JSON is never overwritten ---
$brokenDir = Join-Path $root 'broken/.config/opencode'
New-Item -ItemType Directory -Path $brokenDir -Force | Out-Null
$broken = '{ "mcp": { "grimorio": '
$brokenPath = Join-Path $brokenDir 'opencode.json'
[System.IO.File]::WriteAllText($brokenPath, $broken, (New-Object System.Text.UTF8Encoding($false)))
$OpenCodeDir = $brokenDir
$CommandsDir = Join-Path $brokenDir 'commands'
Merge-OpenCodeConfig
Check 'invalid-json-untouched' ((Get-Content $brokenPath -Raw) -eq $broken)

Write-Output "FAILURES=$failures"
`

func TestInstallPs1_ConfigMergeV2(t *testing.T) {
	ps := powershellPath()
	if ps == "" {
		t.Skip("PowerShell is not available on this machine")
	}

	root := t.TempDir()
	driverPath := filepath.Join(root, "driver.ps1")
	if err := os.WriteFile(driverPath, []byte(installPs1Driver), 0644); err != nil {
		t.Fatalf("writing driver: %v", err)
	}

	cmd := exec.Command(ps, "-NoLogo", "-NoProfile", "-NonInteractive", "-File", driverPath)
	cmd.Env = append(os.Environ(),
		"GRIMORIO_TEST_ROOT="+root,
		"GRIMORIO_TEST_SCRIPT="+installPs1Path(),
	)
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		t.Fatalf("PowerShell driver failed: %v\n%s", err, output)
	}
	if strings.Contains(output, "FAIL ") {
		t.Errorf("install.ps1 contract check failed:\n%s", output)
	}
	if !strings.Contains(output, "FAILURES=0") {
		t.Errorf("expected FAILURES=0 in driver output:\n%s", output)
	}
	for _, want := range []string{"OK config-no-bom", "OK command-no-bom", "OK v2-server-entry", "OK drops-legacy-mcp"} {
		if !strings.Contains(output, want) {
			t.Errorf("missing %q in driver output:\n%s", want, output)
		}
	}
}
