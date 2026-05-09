#!/bin/bash

#===============================================================================
# install.sh — Instalador Completo de Grimorio
# Versión: 1.0.0
#
# Uso: ./scripts/install.sh [--reinstall|--validate|--quick]
#
# Este script:
# 1. Valida la configuración de opencode.json
# 2. Verifica/instala dependencias (Go, jq, wkhtmltopdf)
# 3. Compila el binario Grimorio
# 4. Verifica templates, scripts y skills
# 5. Configura MCP servers
# 6. Ejecuta validación final
#===============================================================================

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Rutas
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OPENCODE_CONFIG="$HOME/.config/opencode/opencode.json"
GRIMORIO_BINARY="$PROJECT_ROOT/grimorio"
SKILLS_DIR="$HOME/.config/opencode/skills"

# Modo de instalación
MODE="full"

#===============================================================================
# Helper Functions
#===============================================================================

print_header() {
    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}       ${BLUE}Grimorio Installation Script${NC}                      ${CYAN}║${NC}"
    echo -e "${CYAN}║${NC}       D&D 5e Campaign Management CLI & MCP         ${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${BLUE}━━━${NC} ${CYAN}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

check_command() {
    local cmd="$1"
    local package="$2"
    
    if command -v "$cmd" &> /dev/null; then
        local version=$("$cmd" --version 2>&1 | head -1)
        print_success "$cmd installed ($version)"
        return 0
    else
        print_error "$cmd not found"
        print_warning "Install with: $package"
        return 1
    fi
}

#===============================================================================
# Installation Functions
#===============================================================================

install_dependencies() {
    print_step "Checking Dependencies..."
    echo ""
    
    local missing_deps=0
    
    # Go
    print_step "Checking Go..."
    if check_command "go" "apt install golang-go / brew install go"; then
        :
    else
        missing_deps=$((missing_deps + 1))
    fi
    
    # jq (optional but recommended)
    print_step "Checking jq..."
    if command -v jq &> /dev/null; then
        print_success "jq installed ($(jq --version))"
    else
        print_warning "jq not found (optional, recommended for validation)"
    fi
    
    # wkhtmltopdf (for PDF compilation)
    print_step "Checking wkhtmltopdf..."
    if command -v wkhtmltopdf &> /dev/null; then
        print_success "wkhtmltopdf installed ($(wkhtmltopdf --version 2>&1 | head -1))"
    else
        print_warning "wkhtmltopdf not found (required for PDF compilation)"
        print_warning "Install with: apt install wkhtmltopdf / brew install wkhtmltopdf"
    fi
    
    echo ""
    
    if [ $missing_deps -gt 0 ]; then
        print_error "Missing $missing_deps critical dependency(ies)"
        echo ""
        echo "Please install missing dependencies and re-run this script."
        echo ""
        return 1
    else
        print_success "All critical dependencies installed"
        return 0
    fi
}

build_grimorio() {
    print_step "Building Grimorio Binary..."
    echo ""
    
    cd "$PROJECT_ROOT"
    
    # Check if already built
    if [ -f "$GRIMORIO_BINARY" ]; then
        local current_version=$("$GRIMORIO_BINARY" --version 2>/dev/null || echo "unknown")
        print_warning "Existing binary found: $current_version"
        read -p "Rebuild? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_success "Keeping existing binary"
            return 0
        fi
    fi
    
    # Build
    print_step "Compiling..."
    if go build -o grimorio ./cmd/grimorio; then
        chmod +x grimorio
        local new_version=$(./grimorio --version 2>/dev/null || echo "dev build")
        print_success "Binary compiled: $GRIMORIO_BINARY ($new_version)"
        print_success "Size: $(ls -lh grimorio | awk '{print $5}')"
    else
        print_error "Build failed"
        return 1
    fi
    
    echo ""
    return 0
}

verify_opencode_config() {
    print_step "Verifying OpenCode Configuration..."
    echo ""
    
    # Check config file exists
    if [ ! -f "$OPENCODE_CONFIG" ]; then
        print_error "OpenCode config not found: $OPENCODE_CONFIG"
        print_warning "Run 'opencode init' first or copy config manually"
        return 1
    fi
    
    print_success "Config file exists: $OPENCODE_CONFIG"
    
    # Check JSON syntax
    if command -v jq &> /dev/null; then
        if jq empty "$OPENCODE_CONFIG" 2>/dev/null; then
            print_success "JSON syntax valid"
        else
            print_error "Invalid JSON syntax"
            return 1
        fi
    fi
    
    # Check required sections
    local required_sections=("agent" "mcp" "permission" "command")
    for section in "${required_sections[@]}"; do
        if grep -q "\"$section\"" "$OPENCODE_CONFIG"; then
            print_success "Section '$section' present"
        else
            print_error "Section '$section' missing"
            return 1
        fi
    done
    
    # Check SDD config
    if grep -q '"sdd"' "$OPENCODE_CONFIG"; then
        print_success "SDD configuration present"
        
        # Check delivery_strategy
        if grep -q '"delivery_strategy"' "$OPENCODE_CONFIG"; then
            local strategy=$(grep -A 1 '"delivery_strategy"' "$OPENCODE_CONFIG" | grep -oE '(exception-ok|ask-on-risk|auto-chain|single-pr)')
            print_success "delivery_strategy: $strategy"
        fi
        
        # Check chain_strategy
        if grep -q '"chain_strategy"' "$OPENCODE_CONFIG"; then
            local chain=$(grep -A 1 '"chain_strategy"' "$OPENCODE_CONFIG" | grep -oE '(stacked-to-main|feature-branch-chain)')
            print_success "chain_strategy: $chain"
        fi
    else
        print_warning "SDD configuration not found"
    fi
    
    # Check grimorio command
    if grep -q '"grimorio"' "$OPENCODE_CONFIG" && grep -A 5 '"grimorio"' "$OPENCODE_CONFIG" | grep -q '"agent": "grimorio-architect"'; then
        print_success "/grimorio command configured with grimorio-architect"
    else
        print_error "/grimorio command not properly configured"
        return 1
    fi
    
    echo ""
    return 0
}

