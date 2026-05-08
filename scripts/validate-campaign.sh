#!/bin/bash

#===============================================================================
# validate-campaign.sh — WotC Validation Script for Grimorio Campaigns
# Version: 2.4.0
#
# Usage: ./scripts/validate-campaign.sh {campaign-path} [--check=structure|wotc|references|all]
#
# Exit Codes:
#   0 - All validations passed
#   1 - One or more validations failed
#   2 - Error (invalid arguments, campaign not found, etc.)
#===============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASS_COUNT=0
FAIL_COUNT=0
TOTAL_CHECKS=0

#===============================================================================
# Helper Functions
#===============================================================================

print_header() {
    echo "====================================="
    echo "  Campaign Validation Report"
    echo "====================================="
    echo "Campaign: $(basename "$CAMPAIGN_PATH")"
    echo "Date: $(date +%Y-%m-%d)"
    echo ""
}

print_check() {
    local check_name="$1"
    local status="$2"
    local details="$3"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✅${NC} $check_name: $status"
        if [ -n "$details" ]; then
            echo "   $details"
        fi
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "${RED}❌${NC} $check_name: $status"
        if [ -n "$details" ]; then
            echo "   $details"
        fi
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

print_footer() {
    echo ""
    echo "====================================="
    if [ $FAIL_COUNT -eq 0 ]; then
        echo -e "${GREEN}  VALIDATION PASSED${NC}"
        echo "====================================="
        echo "Exit code: 0"
        exit 0
    else
        echo -e "${RED}  VALIDATION FAILED: $FAIL_COUNT issues${NC}"
        echo "====================================="
        echo ""
        echo "Remediation:"
        echo "$REMEDIATION_STEPS"
        echo ""
        echo "Exit code: 1"
        exit 1
    fi
}

#===============================================================================
# Validation Functions
#===============================================================================

validate_structure() {
    echo "Running Structure Validation..."
    echo ""
    
    local required_files=(
        "lore.md"
        "npcs/npcs_and_factions.md"
        "bestiary/bestiary.md"
        "encounters/encounters.md"
        "maps/maps_and_scenes.md"
    )
    
    local required_dirs=(
        "acts"
        "npcs"
        "bestiary"
        "encounters"
        "maps"
        "assets"
    )
    
    # Check directories
    for dir in "${required_dirs[@]}"; do
        if [ -d "$CAMPAIGN_PATH/$dir" ]; then
            file_count=$(find "$CAMPAIGN_PATH/$dir" -type f | wc -l)
            print_check "Directory: $dir" "PASS" "$file_count files found"
        else
            print_check "Directory: $dir" "FAIL" "Missing directory"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create missing directory: $dir"
        fi
    done
    
    # Check files
    for file in "${required_files[@]}"; do
        if [ -f "$CAMPAIGN_PATH/$file" ]; then
            file_size=$(wc -c < "$CAMPAIGN_PATH/$file")
            print_check "File: $file" "PASS" "$file_size bytes"
        else
            print_check "File: $file" "FAIL" "Missing file"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create missing file: $file"
        fi
    done
    
    # Check acts
    act_count=$(find "$CAMPAIGN_PATH/acts" -name "*.md" -type f 2>/dev/null | wc -l)
    if [ $act_count -ge 1 ]; then
        print_check "Acts" "PASS" "$act_count act files found"
    else
        print_check "Acts" "FAIL" "No act files found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Generate at least 1 act file"
    fi
    
    # Check assets
    asset_count=$(find "$CAMPAIGN_PATH/assets" -type f \( -name "*.png" -o -name "*.svg" \) 2>/dev/null | wc -l)
    if [ $asset_count -ge 1 ]; then
        print_check "Assets" "PASS" "$asset_count asset files found"
    else
        print_check "Assets" "FAIL" "No asset files found (warning: may be OK if not yet generated)"
    fi
    
    echo ""
}

validate_wotc_format() {
    echo "Running WotC Format Validation..."
    echo ""
    
    # Boxed Text Validation
    boxed_count=$(grep -c '^>>' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
    if [ "$boxed_count" -ge 1 ]; then
        # Validate word count per boxed text (sample check)
        invalid_boxed=0
        for act_file in "$CAMPAIGN_PATH"/acts/*.md; do
            while IFS= read -r line; do
                if [[ $line == ">>"* ]]; then
                    # Extract boxed text (simplified - just check next few lines)
                    word_count=$(grep -A 10 "^>>" "$act_file" | wc -w)
                    if [ $word_count -lt 100 ] || [ $word_count -gt 600 ]; then
                        invalid_boxed=$((invalid_boxed + 1))
                    fi
                fi
            done < "$act_file"
        done
        
        if [ $invalid_boxed -eq 0 ]; then
            print_check "Boxed Text" "PASS" "$boxed_count sections (100-600 words each)"
        else
            print_check "Boxed Text" "FAIL" "$invalid_boxed sections outside 100-600 word range"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Adjust boxed text word count in affected areas"
        fi
    else
        print_check "Boxed Text" "FAIL" "No boxed text sections found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add boxed text (>>) to each area"
    fi
    
    # Character Hooks Validation
    hook_count=$(grep -ci 'hook\|gancho' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
    area_count=$(grep -c '^## Area' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
    
    if [ $area_count -gt 0 ]; then
        hooks_per_area=$((hook_count / area_count))
        if [ $hooks_per_area -ge 2 ]; then
            print_check "Character Hooks" "PASS" "$hooks_per_area per area (≥2 required)"
        else
            print_check "Character Hooks" "FAIL" "$hooks_per_area per area (required: ≥2)"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add character hooks to areas with <2 hooks (tie to PC backgrounds/classes)"
        fi
    else
        print_check "Character Hooks" "FAIL" "No areas found to validate"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Generate area content first"
    fi
    
    # Developments Validation
    dev_count=$(grep -ci 'development\|desarrollo' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
    
    if [ $area_count -gt 0 ]; then
        devs_per_area=$((dev_count / area_count))
        if [ $devs_per_area -ge 3 ]; then
            # Check recovery paths
            recovery_count=$(grep -ci 'if.*fail\|si.*fallan\|recovery\|recuperación' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
            if [ $recovery_count -ge $dev_count ]; then
                print_check "Developments" "PASS" "$devs_per_area per area with recovery paths (≥3 required)"
            else
                print_check "Developments" "FAIL" "$recovery_count/$dev_count developments have recovery paths"
                REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add recovery paths to developments (add 'If PCs fail...' clause)"
            fi
        else
            print_check "Developments" "FAIL" "$devs_per_area per area (required: ≥3)"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add 3+ development branches to each area"
        fi
    else
        print_check "Developments" "FAIL" "No areas found to validate"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Generate area content first"
    fi
    
    # Running Guidance Validation
    guidance_count=$(grep -c 'Cómo Dirigir esta Escena\|Running the Scene' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
    
    if [ $guidance_count -ge 1 ]; then
        print_check "Running Guidance" "PASS" "$guidance_count sections found"
    else
        print_check "Running Guidance" "FAIL" "No running guidance sections found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add 'Cómo Dirigir esta Escena' section to each area"
    fi
    
    # Sidebars Validation
    sidebar_count=$(grep -c '^> #####' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
    act_count_files=$(find "$CAMPAIGN_PATH/acts" -name "*.md" -type f 2>/dev/null | wc -l)
    
    if [ $act_count_files -gt 0 ]; then
        sidebars_per_act=$((sidebar_count / act_count_files))
        if [ $sidebars_per_act -ge 1 ]; then
            print_check "Sidebars" "PASS" "$sidebar_count total (≥1 per act)"
        else
            print_check "Sidebars" "FAIL" "$sidebars_per_act per act (required: ≥1)"
            REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add at least 1 sidebar to each act (rules clarification, DM tip, or lore excerpt)"
        fi
    else
        print_check "Sidebars" "FAIL" "No act files found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Generate act content first"
    fi
    
    echo ""
}

validate_references() {
    echo "Running Cross-Reference Validation..."
    echo ""
    
    # Check creature references
    creature_errors=0
    if [ -f "$CAMPAIGN_PATH/bestiary/bestiary.md" ]; then
        creature_refs=$(grep -oE '\[[A-Z][a-z]+[A-Z][a-z]+\]' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | sort -u || true)
        for creature in $creature_refs; do
            creature_clean=$(echo "$creature" | tr -d '[]')
            if ! grep -q "$creature_clean" "$CAMPAIGN_PATH/bestiary/bestiary.md" 2>/dev/null; then
                creature_errors=$((creature_errors + 1))
                if [ $creature_errors -le 5 ]; then
                    REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add creature '$creature_clean' to bestiary OR fix reference in acts"
                fi
            fi
        done
        
        creature_total=$(echo "$creature_refs" | wc -w)
        if [ $creature_errors -eq 0 ]; then
            print_check "Creature References" "PASS" "$creature_total references validated"
        else
            print_check "Creature References" "FAIL" "$creature_errors/$creature_total creatures not found in bestiary"
        fi
    else
        print_check "Creature References" "FAIL" "Bestiary file not found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create bestiary/bestiary.md first"
    fi
    
    # Check NPC references
    npc_errors=0
    if [ -f "$CAMPAIGN_PATH/npcs/npcs_and_factions.md" ]; then
        npc_refs=$(grep -oE '\*[A-Z][a-z]+[A-Z][a-z]+\*' "$CAMPAIGN_PATH"/acts/*.md 2>/dev/null | sort -u || true)
        for npc in $npc_refs; do
            npc_clean=$(echo "$npc" | tr -d '*')
            if ! grep -q "$npc_clean" "$CAMPAIGN_PATH/npcs/npcs_and_factions.md" 2>/dev/null; then
                npc_errors=$((npc_errors + 1))
                if [ $npc_errors -le 5 ]; then
                    REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Add NPC '$npc_clean' to npcs_and_factions.md OR fix reference in acts"
                fi
            fi
        done
        
        npc_total=$(echo "$npc_refs" | wc -w)
        if [ $npc_errors -eq 0 ]; then
            print_check "NPC References" "PASS" "$npc_total references validated"
        else
            print_check "NPC References" "FAIL" "$npc_errors/$npc_total NPCs not found in npcs_and_factions.md"
        fi
    else
        print_check "NPC References" "FAIL" "NPCs file not found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Create npcs/npcs_and_factions.md first"
    fi
    
    echo ""
}

validate_completeness() {
    echo "Running Content Completeness Validation..."
    echo ""
    
    # Check for canon.json
    if [ -f "$CAMPAIGN_PATH/canon.json" ]; then
        print_check "Canon File" "PASS" "canon.json exists"
    else
        print_check "Canon File" "FAIL" "canon.json not found"
        REMEDIATION_STEPS="${REMEDIATION_STEPS}\n- Generate Adventure Bible (canon.json) first"
    fi
    
    # Check for narrative_state.json
    if [ -f "$CAMPAIGN_PATH/narrative_state.json" ]; then
        print_check "Narrative State" "PASS" "narrative_state.json exists"
    else
        print_check "Narrative State" "WARN" "narrative_state.json not found (optional for new campaigns)"
    fi
    
    # Check quests
    quest_count=$(find "$CAMPAIGN_PATH/quests" -name "*.md" -type f 2>/dev/null | wc -l)
    if [ $quest_count -ge 1 ]; then
        print_check "Quests" "PASS" "$quest_count quest files found"
    else
        print_check "Quests" "WARN" "No quest files found (optional)"
    fi
    
    # Check characters
    char_count=$(find "$CAMPAIGN_PATH/characters" -name "*.md" -type f 2>/dev/null | wc -l)
    if [ $char_count -ge 1 ]; then
        print_check "Characters" "PASS" "$char_count character files found"
    else
        print_check "Characters" "WARN" "No character files found (optional)"
    fi
    
    echo ""
}

#===============================================================================
# Main Script
#===============================================================================

# Parse arguments
CHECK_TYPE="all"
CAMPAIGN_PATH=""

for arg in "$@"; do
    case $arg in
        --check=*)
            CHECK_TYPE="${arg#*=}"
            shift
            ;;
        *)
            CAMPAIGN_PATH="$arg"
            ;;
    esac
done

# Validate arguments
if [ -z "$CAMPAIGN_PATH" ]; then
    echo "Error: Campaign path required"
    echo "Usage: $0 {campaign-path} [--check=structure|wotc|references|all]"
    exit 2
fi

if [ ! -d "$CAMPAIGN_PATH" ]; then
    echo "Error: Campaign directory not found: $CAMPAIGN_PATH"
    exit 2
fi

# Initialize remediation steps
REMEDIATION_STEPS=""

# Print header
print_header

# Run validations based on check type
case $CHECK_TYPE in
    structure)
        validate_structure
        ;;
    wotc)
        validate_wotc_format
        ;;
    references)
        validate_references
        ;;
    all)
        validate_structure
        validate_wotc_format
        validate_references
        validate_completeness
        ;;
    *)
        echo "Error: Invalid check type: $CHECK_TYPE"
        echo "Valid options: structure, wotc, references, all"
        exit 2
        ;;
esac

# Print footer and exit
print_footer
