#Requires -Version 5.1
<#
.SYNOPSIS
    Grimorio Installer for Windows
.DESCRIPTION
    Downloads and installs Grimorio from GitHub Releases.
    Supports fresh install and incremental update.
.PARAMETER Update
    Run in update mode (preserves agents/skills, updates binary only).
.EXAMPLE
    .\install.ps1
    .\install.ps1 -Update
#>
[CmdletBinding()]
param(
    [switch]$Update
)

$ErrorActionPreference = 'Stop'

# ============================================================================
# CONFIGURATION
# ============================================================================
$RepoOwner = 'pauvalls'
$RepoName = 'grimorio'
$InstallDir = Join-Path $env:LOCALAPPDATA 'Grimorio'
$OpenCodeDir = Join-Path $env:USERPROFILE '.config\opencode'
$SkillsDir = Join-Path $OpenCodeDir 'skills'
$AgentsDir = Join-Path $OpenCodeDir 'agents'
$CommandsDir = Join-Path $OpenCodeDir 'commands'
$BinaryName = 'grimorio.exe'

# ============================================================================
# LOGGING
# ============================================================================
function Write-Log {
    param([string]$Message)
    Write-Host "[Grimorio] $Message"
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-ErrorAndExit {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

# ============================================================================
# PLATFORM DETECTION
# ============================================================================
function Get-PlatformInfo {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        'AMD64' { $mappedArch = 'amd64' }
        'ARM64' { $mappedArch = 'arm64' }
        default { Write-ErrorAndExit "Unsupported architecture: $arch" }
    }

    return @{
        OS      = 'windows'
        Arch    = $mappedArch
        OSLabel = 'Windows'
    }
}

# ============================================================================
# ARCHIVE NAME (matches GoReleaser template)
# ============================================================================
function Get-ArchiveName {
    param([hashtable]$Platform)
    $archLabel = if ($Platform.Arch -eq 'amd64') { 'x86_64' } else { $Platform.Arch }
    return "grimorio_$($Platform.OSLabel)_${archLabel}.zip"
}

# ============================================================================
# FETCH LATEST RELEASE TAG
# ============================================================================
function Get-LatestTag {
    $apiUrl = "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -Headers @{
            'Accept' = 'application/vnd.github.v3+json'
        } -TimeoutSec 30
        return $response.tag_name
    }
    catch {
        Write-Warn "Could not fetch latest tag from GitHub API: $($_.Exception.Message)"
        return $null
    }
}

# ============================================================================
# DOWNLOAD RELEASE
# ============================================================================
function Download-Release {
    param(
        [string]$Tag,
        [hashtable]$Platform
    )

    $archiveName = Get-ArchiveName -Platform $Platform
    $baseUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Tag"
    $archiveUrl = "$baseUrl/$archiveName"
    $checksumUrl = "$baseUrl/checksums.txt"

    $tmpDir = Join-Path $env:TEMP "grimorio-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    Write-Log "Downloading $archiveName..."
    try {
        Invoke-WebRequest -Uri $archiveUrl -OutFile (Join-Path $tmpDir $archiveName) -TimeoutSec 120
    }
    catch {
        Remove-Item -Recurse -Force $tmpDir -ErrorAction SilentlyContinue
        Write-ErrorAndExit "Failed to download $archiveName`: $($_.Exception.Message)"
    }

    Write-Log "Downloading checksums.txt..."
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile (Join-Path $tmpDir 'checksums.txt') -TimeoutSec 30
    }
    catch {
        Write-Warn "Could not download checksums.txt"
    }

    return @{
        TempDir     = $tmpDir
        ArchivePath = Join-Path $tmpDir $archiveName
    }
}

# ============================================================================
# VERIFY CHECKSUM
# ============================================================================
function Test-Checksum {
    param(
        [string]$ArchivePath,
        [string]$ChecksumsPath
    )

    if (-not (Test-Path $ChecksumsPath)) {
        Write-Warn "No checksums.txt available, skipping verification"
        return
    }

    $archiveName = Split-Path $ArchivePath -Leaf
    $expectedLine = Get-Content $ChecksumsPath | Where-Object { $_ -match "^[a-f0-9]{64}\s+$([regex]::Escape($archiveName))$" }

    if (-not $expectedLine) {
        Write-Warn "No checksum found for $archiveName, skipping verification"
        return
    }

    $expectedHash = ($expectedLine -split '\s+')[0]
    $actualHash = (Get-FileHash -Algorithm SHA256 $ArchivePath).Hash.ToLower()

    if ($actualHash -ne $expectedHash) {
        Write-ErrorAndExit "Checksum mismatch for $archiveName!`nExpected: $expectedHash`nActual:   $actualHash"
    }

    Write-Log "Checksum verified: $archiveName"
}

