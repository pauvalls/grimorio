# SDD Delta Spec: Fix Grimorio Installer

**Date:** 2026-05-12  
**Type:** Bug Fix / Enhancement  
**Scope:** `install.sh`, `scripts/install.sh`  
**Priority:** High

---

## Problem Statement

The current installer (`install.sh` and `scripts/install.sh`) has incomplete copy logic:

1. **Skills**: Only copies `grimorio-*.md` files, but skills are now structured as directories (`skills/grimorio-*/SKILL.md`)
2. **Agents**: Copies to plugin directories but doesn't have a dedicated function with proper error handling
3. **Binary**: Only copies to `~/.local/bin`, missing `~/.config/opencode/plugins/grimorio/`
4. **Templates**: Copy logic exists but lacks proper reporting
5. **MCP Config**: `.mcp.json` is created inline, not copied from source
6. **grimorio-architect skill**: Lives in `~/.config/opencode/skills/` but should be synced from project root
7. **No unified copy functions**: Logic is scattered across `setup_plugin()` with no individual success/failure reporting

---

## Goals

1. Add dedicated `copy_*()` functions for each component type
2. All copy functions must be idempotent (overwrite existing files)
3. All copy functions must report success/failure with colors
4. All copy functions must create target directories with `mkdir -p`
5. Use `cp -r` for recursive copies (directories)
6. Main flow calls copy functions after verify functions
7. Handle `grimorio-architect` skill sync: project root → opencode skills

---

## Delta Specification

### 1. New Function: `copy_skills()`

**Location:** `install.sh` (after `setup_plugin()` function)

**Purpose:** Copy ALL grimorio skills from `$PROJECT_ROOT/skills/grimorio-*/` to `~/.config/opencode/skills/`

**Implementation:**
```bash
copy_skills() {
    log "Copying Grimorio skills..."
    local src_skills_dir="$PROJECT_ROOT/skills"
    local dst_skills_dir="$HOME/.config/opencode/skills"
    local copied=0
    local failed=0

    # Create destination directory
    mkdir -p "$dst_skills_dir" || {
        error "Failed to create skills directory: $dst_skills_dir"
        return 1
    }

    # Copy each grimorio-* skill directory
    for skill_dir in "$src_skills_dir"/grimorio-*/; do
        if [ -d "$skill_dir" ]; then
            local skill_name=$(basename "$skill_dir")
            if cp -r "$skill_dir" "$dst_skills_dir/"; then
                log "Copied skill: $skill_name"
                copied=$((copied + 1))
            else
                warn "Failed to copy skill: $skill_name"
                failed=$((failed + 1))
            fi
        fi
    done

    if [ $copied -gt 0 ]; then
        success "Copied $copied skill(s) to $dst_skills_dir"
    fi

    [ $failed -gt 0 ] && {
        warn "$failed skill(s) failed to copy"
        return 1
    }

    return 0
}
```

**Behavior:**
- Creates `~/.config/opencode/skills/` if missing
- Copies all `skills/grimorio-*/` directories recursively
- Reports count of copied/failed skills
- Returns 0 on success, 1 on any failure

---

### 2. New Function: `copy_agents()`

**Location:** `install.sh` (after `copy_skills()`)

**Purpose:** Copy ALL agents from `$PROJECT_ROOT/agents/*.md` to `~/.config/opencode/plugins/grimorio/agents/`

**Implementation:**
```bash
copy_agents() {
    log "Copying Grimorio agents..."
    local src_agents_dir="$PROJECT_ROOT/agents"
    local dst_agents_dir="$HOME/.config/opencode/plugins/grimorio/agents"
    local copied=0
    local failed=0

    # Create destination directory
    mkdir -p "$dst_agents_dir" || {
        error "Failed to create agents directory: $dst_agents_dir"
        return 1
    }

    # Copy each agent file
    for agent_file in "$src_agents_dir"/*.md; do
        if [ -f "$agent_file" ]; then
            local agent_name=$(basename "$agent_file")
            if cp "$agent_file" "$dst_agents_dir/"; then
                log "Copied agent: $agent_name"
                copied=$((copied + 1))
            else
                warn "Failed to copy agent: $agent_name"
                failed=$((failed + 1))
            fi
        fi
    done

    if [ $copied -gt 0 ]; then
        success "Copied $copied agent(s) to $dst_agents_dir"
    fi

    [ $failed -gt 0 ] && {
        warn "$failed agent(s) failed to copy"
        return 1
    }

    return 0
}
```

