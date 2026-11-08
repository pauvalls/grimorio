#!/bin/sh
# Grimorio - Cross-Platform Installer v3
# POSIX-compliant download-extract-configure script
# Usage: curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | sh
#
# Supports: Linux (amd64, arm64), macOS (amd64, arm64)
# Requires: curl, tar, shasum or sha256sum
# Optional: python3 (for opencode.json merge), wkhtmltopdf (for PDF generation)

set -e

REPO_OWNER="pauvalls"
REPO_NAME="grimorio"
INSTALL_DIR="${HOME}/.grimorio"
INSTALL_DIR_LEGACY="${HOME}/.local/share/grimorio"
BINARY_DIR="${HOME}/.local/bin"
OPENCODE_CONFIG_DIR="${HOME}/.config/opencode"
OPENCODE_SKILLS_DIR="${OPENCODE_CONFIG_DIR}/skills"
OPENCODE_AGENTS_DIR="${OPENCODE_CONFIG_DIR}/agents"
OPENCODE_COMMANDS_DIR="${OPENCODE_CONFIG_DIR}/commands"
CLAUDE_PLUGIN_DIR="${HOME}/.claude/plugins/grimorio"
METADATA_FILE="${HOME}/.config/grimorio/install-meta.json"
TMP_DIR=""

# ============================================================================
# LOGGING
# ============================================================================
log()   { printf "[Grimorio] %s\n" "$1" >&2; }
warn()  { printf "[WARNING] %s\n" "$1" >&2; }
error() { printf "[ERROR] %s\n" "$1" >&2; exit 1; }
success() { printf "[SUCCESS] %s\n" "$1" >&2; }

command_exists() { command -v "$1" >/dev/null 2>&1; }

# ============================================================================
# CLEANUP
# ============================================================================
cleanup() {
    if [ -n "$TMP_DIR" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}

trap cleanup EXIT INT TERM

# ============================================================================
# PLATFORM DETECTION
# ============================================================================
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    case "$OS" in
        linux|darwin) ;;
        *) error "Unsupported operating system: $OS" ;;
    esac
}

# ============================================================================
# ARCHIVE NAME CONSTRUCTION (matches GoReleaser template)
# ============================================================================
archive_name() {
    local title_os
    title_os=$(printf '%s' "$OS" | awk '{print toupper(substr($0,1,1)) tolower(substr($0,2))}')
    local arch_label
    if [ "$ARCH" = "amd64" ]; then
        arch_label="x86_64"
    else
        arch_label="$ARCH"
    fi
    printf "grimorio_%s_%s" "$title_os" "$arch_label"
}

archive_ext() {
    if [ "$OS" = "darwin" ] || [ "$OS" = "linux" ]; then
        printf "tar.gz"
    else
        printf "zip"
    fi
}

# ============================================================================
# GITHUB API — fetch latest release tag
# ============================================================================
fetch_latest_tag() {
    local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    local response
    response=$(curl -sSL -H "Accept: application/vnd.github.v3+json" "$api_url" 2>/dev/null) || true

    if [ -z "$response" ]; then
        return 1
    fi

    # Extract tag_name using sed (no jq)
    local tag
    tag=$(printf '%s' "$response" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n1)

    if [ -z "$tag" ]; then
        return 1
    fi

    printf '%s' "$tag"
}

# ============================================================================
# DOWNLOAD RELEASE
# ============================================================================
download_release() {
    local tag="$1"
    local name
    name=$(archive_name)
    local ext
    ext=$(archive_ext)
    local archive="${name}.${ext}"

    local base_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag}"
    local archive_url="${base_url}/${archive}"
    local checksum_url="${base_url}/checksums.txt"

    log "Downloading ${archive}..."
    curl -sSL -o "${TMP_DIR}/${archive}" "$archive_url" || error "Failed to download ${archive}"

    log "Downloading checksums.txt..."
    curl -sSL -o "${TMP_DIR}/checksums.txt" "$checksum_url" || warn "Failed to download checksums.txt"

    printf '%s' "${TMP_DIR}/${archive}"
}