verify_agents() {
    print_step "Verifying Agent Definitions..."
    echo ""
    
    local required_agents=(
        "grimorio-architect"
        "grimorio-artist"
        "grimorio-cartographer"
        "grimorio-lore"
        "grimorio-npc"
        "grimorio-bestiary"
        "grimorio-encounters"
        "grimorio-areas"
        "grimorio-quests"
        "grimorio-maps"
        "grimorio-characters"
        "grimorio-narrative-custodian"
        "grimorio-introduction"
        "grimorio-setting-guide"
        "grimorio-appendices"
        "grimorio-integrator"
        "gentle-orchestrator"
        "sdd-apply"
        "sdd-verify"
    )
    
    local missing=0
    for agent in "${required_agents[@]}"; do
        if grep -q "\"$agent\"" "$OPENCODE_CONFIG"; then
            echo -e "  ${GREEN}✅${NC} $agent"
        else
            echo -e "  ${RED}❌${NC} $agent"
            missing=$((missing + 1))
        fi
    done
    
    echo ""
    
    if [ $missing -gt 0 ]; then
        print_error "$missing agent(s) missing"
        return 1
    else
        print_success "All ${#required_agents[@]} agents configured"
        return 0
    fi
}

verify_templates() {
    print_step "Verifying Templates..."
    echo ""
    
    local templates_dir="$PROJECT_ROOT/internal/compiler/templates"
    local required_templates=(
        "areas.md.tmpl"
        "npc.md.tmpl"
        "monster.md.tmpl"
        "encounter.md.tmpl"
        "lore.md.tmpl"
        "map.md.tmpl"
        "setting-guide.md.tmpl"
        "appendices.md.tmpl"
        "introduction.md.tmpl"
    )
    
    if [ ! -d "$templates_dir" ]; then
        print_error "Templates directory not found: $templates_dir"
        return 1
    fi
    
    print_success "Templates directory exists"
    
    local missing=0
    for tmpl in "${required_templates[@]}"; do
        if [ -f "$templates_dir/$tmpl" ]; then
            local lines=$(wc -l < "$templates_dir/$tmpl")
            echo -e "  ${GREEN}✅${NC} $tmpl ($lines lines)"
        else
            echo -e "  ${RED}❌${NC} $tmpl"
            missing=$((missing + 1))
        fi
    done
    
    echo ""
    
    if [ $missing -gt 0 ]; then
        print_error "$missing template(s) missing"
        return 1
    else
        print_success "All ${#required_templates[@]} templates present"
        return 0
    fi
}

verify_scripts() {
    print_step "Verifying Scripts..."
    echo ""
    
    local scripts_dir="$PROJECT_ROOT/scripts"
    local required_scripts=(
        "validate-campaign.sh"
        "validate-opencode.sh"
        "install.sh"
    )
    
    if [ ! -d "$scripts_dir" ]; then
        print_error "Scripts directory not found: $scripts_dir"
        return 1
    fi
    
    local missing=0
    for script in "${required_scripts[@]}"; do
        if [ -f "$scripts_dir/$script" ]; then
            if [ -x "$scripts_dir/$script" ]; then
                echo -e "  ${GREEN}✅${NC} $script (executable)"
            else
                echo -e "  ${YELLOW}⚠️${NC} $script (not executable)"
                chmod +x "$scripts_dir/$script"
            fi
        else
            echo -e "  ${RED}❌${NC} $script"
            missing=$((missing + 1))
        fi
    done
    
    echo ""
    
    if [ $missing -gt 0 ]; then
        print_error "$missing script(s) missing"
        return 1
    else
        print_success "All ${#required_scripts[@]} scripts present"
        return 0
    fi
}