# ============================================================================
# EXTRACT ARCHIVE
# ============================================================================
function Expand-GrimorioArchive {
    param(
        [string]$ArchivePath,
        [string]$TempDir
    )

    $extractDir = Join-Path $TempDir 'extracted'
    New-Item -ItemType Directory -Path $extractDir -Force | Out-Null

    try {
        Expand-Archive -Path $ArchivePath -DestinationPath $extractDir -Force
    }
    catch {
        Write-ErrorAndExit "Failed to extract archive: $($_.Exception.Message)"
    }

    # With wrap_in_directory, GoReleaser creates a subdirectory
    $innerDir = Get-ChildItem -Path $extractDir -Directory | Select-Object -First 1
    if ($innerDir -and (Test-Path (Join-Path $innerDir.FullName $BinaryName))) {
        return $innerDir.FullName
    }
    elseif (Test-Path (Join-Path $extractDir $BinaryName)) {
        return $extractDir
    }
    else {
        Write-ErrorAndExit "Archive extraction failed: grimorio.exe not found"
    }
}

# ============================================================================
# INSTALL BINARY
# ============================================================================
function Install-Binary {
    param([string]$SourceDir)

    $sourceBinary = Join-Path $SourceDir $BinaryName

    if (-not (Test-Path $sourceBinary)) {
        Write-ErrorAndExit "Binary not found in extracted archive: $sourceBinary"
    }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Remove old binary
    $targetBinary = Join-Path $InstallDir $BinaryName
    if (Test-Path $targetBinary) {
        Remove-Item -Force $targetBinary
    }

    Copy-Item -Path $sourceBinary -Destination $targetBinary -Force
    Write-Success "Binary installed: $targetBinary"
}

# ============================================================================
# OPENCODE V2 ASSETS — skills, agents and the /grimorio command
# ============================================================================
function Write-Utf8NoBom {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][AllowEmptyString()][string]$Content
    )

    $parent = Split-Path -Parent $Path
    if ($parent -and -not (Test-Path $parent)) {
        New-Item -ItemType Directory -Path $parent -Force | Out-Null
    }

    # Out-File -Encoding utf8 writes a BOM on Windows PowerShell 5.1, and a
    # BOM in front of the '---' breaks Markdown frontmatter detection.
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($Path, $Content, $utf8NoBom)
}

function Write-GrimorioCommand {
    $commandPath = Join-Path $CommandsDir 'grimorio.md'
    $content = @'
---
description: Generate a complete D&D 5e campaign or one-shot from an idea
agent: grimorio-architect
subagent: false
---

Start the Grimorio campaign generation workflow.

User request: $ARGUMENTS

Follow the mandatory intake (language first), then generate the campaign end-to-end and compile the PDF.
'@
    Write-Utf8NoBom -Path $commandPath -Content ($content + "`n")
    Write-Log "Command installed: $commandPath"
}

function Install-Plugins {
    param([string]$SourceDir)

    # OpenCode V2 loads skills from ~/.config/opencode/skills, agents from
    # ~/.config/opencode/agents and commands from ~/.config/opencode/commands.
    # Nothing is loaded from a plugin directory.
    foreach ($dir in @($SkillsDir, $AgentsDir, $CommandsDir)) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
        }
    }

    $agentsSource = Join-Path $SourceDir 'agents'
    if (Test-Path $agentsSource) {
        # Markdown only: the repository keeps Go test files next to the specs.
        Get-ChildItem -Path $agentsSource -File -Filter '*.md' | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination $AgentsDir -Force
        }
        Write-Log "Agents copied to $AgentsDir"
    }

    $skillsSource = Join-Path $SourceDir 'skills'
    if (Test-Path $skillsSource) {
        Get-ChildItem -Path $skillsSource -Directory | ForEach-Object {
            Copy-Item -Path $_.FullName -Destination $SkillsDir -Recurse -Force
        }
        Write-Log "Skills copied to $SkillsDir"
    }
}