# ============================================================================
# VERIFY CHECKSUM
# ============================================================================
verify_checksum() {
    local archive_path="$1"
    local checksums_file="${TMP_DIR}/checksums.txt"

    if [ ! -f "$checksums_file" ]; then
        warn "No checksums.txt available, skipping verification"
        return 0
    fi

    local archive_name
    archive_name=$(basename "$archive_path")

    local expected_hash
    expected_hash=$(grep "^[[:space:]]*[a-f0-9]\{64\}[[:space:]]\+${archive_name}\$" "$checksums_file" 2>/dev/null | awk '{print $1}' | head -n1)

    if [ -z "$expected_hash" ]; then
        warn "No checksum found for ${archive_name}, skipping verification"
        return 0
    fi

    local actual_hash
    if command_exists sha256sum; then
        actual_hash=$(sha256sum "$archive_path" | awk '{print $1}')
    elif command_exists shasum; then
        actual_hash=$(shasum -a 256 "$archive_path" | awk '{print $1}')
    else
        warn "No SHA256 tool found, skipping verification"
        return 0
    fi

    if [ "$actual_hash" != "$expected_hash" ]; then
        error "Checksum mismatch for ${archive_name}!\nExpected: ${expected_hash}\nActual:   ${actual_hash}"
    fi

    log "Checksum verified: ${archive_name}"
}

# ============================================================================
# EXTRACT ARCHIVE
# ============================================================================
extract_archive() {
    local archive_path="$1"
    local extract_dir="${TMP_DIR}/extracted"

    mkdir -p "$extract_dir"

    case "$archive_path" in
        *.tar.gz)
            tar -xzf "$archive_path" -C "$extract_dir" ;;
        *.zip)
            if command_exists unzip; then
                unzip -q "$archive_path" -d "$extract_dir"
            else
                error "unzip is required to extract .zip archives"
            fi
            ;;
        *)
            error "Unknown archive format: $archive_path" ;;
    esac

    # With wrap_in_directory, GoReleaser creates a subdirectory
    local inner_dir
    inner_dir=$(find "$extract_dir" -maxdepth 1 -mindepth 1 -type d | head -n1)

    if [ -n "$inner_dir" ] && [ -f "${inner_dir}/grimorio" ]; then
        printf '%s' "$inner_dir"
    elif [ -f "${extract_dir}/grimorio" ]; then
        printf '%s' "$extract_dir"
    else
        error "Archive extraction failed: grimorio binary not found"
    fi
}

# ============================================================================
# INSTALL BINARY
# ============================================================================
install_binary() {
    local source_dir="$1"
    local binary_name="grimorio"
    [ "$OS" = "windows" ] && binary_name="grimorio.exe"

    # Migrate legacy install dir if present
    if [ -d "$INSTALL_DIR_LEGACY" ] && [ ! -d "$INSTALL_DIR" ]; then
        log "Migrating legacy install directory..."
        mv "$INSTALL_DIR_LEGACY" "$INSTALL_DIR"
    fi

    # Create install directory
    mkdir -p "$INSTALL_DIR"

    # Remove old binary to avoid conflicts
    rm -f "${INSTALL_DIR}/${binary_name}"

    # Copy binary
    cp "${source_dir}/${binary_name}" "${INSTALL_DIR}/${binary_name}"
    chmod +x "${INSTALL_DIR}/${binary_name}"

    # Create symlink or copy to ~/.local/bin
    mkdir -p "$BINARY_DIR"
    rm -f "${BINARY_DIR}/${binary_name}"

    if [ -w "$BINARY_DIR" ]; then
        ln -s "${INSTALL_DIR}/${binary_name}" "${BINARY_DIR}/${binary_name}" 2>/dev/null || \
            cp "${INSTALL_DIR}/${binary_name}" "${BINARY_DIR}/${binary_name}"
    else
        cp "${INSTALL_DIR}/${binary_name}" "${BINARY_DIR}/${binary_name}"
    fi

    success "Binary installed: ${BINARY_DIR}/${binary_name}"
}