**Behavior:**
- Creates plugin agents directory if missing
- Copies all `agents/*.md` files
- Reports count of copied/failed agents
- Returns 0 on success, 1 on any failure

---

### 3. New Function: `copy_binary()`

**Location:** `install.sh` (after `copy_agents()`)

**Purpose:** Copy built binary to both `~/.config/opencode/plugins/grimorio/` and `~/.local/bin/`

**Implementation:**
```bash
copy_binary() {
    log "Copying Grimorio binary..."
    local src_binary="$PROJECT_ROOT/grimorio"
    local plugin_binary="$HOME/.config/opencode/plugins/grimorio/grimorio"
    local local_bin="$HOME/.local/bin/grimorio"
    local failed=0

    # Check source exists
    if [ ! -f "$src_binary" ]; then
        error "Binary not found: $src_binary (build first)"
        return 1
    fi

    # Copy to plugin directory
    mkdir -p "$(dirname "$plugin_binary")" || {
        error "Failed to create plugin directory"
        return 1
    }
    if cp "$src_binary" "$plugin_binary" && chmod +x "$plugin_binary"; then
        success "Binary copied to: $plugin_binary"
    else
        error "Failed to copy binary to plugin directory"
        failed=$((failed + 1))
    fi

    # Copy to ~/.local/bin (if directory exists or can be created)
    if mkdir -p "$(dirname "$local_bin")" 2>/dev/null; then
        if cp "$src_binary" "$local_bin" && chmod +x "$local_bin"; then
            success "Binary copied to: $local_bin"
        else
            warn "Failed to copy binary to: $local_bin"
            failed=$((failed + 1))
        fi
    else
        warn "Cannot create $HOME/.local/bin, skipping"
    fi

    [ $failed -gt 0 ] && return 1
    return 0
}
```

**Behavior:**
- Verifies source binary exists
- Creates target directories with `mkdir -p`
- Copies to both locations
- Makes binaries executable
- Reports success/failure for each location
- Returns 0 only if both copies succeed (or ~/.local/bin is skipped)

---

### 4. New Function: `copy_templates()`

**Location:** `install.sh` (after `copy_binary()`)

**Purpose:** Copy templates to `~/.config/opencode/plugins/grimorio/internal/compiler/templates/`

**Implementation:**
```bash
copy_templates() {
    log "Copying Grimorio templates..."
    local src_templates="$PROJECT_ROOT/internal/compiler/templates"
    local dst_templates="$HOME/.config/opencode/plugins/grimorio/internal/compiler/templates"
    local copied=0

    # Check source exists
    if [ ! -d "$src_templates" ]; then
        error "Templates directory not found: $src_templates"
        return 1
    fi

    # Create destination directory
    mkdir -p "$dst_templates" || {
        error "Failed to create templates directory: $dst_templates"
        return 1
    }

    # Copy all template files
    if cp -r "$src_templates"/* "$dst_templates/"; then
        copied=$(find "$dst_templates" -type f | wc -l)
        success "Copied $copied template file(s) to $dst_templates"
        return 0
    else
        error "Failed to copy templates"
        return 1
    fi
}
```

**Behavior:**
- Verifies source templates directory exists
- Creates destination directory recursively
- Copies all files recursively
- Reports count of copied files
- Returns 0 on success, 1 on failure

---

### 5. New Function: `copy_mcp_config()`

**Location:** `install.sh` (after `copy_templates()`)

**Purpose:** Copy `.mcp.json` to `~/.config/opencode/plugins/grimorio/.mcp.json`

