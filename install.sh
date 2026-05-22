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
OPENCODE_PLUGIN_DIR="${HOME}/.config/opencode/plugins/grimorio"
OPENCODE_AGENTS_DIR="${HOME}/.config/opencode/agents"
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
# CHECK WKHTMLTOPDF
# ============================================================================
check_wkhtmltopdf() {
    if command_exists wkhtmltopdf; then
        log "wkhtmltopdf found: $(wkhtmltopdf --version 2>/dev/null | head -n1 || echo 'unknown version')"
        return 0
    fi

    warn "wkhtmltopdf not found. PDF generation will not work."
    case "$OS" in
        linux)
            warn "Install with: sudo apt-get install wkhtmltopdf"
            warn "    or:      sudo dnf install wkhtmltopdf"
            warn "    or:      sudo pacman -S wkhtmltopdf"
            ;;
        darwin)
            warn "Install with: brew install --cask wkhtmltopdf"
            ;;
    esac
}

# ============================================================================
# SETUP PLUGINS — Copy agents and skills to plugin directories
# ============================================================================
setup_plugins() {
    local source_dir="$1"

    for plugin_dir in "$OPENCODE_PLUGIN_DIR" "$CLAUDE_PLUGIN_DIR"; do
        log "Setting up plugin: $plugin_dir"
        mkdir -p "$plugin_dir"

        # Copy agents
        if [ -d "${source_dir}/agents" ]; then
            mkdir -p "${plugin_dir}/agents"
            for f in "${source_dir}/agents"/*; do
                [ -f "$f" ] || continue
                cp -f "$f" "${plugin_dir}/agents/"
            done
            log "Agents copied to ${plugin_dir}/agents"
        fi

        # Copy skills
        if [ -d "${source_dir}/skills" ]; then
            mkdir -p "${plugin_dir}/skills"
            for skill_dir in "${source_dir}/skills"/*/; do
                [ -d "$skill_dir" ] || continue
                local sname
                sname=$(basename "$skill_dir")
                mkdir -p "${plugin_dir}/skills/${sname}"
                if [ -f "${skill_dir}/SKILL.md" ]; then
                    cp -f "${skill_dir}/SKILL.md" "${plugin_dir}/skills/${sname}/SKILL.md"
                fi
            done
            log "Skills copied to ${plugin_dir}/skills"
        fi

        # Create .mcp.json
        create_mcp_json "$plugin_dir"

        success "Plugin installed: $plugin_dir"
    done

    # Copy agents to OpenCode global agents directory
    if [ -d "${source_dir}/agents" ]; then
        mkdir -p "$OPENCODE_AGENTS_DIR"
        for f in "${source_dir}/agents"/*; do
            [ -f "$f" ] || continue
            cp -f "$f" "$OPENCODE_AGENTS_DIR/"
        done
        log "Agents copied to $OPENCODE_AGENTS_DIR"
    fi
}

# ============================================================================
# CREATE .MCP.JSON
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
# MERGE OPENCODE.JSON — Python/awk fallback, no jq required
# ============================================================================
merge_opencode_config() {
    local config_file="${HOME}/.config/opencode/opencode.json"
    local plugin_dir="$OPENCODE_PLUGIN_DIR"
    local binary_path="${INSTALL_DIR}/grimorio"
    if [ "$OS" = "windows" ]; then
        binary_path="${INSTALL_DIR}/grimorio.exe"
    fi

    mkdir -p "$(dirname "$config_file")"

    # Build grimorio config fragment
    local grimorio_config
    grimorio_config=$(cat << GRIMJSON
{
  "mcp": {
    "grimorio": {
      "command": ["${binary_path}"],
      "type": "local",
      "enabled": true
    }
  },
  "command": {
    "grimorio": {
      "description": "Generate a complete D&D 5e campaign or one-shot from an idea (executes in main thread)",
      "subtask": false,
      "template": "You are Grimorio, a D&D 5e campaign generator."
    }
  }
}
GRIMJSON
)

    if [ ! -f "$config_file" ]; then
        printf '%s\n' "$grimorio_config" > "$config_file"
        success "Created opencode.json with Grimorio config"
        return 0
    fi

    # Backup existing config
    cp "$config_file" "${config_file}.backup.$(date +%Y%m%d%H%M%S)"

    # Try python3 first for reliable JSON merge
    if command_exists python3; then
        python3 -c "
import json, sys, re
config_path = '${config_file}'
with open(config_path, 'r') as f:
    content = f.read()
try:
    data = json.loads(content)
except json.JSONDecodeError as e:
    print(f'Warning: opencode.json is invalid JSON: {e}', file=sys.stderr)
    sys.exit(1)

# Load grimorio config
grimorio = json.loads('''${grimorio_config}''')

# Remove old auto-generated grimorio entries from mcp, command, agent
for key in ['mcp', 'command', 'agent']:
    if key in data and isinstance(data[key], dict):
        data[key] = {k: v for k, v in data[key].items()
                     if not isinstance(v, dict) or v.get('grimorio_auto_generated') != True}

# Merge grimorio config
for key in ['mcp', 'command']:
    if key in grimorio:
        if key not in data:
            data[key] = {}
        data[key].update(grimorio[key])

with open(config_path, 'w') as f:
    json.dump(data, f, indent=2)
" 2>/dev/null && {
            success "Merged Grimorio config into opencode.json (python3)"
            return 0
        }
    fi

    # Fallback: try python
    if command_exists python; then
        python -c "
import json, sys
config_path = '${config_file}'
with open(config_path, 'r') as f:
    data = json.load(f) if f.read().strip() else {}
if not data:
    data = {}
grimorio = json.loads('''${grimorio_config}''')
for key in ['mcp', 'command']:
    if key in grimorio:
        if key not in data:
            data[key] = {}
        data[key].update(grimorio[key])
with open(config_path, 'w') as f:
    json.dump(data, f, indent=2)
" 2>/dev/null && {
            success "Merged Grimorio config into opencode.json (python)"
            return 0
        }
    fi

    # Last resort: awk-based merge for simple cases
    if command_exists awk; then
        local tmp_file="${config_file}.tmp.$$"
        awk '
            BEGIN { in_mcp=0; in_command=0; brace_depth=0 }
            /"mcp"[[:space:]]*:/ { in_mcp=1 }
            /"command"[[:space:]]*:/ { in_command=1 }
            /"grimorio"[[:space:]]*:/ {
                if (in_mcp || in_command) {
                    # Skip this entry and its value
                    getline
                    while (match($0, /{/) || brace_depth > 0) {
                        if (match($0, /{/)) brace_depth++
                        if (match($0, /}/)) brace_depth--
                        if (brace_depth <= 0) break
                        getline
                    }
                    next
                }
            }
            /}[[:space:]]*,?[[:space:]]*$/ {
                if (in_mcp) in_mcp=0
                if (in_command) in_command=0
            }
            { print }
        ' "$config_file" > "$tmp_file" 2>/dev/null || cp "$config_file" "$tmp_file"

        # Append grimorio entries before final closing brace
        awk -v grimorio_mcp='"grimorio": {"command": ["'"$binary_path"'"], "type": "local", "enabled": true}' \
            -v grimorio_cmd='"grimorio": {"description": "Generate a complete D&D 5e campaign or one-shot", "subtask": false, "template": "You are Grimorio..."}' \
            '
            { print }
            END {
                print ","
                print "  \"mcp\": {"
                print "    " grimorio_mcp
                print "  },"
                print "  \"command\": {"
                print "    " grimorio_cmd
                print "  }"
            }
        ' "$tmp_file" > "${config_file}.new" 2>/dev/null && mv "${config_file}.new" "$config_file"
        rm -f "$tmp_file"
        success "Merged Grimorio config into opencode.json (awk fallback)"
        return 0
    fi

    warn "Could not merge opencode.json — install python3 for automatic merge, or merge manually"
    return 1
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
    printf "   Binary: %s/grimorio\n" "$INSTALL_DIR"
    printf "   Symlink: %s/grimorio\n" "$BINARY_DIR"
    printf "   OpenCode plugin: %s\n" "$OPENCODE_PLUGIN_DIR"
    printf "   Claude plugin: %s\n" "$CLAUDE_PLUGIN_DIR"
    printf "   MCP configured in opencode.json\n"
    printf "\n"
    printf "Next steps:\n"
    printf "   1. Restart your terminal or add %s to PATH\n" "$BINARY_DIR"
    printf "   2. Run: grimorio --version\n"
    printf "   3. Use: grimorio create_campaign <name>\n"
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

    # Step 8: Merge opencode.json
    merge_opencode_config

    # Step 9: Check wkhtmltopdf
    check_wkhtmltopdf

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
    current_version=$("${INSTALL_DIR}/grimorio" --version 2>/dev/null | head -n1 || echo "unknown")
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

    # Update opencode.json
    merge_opencode_config

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
