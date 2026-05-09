#!/bin/bash
set -e

# Grimorio - Clean Installer v2
# Complete MCP installation - removes old, installs fresh
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
NC='\033[0m'

log() { echo -e "${BLUE}[Grimorio]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }

command_exists() { command -v "$1" >/dev/null 2>&1; }

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
    git clone --depth 1 "$REPO_URL" "$INSTALL_DIR" 2>/dev/null || \
        curl -sSL "${REPO_URL}/archive/refs/heads/main.tar.gz" | tar -xzf - -C "$INSTALL_DIR" --strip-components=1
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
        cp -f "$BINARY_DIR"/{grimorio,migrate-v1-to-v2} "$plugin_dir/"

        # Copy agents (if exist)
        if [ -d "$INSTALL_DIR/agents" ]; then
            for f in "$INSTALL_DIR/agents"/grimorio-*.md; do
                [ -f "$f" ] && cp -f "$f" "$plugin_dir/agents/"
            done
        fi

        # Copy skills (if exist)
        if [ -d "$INSTALL_DIR/skills" ]; then
            for f in "$INSTALL_DIR/skills"/grimorio-*.md; do
                [ -f "$f" ] && cp -f "$f" "$plugin_dir/skills/"
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
# CONFIGURE MCP IN OPENCODE.JSON
# ============================================================================
configure_opencode_mcp() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"
    [ ! -f "$OPENCODE_CONFIG" ] && { log "No opencode.json, skipping MCP config"; return 0; }

    log "Configuring MCP in opencode.json..."
    command_exists jq || { warn "jq not found, manual config required"; return 1; }

    jq '.mcp.grimorio = {
        "command": ["'"$OPENCODE_PLUGIN_DIR/grimorio"'"],
        "type": "local",
        "enabled": true
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && \
        mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    success "MCP configured"
}

# ============================================================================
# CONFIGURE COMMAND IN OPENCODE.JSON
# ============================================================================
configure_opencode_command() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"
    [ ! -f "$OPENCODE_CONFIG" ] && return 0

    log "Configuring grimorio command..."
    command_exists jq || return 1

    # Create template
    local TEMPLATE_FILE=$(mktemp)
    cat > "$TEMPLATE_FILE" << 'TEMPLATE_EOF'
Generate a D&D 5e campaign or one-shot from the user's idea.

## EXECUTION MODE: Main Thread Orchestration

**You are the orchestrator.** Execute this workflow directly in the main thread. Use MCP tools and delegate to sub-agents as specified below.

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. **Campaign idea / brief description?** (What story do you want to tell? 2-3 sentences describing the main plot)
4. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
5. Desired tone? (heroic, dark, humorous, political intrigue)
6. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool `create_campaign` to create the structure.

### Phase 3-13: End-to-End Orchestration (sequential batches)
Follow strict batch ordering — each batch waits for the previous:

- **Batch 1** (parallel delegate): NPCs, bestiary, maps → Consistency Gate
- **Batch 2** (parallel delegate): lore, quests, encounters, characters → Consistency Gate → Update Narrative State
- **Batch 3** (parallel delegate): SVG maps, areas → Consistency Gate
- **Phase 6**: Artist batch-spec (cover + NPCs + scenes + monsters)
- **Phase 7**: Generate AI images (1x1 sequential, retry missing)
- **Phase 8**: Update ALL markdown references
- **Phase 9**: Living World tools (factions, random tables, handouts, consequences) → Consistency Gate
- **Phase 10**: DM Experience tools (session prep, flowchart)
- **Phase 11**: Final consistency check
- **Phase 12**: Compile PDF (embeds all images + flowchart)
- **Phase 13**: Final report

Report progress to the user after each phase.

### Final: Report
After completion, report to the user:
- Where the PDF was saved
- What content was generated
- Any issues encountered

**Use delegate for content generation sub-agents. Execute orchestration logic in main thread.**
TEMPLATE_EOF

    local TEMPLATE_JSON
    TEMPLATE_JSON=$(cat "$TEMPLATE_FILE" | jq -Rs '.')
    rm -f "$TEMPLATE_FILE"

    jq --argjson template "$TEMPLATE_JSON" '.command.grimorio = {
        "description": "Generate a complete D&D 5e campaign or one-shot from an idea (executes in main thread)",
        "subtask": false,
        "template": $template
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && \
        mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    success "Command configured"
}