**Implementation:**
```bash
copy_mcp_config() {
    log "Copying MCP configuration..."
    local src_mcp="$PROJECT_ROOT/.mcp.json"
    local dst_mcp="$HOME/.config/opencode/plugins/grimorio/.mcp.json"

    # Check source exists
    if [ ! -f "$src_mcp" ]; then
        error "MCP config not found: $src_mcp"
        return 1
    fi

    # Create destination directory if needed
    mkdir -p "$(dirname "$dst_mcp")" || {
        error "Failed to create plugin directory"
        return 1
    }

    # Copy config file
    if cp "$src_mcp" "$dst_mcp"; then
        success "MCP config copied to: $dst_mcp"
        return 0
    else
        error "Failed to copy MCP config"
        return 1
    fi
}
```

**Behavior:**
- Verifies source `.mcp.json` exists
- Creates destination directory if needed
- Copies config file (overwrites existing)
- Reports success/failure
- Returns 0 on success, 1 on failure

---

### 6. New Function: `sync_grimorio_architect_skill()`

**Location:** `install.sh` (after `copy_mcp_config()`, before `copy_skills()`)

**Purpose:** Move `grimorio-architect` skill from `~/.config/opencode/skills/` to `$PROJECT_ROOT/skills/` FIRST, then `copy_skills()` will handle distribution

**Implementation:**
```bash
sync_grimorio_architect_skill() {
    log "Syncing grimorio-architect skill..."
    local src_skill="$HOME/.config/opencode/skills/grimorio-architect"
    local dst_skill="$PROJECT_ROOT/skills/grimorio-architect"
    local backup_dir="$PROJECT_ROOT/skills/.grimorio-architect-backup"

    # If source doesn't exist, nothing to sync
    if [ ! -d "$src_skill" ]; then
        log "grimorio-architect skill not found in opencode skills (may be first install)"
        return 0
    fi

    # If destination exists, backup before overwrite
    if [ -d "$dst_skill" ]; then
        log "Backing up existing grimorio-architect skill..."
        rm -rf "$backup_dir"
        mv "$dst_skill" "$backup_dir"
    fi

    # Copy from opencode skills to project root
    mkdir -p "$(dirname "$dst_skill")"
    if cp -r "$src_skill" "$dst_skill"; then
        success "Synced grimorio-architect skill to project root"
        log "Backup preserved at: $backup_dir (remove after verification)"
        return 0
    else
        error "Failed to sync grimorio-architect skill"
        # Restore backup if copy failed
        if [ -d "$backup_dir" ]; then
            rm -rf "$dst_skill"
            mv "$backup_dir" "$dst_skill"
        fi
        return 1
    fi
}
```

