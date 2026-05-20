#!/bin/bash
# Grimorio - Clean Installer v2
# Complete MCP installation - removes old, installs fresh
# Usage: curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash

REPO_URL="https://github.com/pauvalls/grimorio"
INSTALL_DIR="${HOME}/.local/share/grimorio"
CLAUDE_PLUGIN_DIR="${HOME}/.claude/plugins/grimorio"
OPENCODE_PLUGIN_DIR="${HOME}/.config/opencode/plugins/grimorio"
BINARY_DIR="${HOME}/.local/bin"
GLOBAL_SKILLS_DIR="${HOME}/.config/opencode/skills"
METADATA_FILE="${HOME}/.config/grimorio/install-meta.json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${BLUE}[Grimorio]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

command_exists() { command -v "$1" >/dev/null 2>&1; }

# ============================================================================
# SHA256 HELPERS
# ============================================================================
sha256_file() {
    sha256sum "$1" | cut -d' ' -f1
}

copy_if_changed() {
    local src="$1" dst="$2" current_hash="$3"
    # Always copy if destination does not exist
    if [ ! -f "$dst" ]; then
        mkdir -p "$(dirname "$dst")"
        cp "$src" "$dst"
        log "Copied: $src -> $dst"
        return 1
    fi
    local file_hash
    file_hash=$(sha256_file "$src" 2>/dev/null || echo "")
    [ -n "$current_hash" ] && [ "$file_hash" = "$current_hash" ] && return 0
    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
    log "Copied: $src -> $dst"
    return 1
}

# ============================================================================
# METADATA MANAGEMENT
# ============================================================================
read_meta() {
    cat "$METADATA_FILE" 2>/dev/null || echo "{}"
}

