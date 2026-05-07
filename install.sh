#!/bin/bash
set -e

# Grimorio - One Command Installer
# Usage: curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash

REPO_URL="https://github.com/pauvalls/grimorio"
INSTALL_DIR="${HOME}/.local/share/grimorio"
CLAUDE_PLUGIN_DIR="${HOME}/.claude/plugins/grimorio"
OPENCODE_PLUGIN_DIR="${HOME}/.config/opencode/plugins/grimorio"
BINARY_DIR="${HOME}/.local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[Grimorio]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    case "$OS" in
        linux)
            WKHTMLTOPDF_URL="https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltopdf_0.12.6.1-3.${OS}_${ARCH}.deb"
            ;;
        darwin)
            WKHTMLTOPDF_URL="https://github.com/wkhtmltopdf/packaging/releases/download/0.12.6.1-3/wkhtmltopdf-0.12.6.1-3.macos-cocoa.pkg"
            ;;
        *)
            warn "Automatic wkhtmltopdf installation not supported for $OS."
            WKHTMLTOPDF_URL=""
            ;;
    esac
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

install_go() {
    if command_exists go; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        log "Go found: $GO_VERSION"
        return 0
    fi

    log "Go not found. Installing Go..."

    GO_VERSION="1.23.4"
    GO_TARBALL="go${GO_VERSION}.${OS}-${ARCH}.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TARBALL}"

    mkdir -p "${HOME}/.local"
    curl -L "$GO_URL" -o "/tmp/${GO_TARBALL}"
    tar -C "${HOME}/.local" -xzf "/tmp/${GO_TARBALL}"

    export PATH="${HOME}/.local/go/bin:$PATH"

    if ! command_exists go; then
        error "Go installation failed"
    fi

    success "Go installed successfully"
}

install_wkhtmltopdf() {
    if command_exists wkhtmltopdf; then
        log "wkhtmltopdf found: $(wkhtmltopdf --version | head -n1)"
        return 0
    fi

    if [ -z "$WKHTMLTOPDF_URL" ]; then
        warn "Please install wkhtmltopdf manually for your OS"
        warn "macOS: brew install --cask wkhtmltopdf"
        warn "Ubuntu/Debian: sudo apt install wkhtmltopdf"
        return 1
    fi

    log "Installing wkhtmltopdf..."

    case "$OS" in
        linux)
            if command_exists dpkg; then
                curl -L "$WKHTMLTOPDF_URL" -o /tmp/wkhtmltopdf.deb
                dpkg -x /tmp/wkhtmltopdf.deb /tmp/wkhtmltopdf
                mkdir -p "${HOME}/.local/bin"
                cp /tmp/wkhtmltopdf/usr/local/bin/wkhtmltopdf "${HOME}/.local/bin/"
                chmod +x "${HOME}/.local/bin/wkhtmltopdf"
            elif command_exists rpm; then
                warn "RPM-based distros: please install wkhtmltopdf manually"
            else
                warn "Unknown package manager. Please install wkhtmltopdf manually."
            fi
            ;;
        darwin)
            curl -L "$WKHTMLTOPDF_URL" -o /tmp/wkhtmltopdf.pkg
            log "Please run: sudo installer -pkg /tmp/wkhtmltopdf.pkg -target /"
            ;;
    esac

    export PATH="${HOME}/.local/bin:$PATH"

    if command_exists wkhtmltopdf; then
        success "wkhtmltopdf installed"
    else
        warn "wkhtmltopdf installation may require manual steps"
    fi
}

setup_repo() {
    log "Setting up Grimorio repository..."

    if [ -d "$INSTALL_DIR" ]; then
        log "Updating existing installation..."
        cd "$INSTALL_DIR"
        git fetch origin main 2>/dev/null || true
        git reset --hard origin/main 2>/dev/null || true
        log "Updated to latest version"
    else
        log "Cloning repository..."
        if command_exists git; then
            git clone "$REPO_URL" "$INSTALL_DIR"
        else
            log "Git not found. Downloading source archive..."
            curl -L "${REPO_URL}/archive/refs/heads/main.tar.gz" -o /tmp/grimorio.tar.gz
            mkdir -p "$INSTALL_DIR"
            tar -xzf /tmp/grimorio.tar.gz -C "$INSTALL_DIR" --strip-components=1
        fi
    fi

    success "Repository ready at $INSTALL_DIR"
}

