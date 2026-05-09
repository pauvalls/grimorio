# Design: Clean Installer v2

## Architecture

```
install.sh (single file, self-contained)
│
├── Helper Functions
│   ├── log(), warn(), error(), success()
│   ├── command_exists()
│   └── detect_platform()
│
├── Installation Functions (called in order)
│   ├── clean_installation()        # Remove everything
│   ├── detect_platform()           # OS detection
│   ├── install_go()                # Install Go if needed
│   ├── install_wkhtmltopdf()       # Install PDF tool if needed
│   ├── setup_repo()                # Clone repository
│   ├── build_binary()              # Compile Go binaries
│   ├── migrate_existing_campaigns() # Migrate v1 campaigns
│   ├── setup_plugin()              # Copy all plugin files
│   ├── configure_opencode_mcp()    # Configure MCP in opencode.json
│   ├── configure_opencode_command() # Configure command template
│   ├── configure_opencode_agents() # Configure all agents
│   └── configure_shell()           # Add PATH to shell
│
└── main()
    ├── Print banner
    ├── Call all functions in order
    └── Print completion message
```

## Key Design Decisions

### 1. Single File, Self-Contained
- **Why**: Easy to curl | bash, no dependencies
- **Tradeoff**: Larger file (~900 lines)
- **Mitigation**: Clear section comments

### 2. Clean First, Then Install
- **Why**: Prevents stale files, version conflicts
- **Risk**: User loses custom configs
- **Mitigation**: Backup campaigns, document reset behavior

### 3. Clone Fresh Every Time
- **Why**: Always gets latest code
- **Risk**: Slower than git pull
- **Mitigation**: Acceptable for one-time install

### 4. All-in-One Script
- **Why**: Single point of truth
- **Risk**: Harder to maintain
- **Mitigation**: Clear function boundaries, comments

## Function Signatures

```bash
clean_installation() -> void
# Removes ALL previous installation

detect_platform() -> void
# Sets OS, ARCH, WKHTMLTOPDF_URL globals

install_go() -> void
# Installs Go 1.23+ if not present

install_wkhtmltopdf() -> void
# Installs wkhtmltopdf if not present

setup_repo() -> void
# Clones repo to INSTALL_DIR

build_binary() -> void
# Compiles Go binaries

migrate_existing_campaigns() -> void
# Migrates v1 campaigns to v2

setup_plugin() -> void
# Copies all files to plugin directories

configure_opencode_mcp() -> void
# Configures .mcp.grimorio in opencode.json

configure_opencode_command() -> void
# Configures .command.grimorio with template

configure_opencode_agents() -> void
# Configures ALL .agent.grimorio-* entries

configure_shell() -> void
# Adds PATH to shell config

main() -> void
# Entry point, calls all functions
```

## File Structure

```
install.sh (884 lines)
├── Shebang & setup (lines 1-20)
├── Helper functions (lines 21-60)
├── clean_installation (lines 61-130)
├── detect_platform (lines 131-160)
├── install_go (lines 161-200)
├── install_wkhtmltopdf (lines 201-240)
├── setup_repo (lines 241-270)
├── build_binary (lines 271-300)
├── migrate_existing_campaigns (lines 301-340)
├── setup_plugin (lines 341-420)
├── configure_opencode_mcp (lines 421-450)
├── configure_opencode_command (lines 451-550)
├── configure_opencode_agents (lines 551-750)
├── configure_shell (lines 751-790)
├── print_instructions (lines 791-850)
└── main (lines 851-884)
```

## Testing Strategy

1. **Syntax Check**: `bash -n install.sh`
2. **Dry Run**: `./install.sh` on clean VM
3. **Re-run Test**: Run twice, verify idempotency
4. **Component Check**: Verify all files installed
5. **Functional Test**: Run `grimorio --help`