# ============================================================================
# CHECK PDF ENGINE
# ============================================================================
check_pdf_engine() {
    # Prefer Chromium/Chrome headless over legacy wkhtmltopdf
    local engines="chromium chrome google-chrome google-chrome-stable wkhtmltopdf"
    local found=""

    for engine in $engines; do
        if command_exists "$engine"; then
            found="$engine"
            break
        fi
    done

    if [ -n "$found" ]; then
        log "PDF engine found: $found"
        return 0
    fi

    warn "No PDF engine found. PDF generation will not work."
    warn "Install one of the following:"
    case "$OS" in
        linux)
            warn "  Chromium:    sudo pacman -S chromium       # Arch/CachyOS"
            warn "  Chromium:    sudo apt-get install chromium  # Debian/Ubuntu"
            warn "  Chrome:      Download from google.com/chrome"
            warn "  (legacy)     sudo pacman -S wkhtmltopdf     # if you must"
            ;;
        darwin)
            warn "  Chrome:      brew install --cask google-chrome"
            warn "  Chromium:    brew install --cask chromium"
            ;;
    esac
}

# ============================================================================
# SETUP PLUGINS — OpenCode V2 assets + Claude plugin
# ============================================================================
setup_plugins() {
    local source_dir="$1"

    # --- OpenCode V2: global skill and agent directories ---
    # V2 discovers skills in ~/.config/opencode/skills and agents in
    # ~/.config/opencode/agents. The old plugin directory is not scanned.
    mkdir -p "$OPENCODE_SKILLS_DIR" "$OPENCODE_AGENTS_DIR"

    if [ -d "${source_dir}/skills" ]; then
        for skill_dir in "${source_dir}/skills"/*/; do
            [ -d "$skill_dir" ] || continue
            local sname
            sname=$(basename "$skill_dir")
            mkdir -p "${OPENCODE_SKILLS_DIR}/${sname}"
            cp -Rf "${skill_dir}." "${OPENCODE_SKILLS_DIR}/${sname}/"
        done
        log "Skills installed in ${OPENCODE_SKILLS_DIR}"
    fi

    if [ -d "${source_dir}/agents" ]; then
        for f in "${source_dir}/agents"/*.md; do
            [ -f "$f" ] || continue
            cp -f "$f" "${OPENCODE_AGENTS_DIR}/"
        done
        log "Agents installed in ${OPENCODE_AGENTS_DIR}"
    fi

    # --- Claude Code plugin (unchanged layout) ---
    local plugin_dir="$CLAUDE_PLUGIN_DIR"
    mkdir -p "${plugin_dir}/agents" "${plugin_dir}/skills"

    if [ -d "${source_dir}/agents" ]; then
        for f in "${source_dir}/agents"/*.md; do
            [ -f "$f" ] || continue
            cp -f "$f" "${plugin_dir}/agents/"
        done
    fi

    if [ -d "${source_dir}/skills" ]; then
        for skill_dir in "${source_dir}/skills"/*/; do
            [ -d "$skill_dir" ] || continue
            local sname
            sname=$(basename "$skill_dir")
            mkdir -p "${plugin_dir}/skills/${sname}"
            cp -Rf "${skill_dir}." "${plugin_dir}/skills/${sname}/"
        done
    fi

    create_mcp_json "$plugin_dir"
    success "Claude plugin installed: ${plugin_dir}"
}

# ============================================================================
# CREATE .MCP.JSON — Claude Code plugin descriptor
# ============================================================================
create_mcp_json() {
    local plugin_dir="$1"
    local binary_path="${INSTALL_DIR}/grimorio"
    if [ "$OS" = "windows" ]; then
        binary_path="${INSTALL_DIR}/grimorio.exe"
    fi

    cat > "${plugin_dir}/.mcp.json" << EOF
{
  "grimorio": {
    "command": "${binary_path}",
    "args": [],
    "env": {}
  }
}
EOF
}

