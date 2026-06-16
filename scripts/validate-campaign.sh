#!/bin/bash
#===============================================================================
# validate-campaign.sh — DEPRECATED SHIM → use `grimorio validate`
# Version: 2.5.0 (deprecated)
#
# This script is a thin wrapper around `grimorio validate <name>`. It exists
# for backward compatibility with previous Grimorio releases. It will be
# REMOVED in v5.0.0.
#
# Migration:
#   scripts/validate-campaign.sh /path/to/campaign --check=wotc
#                                 ↓
#   grimorio validate --scope=wotc <campaign-name>
#
# Flags:
#   --check=structure|wotc|references|all  →  --scope=structure|wotc|references|all
#   campaign-path (positional)             →  campaign name (basename of path)
#===============================================================================

set -e

# Emit a deprecation notice to stderr exactly once per invocation.
echo "[DEPRECATION] scripts/validate-campaign.sh is deprecated and will be REMOVED in v5.0.0." >&2
echo "[DEPRECATION] Use: grimorio validate --scope=<scope> <campaign-name>" >&2

# Translate arguments: --check=... → --scope=..., and extract basename from path.
ARGS=()
for arg in "$@"; do
    case "$arg" in
        --check=*)
            ARGS+=("--scope=${arg#--check=}")
            ;;
        --help|-h)
            cat <<'USAGE'
Usage: grimorio validate [options] <campaign-name>

Options:
  --scope=structure|wotc|references|all   Validation scope (default: all)
  --json                                  Emit machine-readable JSON
  -h, --help                              Show this help
USAGE
            exit 0
            ;;
        *)
            # If the argument looks like a directory path, convert to basename
            # so the new CLI can resolve it under CAMPAIGN_ROOT/~/campaigns.
            if [ -d "$arg" ]; then
                ARGS+=("$(basename "$arg")")
            else
                ARGS+=("$arg")
            fi
            ;;
    esac
done

# Forward to the real CLI.
exec grimorio validate "${ARGS[@]}"