build_binary() {
    log "Building Grimorio binary..."

    cd "$INSTALL_DIR"
    export PATH="${HOME}/.local/go/bin:$PATH"

    go build -o grimorio ./cmd/grimorio
    go build -o migrate-v1-to-v2 ./cmd/migrate-v1-to-v2

    mkdir -p "$BINARY_DIR"
    cp grimorio "$BINARY_DIR/"
    cp migrate-v1-to-v2 "$BINARY_DIR/"
    chmod +x "$BINARY_DIR/grimorio"
    chmod +x "$BINARY_DIR/migrate-v1-to-v2"

    success "Binary built and installed to $BINARY_DIR/grimorio"
    success "Migration tool built and installed to $BINARY_DIR/migrate-v1-to-v2"
}

setup_plugin() {
    # Install for Claude Code
    log "Setting up Claude Code plugin..."
    mkdir -p "$CLAUDE_PLUGIN_DIR"

    cp -rf "$INSTALL_DIR/.claude-plugin" "$CLAUDE_PLUGIN_DIR/"
    cp -rf "$INSTALL_DIR/commands" "$CLAUDE_PLUGIN_DIR/"
    cp -rf "$INSTALL_DIR/agents" "$CLAUDE_PLUGIN_DIR/"
    cp -rf "$INSTALL_DIR/skills" "$CLAUDE_PLUGIN_DIR/"
    cp -f "$BINARY_DIR/grimorio" "$CLAUDE_PLUGIN_DIR/"
    cp -f "$BINARY_DIR/migrate-v1-to-v2" "$CLAUDE_PLUGIN_DIR/"

    # Always update commands/grimorio.md to latest version
    if [ -f "$INSTALL_DIR/commands/grimorio.md" ]; then
        cp -f "$INSTALL_DIR/commands/grimorio.md" "$CLAUDE_PLUGIN_DIR/commands/"
    fi

    # Copy new cartographer agent if it exists in repo but not in plugin
    if [ -f "$INSTALL_DIR/agents/grimorio-cartographer.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-cartographer.md" "$CLAUDE_PLUGIN_DIR/agents/"
    fi

    # Copy artist agent
    if [ -f "$INSTALL_DIR/agents/grimorio-artist.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-artist.md" "$CLAUDE_PLUGIN_DIR/agents/"
    fi

    # Fix .mcp.json for Claude Code (uses ${CLAUDE_PLUGIN_ROOT})
    cat > "$CLAUDE_PLUGIN_DIR/.mcp.json" << 'EOF'
{
  "grimorio": {
    "command": "${CLAUDE_PLUGIN_ROOT}/grimorio",
    "args": [],
    "env": {}
  }
}
EOF

    success "Plugin installed to $CLAUDE_PLUGIN_DIR"

    # Install for OpenCode
    log "Setting up OpenCode plugin..."
    mkdir -p "$OPENCODE_PLUGIN_DIR"

    cp -rf "$INSTALL_DIR/.claude-plugin" "$OPENCODE_PLUGIN_DIR/"
    cp -rf "$INSTALL_DIR/commands" "$OPENCODE_PLUGIN_DIR/"
    cp -rf "$INSTALL_DIR/agents" "$OPENCODE_PLUGIN_DIR/"
    cp -rf "$INSTALL_DIR/skills" "$OPENCODE_PLUGIN_DIR/"
    cp -f "$BINARY_DIR/grimorio" "$OPENCODE_PLUGIN_DIR/"
    cp -f "$BINARY_DIR/migrate-v1-to-v2" "$OPENCODE_PLUGIN_DIR/"

    # Always update commands/grimorio.md to latest version
    if [ -f "$INSTALL_DIR/commands/grimorio.md" ]; then
        cp -f "$INSTALL_DIR/commands/grimorio.md" "$OPENCODE_PLUGIN_DIR/commands/"
    fi

    # Copy new cartographer agent if it exists in repo but not in plugin
    if [ -f "$INSTALL_DIR/agents/grimorio-cartographer.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-cartographer.md" "$OPENCODE_PLUGIN_DIR/agents/"
    fi

    # Copy artist agent
    if [ -f "$INSTALL_DIR/agents/grimorio-artist.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-artist.md" "$OPENCODE_PLUGIN_DIR/agents/"
    fi

    # Fix .mcp.json for OpenCode (uses absolute path, not ${CLAUDE_PLUGIN_ROOT})
    cat > "$OPENCODE_PLUGIN_DIR/.mcp.json" << EOF
{
  "grimorio": {
    "command": "$OPENCODE_PLUGIN_DIR/grimorio",
    "args": [],
    "env": {}
  }
}
EOF

    success "Plugin installed to $OPENCODE_PLUGIN_DIR"

    # Clean up stale grimorio.md files from old installation locations
    log "Cleaning up stale template files..."
    local stale_files=(
        "${HOME}/.config/opencode/commands/grimorio.md"
        "${HOME}/.local/share/grimorio/.opencode/commands/grimorio.md"
        "${HOME}/.local/share/grimorio/commands/grimorio.md"
        "${HOME}/Grimorio/.opencode/commands/grimorio.md"
    )
    for stale in "${stale_files[@]}"; do
        if [ -f "$stale" ]; then
            rm -f "$stale"
            log "Removed stale file: $stale"
        fi
    done

    # Configure grimorio in opencode.json for versions that don't support .mcp.json
    configure_opencode_mcp
}

configure_shell() {
    log "Configuring shell..."

    SHELL_RC=""
    case "$(basename "$SHELL")" in
        bash) SHELL_RC="${HOME}/.bashrc" ;;
        zsh)  SHELL_RC="${HOME}/.zshrc" ;;
        fish) SHELL_RC="${HOME}/.config/fish/config.fish" ;;
        *)    SHELL_RC="${HOME}/.profile" ;;
    esac

    if [ -n "$SHELL_RC" ] && [ -f "$SHELL_RC" ]; then
        # 1. Clean up legacy standalone lines (re-install compatibility)
        sed -i '/^# Grimorio$/d' "$SHELL_RC"
        sed -i '/^export PATH="\$HOME\/\.local\/go\/bin:\$PATH"$/d' "$SHELL_RC"
        sed -i '/^export PATH="\$HOME\/\.local\/bin:\$PATH"$/d' "$SHELL_RC"

        # 2. Clean up any existing marked block to prevent duplicates
        awk '
            /^# === GRIMORIO CONFIG BEGIN ===$/ { in_block=1; next }
            /^# === GRIMORIO CONFIG END ===$/   { in_block=0; next }
            !in_block { print }
        ' "$SHELL_RC" > "${SHELL_RC}.tmp" && mv "${SHELL_RC}.tmp" "$SHELL_RC"

        # 3. Add fresh marked block with both paths
        cat >> "$SHELL_RC" << 'EOF'

# === GRIMORIO CONFIG BEGIN ===
export PATH="$HOME/.local/bin:$PATH"
export PATH="$HOME/.local/go/bin:$PATH"
# === GRIMORIO CONFIG END ===
EOF

        log "Shell configured at $SHELL_RC"
    fi

    success "Shell configured"
}