# ============================================================================
# MERGE OPENCODE.JSON — OpenCode V2 native shape
# ============================================================================
function ConvertTo-HashtableDeep {
    param($InputObject)

    if ($null -eq $InputObject) { return $null }

    if ($InputObject -is [System.Collections.IDictionary]) {
        $result = @{}
        foreach ($key in @($InputObject.Keys)) {
            $result[$key] = ConvertTo-HashtableDeep $InputObject[$key]
        }
        return $result
    }

    if ($InputObject -is [System.Management.Automation.PSCustomObject]) {
        $result = @{}
        foreach ($prop in $InputObject.PSObject.Properties) {
            $result[$prop.Name] = ConvertTo-HashtableDeep $prop.Value
        }
        return $result
    }

    if ($InputObject -is [System.Collections.IEnumerable] -and -not ($InputObject -is [string])) {
        return @($InputObject | ForEach-Object { ConvertTo-HashtableDeep $_ })
    }

    return $InputObject
}

function Build-McpEnvironment {
    param([hashtable]$Existing)

    $environment = @{}
    if ($Existing) {
        foreach ($key in $Existing.Keys) {
            if ($null -ne $Existing[$key]) {
                $environment[$key] = [string]$Existing[$key]
            }
        }
    }

    if (-not $environment.ContainsKey('PATH')) {
        if ($env:PATH) {
            $environment['PATH'] = $env:PATH
        }
        else {
            $environment['PATH'] = "$env:SystemRoot\System32;$env:SystemRoot"
        }
    }

    if (-not $environment.ContainsKey('PIPER_MODEL_PATH')) {
        $piperDir = Join-Path $env:USERPROFILE '.local\share\piper'
        if (Test-Path $piperDir) {
            $voice = Get-ChildItem -Path $piperDir -Filter '*.onnx' -File | Select-Object -First 1
            if ($voice) {
                $environment['PIPER_MODEL_PATH'] = $voice.FullName
                if (Test-Path ($voice.FullName + '.json')) {
                    $environment['PIPER_CONFIG_PATH'] = $voice.FullName + '.json'
                }
            }
        }
    }

    $defaults = [ordered]@{
        'PIPER_PORT'           = '5000'
        'PIPER_HOST'           = '127.0.0.1'
        'PIPER_LENGTH_SCALE'   = '1.0'
        'PIPER_VOLUME'         = '0.8'
        'GRIMORIO_TTS_ENABLED' = 'true'
        'AUDIO_PLAYER'         = 'auto'
    }
    foreach ($key in $defaults.Keys) {
        if (-not $environment.ContainsKey($key)) {
            $environment[$key] = $defaults[$key]
        }
    }

    return $environment
}

function Merge-OpenCodeConfig {
    $configFile = Join-Path $OpenCodeDir 'opencode.json'
    $binaryPath = Join-Path $InstallDir $BinaryName

    if (-not (Test-Path $OpenCodeDir)) {
        New-Item -ItemType Directory -Path $OpenCodeDir -Force | Out-Null
    }

    Write-GrimorioCommand

    $data = @{}

    if (Test-Path $configFile) {
        $backupPath = "${configFile}.backup.$(Get-Date -Format 'yyyyMMddHHmmss')"
        Copy-Item -Path $configFile -Destination $backupPath -Force

        try {
            $raw = Get-Content $configFile -Raw
            if ($raw -and $raw.Trim().Length -gt 0) {
                $data = ConvertTo-HashtableDeep ($raw | ConvertFrom-Json)
            }
        }
        catch {
            # Never overwrite a config we cannot parse: hand-edited files and
            # JSONC are common. Leave it untouched and print the entry instead.
            Write-Warn "opencode.json is not valid JSON; leaving it untouched."
            Write-Warn "Add this entry under mcp.servers manually:"
            Write-Warn ('  "grimorio": { "type": "local", "command": ["' + $binaryPath + '"] }')
            return
        }
    }

    if (-not ($data -is [hashtable])) { $data = @{} }
    if (-not $data.ContainsKey('mcp') -or -not ($data['mcp'] -is [hashtable])) { $data['mcp'] = @{} }
    if (-not $data['mcp'].ContainsKey('servers') -or -not ($data['mcp']['servers'] -is [hashtable])) {
        $data['mcp']['servers'] = @{}
    }

    # Preserve environment values the user already configured.
    $previousEnv = @{}
    $existingServer = $data['mcp']['servers']['grimorio']
    if ($existingServer -is [hashtable]) {
        foreach ($sourceKey in @('environment', 'env')) {
            if ($existingServer.ContainsKey($sourceKey) -and $existingServer[$sourceKey] -is [hashtable]) {
                foreach ($key in $existingServer[$sourceKey].Keys) {
                    $previousEnv[$key] = $existingServer[$sourceKey][$key]
                }
            }
        }
    }

    $data['mcp']['servers']['grimorio'] = @{
        type        = 'local'
        command     = @($binaryPath)
        disabled    = $false
        environment = Build-McpEnvironment -Existing $previousEnv
    }

    # Migrate the V1 shapes so the V2 entry is unambiguous.
    $data['mcp'].Remove('grimorio')

    if ($data.ContainsKey('command') -and $data['command'] -is [hashtable]) {
        $data['command'].Remove('grimorio')
        if ($data['command'].Count -eq 0) {
            $data.Remove('command')
        }
    }

    $json = $data | ConvertTo-Json -Depth 20
    Write-Utf8NoBom -Path $configFile -Content ($json + "`n")

    Write-Success "Merged Grimorio config into opencode.json (OpenCode V2)"
    Write-Log "MCP server: mcp.servers.grimorio -> $binaryPath"
}

