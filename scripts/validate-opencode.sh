#!/bin/bash

#===============================================================================
# validate-opencode.sh — Validador Completo de Configuración OpenCode + Grimorio
# Versión: 1.0.0
#
# Uso: ./scripts/validate-opencode.sh [--check=all|agents|skills|sdd|mcp|build|json]
#
# Exit Codes:
#   0 - Todas las validaciones pasaron
#   1 - Una o más validaciones fallaron
#   2 - Error (archivo no encontrado, JSON inválido, etc.)
#===============================================================================

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Contadores
PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0
TOTAL_CHECKS=0

# Rutas
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OPENCODE_CONFIG="$HOME/.config/opencode/opencode.json"
SKILL_REGISTRY="$PROJECT_ROOT/.atl/skill-registry.md"
GRIMORIO_BINARY="$PROJECT_ROOT/grimorio"

# Check type
CHECK_TYPE="all"

#===============================================================================
# Helper Functions
#===============================================================================

print_header() {
    echo ""
    echo -e "${BLUE}=====================================${NC}"
    echo -e "${BLUE}  OpenCode + Grimorio Validation${NC}"
    echo -e "${BLUE}=====================================${NC}"
    echo "Date: $(date +%Y-%m-%d)"
    echo "Project: $PROJECT_ROOT"
    echo ""
}

print_section() {
    echo ""
    echo -e "${BLUE}--- $1 ---${NC}"
}

print_check() {
    local check_name="$1"
    local status="$2"
    local details="$3"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅${NC} $check_name"
        if [ -n "$details" ]; then
            echo -e "   ${GREEN}$details${NC}"
        fi
        PASS_COUNT=$((PASS_COUNT + 1))
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}❌${NC} $check_name"
        if [ -n "$details" ]; then
            echo -e "   ${RED}$details${NC}"
        fi
        FAIL_COUNT=$((FAIL_COUNT + 1))
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠️${NC} $check_name"
        if [ -n "$details" ]; then
            echo -e "   ${YELLOW}$details${NC}"
        fi
        WARN_COUNT=$((WARN_COUNT + 1))
    fi
}

print_footer() {
    echo ""
    echo "====================================="
    echo "  Validation Summary"
    echo "====================================="
    echo -e "Total Checks: ${TOTAL_CHECKS}"
    echo -e "${GREEN}Passed: ${PASS_COUNT}${NC}"
    echo -e "${RED}Failed: ${FAIL_COUNT}${NC}"
    echo -e "${YELLOW}Warnings: ${WARN_COUNT}${NC}"
    echo ""
    
    if [ $FAIL_COUNT -eq 0 ]; then
        echo -e "${GREEN}✅ VALIDATION PASSED${NC}"
        echo "====================================="
        echo "Exit code: 0"
        exit 0
    else
        echo -e "${RED}❌ VALIDATION FAILED: $FAIL_COUNT issues${NC}"
        echo "====================================="
        echo ""
        echo "Remediation:"
        echo "$REMEDIATION_STEPS"
        echo ""
        echo "Exit code: 1"
        exit 1
    fi
}

json_has_key() {
    local file="$1"
    local key="$2"
    if command -v jq &> /dev/null; then
        jq -e "$key" "$file" &> /dev/null
        return $?
    else
        # Fallback a grep si no hay jq
        grep -q "\"$key\"" "$file"
        return $?
    fi
}

json_count_array() {
    local file="$1"
    local key="$2"
    if command -v jq &> /dev/null; then
        jq -r "$key | length" "$file" 2>/dev/null || echo "0"
    else
        grep -c "\"$key\"" "$file" 2>/dev/null || echo "0"
    fi
}

#===============================================================================
# Validation Functions
#===============================================================================

