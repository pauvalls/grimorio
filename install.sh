#!/bin/bash
set -e

# Grimorio - One Command Installer
# Usage: curl -sSL https://raw.githubusercontent.com/paupena/grimorio/main/install.sh | bash

REPO_URL="https://github.com/pauvalls/Grimorio"
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
        git pull origin main 2>/dev/null || true
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

    cp -r "$INSTALL_DIR/.claude-plugin" "$CLAUDE_PLUGIN_DIR/"
    cp -r "$INSTALL_DIR/commands" "$CLAUDE_PLUGIN_DIR/"
    cp -r "$INSTALL_DIR/agents" "$CLAUDE_PLUGIN_DIR/"
    cp -r "$INSTALL_DIR/skills" "$CLAUDE_PLUGIN_DIR/"
    cp "$BINARY_DIR/grimorio" "$CLAUDE_PLUGIN_DIR/"

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

    cp -r "$INSTALL_DIR/.claude-plugin" "$OPENCODE_PLUGIN_DIR/"
    cp -r "$INSTALL_DIR/commands" "$OPENCODE_PLUGIN_DIR/"
    cp -r "$INSTALL_DIR/agents" "$OPENCODE_PLUGIN_DIR/"
    cp -r "$INSTALL_DIR/skills" "$OPENCODE_PLUGIN_DIR/"
    cp "$BINARY_DIR/grimorio" "$OPENCODE_PLUGIN_DIR/"

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

    log "Adding grimorio MCP to opencode.json..."

    # Check if grimorio MCP is already configured
    if command_exists jq && jq -e '.mcp.grimorio' "$OPENCODE_CONFIG" > /dev/null 2>&1; then
        log "grimorio MCP already configured in opencode.json"
        return 0
    fi

    # Use jq if available
    if command_exists jq; then
        # Add grimorio to the mcp section using jq
        jq '.mcp.grimorio = {
            "command": ["'"$OPENCODE_PLUGIN_DIR/grimorio"'"],
            "type": "local",
            "enabled": true
        }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
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

    success "grimorio MCP added to opencode.json"
}

configure_opencode_command() {
    local OPENCODE_CONFIG="${HOME}/.config/opencode/opencode.json"

    if [ ! -f "$OPENCODE_CONFIG" ]; then
        log "No opencode.json found, skipping command config"
        return 0
    fi

    # Add grimorio-architect agent if not present
    log "Configuring grimorio-architect agent..."
    if command_exists jq; then
        if ! jq -e '.agent["grimorio-architect"]' "$OPENCODE_CONFIG" > /dev/null 2>&1; then
            jq '.agent["grimorio-architect"] = {
                "description": "Expert Dungeon Master agent for D&D 5e campaign generation",
                "mode": "primary",
                "prompt": "You are an expert Dungeon Master and campaign designer. Your job is to:\n1. Ask the user clarifying questions about their campaign idea (level, tone, duration, name)\n2. After gathering all requirements, use subagents to generate each component\n3. When all subagents complete, compile the final PDF\n\nDO NOT edit files in the main thread. Always delegate generation work to subagents.",
                "tools": {
                    "bash": true,
                    "delegate": true,
                    "delegation_list": true,
                    "delegation_read": true,
                    "edit": true,
                    "read": true,
                    "write": true
                },
                "options": {}
            }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
            success "grimorio-architect agent added to opencode.json"
        else
            log "grimorio-architect agent already configured"
        fi
    fi

    # Add grimorio command if not present
    log "Configuring grimorio command..."
    if command_exists jq; then
        if ! jq -e '.command.grimorio' "$OPENCODE_CONFIG" > /dev/null 2>&1; then
            jq '.command.grimorio = {
                "description": "Generate a complete D&D 5e campaign or one-shot from an idea",
                "agent": "grimorio-architect",
                "subtask": false,
                "template": "Generate a D&D 5e campaign or one-shot from the user'\''s idea.\n\n## Workflow\n\n### Phase 1: Gather Requirements\nAsk the user these questions (one at a time, interactively):\n1. What'\''s the campaign name? (kebab-case, e.g. \"sunken-city\")\n2. One-shot or full campaign?\n3. Player level? (1-3, 4-6, 7-10, 11-15, 16-20)\n4. Desired tone? (heroic, dark, humorous, political intrigue)\n5. Duration? (one-shot, 3-5 sessions, long campaign)\n\n### Phase 2: Create Campaign Structure\nUse the grimorio MCP tool `create_campaign` to create the structure.\n\n### Phase 3: Generate Content via Subagents\nLaunch subagents in PARALLEL using `delegate` for each of these tasks:\n- Delegate lore generation: generate world backstory, setting, conflict\n- Delegate acts generation: generate act 1, act 2, act 3\n- Delegate NPCs generation: generate 5+ NPCs with factions\n- Delegate bestiary generation: generate 3-5 monsters with stat blocks\n- Delegate encounters generation: generate 3-5 encounters\n- Delegate maps generation: generate scene descriptions\n\nEach subagent must use the grimorio MCP tools to save their output.\n\n### Phase 4: Compile PDF\nAfter ALL subagents complete, use grimorio MCP tool `compile_pdf` to generate the final PDF.\n\n### Phase 5: Report\nTell the user where the PDF and markdown files were saved."
            }' "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
            success "grimorio command added to opencode.json"
        else
            log "grimorio command already configured"
        fi
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
    echo -e "   • Agent       → grimorio-architect (with delegate/subagent support)"
    echo -e "   • Command     → /grimorio (interactive Q&A + parallel subagents)"
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
    echo -e "   • DALL-E images     → Set OPENAI_API_KEY for cover art/portraits"
    echo ""
    echo -e "${BLUE}Manual usage (without AI tools):${NC}"
    echo -e "   ${GREEN}grimorio${NC} - Runs the MCP server"
    echo ""
    echo -e "${YELLOW}Need help?${NC} Check the README at: ${GREEN}${INSTALL_DIR}/README.md${NC}"
    echo ""
}

main() {
    echo -e "${GREEN}"
    echo -e "   ____                _       _    ___ "
    echo -e "  / ___| _   _ _ __ __| | __ _| |  |_ _|"
    echo -e " | |  _ | | | | '_ \(_)/ _\` | |   | | "
    echo -e " | |_| || |_| | | | | | (_| | |   | | "
    echo -e "  \____| \__, |_| |_| |_|\__,_|_|  |___|"
    echo -e "         |___/                          "
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