# ============================================================================
# MERGE OPENCODE.JSON — OpenCode V2 native shape
# ============================================================================
write_grimorio_command() {
    mkdir -p "$OPENCODE_COMMANDS_DIR"
    cat > "${OPENCODE_COMMANDS_DIR}/grimorio.md" << 'GRIMCMD'
---
description: Generate a complete D&D 5e campaign or one-shot from an idea
agent: grimorio-architect
subagent: false
---

Start the Grimorio campaign generation workflow.

User request: $ARGUMENTS

Follow the mandatory intake (language first), then generate the campaign end-to-end and compile the PDF.
GRIMCMD
    log "Command installed: ${OPENCODE_COMMANDS_DIR}/grimorio.md"
}

merge_opencode_config() {
    local config_file="${OPENCODE_CONFIG_DIR}/opencode.json"
    local binary_path="${INSTALL_DIR}/grimorio"
    if [ "$OS" = "windows" ]; then
        binary_path="${INSTALL_DIR}/grimorio.exe"
    fi

    mkdir -p "$OPENCODE_CONFIG_DIR"
    write_grimorio_command

    if ! command_exists python3; then
        warn "python3 not found; cannot merge ${config_file} automatically."
        warn "Add this entry under mcp.servers in ${config_file}:"
        warn "  \"grimorio\": {\"type\": \"local\", \"command\": [\"${binary_path}\"]}"
        return 1
    fi

    if [ -f "$config_file" ]; then
        cp "$config_file" "${config_file}.backup.$(date +%Y%m%d%H%M%S)"
    fi

    if ! GRIMORIO_BINARY="$binary_path" GRIMORIO_CONFIG="$config_file" python3 - << 'GRIMPY'
import json, os, sys
from pathlib import Path

config_path = Path(os.environ['GRIMORIO_CONFIG'])
binary = os.environ['GRIMORIO_BINARY']
home = Path.home()

data = {}
if config_path.exists():
    raw = config_path.read_text()
    if raw.strip():
        try:
            data = json.loads(raw)
        except json.JSONDecodeError as exc:
            print('opencode.json is not valid JSON: %s' % exc, file=sys.stderr)
            sys.exit(1)

mcp = data.get('mcp')
if not isinstance(mcp, dict):
    mcp = {}
    data['mcp'] = mcp

servers = mcp.get('servers')
if not isinstance(servers, dict):
    servers = {}
    mcp['servers'] = servers

previous = servers.get('grimorio')
if not isinstance(previous, dict):
    previous = {}

environment = {}
for source in (previous.get('environment'), previous.get('env')):
    if isinstance(source, dict):
        for key, value in source.items():
            if isinstance(value, str):
                environment[key] = value

if not environment.get('PATH'):
    environment['PATH'] = os.environ.get('PATH') or '/usr/local/bin:/usr/bin:/bin'

if not environment.get('PIPER_MODEL_PATH'):
    voices = sorted((home / '.local' / 'share' / 'piper').glob('*.onnx'))
    if voices:
        environment['PIPER_MODEL_PATH'] = str(voices[0])
        if Path(str(voices[0]) + '.json').exists():
            environment.setdefault('PIPER_CONFIG_PATH', str(voices[0]) + '.json')

defaults = {
    'PIPER_PORT': '5000',
    'PIPER_HOST': '127.0.0.1',
    'PIPER_LENGTH_SCALE': '1.0',
    'PIPER_VOLUME': '0.8',
    'GRIMORIO_TTS_ENABLED': 'true',
    'AUDIO_PLAYER': 'auto',
}
for key, value in defaults.items():
    environment.setdefault(key, value)

servers['grimorio'] = {
    'type': 'local',
    'command': [binary],
    'disabled': False,
    'environment': environment,
}

# Drop the V1 shapes so the V2 entry is unambiguous.
mcp.pop('grimorio', None)
legacy = data.get('command')
if isinstance(legacy, dict):
    legacy.pop('grimorio', None)
    if not legacy:
        data.pop('command', None)

rendered = json.dumps(data, indent=2, ensure_ascii=False) + '\n'
if not config_path.exists() or config_path.read_text() != rendered:
    config_path.write_text(rendered)
    print('opencode.json updated')
else:
    print('opencode.json already up to date')
GRIMPY
    then
        warn "Could not merge ${config_file} automatically."
        warn "Add this entry under mcp.servers in ${config_file}:"
        warn "  \"grimorio\": {\"type\": \"local\", \"command\": [\"${binary_path}\"]}"
        return 1
    fi

    success "Merged Grimorio config into opencode.json (OpenCode V2)"
    return 0
}