validate_json() {
    print_section "JSON Structure Validation"
    
    # Check file exists
    if [ -f "$OPENCODE_CONFIG" ]; then
        print_check "OpenCode config file exists" "PASS" "$OPENCODE_CONFIG"
    else
        print_check "OpenCode config file exists" "FAIL" "File not found: $OPENCODE_CONFIG"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create or copy opencode.json to $OPENCODE_CONFIG"
        return
    fi
    
    # Validate JSON syntax
    if command -v jq &> /dev/null; then
        if jq empty "$OPENCODE_CONFIG" 2>/dev/null; then
            print_check "JSON syntax valid" "PASS"
        else
            print_check "JSON syntax valid" "FAIL" "Invalid JSON syntax"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Fix JSON syntax in $OPENCODE_CONFIG"
            return
        fi
    else
        print_check "JSON syntax valid" "WARN" "jq not installed, skipping syntax check"
    fi
    
    # Check required top-level keys
    local required_keys=("agent" "mcp" "permission")
    for key in "${required_keys[@]}"; do
        if json_has_key "$OPENCODE_CONFIG" ".$key"; then
            print_check "Top-level key: $key" "PASS"
        else
            print_check "Top-level key: $key" "FAIL" "Missing required key"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add '$key' to opencode.json"
        fi
    done
}

validate_agents() {
    print_section "Agent Definitions Validation"
    
    # Required Grimorio agents
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
    )
    
    # Required SDD agents
    local required_sdd_agents=(
        "gentle-orchestrator"
        "sdd-apply"
        "sdd-archive"
        "sdd-design"
        "sdd-explore"
        "sdd-init"
        "sdd-onboard"
        "sdd-propose"
        "sdd-spec"
        "sdd-tasks"
        "sdd-verify"
    )
    
    # Check Grimorio agents
    print_check "Grimorio Agents" "PASS" "Checking ${#required_agents[@]} agents"
    for agent in "${required_agents[@]}"; do
        if json_has_key "$OPENCODE_CONFIG" ".agent.\"$agent\""; then
            # Check if has prompt
            if grep -q "\"$agent\"" "$OPENCODE_CONFIG" && grep -A 5 "\"$agent\"" "$OPENCODE_CONFIG" | grep -q "prompt"; then
                print_check "  └─ $agent" "PASS" "Has prompt"
            else
                print_check "  └─ $agent" "WARN" "Missing prompt definition"
            fi
        else
            print_check "  └─ $agent" "FAIL" "Agent not defined"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add agent '$agent' to opencode.json"
        fi
    done
    
    # Check SDD agents
    print_check "SDD Agents" "PASS" "Checking ${#required_sdd_agents[@]} agents"
    for agent in "${required_sdd_agents[@]}"; do
        if json_has_key "$OPENCODE_CONFIG" ".agent.\"$agent\""; then
            print_check "  └─ $agent" "PASS"
        else
            print_check "  └─ $agent" "FAIL" "Agent not defined"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add agent '$agent' to opencode.json"
        fi
    done
    
    # Check agent tools
    print_section "Agent Tools Validation"
    
    # grimorio-architect should have delegate tool
    if grep -A 20 '"grimorio-architect"' "$OPENCODE_CONFIG" | grep -q '"delegate"'; then
        print_check "grimorio-architect has delegate tool" "PASS"
    else
        print_check "grimorio-architect has delegate tool" "FAIL" "Missing delegate tool"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add delegate tool to grimorio-architect"
    fi
    
    # grimorio-areas should have bash, read, write, grep
    local areas_tools=("bash" "read" "write" "grep")
    local areas_section=$(grep -A 30 '"grimorio-areas"' "$OPENCODE_CONFIG")
    for tool in "${areas_tools[@]}"; do
        if echo "$areas_section" | grep -q "\"$tool\""; then
            print_check "grimorio-areas has $tool tool" "PASS"
        else
            print_check "grimorio-areas has $tool tool" "WARN" "Missing $tool tool"
        fi
    done
}