**Behavior:**
- Checks if skill exists in opencode skills
- Backs up existing project skill if present
- Copies from opencode to project root
- Preserves backup for manual verification
- Returns 0 on success (or if source doesn't exist), 1 on failure

---

### 7. Updated Main Flow

**Location:** `install.sh` - `main()` function

**Current flow (simplified):**
```bash
main() {
    clean_installation
    detect_platform
    install_go
    install_wkhtmltopdf
    setup_repo
    build_binary
    migrate_existing_campaigns
    setup_plugin          # ← Has partial copy logic
    configure_opencode_mcp
    configure_opencode_command
    copy_commands
    configure_opencode_agents
    configure_shell
    print_instructions
}
```

**New flow (with copy functions):**
```bash
main() {
    clean_installation
    detect_platform
    install_go
    install_wkhtmltopdf
    setup_repo
    build_binary
    migrate_existing_campaigns
    
    # NEW: Sync grimorio-architect skill FIRST
    sync_grimorio_architect_skill
    
    # NEW: Copy all components with dedicated functions
    copy_binary
    copy_agents
    copy_skills
    copy_templates
    copy_mcp_config
    
    # Existing plugin setup (now simplified or removed if redundant)
    setup_plugin
    
    configure_opencode_mcp
    configure_opencode_command
    copy_commands
    configure_opencode_agents
    configure_shell
    print_instructions
}
```

**Changes:**
1. `sync_grimorio_architect_skill()` called after `build_binary`
2. All `copy_*()` functions called in sequence after sync
3. `setup_plugin()` kept for backward compatibility but can be simplified

---

## Expected Behavior

### Scenario 1: Fresh Install
- All copy functions succeed
- Skills: All `skills/grimorio-*/` copied to `~/.config/opencode/skills/`
- Agents: All `agents/*.md` copied to plugin directory
- Binary: Copied to both plugin directory and `~/.local/bin`
- Templates: All template files copied
- MCP Config: `.mcp.json` copied to plugin directory
- Result: Complete installation, all components in place

### Scenario 2: Reinstall
- All copy functions succeed with overwrite
- Existing files replaced with latest versions
- Backup created for `grimorio-architect` skill if it existed
- Result: Latest version in place, backups available

### Scenario 3: Partial Install
- Missing components get copied
- Existing components get updated
- Failed copies reported with warnings
- Result: Best-effort installation, user informed of failures

---

## Acceptance Criteria

- [ ] `copy_skills()` function exists and copies all `skills/grimorio-*/` directories
- [ ] `copy_agents()` function exists and copies all `agents/*.md` files
- [ ] `copy_binary()` function exists and copies to both locations
- [ ] `copy_templates()` function exists and copies all template files
- [ ] `copy_mcp_config()` function exists and copies `.mcp.json`
- [ ] `sync_grimorio_architect_skill()` function exists and syncs skill to project root
- [ ] All copy functions use `mkdir -p` for directory creation
- [ ] All copy functions use `cp -r` for recursive copies
- [ ] All copy functions report success/failure with colors
- [ ] All copy functions are idempotent (overwrite existing)
- [ ] Main flow calls all copy functions after verify/build functions
- [ ] Fresh install scenario works end-to-end
- [ ] Reinstall scenario works with overwrites
- [ ] Partial install scenario reports failures gracefully

---

## Files to Modify

1. **`install.sh`** - Main installer script
   - Add 6 new `copy_*()` functions
   - Update `main()` flow to call copy functions
   - Optionally simplify `setup_plugin()` to avoid duplication

2. **`scripts/install.sh`** - Development installer (optional)
   - Same copy functions can be added for consistency
   - Or keep as validation-only script

---

## Testing Strategy

1. **Fresh Install Test:**
   ```bash
   # Clean everything
   rm -rf ~/.config/opencode/skills/grimorio-*
   rm -rf ~/.config/opencode/plugins/grimorio
   rm -rf ~/.local/bin/grimorio
   
   # Run installer
   ./install.sh
   
   # Verify all components
   ls ~/.config/opencode/skills/grimorio-*/
   ls ~/.config/opencode/plugins/grimorio/agents/
   ls ~/.config/opencode/plugins/grimorio/internal/compiler/templates/
   ```

2. **Reinstall Test:**
   ```bash
   # Modify a skill file
   echo "# Modified" >> ~/.config/opencode/skills/grimorio-architect/SKILL.md
   
   # Run installer
   ./install.sh
   
   # Verify file was overwritten
   grep -c "# Modified" ~/.config/opencode/skills/grimorio-architect/SKILL.md
   # Should be 0 (overwritten)
   ```

3. **Partial Install Test:**
   ```bash
   # Remove templates directory
   rm -rf ~/.config/opencode/plugins/grimorio/internal
   
   # Run installer
   ./install.sh
   
   # Verify templates were copied
   ls ~/.config/opencode/plugins/grimorio/internal/compiler/templates/
   ```

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Copy overwrites user modifications | Medium | Create backups before overwrite (implemented for grimorio-architect) |
| Permission errors on copy | High | Use `mkdir -p` and report failures clearly |
| Source files missing | Medium | Check existence before copy, warn user |
| Disk space issues | Low | Unlikely for text files, binary already built |

---

## Dependencies

- `bash` 4.0+ (for associative arrays if needed)
- `cp`, `mkdir`, `chmod` (standard Unix utilities)
- Existing `build_binary()` function must run first

---

## Out of Scope

- Modifying `scripts/install.sh` (can be done in follow-up)
- Changing clean installation logic
- Modifying MCP configuration in `opencode.json`
- Shell configuration changes
