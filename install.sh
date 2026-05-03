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

    mkdir -p "$BINARY_DIR"
    cp grimorio "$BINARY_DIR/"
    chmod +x "$BINARY_DIR/grimorio"

    success "Binary built and installed to $BINARY_DIR/grimorio"
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

    # Always update commands/grimorio.md to latest version
    if [ -f "$INSTALL_DIR/commands/grimorio.md" ]; then
        cp -f "$INSTALL_DIR/commands/grimorio.md" "$CLAUDE_PLUGIN_DIR/commands/"
    fi

    # Copy new cartographer agent if it exists in repo but not in plugin
    if [ -f "$INSTALL_DIR/agents/grimorio-cartographer.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-cartographer.md" "$CLAUDE_PLUGIN_DIR/agents/"
    fi

    # Copy orchestrator agent
    if [ -f "$INSTALL_DIR/agents/grimorio-orchestrator.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-orchestrator.md" "$CLAUDE_PLUGIN_DIR/agents/"
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

    # Always update commands/grimorio.md to latest version
    if [ -f "$INSTALL_DIR/commands/grimorio.md" ]; then
        cp -f "$INSTALL_DIR/commands/grimorio.md" "$OPENCODE_PLUGIN_DIR/commands/"
    fi

    # Copy new cartographer agent if it exists in repo but not in plugin
    if [ -f "$INSTALL_DIR/agents/grimorio-cartographer.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-cartographer.md" "$OPENCODE_PLUGIN_DIR/agents/"
    fi

    # Copy orchestrator agent
    if [ -f "$INSTALL_DIR/agents/grimorio-orchestrator.md" ]; then
        cp -f "$INSTALL_DIR/agents/grimorio-orchestrator.md" "$OPENCODE_PLUGIN_DIR/agents/"
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

    if [ -n "$SHELL_RC" ]; then
        if ! grep -q "\.local/bin" "$SHELL_RC" 2>/dev/null; then
            echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC"
            log "Added ~/.local/bin to PATH in $SHELL_RC"
        fi

        if ! grep -q "grimorio" "$SHELL_RC" 2>/dev/null; then
            echo '# Grimorio' >> "$SHELL_RC"
            echo 'export PATH="$HOME/.local/go/bin:$PATH"' >> "$SHELL_RC"
        fi
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

    # Always update agent (not just add) to ensure latest prompt
    log "Configuring grimorio-architect agent..."
    if command_exists jq; then
        jq '.agent["grimorio-architect"] = {
            "description": "Expert Dungeon Master agent for D&D 5e campaign generation",
            "mode": "primary",
            "prompt": "You are an expert Dungeon Master and campaign designer. Your job is to:\n1. Ask the user clarifying questions about their campaign idea (level, tone, duration, name)\n2. After gathering all requirements, launch grimorio-orchestrator with a single delegate call\n3. Report the final result to the user\n\nDO NOT edit files in the main thread. Always delegate to grimorio-orchestrator.",
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

    # Configure grimorio-orchestrator subagent
    log "Configuring grimorio-orchestrator agent..."
    if command_exists jq; then
        jq '.agent["grimorio-orchestrator"] = {
            "description": "Internal coordinator for grimorio campaigns with MCP tool access",
            "mode": "subagent",
            "prompt": "You are the Grimorio Orchestrator. Your ONLY job is to coordinate subagent execution.\n\nWorkflow:\n1. Launch content subagents (lore, NPCs, bestiary, encounters, maps) in parallel\n2. Monitor completion via delegation_list\n3. Launch acts subagent (uses [SCENE: ...] placeholders)\n4. Launch cartographer (SVGs) + artist (batch-spec.json) in parallel\n5. Use generate_images_batch MCP tool to generate ALL AI images at once\n6. Retry failed images individually with generate_image\n7. Launch artist again to update markdown references\n8. Compile PDF with compile_pdf\n9. Report to parent",
            "tools": {
                "bash": true,
                "delegate": true,
                "delegation_list": true,
                "delegation_read": true,
                "edit": true,
                "read": true,
                "write": true,
                "generate_image": true,
                "generate_images_batch": true,
                "generate_map": true,
                "generate_divider": true,
                "compile_pdf": true
            },
            "options": {}
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
        success "grimorio-orchestrator agent configured"
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

    # Always update command (not just add) to ensure latest template with image generation
    log "Configuring grimorio command..."
    if command_exists jq; then
        # Create template in temp file to avoid bash parenthesis issues
        local TEMPLATE_FILE=$(mktemp)
        cat > "$TEMPLATE_FILE" << 'TEMPLATE_EOF'
Generate a D&D 5e campaign or one-shot from the user's idea.

## IMPORTANT: Use `delegate` tool to launch ALL subagents. NEVER do the work yourself.

## Workflow

### Phase 1: Gather Requirements
Ask the user these questions (one at a time, interactively):
1. What's the campaign name? (kebab-case, e.g. "sunken-city")
2. One-shot or full campaign?
3. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)
4. Desired tone? (heroic, dark, humorous, political intrigue)
5. Duration? (one-shot, 3-5 sessions, long campaign)

### Phase 2: Create Campaign Structure
Use the grimorio MCP tool `create_campaign` to create the structure.

### Phase 3: Launch Orchestrator (SINGLE delegate call)
Launch the **grimorio-orchestrator** subagent with ALL campaign parameters.

You MUST pass these parameters in the prompt:
- `campaign_path` — the full path returned by create_campaign
- `campaign_name` — the kebab-case campaign name
- `setting` — the campaign description/setting
- `level_range` — e.g., "1-3", "4-6"
- `tone` — e.g., "heroic", "dark"
- `duration` — e.g., "one-shot", "3-5 sessions"
- `is_oneshot` — true if one-shot, false if campaign

Example:
```
delegate(
  agent="grimorio-orchestrator",
  prompt="Coordinate campaign generation for 'sunken-city'.\n\ncampaign_path: /home/pau/campaigns/sunken-city\ncampaign_name: sunken-city\nsetting: A sunken city where nobles are aquatic vampires...\nlevel_range: 4-6\ntone: dark\nduration: 3-5 sessions\nis_oneshot: false"
)
```

**CRITICAL:** This is the ONLY `delegate` call you make. The orchestrator handles ALL other subagents internally. Do NOT launch any other subagents from this thread.

### Phase 4: Report
After the orchestrator completes, report to the user:
- Where the PDF was saved
- What content was generated
- Any issues encountered

**DO NOT call `delegation_list` repeatedly. Launch the orchestrator once and wait for it to complete.**
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
    echo -e "   • Architect   → grimorio-architect (handles user Q&A)"
    echo -e "   • Orchestrator→ grimorio-orchestrator (coordinates all subagents + MCP tools)"
    echo -e "   • Artist      → grimorio-artist (prepares image specs + updates references)"
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
    echo -e "   • AI images         → FREE via Pollinations.ai (no API key needed)"
    echo -e "   • DALL-E (optional) → Set OPENAI_API_KEY for higher quality"
    echo ""
    echo -e "7. ${YELLOW}Update grimorio later:${NC}"
    echo -e "   Just re-run: ${GREEN}curl -sSL ${REPO_URL}/raw/main/install.sh | bash${NC}"
    echo ""
    echo -e "${BLUE}Manual usage (without AI tools):${NC}"
    echo -e "   ${GREEN}grimorio${NC} - Runs the MCP server"
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
    setup_plugin
    configure_shell
    configure_opencode_command

    success "Installation complete!"
    print_instructions
}

main "$@"