# ============================================================================
# UPDATE PATH
# ============================================================================
update_path() {
    case ":${PATH}:" in
        *:"${BINARY_DIR}":*)
            log "${BINARY_DIR} is already in PATH"
            return 0
            ;;
    esac

    warn "${BINARY_DIR} is not in your PATH."
    warn "Add it by running one of the following:"
    warn "  echo 'export PATH=\"${BINARY_DIR}:\$PATH\"' >> ~/.bashrc"
    warn "  echo 'export PATH=\"${BINARY_DIR}:\$PATH\"' >> ~/.zshrc"
    warn "Then restart your terminal or source your shell config."
}

# ============================================================================
# WRITE METADATA
# ============================================================================
write_meta() {
    mkdir -p "$(dirname "$METADATA_FILE")"
    local binary_path="${INSTALL_DIR}/grimorio"
    local version
    version=$("$binary_path" --version 2>/dev/null | head -n1 || echo "unknown")

    cat > "$METADATA_FILE" << EOF
{
  "version": "${version}",
  "installedAt": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "installDir": "${INSTALL_DIR}",
  "os": "${OS}",
  "arch": "${ARCH}"
}
EOF
    success "Install metadata written to $METADATA_FILE"
}

# ============================================================================
# PRINT INSTRUCTIONS
# ============================================================================
print_instructions() {
    printf "\n"
    printf "========================================\n"
    printf "  Grimorio Installation Complete!\n"
    printf "========================================\n"
    printf "\n"
    printf "What was installed:\n"
    printf "   Binary:        %s/grimorio\n" "$INSTALL_DIR"
    printf "   Symlink:       %s/grimorio\n" "$BINARY_DIR"
    printf "   OpenCode MCP:  mcp.servers.grimorio in %s/opencode.json\n" "$OPENCODE_CONFIG_DIR"
    printf "   Skills:        %s\n" "$OPENCODE_SKILLS_DIR"
    printf "   Agents:        %s\n" "$OPENCODE_AGENTS_DIR"
    printf "   Command:       %s/grimorio.md\n" "$OPENCODE_COMMANDS_DIR"
    printf "   Claude plugin: %s\n" "$CLAUDE_PLUGIN_DIR"
    printf "\n"
    printf "Next steps:\n"
    printf "   1. Restart your terminal, or add %s to PATH\n" "$BINARY_DIR"
    printf "   2. Verify the binary:  grimorio --version\n"
    printf "   3. Verify OpenCode:    opencode mcp list\n"
    printf "   4. Run /grimorio <your idea> inside OpenCode\n"
    printf "\n"
    printf "To update later: grimorio update all\n"
    printf "\n"
}


# ============================================================================
# VERIFY INSTALLATION
# ============================================================================
verify_installation() {
    local binary_name="grimorio"
    if [ -f "${INSTALL_DIR}/${binary_name}" ]; then
        local version
        version=$("${INSTALL_DIR}/${binary_name}" --version 2>/dev/null | head -n1 || echo "unknown")
        log "Installed version: $version"
    else
        error "Installation verification failed: binary not found at ${INSTALL_DIR}/${binary_name}"
    fi
}

