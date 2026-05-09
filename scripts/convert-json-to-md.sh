#!/bin/bash
# convert-json-to-md.sh — Converts characters/*.json → characters.md and quests/*.json → quests.md

set -e

CAMPAIGN=$1
BASE="/home/pau/campaigns/$CAMPAIGN"

if [ -z "$CAMPAIGN" ]; then
    echo "Usage: ./convert-json-to-md.sh {campaign_name}"
    exit 1
fi

if [ ! -d "$BASE" ]; then
    echo "Error: Campaign directory not found: $BASE"
    exit 1
fi

echo "=== Converting JSON to Markdown for: $CAMPAIGN ==="
echo ""

# Characters
if [ -d "$BASE/characters" ] && ls "$BASE/characters"/*.json >/dev/null 2>&1; then
    echo "Converting characters..."
    
    echo "# Personajes Jugadores" > "$BASE/characters/characters.md"
    echo "" >> "$BASE/characters/characters.md"
    echo "*Fichas completas de los personajes de los jugadores.*" >> "$BASE/characters/characters.md"
    echo "" >> "$BASE/characters/characters.md"
    echo "---" >> "$BASE/characters/characters.md"
    
    for json in "$BASE/characters"/*.json; do
        name=$(jq -r '.name // "Unknown"' "$json")
        race=$(jq -r '.race // "Unknown"' "$json")
        class=$(jq -r '.class // "Unknown"' "$json")
        level=$(jq -r '.level // 1' "$json")
        background=$(jq -r '.background // "Unknown"' "$json")
        alignment=$(jq -r '.alignment // "Unknown"' "$json")
        hp_max=$(jq -r '.hp.maximum // 0' "$json")
        ac=$(jq -r '.ac // 10' "$json")
        str=$(jq -r '.stats.str // 10' "$json")
        dex=$(jq -r '.stats.dex // 10' "$json")
        con=$(jq -r '.stats.con // 10' "$json")
        int=$(jq -r '.stats.int // 10' "$json")
        wis=$(jq -r '.stats.wis // 10' "$json")
        cha=$(jq -r '.stats.cha // 10' "$json")
        
        echo "" >> "$BASE/characters/characters.md"
        echo "## $name" >> "$BASE/characters/characters.md"
        echo "" >> "$BASE/characters/characters.md"
        echo "- **Raza:** $race" >> "$BASE/characters/characters.md"
        echo "- **Clase:** $class $level" >> "$BASE/characters/characters.md"
        echo "- **Antecedente:** $background" >> "$BASE/characters/characters.md"
        echo "- **Alineamiento:** $alignment" >> "$BASE/characters/characters.md"
        echo "- **AC:** $ac" >> "$BASE/characters/characters.md"
        echo "- **HP:** $hp_max" >> "$BASE/characters/characters.md"
        echo "" >> "$BASE/characters/characters.md"
        echo "### Estadísticas" >> "$BASE/characters/characters.md"
        echo "| STR | DEX | CON | INT | WIS | CHA |" >> "$BASE/characters/characters.md"
        echo "|-----|-----|-----|-----|-----|-----|" >> "$BASE/characters/characters.md"
        echo "| $str | $dex | $con | $int | $wis | $cha |" >> "$BASE/characters/characters.md"
        echo "" >> "$BASE/characters/characters.md"
        echo "---" >> "$BASE/characters/characters.md"
    done
    
    count=$(ls -1 "$BASE/characters"/*.json | wc -l)
    echo "  ✅ Characters: $count files converted"
else
    echo "  ℹ️  No character JSON files found"
fi

# Quests
if [ -d "$BASE/quests" ] && ls "$BASE/quests"/*.json >/dev/null 2>&1; then
    echo "Converting quests..."
    
    echo "# Misiones y Quests" > "$BASE/quests/quests.md"
    echo "" >> "$BASE/quests/quests.md"
    echo "*Lista de misiones activas, completadas y fallidas.*" >> "$BASE/quests/quests.md"
    echo "" >> "$BASE/quests/quests.md"
    echo "---" >> "$BASE/quests/quests.md"
    
    for json in "$BASE/quests"/*.json; do
        title=$(jq -r '.title // "Untitled Quest"' "$json")
        type=$(jq -r '.type // "unknown"' "$json")
        status=$(jq -r '.status // "unknown"' "$json")
        hook=$(jq -r '.hook // "No hook provided"' "$json")
        stakes=$(jq -r '.stakes // "No stakes defined"' "$json")
        reward_desc=$(jq -r '.reward.description // "No reward"' "$json")
        reward_value=$(jq -r '.reward.value // "0"' "$json")
        
        echo "" >> "$BASE/quests/quests.md"
        echo "## $title" >> "$BASE/quests/quests.md"
        echo "" >> "$BASE/quests/quests.md"
        echo "- **Tipo:** $type" >> "$BASE/quests/quests.md"
        echo "- **Estado:** $status" >> "$BASE/quests/quests.md"
        echo "" >> "$BASE/quests/quests.md"
        echo "### Gancho" >> "$BASE/quests/quests.md"
        echo "$hook" >> "$BASE/quests/quests.md"
        echo "" >> "$BASE/quests/quests.md"
        echo "### stakes" >> "$BASE/quests/quests.md"
        echo "$stakes" >> "$BASE/quests/quests.md"
        echo "" >> "$BASE/quests/quests.md"
        echo "### Recompensas" >> "$BASE/quests/quests.md"
        echo "- $reward_desc ($reward_value)" >> "$BASE/quests/quests.md"
        echo "" >> "$BASE/quests/quests.md"
        echo "---" >> "$BASE/quests/quests.md"
    done
    
    count=$(ls -1 "$BASE/quests"/*.json | wc -l)
    echo "  ✅ Quests: $count files converted"
else
    echo "  ℹ️  No quest JSON files found"
fi

echo ""
echo "=== Conversion Complete ==="
echo "Ready to compile PDF: grimorio compile-pdf --campaign $CAMPAIGN"