# ============================================================================
# UPDATE PATH
# ============================================================================
function Update-Path {
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -split ';' | Where-Object { $_ -eq $InstallDir }) {
        Write-Log "$InstallDir is already in PATH"
        return
    }

    $newPath = "$userPath;$InstallDir"
    [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
    Write-Success "Added $InstallDir to user PATH"
    Write-Log "You may need to restart your terminal for PATH changes to take effect."
}

# ============================================================================
# CHECK PDF ENGINE
# ============================================================================
function Test-PDFEngine {
    $engines = @('chrome', 'chromium', 'msedge', 'wkhtmltopdf')
    $found = $null

    foreach ($engine in $engines) {
        $cmd = Get-Command $engine -ErrorAction SilentlyContinue
        if ($cmd) {
            $found = $cmd.Source
            break
        }
    }

    if ($found) {
        Write-Log "PDF engine found: $found"
        return
    }

    Write-Warn "No PDF engine found. PDF generation will not work."
    Write-Warn "Install one of the following:"
    Write-Warn "  Chrome:      winget install Google.Chrome"
    Write-Warn "  Chrome:      choco install googlechrome"
    Write-Warn "  Edge:        Already installed on Windows 10/11"
    Write-Warn "  (legacy)     winget install wkhtmltopdf"
}

# ============================================================================
# VERIFY INSTALLATION
# ============================================================================
function Test-Installation {
    $binary = Join-Path $InstallDir $BinaryName
    if (-not (Test-Path $binary)) {
        Write-ErrorAndExit "Installation verification failed: binary not found at $binary"
    }

    try {
        $version = & $binary --version 2>$null | Select-Object -First 1
        Write-Log "Installed version: $version"
    }
    catch {
        Write-Warn "Could not verify version: $($_.Exception.Message)"
    }
}

