# Design: Install Command Sync

## Architecture

```
install.sh
├── configure_opencode_command()
│   ├── Clean up deprecated agents
│   ├── sync_command_from_agent($OPENCODE_CONFIG) ← NEW
│   │   ├── Read agents/grimorio-architect.md
│   │   ├── Extract Phase 1 section (sed/awk)
│   │   ├── Format as template text
│   │   └── Update .command.grimorio.template via jq
│   └── Configure grimorio-architect agent
└── configure_opencode_mcp()
```

## Implementation Details

### sync_command_from_agent Function

```bash
sync_command_from_agent() {
    local OPENCODE_CONFIG="$1"
    local AGENT_FILE="${INSTALL_DIR}/agents/grimorio-architect.md"
    
    if [ ! -f "$AGENT_FILE" ]; then
        log "Agent file not found, using hardcoded template"
        return 0
    fi
    
    # Extract Phase 1 questions (between "### Phase 1" and "### Phase 2")
    local PHASE1_TEXT
    PHASE1_TEXT=$(sed -n '/^### Phase 1: Gather Requirements/,/^### Phase 2/p' "$AGENT_FILE" | \
                  head -n -1 | \
                  sed 's/^> //g' | \
                  sed '/^Ask the user/d')
    
    # Create template file
    local TEMPLATE_FILE=$(mktemp)
    cat > "$TEMPLATE_FILE" << EOF
Generate a D&D 5e campaign or one-shot from the user's idea.

## IMPORTANT: Use the grimorio-architect agent. It handles everything end-to-end.

## Workflow (followed by grimorio-architect)

${PHASE1_TEXT}

[Rest of workflow from agent file...]
EOF
    
    # Escape for JSON and update config
    local TEMPLATE_JSON
    TEMPLATE_JSON=$(cat "$TEMPLATE_FILE" | jq -Rs '.')
    rm -f "$TEMPLATE_FILE"
    
    jq --argjson template "$TEMPLATE_JSON" '.command.grimorio.template = $template' \
       "$OPENCODE_CONFIG" > "${OPENCODE_CONFIG}.tmp" && \
       mv "${OPENCODE_CONFIG}.tmp" "$OPENCODE_CONFIG"
    
    success "Command template synced from grimorio-architect.md"
}
```

## File Changes

| File | Change Type | Description |
|------|-------------|-------------|
| `install.sh` | Add function | `sync_command_from_agent()` |
| `install.sh` | Modify | `configure_opencode_command()` to call sync function |
| `install.sh` | Update | Hardcoded template (fallback) with Question 3 |

## Tradeoffs

### Option A: Full Sync (Chosen)
- **Pros**: Single source of truth, automatic updates
- **Cons**: More complex bash parsing, potential for extraction errors

### Option B: Partial Sync
- Only sync Phase 1 questions, keep rest hardcoded
- **Pros**: Simpler parsing, less risk
- **Cons**: Still some duplication

### Option C: No Sync (Rejected)
- Keep hardcoded template only
- **Pros**: Simple, reliable
- **Cons**: Drift between agent and command, manual updates required

## Testing Strategy

1. **Unit Test**: Run sync_command_from_agent in isolation
2. **Integration Test**: Full install.sh run
3. **Regression Test**: Verify existing installations update correctly
4. **Edge Case**: Missing agent file, malformed agent file

## Rollback Plan

If sync fails:
1. Hardcoded template still works as fallback
2. Manual fix: `jq '.command.grimorio.template = "..."` on opencode.json
3. Re-run install.sh after fixing agent file
