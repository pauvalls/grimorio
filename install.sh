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

reexec_from_clone() {
    if [ -z "$GRIMORIO_REEXEC" ] && [ -f "$INSTALL_DIR/install.sh" ]; then
        log "Re-executing from latest cloned repository..."
        export GRIMORIO_REEXEC=1
        exec "$INSTALL_DIR/install.sh" "$@"
    fi
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
    for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
        if [ "$plugin_dir" = "$CLAUDE_PLUGIN_DIR" ]; then
            log "Setting up Claude Code plugin..."
        else
            log "Setting up OpenCode plugin..."
        fi
        mkdir -p "$plugin_dir"

        [ -f "$BINARY_DIR/grimorio" ] && cp -f "$BINARY_DIR/grimorio" "$plugin_dir/"
        [ -f "$BINARY_DIR/migrate-v1-to-v2" ] && cp -f "$BINARY_DIR/migrate-v1-to-v2" "$plugin_dir/"

        if [ -d "$INSTALL_DIR/agents" ]; then
            mkdir -p "$plugin_dir/agents"
            for agent_file in "$INSTALL_DIR/agents"/grimorio-*.md; do
                [ -f "$agent_file" ] && cp -f "$agent_file" "$plugin_dir/agents/"
            done
        fi

        if [ -d "$INSTALL_DIR/skills" ]; then
            mkdir -p "$plugin_dir/skills"
            for skill_file in "$INSTALL_DIR/skills"/grimorio-*.md; do
                [ -f "$skill_file" ] && cp -f "$skill_file" "$plugin_dir/skills/"
            done
        fi

        if [ "$plugin_dir" = "$CLAUDE_PLUGIN_DIR" ]; then
            cat > "$plugin_dir/.mcp.json" << 'EOF'
{
  "grimorio": {
    "command": "${CLAUDE_PLUGIN_ROOT}/grimorio",
    "args": [],
    "env": {}
  }
}
EOF
        else
            cat > "$plugin_dir/.mcp.json" << EOF
{
  "grimorio": {
    "command": "$plugin_dir/grimorio",
    "args": [],
    "env": {}
  }
}
EOF
        fi

        success "Plugin installed to $plugin_dir"
    done

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

clean_installation() {
    log "Cleaning previous Grimorio installation..."
    local cleaned=false

    for plugin_dir in "$CLAUDE_PLUGIN_DIR" "$OPENCODE_PLUGIN_DIR"; do
        if [ -d "$plugin_dir" ]; then
            rm -f "$plugin_dir/grimorio"
            rm -f "$plugin_dir/migrate-v1-to-v2"
            rm -f "$plugin_dir/.mcp.json"
            rm -rf "$plugin_dir/.claude-plugin"
            [ -d "$plugin_dir/commands" ] && rm -f "$plugin_dir/commands/grimorio.md"
            [ -d "$plugin_dir/agents" ] && rm -f "$plugin_dir/agents/grimorio-*.md"
            [ -d "$plugin_dir/skills" ] && rm -f "$plugin_dir/skills/grimorio-*.md"
            log "Cleaned Grimorio files from: $plugin_dir"
            cleaned=true
        fi
    done

    [ -f "$BINARY_DIR/grimorio" ] && rm -f "$BINARY_DIR/grimorio" && log "Removed: $BINARY_DIR/grimorio" && cleaned=true
    [ -f "$BINARY_DIR/migrate-v1-to-v2" ] && rm -f "$BINARY_DIR/migrate-v1-to-v2" && log "Removed: $BINARY_DIR/migrate-v1-to-v2" && cleaned=true

    if [ -d "$INSTALL_DIR" ]; then
        rm -rf "$INSTALL_DIR"
        log "Removed: $INSTALL_DIR"
        cleaned=true
    fi

    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"
    if [ -f "$OPENCODE_CONFIG" ] && command_exists jq; then
        jq 'del(.mcp.grimorio, .agent["grimorio-architect"], .agent["grimorio-artist"], .agent["grimorio-cartographer"], .agent["grimorio-lore"], .agent["grimorio-npc"], .agent["grimorio-bestiary"], .agent["grimorio-encounters"], .agent["grimorio-areas"], .agent["grimorio-quests"], .agent["grimorio-maps"], .agent["grimorio-characters"], .agent["grimorio-narrative-custodian"], .agent["grimorio-introduction"], .agent["grimorio-setting-guide"], .agent["grimorio-appendices"], .agent["grimorio-integrator"], .agent["grimorio-orchestrator"], .command.grimorio)' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        log "Cleaned grimorio entries from opencode.json"
        cleaned=true
    fi

    local shell_rcs=("${HOME}/.bashrc" "${HOME}/.zshrc" "${HOME}/.config/fish/config.fish" "${HOME}/.profile")
    for rc in "${shell_rcs[@]}"; do
        if [ -f "$rc" ]; then
            awk '
                /^# === GRIMORIO CONFIG BEGIN ===$/ { in_block=1; next }
                /^# === GRIMORIO CONFIG END ===$/   { in_block=0; next }
                !in_block { print }
            ' "$rc" > "${rc}.tmp" && mv "${rc}.tmp" "$rc"
            sed -i '/^# Grimorio$/d' "$rc"
            sed -i '/^export PATH="\$HOME\/\.local\/go\/bin:\$PATH"$/d' "$rc"
            sed -i '/^export PATH="\$HOME\/\.local\/bin:\$PATH"$/d' "$rc"
        fi
    done
    log "Cleaned shell configuration files"

    [ "$cleaned" = true ] && success "Previous Grimorio installation cleaned" || log "No previous installation found"
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
            "prompt": "You are the Grimorio Artist. Prepare image batch specifications and update markdown references.\\n\\nPhase A: Read NPCs, bestiary, and area chapters. Create batch-spec.json with all image prompts.\\nPhase B: After images are generated, update all markdown files with ![alt](assets/filename.png) references.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_generate_image to generate images. If MCP tools are unavailable, use the write tool to create files directly at ~/campaigns/{campaign_name}/assets/.",
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
            "prompt": "You are the Grimorio Cartographer. Generate ALL SVG assets for a campaign: battle maps, decorative dividers, and stat block borders. Use grimorio_generate_map and grimorio_generate_divider tools. Reference all SVGs in markdown files.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_generate_map and grimorio_generate_divider. If MCP tools are unavailable, use the write tool to create SVG files directly at ~/campaigns/{campaign_name}/assets/.",
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
            "prompt": "You are the Grimorio Lore Master. Generate world lore, backstory, setting, and atmosphere for a D&D 5e campaign. Use grimorio_save_lore tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_lore to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/lore.md.",
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
            "prompt": "You are the Grimorio NPC Designer. Generate NPCs, factions, and social entities for a D&D 5e campaign. Use grimorio_save_npcs tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_npcs to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/npcs/npcs_and_factions.md.",
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
            "prompt": "You are the Grimorio Bestiary Designer. Generate monsters, creatures, and stat blocks for a D&D 5e campaign. Use grimorio_save_bestiary tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_bestiary to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/bestiary/bestiary.md.",
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
            "prompt": "You are the Grimorio Encounter Designer. Generate balanced encounters and challenges for a D&D 5e campaign. Use grimorio_save_encounters tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_encounters to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/encounters/encounters.md.",
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

    # Configure grimorio-areas subagent
    log "Configuring grimorio-areas agent..."
    if command_exists jq; then
        jq '.agent["grimorio-areas"] = {
            "description": "Campaign areas designer — numbered playable areas (10-15 per act, WotC format) with DCs, treasure, and mechanics",
            "mode": "subagent",
            "prompt": "You are the Grimorio Areas Designer. Generate numbered playable areas for a D&D 5e campaign in WotC format. Each area has 150-200 words with specific DCs, treasure, and mechanics. Read ALL source files first (lore, NPCs, bestiary, maps, quests, encounters, characters). Use grimorio_save_areas tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_areas to save each chapter. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/areas/chapter_XX_title.md.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-areas agent configured"
    fi

    # Configure grimorio-quests subagent
    log "Configuring grimorio-quests agent..."
    if command_exists jq; then
        jq '.agent["grimorio-quests"] = {
            "description": "Campaign quest designer — personal quests, side missions, narrative hooks",
            "mode": "subagent",
            "prompt": "You are the Grimorio Quest Designer. Generate personal quests, side missions, and narrative hooks for a D&D 5e campaign. Use grimorio_create_personal_quest tool.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_create_personal_quest to create quests. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/quests/.",
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
            "prompt": "You are the Grimorio Map Describer. Generate location descriptions and zone breakdowns for a D&D 5e campaign. Use grimorio_save_maps tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_maps to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/maps/maps.md.",
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
            "prompt": "You are the Grimorio Character Builder. Generate pre-generated player characters with backstories for a D&D 5e campaign. Use grimorio_save_characters or grimorio_generate_character tools to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_characters to save characters. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/characters/.",
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

    # Configure grimorio-narrative-custodian subagent
    log "Configuring grimorio-narrative-custodian agent..."
    if command_exists jq; then
        jq '.agent["grimorio-narrative-custodian"] = {
            "description": "Campaign narrative custodian — validates canon consistency, checks cross-references, and manages narrative state",
            "mode": "subagent",
            "prompt": "You are the Grimorio Narrative Custodian. You validate campaign content for narrative coherence, check canon consistency, and manage narrative state. You NEVER generate creative content — only validate, check, and fix inconsistencies. Use grimorio_validate_canon, grimorio_check_consistency, grimorio_process_consistency_gate, grimorio_update_narrative_state, grimorio_evaluate_consequences, and other coherence tools.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Always use them for validation and state updates.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-narrative-custodian agent configured"
    fi

    # Configure grimorio-introduction subagent
    log "Configuring grimorio-introduction agent..."
    if command_exists jq; then
        jq '.agent["grimorio-introduction"] = {
            "description": "Campaign introduction — overview, hooks, and campaign summary for players",
            "mode": "subagent",
            "prompt": "You are the Grimorio Introduction Designer. Generate the campaign introduction and overview for players. Create compelling hooks and summarize the campaign arc. Use grimorio_save_introduction tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_introduction to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/introduction.md.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-introduction agent configured"
    fi

    # Configure grimorio-setting-guide subagent
    log "Configuring grimorio-setting-guide agent..."
    if command_exists jq; then
        jq '.agent["grimorio-setting-guide"] = {
            "description": "DM-only setting reference — geography, history, culture, factions, and secrets",
            "mode": "subagent",
            "prompt": "You are the Grimorio Setting Guide Designer. Generate DM-only reference material with spoilers. Include geography, history, culture, factions, and secrets. Read canon.json and lore.md first. Use grimorio_save_setting_guide tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_setting_guide to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/setting-guide.md.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-setting-guide agent configured"
    fi

    # Configure grimorio-appendices subagent
    log "Configuring grimorio-appendices agent..."
    if command_exists jq; then
        jq '.agent["grimorio-appendices"] = {
            "description": "Campaign appendices — consolidated reference material (magic items, stat blocks, handouts, maps)",
            "mode": "subagent",
            "prompt": "You are the Grimorio Appendices Designer. Generate consolidated reference material: Appendix A (Magic Items), Appendix B (NPCs and Monsters), Appendix C (Handouts), Appendix D (Maps), Appendix E (Reference Tables). Read ALL source files. Use grimorio_save_appendices tool to persist content.\\n\\nIMPORTANT: You have access to grimorio MCP tools. Use grimorio_save_appendices to save content. If MCP tools are unavailable, use the write tool to save content directly to ~/campaigns/{campaign_name}/appendices.md.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-appendices agent configured"
    fi

    # Configure grimorio-integrator subagent
    log "Configuring grimorio-integrator agent..."
    if command_exists jq; then
        jq '.agent["grimorio-integrator"] = {
            "description": "Campaign integrator — cross-references, finds inconsistencies, and finalizes content",
            "mode": "subagent",
            "prompt": "You are the Grimorio Integrator. Cross-reference all campaign content, find inconsistencies, and finalize. Check that all references between files are valid.\\n\\nIMPORTANT: You have access to grimorio MCP tools including grimorio_validate_canon, grimorio_check_consistency, grimorio_process_consistency_gate, grimorio_save_areas, grimorio_save_npcs, grimorio_save_encounters. Use them for validation and persistence. If MCP tools are unavailable, use the write tool to save content directly.",
            "tools": {
                "bash": true,
                "edit": true,
                "read": true,
                "write": true,
                "grep": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-integrator agent configured"
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

### Phase 3-13: End-to-End Orchestration (sequential batches)
The architect follows strict batch ordering — each batch waits for the previous:

- **Batch 1** (parallel): NPCs, bestiary, maps → Consistency Gate
- **Batch 2** (parallel): lore, quests, encounters, characters → Consistency Gate → Update Narrative State
- **Batch 3** (parallel): SVG maps, areas → Consistency Gate
- **Phase 6**: Artist batch-spec (cover + NPCs + scenes + monsters)
- **Phase 7**: Generate AI images (1x1 sequential, retry missing)
- **Phase 8**: Update ALL markdown references
- **Phase 9**: Living World tools (factions, random tables, handouts, consequences) → Consistency Gate
- **Phase 10**: DM Experience tools (session prep, flowchart)
- **Phase 11**: Final consistency check
- **Phase 12**: Compile PDF (embeds all images + flowchart)
- **Phase 13**: Final report

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

get_version() {
    if [ -d "$INSTALL_DIR" ] && command_exists git; then
        git -C "$INSTALL_DIR" tag --sort=-v:refname 2>/dev/null | head -1 || echo "dev"
    else
        echo "dev"
    fi
}

print_instructions() {
    local VERSION=$(get_version)
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║         Grimorio ${VERSION} - Installed Successfully!           ║${NC}"
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
    echo -e "   • Command     → /grimorio (orchestrated by grimorio-architect)"
    echo -e "   • Agents configured:"
    echo -e "     - grimorio-architect     (orchestrates all phases)"
    echo -e "     - grimorio-lore          (world backstory & atmosphere)"
    echo -e "     - grimorio-npc           (NPCs & factions)"
    echo -e "     - grimorio-bestiary      (monster stat blocks)"
    echo -e "     - grimorio-encounters    (combat & exploration challenges)"
    echo -e "     - grimorio-maps          (location & zone descriptions)"
    echo -e "     - grimorio-areas         (numbered playable areas, WotC format)"
    echo -e "     - grimorio-quests        (personal quests & side missions)"
    echo -e "     - grimorio-characters    (pre-generated character sheets)"
    echo -e "     - grimorio-narrative-custodian (canon validation + state tracking)"
    echo -e "     - grimorio-introduction  (campaign overview & hooks)"
    echo -e "     - grimorio-setting-guide (DM-only setting reference)"
    echo -e "     - grimorio-appendices    (consolidated reference material)"
    echo -e "     - grimorio-integrator    (cross-references & finalization)"
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
echo -e "7. ${YELLOW}Narrative Coherence Tools (v2.0):${NC}"
   echo -e "   • generate_adventure_bible → Creates canon.json with facts, entities, rules"
   echo -e "   • validate_canon → Validates content against canon (prevents NPC resurrections!)"
   echo -e "   • update_narrative_state → Track session state (clues, quests, deaths)"
   echo -e "   • check_consistency → Full campaign validation before PDF"
   echo -e "   • process_consistency_gate → Batch validation gate (approve/reject/retry)"
   echo -e ""
   echo -e "8. ${YELLOW}Living World Tools (NEW v2.1):${NC}"
   echo -e "   • update_faction_reputation → Modify faction reputation with ally/enemy propagation"
   echo -e "   • generate_random_tables → Contextual encounter, rumor, weather, treasure tables"
   echo -e "   • generate_handouts → Player-facing + DM-only handouts (letters, maps, codes)"
   echo -e "   • evaluate_consequences → Evaluate consequence rules against narrative state"
   echo -e ""
   echo -e "9. ${YELLOW}Update grimorio later:${NC}"
   echo -e "   Just re-run: ${GREEN}curl -sSL ${REPO_URL}/raw/main/install.sh | bash${NC}"
   echo -e ""
   echo -e "10. ${YELLOW}Migration from v1:${NC}"
   echo -e "   If you have old campaigns: ${GREEN}migrate-v1-to-v2 ~/campaigns${NC}"
   echo -e ""
   echo -e "${BLUE}Manual usage (without AI tools):${NC}"
   echo -e "   ${GREEN}grimorio${NC} - Runs the MCP server"
   echo -e "   ${GREEN}migrate-v1-to-v2 ~/campaigns${NC} - Migrate old campaigns"
   echo -e ""
   echo -e "${YELLOW}Need help?${NC} Check the README at: ${GREEN}${INSTALL_DIR}/README.md${NC}"
   echo -e ""
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

    # Always clean first to ensure no stale files from previous installs
    clean_installation

    detect_platform
    install_go
    install_wkhtmltopdf
    setup_repo
    reexec_from_clone "$@"
    build_binary
    migrate_existing_campaigns
    setup_plugin
    configure_shell
    configure_opencode_command

    success "Installation complete!"
    print_instructions
}

main "$@"
