# Specification: Clean Installer v2

## Functional Requirements

### FR1: Complete Cleanup
The installer MUST remove ALL previous installation artifacts:
- Plugin directories (`~/.claude/plugins/grimorio`, `~/.config/opencode/plugins/grimorio`)
- Binary files (`~/.local/bin/grimorio`, `~/.local/bin/migrate-v1-to-v2`)
- Install directory (`~/.local/share/grimorio`)
- opencode.json entries (`.mcp.grimorio`, `.agent.grimorio-*`, `.command.grimorio`)
- Shell configuration (PATH entries in .bashrc, .zshrc, etc.)

### FR2: Fresh Clone
The installer MUST clone the repository fresh:
- Remove existing `~/.local/share/grimorio` before cloning
- Clone from `https://github.com/pauvalls/grimorio.git` main branch
- Use the cloned repository for all subsequent steps

### FR3: Build Binaries
The installer MUST compile Go binaries:
- Install Go if not present (version 1.23+)
- Build `grimorio` binary from `./cmd/grimorio`
- Build `migrate-v1-to-v2` binary from `./cmd/migrate-v1-to-v2`
- Install to `~/.local/bin/`

### FR4: Install All Components
The installer MUST copy ALL components to plugin directories:
- Binary (`grimorio`, `migrate-v1-to-v2`)
- Agents (`agents/grimorio-*.md`)
- Skills (`skills/grimorio-*.md`)
- Templates (`internal/compiler/templates/`)
- MCP config (`.mcp.json`)

### FR5: Configure opencode.json
The installer MUST update opencode.json with:
- `.mcp.grimorio` configuration
- ALL `.agent.grimorio-*` entries (architect, artist, cartographer, lore, npc, bestiary, encounters, areas, quests, maps, characters, narrative-custodian, introduction, setting-guide, appendices, integrator)
- `.command.grimorio` with full template

### FR6: Configure Shell
The installer MUST add PATH entries to shell config:
- `~/.local/bin` to PATH
- `~/.local/go/bin` to PATH (if Go installed)
- Use marked block for easy cleanup

### FR7: Migrate Campaigns
The installer MUST migrate existing v1 campaigns:
- Detect campaigns without `canon.json`
- Run `migrate-v1-to-v2` automatically
- Create backups in `.v1-backup/`

## Non-Functional Requirements

### NFR1: Idempotency
Running the installer multiple times MUST produce the same result.

### NFR2: Error Handling
The installer MUST:
- Continue on non-critical errors
- Report all errors clearly
- Exit with appropriate codes

### NFR3: Progress Reporting
The installer MUST show progress for each major step.

### NFR4: Backup Safety
The installer MUST backup user data (campaigns) before any destructive operation.

## Acceptance Criteria

### AC1: Clean Run
```bash
./install.sh 2>&1 | grep -E "\[SUCCESS\]|\[ERROR\]" | wc -l
# Expected: Multiple SUCCESS, zero ERROR
```

### AC2: All Components Installed
```bash
# Check binaries
test -x ~/.local/bin/grimorio && echo "✅ Binary"
test -x ~/.local/bin/migrate-v1-to-v2 && echo "✅ Migration tool"

# Check plugins
test -d ~/.config/opencode/plugins/grimorio/agents && echo "✅ Agents"
test -f ~/.config/opencode/plugins/grimorio/.mcp.json && echo "✅ MCP config"

# Check opencode.json
jq '.mcp.grimorio' ~/.config/opencode/opencode.json && echo "✅ MCP configured"
jq '.agent | keys | map(select(startswith("grimorio")))' ~/.config/opencode/opencode.json && echo "✅ Agents configured"
jq '.command.grimorio' ~/.config/opencode/opencode.json && echo "✅ Command configured"
```

### AC3: Functional Test
```bash
~/.local/bin/grimorio --help
# Expected: Help output, exit code 0
```

### AC4: Re-run Safety
```bash
./install.sh && ./install.sh
# Expected: Both runs succeed, no errors
```
