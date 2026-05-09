# Specification: Install Command Sync

## Requirements

### R1: sync_command_from_agent Function
The install.sh script MUST include a function `sync_command_from_agent` that:
- Reads the Phase 1 questions from `agents/grimorio-architect.md`
- Extracts the workflow section (lines between "### Phase 1" and "### Phase 2")
- Generates a template JSON snippet for the command configuration
- Falls back to hardcoded template if the agent file is unavailable

### R2: Updated Hardcoded Template
The hardcoded template in `configure_opencode_command` MUST include:
- Question 1: Campaign name (kebab-case)
- Question 2: One-shot or full campaign
- **Question 3: Campaign idea / brief description** (NEW - 2-3 sentences)
- Question 4: Player level range
- Question 5: Desired tone
- Question 6: Duration

### R3: Integration Point
The `configure_opencode_command` function MUST call `sync_command_from_agent` BEFORE configuring the grimorio-architect agent.

### R4: Backward Compatibility
If `agents/grimorio-architect.md` does not exist or cannot be read, the script MUST fall back to the hardcoded template without failing.

## Acceptance Criteria

### AC1: Function Exists
```bash
grep -q "sync_command_from_agent" install.sh
```

### AC2: Question 3 Present
```bash
grep -q "Campaign idea / brief description" install.sh
```

### AC3: Function Called
```bash
grep -q "sync_command_from_agent.*OPENCODE_CONFIG" install.sh
```

### AC4: Template Extraction
The sync_command_from_agent function MUST extract text from grimorio-architect.md using sed/awk/grep.

### AC5: JSON Escaping
The extracted template MUST be properly escaped for JSON inclusion (using jq -Rs).

## Test Scenarios

### Scenario 1: Fresh Install
- Run: `./install.sh`
- Expected: Command configured with Question 3 included

### Scenario 2: Update Install
- Run: `./install.sh` on existing installation
- Expected: Command template updated, no duplicates

### Scenario 3: Missing Agent File
- Temporarily rename: `agents/grimorio-architect.md`
- Run: `./install.sh`
- Expected: Falls back to hardcoded template, no error

### Scenario 4: Agent File Updated
- Modify: `agents/grimorio-architect.md` Phase 1 section
- Run: `./install.sh`
- Expected: Command template reflects the changes