write_meta() {
    mkdir -p "$(dirname "$METADATA_FILE")"
    command_exists jq || return 1

    local binary_hash
    binary_hash=$(sha256_file "$BINARY_DIR/grimorio" 2>/dev/null || echo "unknown")

    local version commit build_date installed_at
    version=$(git -C "$INSTALL_DIR" describe --tags --always --dirty 2>/dev/null || echo "v3.5.0")
    commit=$(git -C "$INSTALL_DIR" rev-parse --short HEAD 2>/dev/null || echo "unknown")
    build_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    installed_at=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Hash agents
    local agents_json="{}"
    if [ -d "$INSTALL_DIR/agents" ]; then
        for f in "$INSTALL_DIR/agents"/grimorio-*.md; do
            [ -f "$f" ] || continue
            local name hash
            name=$(basename "$f")
            hash=$(sha256_file "$f")
            agents_json=$(jq --arg name "$name" --arg hash "sha256-$hash" '.[$name] = $hash' <<< "$agents_json" 2>/dev/null)
        done
    fi

    # Hash skills
    local skills_json="{}"
    if [ -d "$INSTALL_DIR/skills" ]; then
        for skill_dir in "$INSTALL_DIR/skills"/*/; do
            [ -d "$skill_dir" ] || continue
            local skill_name="${skill_dir%/}"; skill_name="${skill_name##*/}"
            local skill_file="$skill_dir/SKILL.md"
            if [ -f "$skill_file" ]; then
                local hash
                hash=$(sha256_file "$skill_file")
                skills_json=$(jq --arg name "$skill_name" --arg hash "sha256-$hash" '.[$name] = $hash' <<< "$skills_json" 2>/dev/null)
            fi
        done
    fi

    # Hash templates
    local templates_json="{}"
    local tmpl_dir="$INSTALL_DIR/internal/compiler/templates"
    if [ -d "$tmpl_dir" ]; then
        for f in "$tmpl_dir"/*.tmpl; do
            [ -f "$f" ] || continue
            local tmpl_name hash
            tmpl_name=$(basename "$f")
            hash=$(sha256_file "$f")
            templates_json=$(jq --arg name "$tmpl_name" --arg hash "sha256-$hash" '.[$name] = $hash' <<< "$templates_json" 2>/dev/null)
        done
    fi

    jq -n \
      --arg version "$version" \
      --arg commit "$commit" \
      --arg buildDate "$build_date" \
      --arg installedAt "$installed_at" \
      --argjson agents "$agents_json" \
      --argjson skills "$skills_json" \
      --argjson templates "$templates_json" \
      --arg binaryHash "sha256-$binary_hash" \
      '{
        "version": $version,
        "commit": $commit,
        "buildDate": $buildDate,
        "installedAt": $installedAt,
        "agents": $agents,
        "skills": $skills,
        "templates": $templates,
        "binaryHash": $binaryHash
      }' > "$METADATA_FILE"

    success "Install metadata written to $METADATA_FILE"
}

# ============================================================================
# CLEAN INSTALLATION - Remove EVERYTHING from previous installs
# ============================================================================
clean_installation() {
    log "Cleaning previous Grimorio installation..."
    local cleaned=false

    # Clean plugin directories
    for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
        if [ -d "$plugin_dir" ]; then
            rm -f "$plugin_dir/grimorio" "$plugin_dir/migrate-v1-to-v2"
            rm -f "$plugin_dir/.mcp.json"
            rm -rf "$plugin_dir/agents" "$plugin_dir/skills" "$plugin_dir/internal"
            rm -rf "$plugin_dir/.claude-plugin" "$plugin_dir/commands"
            log "Cleaned: $plugin_dir"
            cleaned=true
        fi
    done

    # Clean binaries
    for bin in grimorio migrate-v1-to-v2; do
        if [ -f "$BINARY_DIR/$bin" ]; then
            rm -f "$BINARY_DIR/$bin"
            log "Removed: $BINARY_DIR/$bin"
            cleaned=true
        fi
    done

    # Clean install directory
    if [ -d "$INSTALL_DIR" ]; then
        rm -rf "$INSTALL_DIR"
        log "Removed: $INSTALL_DIR"
        cleaned=true
    fi

    # Clean opencode.json
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"
    if [ -f "$OPENCODE_CONFIG" ] && command_exists jq; then
        jq 'del(.mcp.grimorio) | del(.command.grimorio) | 
            del(.agent | with_entries(select(.key | startswith("grimorio"))))' \
            "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" 2>/dev/null && \
            mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        log "Cleaned opencode.json"
        cleaned=true
    fi

    # Clean shell configs
    local shell_rcs=("${HOME}/.bashrc" "${HOME}/.zshrc" "${HOME}/.config/fish/config.fish" "${HOME}/.profile")
    for rc in "${shell_rcs[@]}"; do
        if [ -f "$rc" ]; then
            awk '/^# === GRIMORIO CONFIG BEGIN ===$/{skip=1} /^# === GRIMORIO CONFIG END ===$/{skip=0} !skip' \
                "$rc" > "${rc}.tmp" && mv "${rc}.tmp" "$rc"
            grep -v "^export PATH.*\.local.*bin.*PATH" "$rc" > "${rc}.tmp" 2>/dev/null && mv "${rc}.tmp" "$rc" || true
        fi
    done
    log "Cleaned shell configs"

    [ "$cleaned" = true ] && success "Previous installation cleaned" || log "No previous installation found"
}

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
        linux) WKHTMLTOPDF_URL="https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltopdf_0.12.6.1-3.linux-${ARCH}.deb" ;;
        darwin) WKHTMLTOPDF_URL="https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltopdf-0.12.6.1-3.macos-cocoa.pkg" ;;
        *) warn "Manual wkhtmltopdf install required for $OS"; WKHTMLTOPDF_URL="" ;;
    esac
}

# ============================================================================
# INSTALL GO
# ============================================================================
install_go() {
    if command_exists go; then
        log "Go found: $(go version | awk '{print $3}')"
        return 0
    fi

    log "Installing Go 1.23.4..."
    GO_VERSION="1.23.4"
    curl -L "https://go.dev/dl/go${GO_VERSION}.${OS}-${ARCH}.tar.gz" -o "/tmp/go.tar.gz"
    tar -C "${HOME}/.local" -xzf "/tmp/go.tar.gz"
    export PATH="${HOME}/.local/go/bin:$PATH"
    command_exists go || error "Go installation failed"
    success "Go installed"
}

# ============================================================================
# INSTALL WKHTMLTOPDF
# ============================================================================
install_wkhtmltopdf() {
    if command_exists wkhtmltopdf; then
        log "wkhtmltopdf found: $(wkhtmltopdf --version | head -n1)"
        return 0
    fi

    [ -z "$WKHTMLTOPDF_URL" ] && { warn "Install wkhtmltopdf manually"; return 0; }

    log "Installing wkhtmltopdf..."
    case "$OS" in
        linux)
            if command_exists dpkg; then
                curl -L "$WKHTMLTOPDF_URL" -o /tmp/wkhtmltopdf.deb
                dpkg -x /tmp/wkhtmltopdf.deb /tmp/wkhtmltopdf
                mkdir -p "$BINARY_DIR"
                cp /tmp/wkhtmltopdf/usr/local/bin/wkhtmltopdf "$BINARY_DIR/"
                chmod +x "$BINARY_DIR/wkhtmltopdf"
            fi ;;
        darwin)
            curl -L "$WKHTMLTOPDF_URL" -o /tmp/wkhtmltopdf.pkg
            log "Run: sudo installer -pkg /tmp/wkhtmltopdf.pkg -target /" ;;
    esac

    command_exists wkhtmltopdf && success "wkhtmltopdf installed" || warn "Manual install may be required"
}

# ============================================================================
# SETUP REPOSITORY
# ============================================================================
setup_repo() {
    log "Setting up repository..."
    [ -d "$INSTALL_DIR" ] && rm -rf "$INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
    git clone --depth 200 "$REPO_URL" "$INSTALL_DIR" 2>/dev/null || \
        (curl -sSL "${REPO_URL}/archive/refs/heads/main.tar.gz" | tar -xzf - -C "$INSTALL_DIR" --strip-components=1)
    # Fetch tags for proper version detection
    git -C "$INSTALL_DIR" fetch --tags 2>/dev/null || true
    success "Repository ready at $INSTALL_DIR"
}

# ============================================================================
# BUILD BINARIES
# ============================================================================
build_binary() {
    log "Building Grimorio binaries..."
    cd "$INSTALL_DIR"
    export PATH="${HOME}/.local/go/bin:$PATH"

    # Get version info from git
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v3.0.0-dev")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

    go build -ldflags "$LDFLAGS" -o grimorio ./cmd/grimorio
    go build -ldflags "$LDFLAGS" -o migrate-v1-to-v2 ./cmd/migrate-v1-to-v2

    mkdir -p "$BINARY_DIR"
    cp grimorio migrate-v1-to-v2 "$BINARY_DIR/"
    chmod +x "$BINARY_DIR/grimorio" "$BINARY_DIR/migrate-v1-to-v2"

    success "Binaries installed to $BINARY_DIR"
}

# ============================================================================
# MIGRATE CAMPAIGNS
# ============================================================================
migrate_existing_campaigns() {
    local CAMPAIGNS_DIR="${HOME}/campaigns"
    [ ! -d "$CAMPAIGNS_DIR" ] && return 0

    local needs_migration=false
    for campaign_dir in "$CAMPAIGNS_DIR"/*/; do
        [ -d "$campaign_dir" ] && [ ! -f "$campaign_dir/canon.json" ] && needs_migration=true && break
    done

    [ "$needs_migration" = false ] && return 0

    log "Migrating v1 campaigns..."
    "$BINARY_DIR/migrate-v1-to-v2" "$CAMPAIGNS_DIR" 2>/dev/null && \
        success "Migration complete. Backups in .v1-backup/" || \
        warn "Migration had issues"
}