configure_opencode_mcp() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"

    if [ ! -f "$OPENCODE_CONFIG" ]; then
        log "No opencode.json found, skipping MCP config"
        return 0
    fi

    log "Configuring grimorio MCP in opencode.json..."

    # Always update (not just add) to ensure latest config
    if command_exists jq; then
        jq '.mcp.grimorio = {
            "command": ["'"$OPENCODE_PLUGIN_DIR/grimorio"'"],
            "type": "local",
            "enabled": true
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio MCP configured in opencode.json"
    else
        warn "jq not found. Please manually add grimorio to your opencode.json mcp section:"
        warn '{
  "grimorio": {
    "command": ["'"$OPENCODE_PLUGIN_DIR/grimorio"'"],
    "type": "local",
    "enabled": true
  }
}'
        return 1
    fi
}

configure_opencode_command() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"

    if [ ! -f "$OPENCODE_CONFIG" ]; then
        log "No opencode.json found, skipping command config"
        return 0
    fi

    # Clean up deprecated orchestrator agent from previous installations
    if command_exists jq; then
        jq 'del(.agent["grimorio-orchestrator"])' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" 2>/dev/null && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG" || true
    fi

    # Always update agent (not just add) to ensure latest prompt
    log "Configuring grimorio-architect agent..."
    if command_exists jq; then
        jq '.agent["grimorio-architect"] = {
            "description": "Expert Dungeon Master agent for D&D 5e campaign generation",
            "mode": "primary",
            "prompt": "You are an expert Dungeon Master and campaign designer. Your job is to:\n1. Ask the user clarifying questions about their campaign idea (level, tone, duration, name)\n2. After gathering all requirements, create the campaign structure and orchestrate ALL phases directly via delegate and MCP tools\n3. Report progress to the user after each phase\n4. Report the final result\n\nDO NOT edit files in the main thread. Always use delegate for content generation.",
            "tools": {
                "bash": true,
                "delegate": true,
                "edit": true,
                "read": true,
                "write": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-architect agent configured"
    fi

    # Configure grimorio-artist subagent
    log "Configuring grimorio-artist agent..."
    if command_exists jq; then
        jq '.agent["grimorio-artist"] = {
            "description": "Campaign artist — prepares image specs and updates markdown references",
            "mode": "subagent",
            "prompt": "You are the Grimorio Artist. Prepare image batch specifications and update markdown references.\n\nPhase A: Read NPCs, bestiary, and acts. Create batch-spec.json with all image prompts.\nPhase B: After images are generated, update all markdown files with ![alt](assets/filename.png) references.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-artist agent configured"
    fi

    # Configure grimorio-cartographer subagent
    log "Configuring grimorio-cartographer agent..."
    if command_exists jq; then
        jq '.agent["grimorio-cartographer"] = {
            "description": "Campaign cartographer — generates SVG battle maps and decorative dividers",
            "mode": "subagent",
            "prompt": "You are the Grimorio Cartographer. Generate ALL SVG assets for a campaign: battle maps, decorative dividers, and stat block borders. Use generate_map and generate_divider tools. Reference all SVGs in markdown files.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-cartographer agent configured"
    fi

    # Configure grimorio-lore subagent
    log "Configuring grimorio-lore agent..."
    if command_exists jq; then
        jq '.agent["grimorio-lore"] = {
            "description": "Campaign lore writer — world backstory, setting, history, and atmosphere",
            "mode": "subagent",
            "prompt": "You are the Grimorio Lore Master. Generate world lore, backstory, setting, and atmosphere for a D&D 5e campaign. Use grimorio_save_lore tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-lore agent configured"
    fi

    # Configure grimorio-npc subagent
    log "Configuring grimorio-npc agent..."
    if command_exists jq; then
        jq '.agent["grimorio-npc"] = {
            "description": "Campaign NPC designer — characters, factions, and social relationships",
            "mode": "subagent",
            "prompt": "You are the Grimorio NPC Designer. Generate NPCs, factions, and social entities for a D&D 5e campaign. Use grimorio_save_npcs tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-npc agent configured"
    fi

    # Configure grimorio-bestiary subagent
    log "Configuring grimorio-bestiary agent..."
    if command_exists jq; then
        jq '.agent["grimorio-bestiary"] = {
            "description": "Campaign bestiary designer — monster stat blocks, abilities, and tactics",
            "mode": "subagent",
            "prompt": "You are the Grimorio Bestiary Designer. Generate monsters, creatures, and stat blocks for a D&D 5e campaign. Use grimorio_save_bestiary tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-bestiary agent configured"
    fi

    # Configure grimorio-encounters subagent
    log "Configuring grimorio-encounters agent..."
    if command_exists jq; then
        jq '.agent["grimorio-encounters"] = {
            "description": "Campaign encounter designer — combat, social, exploration challenges",
            "mode": "subagent",
            "prompt": "You are the Grimorio Encounter Designer. Generate balanced encounters and challenges for a D&D 5e campaign. Use grimorio_save_encounters tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-encounters agent configured"
    fi

    # Configure grimorio-acts subagent
    log "Configuring grimorio-acts agent..."
    if command_exists jq; then
        jq '.agent["grimorio-acts"] = {
            "description": "Campaign story architect — narrative acts, scenes, and session structure",
            "mode": "subagent",
            "prompt": "You are the Grimorio Story Architect. Generate narrative acts and scenes for a D&D 5e campaign. Use grimorio_save_act tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-acts agent configured"
    fi

    # Configure grimorio-quests subagent
    log "Configuring grimorio-quests agent..."
    if command_exists jq; then
        jq '.agent["grimorio-quests"] = {
            "description": "Campaign quest designer — personal quests, side missions, narrative hooks",
            "mode": "subagent",
            "prompt": "You are the Grimorio Quest Designer. Generate personal quests, side missions, and narrative hooks for a D&D 5e campaign. Use grimorio_create_personal_quest tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-quests agent configured"
    fi

    # Configure grimorio-maps subagent
    log "Configuring grimorio-maps agent..."
    if command_exists jq; then
        jq '.agent["grimorio-maps"] = {
            "description": "Campaign map describer — location details, zone breakdowns, scene layouts",
            "mode": "subagent",
            "prompt": "You are the Grimorio Map Describer. Generate location descriptions and zone breakdowns for a D&D 5e campaign. Use grimorio_save_maps tool.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-maps agent configured"
    fi

    # Configure grimorio-characters subagent
    log "Configuring grimorio-characters agent..."
    if command_exists jq; then
        jq '.agent["grimorio-characters"] = {
            "description": "Campaign character builder — pre-generated player character sheets and backstories",
            "mode": "subagent",
            "prompt": "You are the Grimorio Character Builder. Generate pre-generated player characters with backstories for a D&D 5e campaign. Write to characters/ directory.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-characters agent configured"
    fi

    # Always update command (not just add) to ensure latest template with image generation
    log "Configuring grimorio command..."
    if command_exists jq; then
        # Create template in temp file to avoid bash parenthesis issues
        local TEMPLATE_FILE=$(mktemp)
        cat > "$TEMPLATE_FILE" << 'TEMPLATE_EOF'
Generate a D&D 5e campaign or one-shot from the user's idea.

## IMPORTANT: Use the grimorio-architect agent. It handles everything end-to-end.

## Workflow (followed by grimorio-architect)

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
4. Desired tone? (heroic, dark, humorous, political intrigue)
5. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool `create_campaign` to create the structure.

### Phase 3-10: End-to-End Orchestration (sequential batches)
The architect follows strict batch ordering — each batch waits for the previous:

- **Batch 1** (parallel): NPCs, bestiary, maps
- **Batch 2** (parallel): lore, quests, encounters, characters
- **Batch 3** (parallel): SVG maps, acts (needs ALL prior content)
- **Sequential**: artist batch-spec (cover + NPCs + scenes + monsters) → generate images (1x1, retry missing) → update ALL references → PDF

The architect reports progress to the user after each phase.

### Final: Report
After completion, report to the user:
- Where the PDF was saved
- What content was generated
- Any issues encountered

**DO NOT launch subagents from the command thread — the architect manages all delegation internally.**
TEMPLATE_EOF

        # Read template and escape for JSON
        local TEMPLATE_JSON
        TEMPLATE_JSON=$(cat "$TEMPLATE_FILE" | jq -Rs '.')
        rm -f "$TEMPLATE_FILE"

        jq --argjson template "$TEMPLATE_JSON" '.command.grimorio = {
            "description": "Generate a complete D&D 5e campaign or one-shot from an idea",
            "agent": "grimorio-architect",
            "subtask": false,
            "template": $template
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio command configured"
    else
        warn "jq not found. Please manually configure grimorio in opencode.json"
        return 1
    fi
}