# ============================================================================
# CONFIGURE ALL AGENTS IN OPENCODE.JSON
# ============================================================================
configure_opencode_agents() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"
    [ ! -f "$OPENCODE_CONFIG" ] && return 0
    command_exists jq || return 1

    log "Configuring grimorio-architect agent..."
    jq '.agent["grimorio-architect"] = {
        "description": "Expert Dungeon Master agent for D&D 5e campaign generation",
        "mode": "primary",
        "prompt": "You are an expert Dungeon Master and campaign designer. Your job is to:\n1. Ask the user clarifying questions about their campaign idea (level, tone, duration, name)\n2. After gathering all requirements, create the campaign structure and orchestrate ALL phases directly via delegate and MCP tools\n3. Report progress to the user after each phase\n4. Report the final result\n\nDO NOT edit files in the main thread. Always use delegate for content generation.",
        "tools": {"bash": true, "delegate": true, "edit": true, "read": true, "write": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-artist agent..."
    jq '.agent["grimorio-artist"] = {
        "description": "Campaign artist — prepares image specs and updates markdown references",
        "mode": "subagent",
        "prompt": "You are the Grimorio Artist. Prepare image batch specifications and update markdown references.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-cartographer agent..."
    jq '.agent["grimorio-cartographer"] = {
        "description": "Campaign cartographer — generates SVG battle maps and decorative dividers",
        "mode": "subagent",
        "prompt": "You are the Grimorio Cartographer. Generate ALL SVG assets for a campaign: battle maps, decorative dividers, and stat block borders.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-lore agent..."
    jq '.agent["grimorio-lore"] = {
        "description": "Campaign lore writer — world backstory, setting, history, and atmosphere",
        "mode": "subagent",
        "prompt": "You are the Grimorio Lore Master. Generate world lore, backstory, setting, and atmosphere for a D&D 5e campaign.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-npc agent..."
    jq '.agent["grimorio-npc"] = {
        "description": "Campaign NPC designer — characters, factions, and social relationships",
        "mode": "subagent",
        "prompt": "You are the Grimorio NPC Designer. Generate NPCs, factions, and social entities for a D&D 5e campaign.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-bestiary agent..."
    jq '.agent["grimorio-bestiary"] = {
        "description": "Campaign bestiary designer — monster stat blocks, abilities, and tactics",
        "mode": "subagent",
        "prompt": "You are the Grimorio Bestiary Designer. Generate monsters, creatures, and stat blocks for a D&D 5e campaign.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-encounters agent..."
    jq '.agent["grimorio-encounters"] = {
        "description": "Campaign encounter designer — combat, social, exploration challenges",
        "mode": "subagent",
        "prompt": "You are the Grimorio Encounter Designer. Generate balanced encounters and challenges for a D&D 5e campaign.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-areas agent..."
    jq '.agent["grimorio-areas"] = {
        "description": "Campaign areas designer — numbered playable areas (10-15 per act, WotC format)",
        "mode": "subagent",
        "prompt": "You are the Grimorio Areas Designer. Generate numbered playable areas for a D&D 5e campaign in WotC format.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-quests agent..."
    jq '.agent["grimorio-quests"] = {
        "description": "Campaign quest designer — personal quests, side missions, narrative hooks",
        "mode": "subagent",
        "prompt": "You are the Grimorio Quest Designer. Generate personal quests, side missions, and narrative hooks.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-maps agent..."
    jq '.agent["grimorio-maps"] = {
        "description": "Campaign map describer — location details, zone breakdowns, scene layouts",
        "mode": "subagent",
        "prompt": "You are the Grimorio Map Describer. Generate location descriptions and zone breakdowns.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-characters agent..."
    jq '.agent["grimorio-characters"] = {
        "description": "Campaign character builder — pre-generated player character sheets",
        "mode": "subagent",
        "prompt": "You are the Grimorio Character Builder. Generate pre-generated player characters with backstories.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-narrative-custodian agent..."
    jq '.agent["grimorio-narrative-custodian"] = {
        "description": "Campaign narrative custodian — validates canon consistency",
        "mode": "subagent",
        "prompt": "You are the Grimorio Narrative Custodian. Validate campaign content for narrative coherence.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-introduction agent..."
    jq '.agent["grimorio-introduction"] = {
        "description": "Campaign introduction writer — overview and hooks",
        "mode": "subagent",
        "prompt": "You are the Grimorio Introduction Writer. Generate campaign introduction and overview.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-setting-guide agent..."
    jq '.agent["grimorio-setting-guide"] = {
        "description": "Campaign setting guide writer — DM-only reference",
        "mode": "subagent",
        "prompt": "You are the Grimorio Setting Guide Writer. Generate DM-only setting reference.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-appendices agent..."
    jq '.agent["grimorio-appendices"] = {
        "description": "Campaign appendices compiler — reference material",
        "mode": "subagent",
        "prompt": "You are the Grimorio Appendices Compiler. Consolidate reference material.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    log "Configuring grimorio-integrator agent..."
    jq '.agent["grimorio-integrator"] = {
        "description": "Campaign integrator — final assembly and PDF",
        "mode": "subagent",
        "prompt": "You are the Grimorio Integrator. Assemble final campaign and compile PDF.",
        "tools": {"bash": true, "edit": true, "read": true, "write": true, "grep": true},
        "options": {}
    }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"

    success "All agents configured"
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
    echo -e "   • compile_pdf, validate_canon, check_consistency"
    echo -e "   • process_consistency_gate, update_narrative_state"
    echo -e "   • evaluate_consequences, update_faction_reputation"
    echo -e "   • generate_random_tables, generate_handouts"
    echo -e "   • generate_session_prep, generate_flowchart"
    echo -e "   • generate_character_hooks, create_personal_quest"
    echo ""
    echo -e "${YELLOW}Need help?${NC} Check: ${GREEN}$INSTALL_DIR/README.md${NC}"
    echo ""
}

# ============================================================================
# MAIN
# ============================================================================
main() {
    echo -e "${GREEN}"
    echo "  ____      _                      _"
    echo " / ___|_ __(_)_ __ ___   ___  _ __(_) ___"
    echo "| |  _| '__| | '_ \` _ \ / _ \| '__| |/ _ \\"
    echo "| |_| | |  | | | | | | | (_) | |  | | (_) |"
    echo " \____|_|  |_|_| |_| |_|\___/|_|  |_|\___/"
    echo -e "${NC}"
    echo "       D&D One-shot & Campaign Generator"
    echo ""

    log "Starting installation..."

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

    # Step 9: Configure MCP
    configure_opencode_mcp

    # Step 10: Configure command
    configure_opencode_command

    # Step 11: Configure agents
    configure_opencode_agents

    # Step 12: Configure shell
    configure_shell

    success "Installation complete!"
    print_instructions
}

main "$@"