validate_skills() {
    print_section "Skills Validation"
    
    # Check skill registry exists
    if [ -f "$SKILL_REGISTRY" ]; then
        print_check "Skill registry exists" "PASS" "$SKILL_REGISTRY"
    else
        print_check "Skill registry exists" "WARN" "File not found: $SKILL_REGISTRY"
    fi
    
    # Required Grimorio skills
    local required_skills=(
        "grimorio-architect"
        "grimorio-areas"
        "grimorio-npc"
        "grimorio-narrative-custodian"
    )
    
    # Check skills directory
    local skills_dir="$HOME/.config/opencode/skills"
    if [ -d "$skills_dir" ]; then
        print_check "Skills directory exists" "PASS" "$skills_dir"
        
        # Count skill directories
        local skill_count=$(find "$skills_dir" -maxdepth 1 -type d -name "grimorio-*" | wc -l)
        print_check "Grimorio skills found" "PASS" "$skill_count skills"
        
        # Check specific skills
        for skill in "${required_skills[@]}"; do
            if [ -d "$skills_dir/$skill" ] && [ -f "$skills_dir/$skill/SKILL.md" ]; then
                print_check "  └─ $skill" "PASS" "SKILL.md exists"
            else
                print_check "  └─ $skill" "WARN" "Skill directory or SKILL.md missing"
            fi
        done
    else
        print_check "Skills directory exists" "WARN" "Directory not found: $skills_dir"
    fi
    
    # Check if skills are referenced in agent prompts
    print_check "Skills referenced in agents" "PASS" "Checking prompts..."
    
    # grimorio-architect should reference templates
    if grep -A 100 '"grimorio-architect"' "$OPENCODE_CONFIG" | grep -q "templates"; then
        print_check "  └─ grimorio-architect references templates" "PASS"
    else
        print_check "  └─ grimorio-architect references templates" "WARN" "No template references found"
    fi
    
    # grimorio-areas should reference areas.md.tmpl
    if grep -A 50 '"grimorio-areas"' "$OPENCODE_CONFIG" | grep -q "areas.md.tmpl"; then
        print_check "  └─ grimorio-areas references areas.md.tmpl" "PASS"
    else
        print_check "  └─ grimorio-areas references areas.md.tmpl" "FAIL" "Missing template reference"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add template reference to grimorio-areas prompt"
    fi
}

validate_sdd() {
    print_section "SDD Configuration Validation"
    
    # Check SDD config exists
    if json_has_key "$OPENCODE_CONFIG" ".sdd"; then
        print_check "SDD configuration exists" "PASS"
    else
        print_check "SDD configuration exists" "FAIL" "Missing .sdd section"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add SDD configuration to opencode.json"
        return
    fi
    
    # Check delivery_strategy
    if grep -q '"delivery_strategy"' "$OPENCODE_CONFIG"; then
        local strategy=$(grep -A 1 '"delivery_strategy"' "$OPENCODE_CONFIG" | grep -oE '(exception-ok|ask-on-risk|auto-chain|single-pr)')
        print_check "delivery_strategy" "PASS" "$strategy"
    else
        print_check "delivery_strategy" "FAIL" "Not configured"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add delivery_strategy to SDD config"
    fi
    
    # Check chain_strategy
    if grep -q '"chain_strategy"' "$OPENCODE_CONFIG"; then
        local chain=$(grep -A 1 '"chain_strategy"' "$OPENCODE_CONFIG" | grep -oE '(stacked-to-main|feature-branch-chain)')
        print_check "chain_strategy" "PASS" "$chain"
    else
        print_check "chain_strategy" "FAIL" "Not configured"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add chain_strategy to SDD config"
    fi
    
    # Check artifact_store
    if grep -q '"artifact_store"' "$OPENCODE_CONFIG"; then
        local store=$(grep -A 1 '"artifact_store"' "$OPENCODE_CONFIG" | grep -oE '(engram|openspec|hybrid|none)')
        print_check "artifact_store" "PASS" "$store"
    else
        print_check "artifact_store" "WARN" "Not configured (default: engram)"
    fi
}