migrate_existing_campaigns() {
    local CAMPAIGNS_DIR="${HOME}/campaigns"
    
    if [ ! -d "$CAMPAIGNS_DIR" ]; then
        return 0
    fi

    # Check if there are any campaigns without canon.json (v1 format)
    local needs_migration=false
    for campaign_dir in "$CAMPAIGNS_DIR"/*/; do
        if [ -d "$campaign_dir" ] && [ ! -f "$campaign_dir/canon.json" ]; then
            needs_migration=true
            break
        fi
    done

    if [ "$needs_migration" = true ]; then
        log "Found existing v1 campaigns. Running migration..."
        if [ -f "$BINARY_DIR/migrate-v1-to-v2" ]; then
            "$BINARY_DIR/migrate-v1-to-v2" "$CAMPAIGNS_DIR" || warn "Migration had issues, but installation continues"
            success "Migration complete. Backups saved as .v1-backup/"
        else
            warn "Migration tool not found. Run manually: migrate-v1-to-v2 ~/campaigns"
        fi
    fi
}

print_instructions() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         Grimorio Installed Successfully!                   ║${NC}"
    echo -e "${GREEN}║         D&D One-shot & Campaign Generator                  ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}What's next:${NC}"
    echo ""
    echo -e "1. ${YELLOW}Restart your terminal${NC} or run:"
    echo -e "   ${GREEN}source ~/.bashrc${NC} (or ~/.zshrc)"
    echo ""
    echo -e "2. ${YELLOW}Plugin installed for:${NC}"
    echo -e "   • Claude Code → ${GREEN}${CLAUDE_PLUGIN_DIR}${NC}"
    echo -e "   • OpenCode    → ${GREEN}${OPENCODE_PLUGIN_DIR}${NC}"
    echo ""
    echo -e "3. ${YELLOW}OpenCode auto-configured:${NC}"
    echo -e "   • MCP server  → ${GREEN}~/.config/opencode/opencode.json${NC} (mcp section)"
    echo -e "   • Architect   → grimorio-architect (orchestrates all phases via delegate)"
    echo -e "   • Content agents (delegated by architect):"
    echo -e "     - grimorio-lore          (world backstory & atmosphere)"
    echo -e "     - grimorio-npc           (NPCs & factions)"
    echo -e "     - grimorio-bestiary      (monster stat blocks)"
    echo -e "     - grimorio-encounters    (combat & exploration challenges)"
    echo -e "     - grimorio-maps          (location & zone descriptions)"
    echo -e "     - grimorio-acts          (narrative acts & scenes)"
    echo -e "     - grimorio-quests        (personal quests & side missions)"
    echo -e "     - grimorio-characters    (pre-generated character sheets)"
    echo -e "   • Artist      → grimorio-artist (image specs + reference updates)"
    echo -e "   • Cartographer→ grimorio-cartographer (SVG maps + dividers)"
    echo -e "   • Command     → /grimorio (single delegate, zero polling)"
    echo ""
    echo -e "4. ${YELLOW}Generate your first campaign:${NC}"
    echo -e "   Type in OpenCode or Claude Code:"
    echo -e "   ${GREEN}/grimorio A sunken city where the nobles are aquatic vampires${NC}"
    echo ""
    echo -e "5. ${YELLOW}Campaigns are saved to:${NC}"
    echo -e "   ${GREEN}~/campaigns/${NC}"
    echo ""
    echo -e "6. ${YELLOW}Image generation:${NC}"
    echo -e "   • SVG maps & dividers → ${GREEN}100% local, no API key needed${NC}"
    echo -e "   • AI images         → FREE with automatic fallback:"
    echo -e "     - Pollinations.ai (primary)"
    echo -e "     - Raphael AI (raphael.app, fallback)"
    echo -e "   • DALL-E (optional) → Set OPENAI_API_KEY for higher quality"
    echo ""
    echo -e "7. ${YELLOW}Narrative Coherence Tools (NEW v2.0):${NC}"
    echo -e "   • generate_adventure_bible → Creates canon.json with facts, entities, rules"
    echo -e "   • validate_canon → Validates content against canon (prevents NPC resurrections!)"
    echo -e "   • update_narrative_state → Track session state (clues, quests, deaths)"
    echo -e "   • check_consistency → Full campaign validation before PDF"
    echo ""
    echo -e "8. ${YELLOW}Update grimorio later:${NC}"
    echo -e "   Just re-run: ${GREEN}curl -sSL ${REPO_URL}/raw/main/install.sh | bash${NC}"
    echo ""
    echo -e "9. ${YELLOW}Migration from v1:${NC}"
    echo -e "   If you have old campaigns: ${GREEN}migrate-v1-to-v2 ~/campaigns${NC}"
    echo ""
    echo -e "${BLUE}Manual usage (without AI tools):${NC}"
    echo -e "   ${GREEN}grimorio${NC} - Runs the MCP server"
    echo -e "   ${GREEN}migrate-v1-to-v2 ~/campaigns${NC} - Migrate old campaigns"
    echo ""
    echo -e "${YELLOW}Need help?${NC} Check the README at: ${GREEN}${INSTALL_DIR}/README.md${NC}"
    echo ""
}

main() {
    echo -e "${GREEN}"
    echo -e "  ____      _                      _"
    echo -e " / ___|_ __(_)_ __ ___   ___  _ __(_) ___"
    echo -e "| |  _| '__| | '_ \` _ \ / _ \| '__| |/ _ \\"
    echo -e "| |_| | |  | | | | | | | (_) | |  | | (_) |"
    echo -e " \____|_|  |_|_| |_| |_|\___/|_|  |_|\___/"
    echo -e "${NC}"
    echo -e "       D&D One-shot & Campaign Generator"
    echo ""

    log "Starting installation..."

    detect_platform
    install_go
    install_wkhtmltopdf
    setup_repo
    build_binary
    migrate_existing_campaigns
    setup_plugin
    configure_shell
    configure_opencode_command

    success "Installation complete!"
    print_instructions
}

main "$@"