# ============================================================================
# WRITE METADATA
# ============================================================================
function Write-Metadata {
    $metaDir = Join-Path $env:USERPROFILE '.config\grimorio'
    if (-not (Test-Path $metaDir)) {
        New-Item -ItemType Directory -Path $metaDir -Force | Out-Null
    }

    $platform = Get-PlatformInfo
    $version = 'unknown'
    try {
        $binary = Join-Path $InstallDir $BinaryName
        $version = & $binary --version 2>$null | Select-Object -First 1
    }
    catch { }

    $metadata = @{
        version     = $version
        installedAt = (Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ')
        installDir  = $InstallDir
        os          = $platform.OS
        arch        = $platform.Arch
    } | ConvertTo-Json -Depth 3

    $metaPath = Join-Path $metaDir 'install-meta.json'
    $metadata | Out-File -FilePath $metaPath -Encoding utf8 -Force
    Write-Log "Install metadata written to $metaPath"
}

# ============================================================================
# PRINT INSTRUCTIONS
# ============================================================================
function Show-Instructions {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  Grimorio Installation Complete!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "What was installed:"
    Write-Host "   Binary:        $InstallDir\$BinaryName"
    Write-Host "   OpenCode MCP:  mcp.servers.grimorio in $OpenCodeDir\opencode.json"
    Write-Host "   Skills:        $SkillsDir"
    Write-Host "   Agents:        $AgentsDir"
    Write-Host "   Command:       $CommandsDir\grimorio.md"
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "   1. Restart your terminal or run: refreshenv"
    Write-Host "   2. Verify the binary:  grimorio --version"
    Write-Host "   3. Verify OpenCode:    opencode mcp list"
    Write-Host "   4. Run /grimorio <your idea> inside OpenCode"
    Write-Host ""
    Write-Host "To update later: grimorio update all"
    Write-Host ""
}

# ============================================================================
# FULL INSTALL
# ============================================================================
function Install-Grimorio {
    Write-Log "Starting Grimorio installation..."

    $platform = Get-PlatformInfo
    Write-Log "Platform: $($platform.OS)/$($platform.Arch)"

    # Get latest tag
    $tag = Get-LatestTag
    if (-not $tag) {
        Write-Warn "Using 'latest' URL fallback"
        $tag = 'latest'
    }
    Write-Log "Release: $tag"

    # Download
    $downloadInfo = Download-Release -Tag $tag -Platform $platform

    # Verify
    $checksumsPath = Join-Path $downloadInfo.TempDir 'checksums.txt'
    Test-Checksum -ArchivePath $downloadInfo.ArchivePath -ChecksumsPath $checksumsPath

    # Extract
    $extractedDir = Expand-GrimorioArchive -ArchivePath $downloadInfo.ArchivePath -TempDir $downloadInfo.TempDir
    Write-Log "Extracted to: $extractedDir"

    # Install
    Install-Binary -SourceDir $extractedDir

    # Plugins
    Install-Plugins -SourceDir $extractedDir

    # Config
    Merge-OpenCodeConfig

    # PATH
    Update-Path

    # PDF engine
    Test-PDFEngine

    # Metadata
    Write-Metadata

    # Verify
    Test-Installation

    # Cleanup
    Remove-Item -Recurse -Force $downloadInfo.TempDir -ErrorAction SilentlyContinue

    Show-Instructions
    Write-Success "Installation complete!"
}

# ============================================================================
# UPDATE MODE
# ============================================================================
function Update-Grimorio {
    Write-Log "Starting Grimorio update..."

    $metaPath = Join-Path $env:USERPROFILE '.config\grimorio\install-meta.json'
    $binary = Join-Path $InstallDir $BinaryName

    if (-not (Test-Path $metaPath) -or -not (Test-Path $binary)) {
        Write-Warn "No previous installation found, running full install..."
        Install-Grimorio
        return
    }

    # Get current version
    $currentVersion = 'unknown'
    try {
        $currentVersion = & $binary --version 2>$null | Select-Object -First 1
    }
    catch { }
    Write-Log "Current version: $currentVersion"

    $platform = Get-PlatformInfo
    $tag = Get-LatestTag
    if (-not $tag) { $tag = 'latest' }

    if ($tag -eq $currentVersion) {
        Write-Success "Already up to date ($currentVersion)"
        return
    }

    Write-Log "Updating to: $tag"

    # Download and extract
    $downloadInfo = Download-Release -Tag $tag -Platform $platform
    $checksumsPath = Join-Path $downloadInfo.TempDir 'checksums.txt'
    Test-Checksum -ArchivePath $downloadInfo.ArchivePath -ChecksumsPath $checksumsPath
    $extractedDir = Expand-GrimorioArchive -ArchivePath $downloadInfo.ArchivePath -TempDir $downloadInfo.TempDir

    # Backup current binary
    $backupPath = Join-Path $InstallDir 'grimorio.backup'
    if (Test-Path $binary) {
        Copy-Item -Path $binary -Destination $backupPath -Force
    }

    # Replace binary
    $sourceBinary = Join-Path $extractedDir $BinaryName
    Copy-Item -Path $sourceBinary -Destination $binary -Force

    # Update plugins
    Install-Plugins -SourceDir $extractedDir

    # Update config
    Merge-OpenCodeConfig

    # Update metadata
    Write-Metadata

    # Clean backup on success
    if (Test-Path $backupPath) {
        Remove-Item -Force $backupPath
    }

    # Cleanup
    Remove-Item -Recurse -Force $downloadInfo.TempDir -ErrorAction SilentlyContinue

    Write-Success "Update complete!"
    Write-Log "Run 'grimorio --version' to verify"
}

# ============================================================================
# EXECUTION POLICY HANDLING
# ============================================================================
try {
    $currentPolicy = Get-ExecutionPolicy -Scope Process
    if ($currentPolicy -eq 'Restricted') {
        Write-Warn "Execution policy is Restricted. This script requires at least RemoteSigned."
        Write-Warn "Run the following command in an Administrator PowerShell:"
        Write-Warn "  Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser"
        exit 1
    }
}
catch {
    Write-Warn "Could not determine execution policy: $($_.Exception.Message)"
}

# ============================================================================
# MAIN
# ============================================================================
if ($Update) {
    Update-Grimorio
}
else {
    Install-Grimorio
}
