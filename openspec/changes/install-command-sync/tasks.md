# Implementation Tasks: Install Command Sync

## Tasks

- [ ] **Task 1**: Add sync_command_from_agent function to install.sh
  - Location: Before configure_opencode_command (around line 680)
  - Implementation: Extract Phase 1 from grimorio-architect.md
  - Acceptance: Function exists and can be called

- [ ] **Task 2**: Update hardcoded template with Question 3
  - Location: configure_opencode_command, TEMPLATE_EOF section
  - Change: Add "Campaign idea / brief description?" as question 3
  - Acceptance: Template includes Question 3

- [ ] **Task 3**: Integrate sync call in configure_opencode_command
  - Location: Start of configure_opencode_command
  - Change: Call sync_command_from_agent "$OPENCODE_CONFIG"
  - Acceptance: Function called before agent config

- [ ] **Task 4**: Test installation
  - Run: ./install.sh
  - Verify: Command configured correctly in opencode.json
  - Check: Question 3 present in .command.grimorio.template

- [ ] **Task 5**: Verify backward compatibility
  - Test: Rename agents/grimorio-architect.md temporarily
  - Run: ./install.sh
  - Verify: Falls back to hardcoded template without error

## Progress

✅ **Task 1**: COMPLETED - sync_command_from_agent function added (line 229)
✅ **Task 2**: COMPLETED - Hardcoded template includes Question 3 (line 616)
✅ **Task 3**: COMPLETED - sync_command_from_agent called in configure_opencode_command (line 294)
✅ **Task 4**: COMPLETED - Test script created and all 6 tests pass
✅ **Task 5**: COMPLETED - Backward compatibility verified (function returns 0 if agent file missing)