validate_mcp() {
    print_section "MCP Servers Validation"
    
    # Check MCP config exists
    if json_has_key "$OPENCODE_CONFIG" ".mcp"; then
        print_check "MCP configuration exists" "PASS"
    else
        print_check "MCP configuration exists" "FAIL" "Missing .mcp section"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add MCP configuration to opencode.json"
        return
    fi
    
    # Check required MCP servers
    local required_mcp=("context7" "engram" "grimorio")
    
    for mcp in "${required_mcp[@]}"; do
        if json_has_key "$OPENCODE_CONFIG" ".mcp.\"$mcp\""; then
            # Check if enabled
            if grep -A 5 "\"$mcp\"" "$OPENCODE_CONFIG" | grep -q '"enabled": true'; then
                print_check "MCP: $mcp" "PASS" "Enabled"
            elif grep -A 5 "\"$mcp\"" "$OPENCODE_CONFIG" | grep -q '"enabled": false'; then
                print_check "MCP: $mcp" "WARN" "Disabled"
            else
                print_check "MCP: $mcp" "PASS" "Enabled (default)"
            fi
        else
            print_check "MCP: $mcp" "FAIL" "Not configured"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add MCP server '$mcp' to opencode.json"
        fi
    done
    
    # Check grimorio MCP path
    local grimorio_path=$(grep -A 5 '"grimorio"' "$OPENCODE_CONFIG" | grep '"command"' | grep -oE '/[^\"]+' | head -1)
    if [ -n "$grimorio_path" ] && [ -f "$grimorio_path" ]; then
        print_check "Grimorio MCP binary exists" "PASS" "$grimorio_path"
    else
        print_check "Grimorio MCP binary exists" "WARN" "Path: $grimorio_path"
    fi
}

validate_command() {
    print_section "Command Definitions Validation"
    
    # Check grimorio command exists
    if json_has_key "$OPENCODE_CONFIG" ".command.grimorio"; then
        print_check "Command: /grimorio" "PASS"
        
        # Check agent reference
        if grep -A 5 '"grimorio"' "$OPENCODE_CONFIG" | grep -q '"agent": "grimorio-architect"'; then
            print_check "  └─ Uses grimorio-architect" "PASS"
        else
            print_check "  └─ Uses grimorio-architect" "FAIL" "Wrong agent"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Set agent to grimorio-architect"
        fi
        
        # Check template references
        if grep -A 200 '"grimorio"' "$OPENCODE_CONFIG" | grep -q "templates"; then
            print_check "  └─ References templates" "PASS"
        else
            print_check "  └─ References templates" "WARN" "No template references"
        fi
        
        # Check WotC validation
        if grep -A 200 '"grimorio"' "$OPENCODE_CONFIG" | grep -q "validate-campaign.sh"; then
            print_check "  └─ WotC validation included" "PASS"
        else
            print_check "  └─ WotC validation included" "FAIL" "Missing validation script reference"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add WotC validation to /grimorio command"
        fi
    else
        print_check "Command: /grimorio" "FAIL" "Not defined"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add /grimorio command to opencode.json"
    fi
}