# ============================================================================
# SETUP PLUGIN - Copy ALL files
# ============================================================================
setup_plugin() {
    for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
        log "Setting up plugin: $plugin_dir"
        mkdir -p "$plugin_dir"/{agents,skills,internal/compiler/templates}

        # Copy binaries
        # Copy binaries (use temp+mv to avoid "text file busy")
        for bin in grimorio migrate-v1-to-v2; do
            if [ -f "$BINARY_DIR/$bin" ]; then
                cp -f "$BINARY_DIR/$bin" "$plugin_dir/$bin.tmp"
                chmod +x "$plugin_dir/$bin.tmp"
                mv -f "$plugin_dir/$bin.tmp" "$plugin_dir/$bin"
            fi
        done

        # Copy agents (if exist)
        if [ -d "$INSTALL_DIR/agents" ]; then
            for f in "$INSTALL_DIR/agents"/grimorio-*.md; do
                [ -f "$f" ] && cp -f "$f" "$plugin_dir/agents/"
            done
        fi

        # Copy skills (if exist) — subdirectories with SKILL.md
        if [ -d "$INSTALL_DIR/skills" ]; then
            mkdir -p "$plugin_dir/skills"
            for skill_dir_entry in "$INSTALL_DIR/skills"/*/; do
                [ -d "$skill_dir_entry" ] || continue
                local sname="${skill_dir_entry%/}"; sname="${sname##*/}"
                local sfile="$skill_dir_entry/SKILL.md"
                [ -f "$sfile" ] || continue
                mkdir -p "$plugin_dir/skills/$sname"
                cp -f "$sfile" "$plugin_dir/skills/$sname/SKILL.md"
                # Also copy to global skills directory
                mkdir -p "$GLOBAL_SKILLS_DIR/$sname"
                cp -f "$sfile" "$GLOBAL_SKILLS_DIR/$sname/SKILL.md"
            done
        fi

        # Copy templates (if exist)
        if [ -d "$INSTALL_DIR/internal/compiler/templates" ]; then
            cp -r "$INSTALL_DIR/internal/compiler/templates"/* "$plugin_dir/internal/compiler/templates/" 2>/dev/null || true
        fi

        # Create .mcp.json
        cat > "$plugin_dir/.mcp.json" << EOF
{
  "grimorio": {
    "command": "$plugin_dir/grimorio",
    "args": [],
    "env": {}
  }
}
EOF
        success "Plugin installed: $plugin_dir"
    done
}

# ============================================================================
# COPY COMMANDS DIRECTORY (for OpenCode/Claude Code)
# ============================================================================
copy_commands() {
    local src_commands="${INSTALL_DIR}/commands"
    
    # Copy commands to plugin directories
    for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
        if [ -d "$src_commands" ]; then
            mkdir -p "$plugin_dir/commands"
            cp -r "$src_commands/"* "$plugin_dir/commands/" 2>/dev/null || true
            log "Commands copied to: $plugin_dir/commands"
        fi
    done
}

# ============================================================================
# PARSE FRONTMATTER TOOLS - Parse YAML tools block from agent frontmatter
# ============================================================================
parse_frontmatter_tools() {
    local file="$1"
    local frontmatter
    frontmatter=$(sed -n '/^---$/,/^---$/p' "$file" 2>/dev/null)

    # Fallback: return default tools JSON if no tools block found
    echo "$frontmatter" | grep -q "^tools:" || {
        echo '{"bash": true, "edit": true, "read": true, "write": true}'
        return
    }

    # Extract tools block content (indented lines under "tools:")
    local tools_yaml
    tools_yaml=$(echo "$frontmatter" | awk '
        /^tools:/ { found=1; next }
        found && /^  [a-z]/ { print; next }
        found && /^    [ -]/ { print; next }
        found && /^[a-z]/ { exit }
        found && /^---/ { exit }
    ')

    # Build JSON from YAML key: value pairs
    local tools_json="{"
    local first=true
    while IFS= read -r line; do
        # Skip "mcp:" header
        [[ "$line" =~ ^[[:space:]]*mcp: ]] && continue

        # Parse "  key: value" (bool tools)
        if [[ "$line" =~ ^[[:space:]]+([a-zA-Z_]+):[[:space:]]*(true|false) ]]; then
            local key="${BASH_REMATCH[1]}"
            local val="${BASH_REMATCH[2]}"
            [ "$first" = true ] || tools_json+=", "
            first=false
            tools_json+="\"$key\": $val"
        fi
    done <<< "$tools_yaml"

    tools_json+="}"

    # Check if there is an mcp array
    if echo "$tools_yaml" | grep -q "^[[:space:]]*mcp:"; then
        # Build mcp array
        local mcp_items="["
        local mcp_first=true
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]+-[[:space:]]+\"?(.*[^\"])?\"?$ ]]; then
                local item="${BASH_REMATCH[1]}"
                [ "$mcp_first" = true ] || mcp_items+=", "
                mcp_first=false
                mcp_items+="\"$item\""
            fi
        done < <(echo "$tools_yaml" | awk '/^  mcp:/,/^$/' | grep "^-")

        mcp_items+="]"
        # Remove trailing comma/space and close, then add mcp
        tools_json="${tools_json%, }"
        tools_json="${tools_json%,}"
        tools_json="${tools_json%}}"
        [ -z "$tools_json" ] && tools_json="{"
        [ "$tools_json" = "{" ] || tools_json+=", "
        tools_json+="\"mcp\": $mcp_items}"
    fi

    # If tools_json is empty or just "{}", use default
    [ "$tools_json" = "{}" ] || [ "$tools_json" = "{" ] || [ -z "$tools_json" ] && tools_json='{"bash": true, "edit": true, "read": true, "write": true}'

    echo "$tools_json"
}

# ============================================================================
# CONFIGURE OPENCODE.JSON — Non-destructive merge with grimorio_auto_generated flag
# ============================================================================
configure_opencode_merge() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"

    # Create default if missing
    if [ ! -f "$OPENCODE_CONFIG" ]; then
        log "No opencode.json found, creating default..."
        mkdir -p "$(dirname "$OPENCODE_CONFIG")"
        echo '{}' > "$OPENCODE_CONFIG"
    fi

    command_exists jq || { warn "jq not found, opencode.json merge skipped"; return 1; }

    # Validate JSON
    if ! jq empty "$OPENCODE_CONFIG" 2>/dev/null; then
        warn "opencode.json is not valid JSON, backing up and recreating"
        cp "$OPENCODE_CONFIG" "${OPENCODE_CONFIG}.backup.$(date +%Y%m%d%H%M%S)"
        echo '{}' > "$OPENCODE_CONFIG"
    fi

    # Backup before modifying
    cp "$OPENCODE_CONFIG" "${OPENCODE_CONFIG}.backup.$(date +%Y%m%d%H%M%S)"

    log "Merging grimorio configuration into opencode.json..."

    # ------------------------------------------------------------------
    # Build agent entries from agents/ directory
    # ------------------------------------------------------------------
    local new_agents="{}"
    if [ -d "$INSTALL_DIR/agents" ]; then
        for f in "$INSTALL_DIR/agents"/grimorio-*.md; do
            [ -f "$f" ] || continue
            local name
            name=$(basename "$f" .md)
            local desc
            desc=$(sed -n '/^---$/,/^---$/p' "$f" | grep "^description:" | sed 's/^description: "\(.*\)"$/\1/')
            local mode
            mode=$(sed -n '/^---$/,/^---$/p' "$f" | grep "^mode:" | awk '{print $2}')
            local prompt
            prompt=$(awk 'BEGIN{n=0} /^---$/{n++; next} n>=2' "$f")

            local agent_json
            local agent_tools
            agent_tools=$(parse_frontmatter_tools "$f")
            # Validate tools JSON before using it
            if ! echo "$agent_tools" | jq empty 2>/dev/null; then
                warn "  Invalid tools JSON for agent '$name', using defaults"
                agent_tools='{"bash": true, "edit": true, "read": true, "write": true}'
            fi

            agent_json=$(jq -n \
              --arg desc "$desc" \
              --arg mode "$mode" \
              --arg prompt "$prompt" \
              --argjson tools "$agent_tools" \
              '{
                "grimorio_auto_generated": true,
                "description": $desc,
                "mode": $mode,
                "prompt": $prompt,
                "tools": $tools,
                "options": {}
              }')

            if [ $? -ne 0 ] || [ -z "$agent_json" ]; then
                warn "  Failed to build agent '$name' (invalid frontmatter)"
                continue
            fi

            new_agents=$(jq --arg name "$name" --argjson agent "$agent_json" \
              '.[$name] = $agent' <<< "$new_agents" 2>/dev/null)

            if [ $? -eq 0 ]; then
                log "  Agent built: $name"
            else
                warn "  Failed to register agent: $name"
            fi
        done
    fi

    # ------------------------------------------------------------------
    # Build command entry from grimorio-architect prompt
    # ------------------------------------------------------------------
    local command_json="{}"
    if [ -f "$INSTALL_DIR/agents/grimorio-architect.md" ]; then
        local architect_prompt
        architect_prompt=$(awk 'BEGIN{n=0} /^---$/{n++; next} n>=2' "$INSTALL_DIR/agents/grimorio-architect.md")
        command_json=$(jq -n \
          --arg template "$architect_prompt" \
          '{
            "grimorio_auto_generated": true,
            "description": "Generate a complete D&D 5e campaign or one-shot from an idea (executes in main thread)",
            "subtask": false,
            "template": $template
          }')
    fi

    # ------------------------------------------------------------------
    # Build MCP entry
    # ------------------------------------------------------------------
    local mcp_json
    mcp_json=$(jq -n \
      --arg path "$OPENCODE_PLUGIN_DIR/grimorio" \
      '{
        "grimorio_auto_generated": true,
        "command": [$path],
        "type": "local",
        "enabled": true
      }')

    # ------------------------------------------------------------------
    # Merge into opencode.json
    # Strategy: for each grimorio-owned key, if it exists WITHOUT the
    # grimorio_auto_generated flag, preserve it (user customized).
    # Otherwise, replace with our value.
    # ------------------------------------------------------------------
    jq \
      --argjson new_agents "$new_agents" \
      --argjson new_command "$command_json" \
      --argjson new_mcp "$mcp_json" \
      '
      # Merge agents: for each grimorio agent, check ownership
      .agent = (.agent // {})
      | .agent = (
          reduce ($new_agents | to_entries[]) as $entry (.agent;
            if .[$entry.key] and (.[$entry.key].grimorio_auto_generated != true) then
              .
            else
              .[$entry.key] = $entry.value
            end
          )
        )
      # Merge command.grimorio
      | .command = (.command // {})
      | if .command.grimorio and (.command.grimorio.grimorio_auto_generated != true) then
          .
        else
          .command.grimorio = $new_command
        end
      # Merge mcp.grimorio
      | .mcp = (.mcp // {})
      | if .mcp.grimorio and (.mcp.grimorio.grimorio_auto_generated != true) then
          .
        else
          .mcp.grimorio = $new_mcp
        end
      ' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && \
        mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    success "OpenCode configuration merged (user-customized grimorio keys preserved)"
}

# ============================================================================
# CONFIGURE SHELL
# ============================================================================
configure_shell() {
    log "Configuring shell..."
    local SHELL_RC=""
    case "$(basename "$SHELL")" in
        bash) SHELL_RC="${HOME}/.bashrc" ;;
        zsh)  SHELL_RC="${HOME}/.zshrc" ;;
        fish) SHELL_RC="${HOME}/.config/fish/config.fish" ;;
        *)    SHELL_RC="${HOME}/.profile" ;;
    esac

    [ ! -f "$SHELL_RC" ] && return 0

    # Remove old config
    awk '/^# === GRIMORIO CONFIG BEGIN ===$/{skip=1} /^# === GRIMORIO CONFIG END ===$/{skip=0} !skip' \
        "$SHELL_RC" > "${SHELL_RC}.tmp" && mv "${SHELL_RC}.tmp" "$SHELL_RC"

    # Add new config
    cat >> "$SHELL_RC" << 'EOF'

# === GRIMORIO CONFIG BEGIN ===
export PATH="$HOME/.local/bin:$PATH"
export PATH="$HOME/.local/go/bin:$PATH"
# === GRIMORIO CONFIG END ===
EOF

    log "Shell configured: $SHELL_RC"
}

# ============================================================================
# UPDATE MODE — Incremental update via git pull + conditional rebuild
# ============================================================================
do_update() {
    log "Starting incremental update..."

    # Check for prior install — fall back to full install if no meta
    [ -f "$METADATA_FILE" ] || { warn "No prior install found, running full install"; do_install; return; }

    log "Reading install metadata from $METADATA_FILE..."

    # Ensure we're in the install dir
    if [ ! -d "$INSTALL_DIR" ]; then
        warn "Install directory not found, running full install"
        do_install
        return
    fi

    # git pull
    log "Pulling latest changes..."
    cd "$INSTALL_DIR"
    local pull_output
    pull_output=$(git pull origin main 2>&1) || {
        error "Update failed: git pull failed."
        error "Git output: $pull_output"
        if echo "$pull_output" | grep -q "would be overwritten"; then
            warn "You have local changes that conflict with the update."
            warn "Run: cd $INSTALL_DIR && git status"
            warn "Then either commit/stash changes or run: git reset --hard && git clean -fd"
        fi
        return 1
    }

    # Check if anything changed
    local changed_files
    changed_files=$(git diff --name-only HEAD@{1} HEAD 2>/dev/null || echo "")
    if [ -z "$changed_files" ]; then
        success "Already up to date"
        return 0
    fi

    log "Changes detected in: $(echo "$changed_files" | wc -l) file(s)"

    # Determine if rebuild is needed
    local needs_build=false
    if echo "$changed_files" | grep -qE '^(cmd/|go\.mod|go\.sum|internal/|Makefile)'; then
        needs_build=true
        log "Source code changes detected, rebuilding binary..."
    fi

    if [ "$needs_build" = true ]; then
        export PATH="${HOME}/.local/go/bin:$PATH"
        make build || { error "Build failed. Check for compilation errors."; return 1; }

        mkdir -p "$BINARY_DIR"
        cp grimorio "$BINARY_DIR/grimorio.tmp"
        chmod +x "$BINARY_DIR/grimorio.tmp"
        mv -f "$BINARY_DIR/grimorio.tmp" "$BINARY_DIR/grimorio"
        log "Binary updated: $BINARY_DIR/grimorio"

        # Sync to plugin dirs
        for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
            mkdir -p "$plugin_dir"
            # Use temp file + mv to avoid "text file busy" when binary is running
            cp grimorio "$plugin_dir/grimorio.tmp"
            chmod +x "$plugin_dir/grimorio.tmp"
            mv -f "$plugin_dir/grimorio.tmp" "$plugin_dir/grimorio"
            log "Plugin binary synced: $plugin_dir/grimorio"

            # Also sync migrate binary if it exists
            if [ -f migrate-v1-to-v2 ]; then
                cp migrate-v1-to-v2 "$plugin_dir/migrate-v1-to-v2.tmp"
                chmod +x "$plugin_dir/migrate-v1-to-v2.tmp"
                mv -f "$plugin_dir/migrate-v1-to-v2.tmp" "$plugin_dir/migrate-v1-to-v2"
            fi
        done
    fi

    # Copy agents with hash comparison
    if [ -d "$INSTALL_DIR/agents" ]; then
        local current_meta
        current_meta=$(read_meta)
        for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
            mkdir -p "$plugin_dir/agents"
            for f in "$INSTALL_DIR/agents"/grimorio-*.md; do
                [ -f "$f" ] || continue
                local fname
                fname=$(basename "$f")
                local current_hash
                current_hash=$(echo "$current_meta" | jq -r ".agents[\"$fname\"] // \"\"" 2>/dev/null)
                # Strip "sha256-" prefix for comparison
                current_hash="${current_hash#sha256-}"
                copy_if_changed "$f" "$plugin_dir/agents/$fname" "$current_hash"
            done
        done
    fi

    # Copy skills with hash comparison
    if [ -d "$INSTALL_DIR/skills" ]; then
        local current_meta
        current_meta=$(read_meta)
        # Copy to plugin directories
        for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
            mkdir -p "$plugin_dir/skills"
            for skill_dir in "$INSTALL_DIR/skills"/*/; do
                [ -d "$skill_dir" ] || continue
                local skill_name="${skill_dir%/}"; skill_name="${skill_name##*/}"
                local skill_file="$skill_dir/SKILL.md"
                [ -f "$skill_file" ] || continue
                local current_hash
                current_hash=$(echo "$current_meta" | jq -r ".skills[\"$skill_name\"] // \"\"" 2>/dev/null)
                current_hash="${current_hash#sha256-}"
                mkdir -p "$plugin_dir/skills/$skill_name"
                copy_if_changed "$skill_file" "$plugin_dir/skills/$skill_name/SKILL.md" "$current_hash"
            done
        done
        # Also copy to global skills directory (OpenCode loads skills from here)
        for skill_dir_entry in "$INSTALL_DIR/skills"/*/; do
            [ -d "$skill_dir_entry" ] || continue
            local sname="${skill_dir_entry%/}"; sname="${sname##*/}"
            local sfile="$skill_dir_entry/SKILL.md"
            [ -f "$sfile" ] || continue
            local shash
            shash=$(echo "$current_meta" | jq -r ".skills[\"$sname\"] // \"\"" 2>/dev/null)
            shash="${shash#sha256-}"
            mkdir -p "$GLOBAL_SKILLS_DIR/$sname"
            copy_if_changed "$sfile" "$GLOBAL_SKILLS_DIR/$sname/SKILL.md" "$shash"
        done
    fi

    # Copy templates with hash comparison
    local tmpl_dir="$INSTALL_DIR/internal/compiler/templates"
    if [ -d "$tmpl_dir" ]; then
        local current_meta
        current_meta=$(read_meta)
        for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
            mkdir -p "$plugin_dir/internal/compiler/templates"
            for f in "$tmpl_dir"/*.tmpl; do
                [ -f "$f" ] || continue
                local tmpl_name
                tmpl_name=$(basename "$f")
                local current_hash
                current_hash=$(echo "$current_meta" | jq -r ".templates[\"$tmpl_name\"] // \"\"" 2>/dev/null)
                current_hash="${current_hash#sha256-}"
                copy_if_changed "$f" "$plugin_dir/internal/compiler/templates/$tmpl_name" "$current_hash"
            done
        done
    fi

    # Copy commands directory
    copy_commands

    # Merge opencode.json (preserve user customizations)
    configure_opencode_merge

    # Update metadata
    write_meta

    success "Update complete!"
}

# ============================================================================
# PRINT INSTRUCTIONS
# ============================================================================
print_instructions() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}  Grimorio Installation Complete!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    echo -e "${BLUE}What was installed:${NC}"
    echo -e "   • Binary: ${GREEN}$BINARY_DIR/grimorio${NC}"
    echo -e "   • Claude plugin: ${GREEN}$CLAUDE_PLUGIN_DIR${NC}"
    echo -e "   • OpenCode plugin: ${GREEN}$OPENCODE_PLUGIN_DIR${NC}"
    echo -e "   • MCP configured in opencode.json"
    echo -e "   • Command /grimorio configured"
    echo -e "   • 16 agents configured"
    echo -e "   • Shell PATH updated"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo -e "   1. Restart your terminal or run: ${GREEN}source $SHELL_RC${NC}"
    echo -e "   2. Use: ${GREEN}/grimorio <idea>${NC}"
    echo -e "   3. Or in opencode: ${GREEN}/grimorio <idea>${NC}"
    echo ""
    echo -e "${BLUE}Available MCP tools:${NC}"
    echo -e "   • create_campaign, generate_adventure_bible"
    echo -e "   • save_introduction, save_setting_guide, save_areas"
    echo -e "   • save_npcs, save_bestiary, save_encounters"
    echo -e "   • save_maps, save_characters, save_appendices"
    echo -e "   • generate_image, generate_map, generate_divider"
    echo -e "   • compile_pdf, get_template"
    echo -e "   • generate_character, get_character, list_characters, save_characters"
    echo -e "   • generate_character_hooks"
    echo -e "   • create_personal_quest, update_quest_status, list_quests"
    echo -e "   • validate_canon, check_consistency, process_consistency_gate"
    echo -e "   • update_narrative_state, evaluate_consequences"
    echo -e "   • update_faction_reputation, generate_random_tables, generate_handouts"
    echo -e "   • generate_session_prep, generate_flowchart"
    echo -e "   • grimorio_generate_prologue"
    echo -e "   • grimorio_generate_tactics, grimorio_get_tactics"
    echo -e "   • grimorio_generate_xp_table, grimorio_track_party_progress"
    echo -e "   • grimorio_generate_player_map, grimorio_export_handout"
    echo ""
    echo -e "${YELLOW}Need help?${NC} Check: ${GREEN}$INSTALL_DIR/README.md${NC}"
    echo ""
}

# ============================================================================
# FULL INSTALL — Fresh installation from scratch
# ============================================================================
do_install() {
    echo -e "${GREEN}"
    echo "  ____      _                      _"
    echo " / ___|_ __(_)_ __ ___   ___  _ __(_) ___"
    echo "| |_| | '__| | '_ \` _ \ / _ \| '__| |/ _ \\"
    echo "| |_| | |  | | | | | | | (_) | |  | | (_) |"
    echo " \____|_|  |_|_| |_| |_|\___/|_|  |_|\___/"
    echo -e "${NC}"
    echo "       D&D One-shot & Campaign Generator"
    echo ""

    log "Starting full installation..."

    # Step 1: Clean everything
    clean_installation

    # Step 2: Detect platform
    detect_platform

    # Step 3: Install Go
    install_go

    # Step 4: Install wkhtmltopdf
    install_wkhtmltopdf

    # Step 5: Setup repository
    setup_repo

    # Step 6: Build binaries
    build_binary

    # Step 7: Migrate campaigns
    migrate_existing_campaigns

    # Step 8: Setup plugin
    setup_plugin

    # Step 9: Copy commands directory
    copy_commands

    # Step 10: Merge opencode.json (reads from agents/ + templates)
    configure_opencode_merge

    # Step 11: Configure shell
    configure_shell

    # Step 12: Write install metadata
    write_meta

    success "Installation complete!"
    print_instructions
}

main() {
    # Parse --update flag
    if [ "$1" = "--update" ]; then
        shift
        do_update "$@"
        return $?
    fi

    do_install
}

main "$@"
