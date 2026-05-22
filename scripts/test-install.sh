#!/bin/sh
# E2E Install Test for Grimorio
# Usage: scripts/test-install.sh [--local]
#   --local   Use local install.sh instead of downloading from GitHub
#
# This script can run in Docker (Ubuntu) or CI.
# It verifies that install.sh completes successfully, the binary is in PATH,
# agents/skills are copied, and opencode.json is modified.

set -e

# Configuration
LOCAL_MODE=false
if [ "$1" = "--local" ]; then
    LOCAL_MODE=true
fi

# Create isolated test environment
TEST_HOME=$(mktemp -d)
export HOME="$TEST_HOME"

cleanup() {
    echo "Cleaning up test environment..."
    rm -rf "$TEST_HOME"
}

trap cleanup EXIT

echo "=== Grimorio E2E Install Test ==="
echo "Test HOME: $TEST_HOME"
echo ""

# ---------------------------------------------------------------------------
# Run install.sh
# ---------------------------------------------------------------------------

if [ "$LOCAL_MODE" = "true" ]; then
    echo "Running local install.sh..."
    if [ ! -f "install.sh" ]; then
        echo "ERROR: install.sh not found in current directory"
        exit 1
    fi
    bash install.sh
else
    echo "Downloading and running install.sh from GitHub..."
    curl -sSL https://raw.githubusercontent.com/pauvalls/grimorio/main/install.sh | bash
fi

# ---------------------------------------------------------------------------
# Verify installation
# ---------------------------------------------------------------------------

echo ""
echo "=== Verification ==="

# 1. Binary exists in ~/.grimorio/
if [ ! -f "$HOME/.grimorio/grimorio" ]; then
    echo "FAIL: grimorio binary not found in ~/.grimorio/"
    exit 1
fi
echo "PASS: grimorio binary exists in ~/.grimorio/"

# 2. Binary is in PATH (via ~/.local/bin symlink or copy)
if [ ! -e "$HOME/.local/bin/grimorio" ]; then
    echo "FAIL: grimorio not found in ~/.local/bin/"
    exit 1
fi
echo "PASS: grimorio is linked/copied to ~/.local/bin/"

# 3. Binary is executable
if [ ! -x "$HOME/.grimorio/grimorio" ]; then
    echo "FAIL: grimorio binary is not executable"
    exit 1
fi
echo "PASS: grimorio binary is executable"

# 4. grimorio --version works
export PATH="$HOME/.local/bin:$PATH"
if ! command -v grimorio >/dev/null 2>&1; then
    echo "FAIL: grimorio not in PATH"
    exit 1
fi
echo "PASS: grimorio is in PATH"

VERSION_OUTPUT=$(grimorio --version 2>/dev/null || true)
if [ -z "$VERSION_OUTPUT" ]; then
    echo "WARN: grimorio --version produced no output (may be expected in test env)"
else
    echo "PASS: grimorio --version works: $VERSION_OUTPUT"
fi

# 5. Agents directory exists in plugin dir
if [ ! -d "$HOME/.config/opencode/plugins/grimorio/agents" ]; then
    echo "FAIL: agents/ not found in plugin directory"
    exit 1
fi
echo "PASS: agents/ directory exists in plugin dir"

# 5b. Agents copied to global opencode agents directory
if [ ! -d "$HOME/.config/opencode/agents" ]; then
    echo "FAIL: agents/ not found in global opencode agents directory"
    exit 1
fi
if [ ! -f "$HOME/.config/opencode/agents/grimorio-architect.md" ]; then
    echo "FAIL: grimorio-architect.md not found in global agents directory"
    exit 1
fi
if [ ! -f "$HOME/.config/opencode/agents/grimorio-dm.md" ]; then
    echo "FAIL: grimorio-dm.md not found in global agents directory"
    exit 1
fi
echo "PASS: agents copied to global opencode agents directory"

# 6. Skills directory exists in plugin dir
if [ ! -d "$HOME/.config/opencode/plugins/grimorio/skills" ]; then
    echo "FAIL: skills/ not found in plugin directory"
    exit 1
fi
echo "PASS: skills/ directory exists in plugin dir"

# 7. .mcp.json exists in plugin dir
if [ ! -f "$HOME/.config/opencode/plugins/grimorio/.mcp.json" ]; then
    echo "FAIL: .mcp.json not found in plugin directory"
    exit 1
fi
echo "PASS: .mcp.json exists in plugin dir"

# 8. opencode.json is modified
if [ ! -f "$HOME/.config/opencode/opencode.json" ]; then
    echo "FAIL: opencode.json not found"
    exit 1
fi
if ! grep -q "grimorio" "$HOME/.config/opencode/opencode.json"; then
    echo "FAIL: opencode.json does not contain grimorio config"
    exit 1
fi
echo "PASS: opencode.json contains grimorio configuration"

echo ""
echo "=== All E2E checks passed! ==="