# ============================================================================
# FULL INSTALL
# ============================================================================
do_install() {
    log "Starting Grimorio installation..."

    TMP_DIR=$(mktemp -d)

    # Step 1: Detect platform
    detect_platform
    log "Platform: ${OS}/${ARCH}"

    # Step 2: Get latest release tag
    local tag
    tag=$(fetch_latest_tag) || {
        warn "Could not fetch latest tag from GitHub API, using 'latest' URL fallback"
        tag="latest"
    }
    log "Release: $tag"

    # Step 3: Download release
    local archive_path
    archive_path=$(download_release "$tag")

    # Step 4: Verify checksum
    verify_checksum "$archive_path"

    # Step 5: Extract archive
    local extracted_dir
    extracted_dir=$(extract_archive "$archive_path")
    log "Extracted to: $extracted_dir"

    # Step 6: Install binary
    install_binary "$extracted_dir"

    # Step 7: Setup plugins (agents, skills, .mcp.json)
    setup_plugins "$extracted_dir"

    # Step 8: Configure OpenCode (safe to fail: the binary is already usable)
    merge_opencode_config || warn "OpenCode integration was not configured automatically"

    # Step 9: Check PDF engine
    check_pdf_engine

    # Step 10: Update PATH warning
    update_path

    # Step 11: Write metadata
    write_meta

    # Step 12: Verify
    verify_installation

    print_instructions
    success "Installation complete!"
}

# ============================================================================
# UPDATE MODE — Incremental update
# ============================================================================
do_update() {
    log "Starting Grimorio update..."

    if [ ! -f "$METADATA_FILE" ]; then
        warn "No previous installation found, running full install..."
        do_install
        return
    fi

    if [ ! -f "${INSTALL_DIR}/grimorio" ]; then
        warn "Grimorio binary not found, running full install..."
        do_install
        return
    fi

    # Get current version
    local current_version
    current_version=$("${INSTALL_DIR}/grimorio" --version 2>/dev/null | sed -n 's/.*version \(v[^ ]*\).*/\1/p' | head -n1)
    [ -n "$current_version" ] || current_version="unknown"
    log "Current version: $current_version"

    TMP_DIR=$(mktemp -d)

    # Detect platform
    detect_platform

    # Get latest tag
    local tag
    tag=$(fetch_latest_tag) || tag="latest"

    # If we can compare versions, skip if already latest
    if [ "$tag" = "$current_version" ]; then
        success "Already up to date ($current_version)"
        return 0
    fi

    log "Updating to: $tag"

    # Download and extract
    local archive_path
    archive_path=$(download_release "$tag")
    verify_checksum "$archive_path"

    local extracted_dir
    extracted_dir=$(extract_archive "$archive_path")

    # Backup current binary
    local backup_path="${INSTALL_DIR}/grimorio.backup"
    cp -f "${INSTALL_DIR}/grimorio" "$backup_path" 2>/dev/null || true

    # Replace binary using temp file to avoid "text file busy"
    local tmp_binary="${INSTALL_DIR}/grimorio.new.$$"
    cp -f "${extracted_dir}/grimorio" "$tmp_binary"
    chmod +x "$tmp_binary"
    mv -f "$tmp_binary" "${INSTALL_DIR}/grimorio"

    # Update symlink
    rm -f "${BINARY_DIR}/grimorio"
    if [ -w "$BINARY_DIR" ]; then
        ln -s "${INSTALL_DIR}/grimorio" "${BINARY_DIR}/grimorio" 2>/dev/null || \
            cp "${INSTALL_DIR}/grimorio" "${BINARY_DIR}/grimorio"
    else
        cp "${INSTALL_DIR}/grimorio" "${BINARY_DIR}/grimorio"
    fi

    # Update plugins (agents/skills only if changed)
    setup_plugins "$extracted_dir"

    # Update OpenCode integration
    merge_opencode_config || warn "OpenCode integration was not updated automatically"

    # Update metadata
    write_meta

    # Clean up backup on success
    rm -f "$backup_path"

    success "Update complete!"
    log "Run 'grimorio --version' to verify"
}

# ============================================================================
# MAIN
# ============================================================================
main() {
    case "${1:-}" in
        --update|-u)
            do_update
            ;;
        *)
            do_install
            ;;
    esac
}

main "$@"