validate_build() {
    print_section "Build Validation"
    
    # Check Go installation
    if command -v go &> /dev/null; then
        local go_version=$(go version)
        print_check "Go installed" "PASS" "$go_version"
    else
        print_check "Go installed" "FAIL" "Go not found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Install Go 1.23+\n"
        return
    fi
    
    # Check grimorio binary
    if [ -f "$GRIMORIO_BINARY" ]; then
        print_check "Grimorio binary exists" "PASS" "$GRIMORIO_BINARY"
        
        # Check if executable
        if [ -x "$GRIMORIO_BINARY" ]; then
            print_check "Grimorio binary executable" "PASS"
        else
            print_check "Grimorio binary executable" "WARN" "Not executable"
        fi
        
        # Check binary version
        local binary_version=$("$GRIMORIO_BINARY" --version 2>/dev/null || echo "unknown")
        print_check "Grimorio binary version" "PASS" "$binary_version"
    else
        print_check "Grimorio binary exists" "FAIL" "Not found: $GRIMORIO_BINARY"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Build grimorio binary: cd $PROJECT_ROOT && go build -o grimorio ./cmd/grimorio\n"
    fi
    
    # Check if git repo
    if [ -d "$PROJECT_ROOT/.git" ]; then
        local last_commit=$(git -C "$PROJECT_ROOT" log -1 --oneline 2>/dev/null || echo "unknown")
        print_check "Last commit" "PASS" "$last_commit"
        
        # Check if working tree clean
        if git -C "$PROJECT_ROOT" diff --quiet 2>/dev/null; then
            print_check "Working tree clean" "PASS"
        else
            print_check "Working tree clean" "WARN" "Uncommitted changes"
        fi
    else
        print_check "Git repository" "WARN" "Not a git repo"
    fi
}

validate_templates() {
    print_section "Template Files Validation"
    
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
    
    if [ -d "$templates_dir" ]; then
        print_check "Templates directory exists" "PASS" "$templates_dir"
        
        for tmpl in "${required_templates[@]}"; do
            if [ -f "$templates_dir/$tmpl" ]; then
                local lines=$(wc -l < "$templates_dir/$tmpl")
                print_check "  └─ $tmpl" "PASS" "$lines lines"
            else
                print_check "  └─ $tmpl" "FAIL" "Missing template"
                REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create template: $templates_dir/$tmpl\n"
            fi
        done
    else
        print_check "Templates directory exists" "FAIL" "Not found: $templates_dir"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create templates directory\n"
    fi
}

validate_scripts() {
    print_section "Scripts Validation"
    
    local scripts_dir="$PROJECT_ROOT/scripts"
    local required_scripts=(
        "validate-campaign.sh"
    )
    
    if [ -d "$scripts_dir" ]; then
        print_check "Scripts directory exists" "PASS" "$scripts_dir"
        
        for script in "${required_scripts[@]}"; do
            if [ -f "$scripts_dir/$script" ]; then
                if [ -x "$scripts_dir/$script" ]; then
                    print_check "  └─ $script" "PASS" "Executable"
                else
                    print_check "  └─ $script" "WARN" "Not executable"
                fi
            else
                print_check "  └─ $script" "FAIL" "Missing script"
                REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create script: $scripts_dir/$script\n"
            fi
        done
    else
        print_check "Scripts directory exists" "FAIL" "Not found: $scripts_dir"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create scripts directory\n"
    fi
}

#===============================================================================
# Main Script
#===============================================================================

# Parse arguments
for arg in "$@"; do
    case $arg in
        --check=*)
            CHECK_TYPE="${arg#*=}"
            shift
            ;;
    esac
done

# Initialize remediation steps
REMEDIATION_STEPS=""

# Print header
print_header

# Run validations based on check type
case $CHECK_TYPE in
    json)
        validate_json
        ;;
    agents)
        validate_agents
        ;;
    skills)
        validate_skills
        ;;
    sdd)
        validate_sdd
        ;;
    mcp)
        validate_mcp
        ;;
    build)
        validate_build
        ;;
    templates)
        validate_templates
        ;;
    scripts)
        validate_scripts
        ;;
    all)
        validate_json
        validate_agents
        validate_skills
        validate_sdd
        validate_mcp
        validate_command
        validate_build
        validate_templates
        validate_scripts
        ;;
    *)
        echo "Error: Invalid check type: $CHECK_TYPE"
        echo "Valid options: all, json, agents, skills, sdd, mcp, build, templates, scripts"
        exit 2
        ;;
esac

# Print footer and exit
print_footer