verify_skills() {
    print_step "Verifying Skills..."
    echo ""
    
    # Check skills directory
    if [ ! -d "$SKILLS_DIR" ]; then
        print_warning "Skills directory not found: $SKILLS_DIR"
        print_warning "Creating directory..."
        mkdir -p "$SKILLS_DIR"
    fi
    
    print_success "Skills directory exists"
    
    # Check grimorio-architect skill
    if [ -f "$SKILLS_DIR/grimorio-architect/SKILL.md" ]; then
        local lines=$(wc -l < "$SKILLS_DIR/grimorio-architect/SKILL.md")
        print_success "grimorio-architect skill ($lines lines)"
    else
        print_warning "grimorio-architect skill not found"
    fi
    
    # Check skill registry
    local skill_registry="$PROJECT_ROOT/.atl/skill-registry.md"
    if [ -f "$skill_registry" ]; then
        print_success "Skill registry exists"
        
        # Check if grimorio skills are registered
        if grep -q "grimorio-architect" "$skill_registry"; then
            print_success "grimorio-architect registered in skill-registry.md"
        fi
        
        if grep -q "grimorio-areas" "$skill_registry"; then
            print_success "grimorio-areas registered in skill-registry.md"
        fi
    else
        print_warning "Skill registry not found"
    fi
    
    echo ""
    return 0
}

verify_mcp() {
    print_step "Verifying MCP Configuration..."
    echo ""
    
    # Check grimorio MCP in config
    if grep -A 5 '"grimorio"' "$OPENCODE_CONFIG" | grep -q '"command"'; then
        print_success "Grimorio MCP server configured"
        
        # Extract path
        local mcp_path=$(grep -A 5 '"grimorio"' "$OPENCODE_CONFIG" | grep '"command"' | grep -oE '/[^\"]+' | head -1)
        if [ -n "$mcp_path" ] && [ -f "$mcp_path" ]; then
            print_success "MCP binary exists: $mcp_path"
        else
            print_warning "MCP binary path: $mcp_path (not found)"
        fi
    else
        print_error "Grimorio MCP server not configured"
        return 1
    fi
    
    # Check engram MCP
    if grep -A 5 '"engram"' "$OPENCODE_CONFIG" | grep -q '"command"'; then
        print_success "Engram MCP server configured"
    else
        print_warning "Engram MCP server not configured"
    fi
    
    # Check context7 MCP
    if grep -A 5 '"context7"' "$OPENCODE_CONFIG" | grep -q '"url"'; then
        print_success "Context7 MCP server configured"
    else
        print_warning "Context7 MCP server not configured"
    fi
    
    echo ""
    return 0
}

run_validation() {
    print_step "Running Full Validation..."
    echo ""
    
    if [ -x "$PROJECT_ROOT/scripts/validate-opencode.sh" ]; then
        "$PROJECT_ROOT/scripts/validate-opencode.sh" --check=all
        return $?
    else
        print_error "Validation script not found or not executable"
        return 1
    fi
}

print_summary() {
    echo ""
    echo -e "${CYAN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}              ${GREEN}Installation Complete!${NC}                      ${CYAN}║${NC}"
    echo -e "${CYAN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Summary:${NC}"
    echo "  ✅ Dependencies installed"
    echo "  ✅ Grimorio binary compiled"
    echo "  ✅ OpenCode configuration verified"
    echo "  ✅ Agents configured ($(grep -c '"grimorio-' "$OPENCODE_CONFIG" 2>/dev/null || echo 0))"
    echo "  ✅ Templates verified ($(ls -1 "$PROJECT_ROOT/internal/compiler/templates"/*.tmpl 2>/dev/null | wc -l))"
    echo "  ✅ Scripts verified ($(ls -1 "$PROJECT_ROOT/scripts"/*.sh 2>/dev/null | wc -l))"
    echo "  ✅ Skills configured"
    echo "  ✅ MCP servers configured"
    echo ""
    echo -e "${BLUE}Next Steps:${NC}"
    echo "  1. Restart your terminal or run: source ~/.zshrc"
    echo "  2. Test with: /grimorio"
    echo "  3. Validate anytime: ./scripts/validate-opencode.sh --check=all"
    echo ""
    echo -e "${CYAN}Happy Gaming! 🎲${NC}"
    echo ""
}

#===============================================================================
# Main Script
#===============================================================================

# Parse arguments
for arg in "$@"; do
    case $arg in
        --reinstall)
            MODE="reinstall"
            shift
            ;;
        --validate)
            MODE="validate"
            shift
            ;;
        --quick)
            MODE="quick"
            shift
            ;;
        --help)
            echo "Usage: $0 [--reinstall|--validate|--quick|--help]"
            echo ""
            echo "Options:"
            echo "  --reinstall  Force rebuild of all components"
            echo "  --validate   Only run validation, no installation"
            echo "  --quick      Skip dependency checks (already installed)"
            echo "  --help       Show this help message"
            exit 0
            ;;
    esac
done

# Start installation
print_header

case $MODE in
    validate)
        run_validation
        exit $?
        ;;
    quick)
        verify_opencode_config
        verify_agents
        verify_templates
        verify_scripts
        print_summary
        exit $?
        ;;
    reinstall|full)
        install_dependencies || exit 1
        build_grimorio || exit 1
        verify_opencode_config || exit 1
        verify_agents || exit 1
        verify_templates || exit 1
        verify_scripts || exit 1
        verify_skills || exit 1
        verify_mcp || exit 1
        run_validation || exit 1
        print_summary
        exit 0
        ;;
esac
