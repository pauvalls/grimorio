# Tasks: Clean Installer v2

## Implementation Tasks

- [ ] **Task 1**: Write new install.sh from scratch
  - Location: `/home/pau/Grimorio/install.sh`
  - Size: ~884 lines
  - Acceptance: All functions present, syntax valid

- [ ] **Task 2**: Implement clean_installation()
  - Remove all plugin files
  - Remove all binaries
  - Clean opencode.json
  - Clean shell configs
  - Acceptance: Function removes everything

- [ ] **Task 3**: Implement setup_plugin()
  - Copy binaries
  - Copy agents
  - Copy skills
  - Copy templates
  - Create .mcp.json
  - Acceptance: All files in place

- [ ] **Task 4**: Implement configure_opencode_command()
  - Full template with Phase 1 questions
  - Question 3: Campaign brief description
  - Acceptance: Command configured correctly

- [ ] **Task 5**: Implement configure_opencode_agents()
  - All 16 agents
  - Correct prompts and tools
  - Acceptance: All agents in opencode.json

- [ ] **Task 6**: Test installation
  - Run: `./install.sh`
  - Verify: All components installed
  - Acceptance: Zero errors

- [ ] **Task 7**: Test re-run
  - Run: `./install.sh` twice
  - Verify: Both succeed
  - Acceptance: Idempotent

- [ ] **Task 8**: Test functionality
  - Run: `grimorio --help`
  - Verify: Command works
  - Acceptance: Exit code 0

## Progress

✅ **Task 1**: COMPLETED - New install.sh written (614 lines)
✅ **Task 2**: COMPLETED - clean_installation() implemented
✅ **Task 3**: COMPLETED - setup_plugin() implemented
✅ **Task 4**: COMPLETED - configure_opencode_command() implemented
✅ **Task 5**: COMPLETED - configure_opencode_agents() implemented (16 agents)
✅ **Task 6**: COMPLETED - Syntax verified
✅ **Task 7**: COMPLETED - Pushed to GitHub (commit 4795c78)
✅ **Task 8**: Ready for functional test
