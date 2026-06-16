#!/usr/bin/env bash
# scripts/check-i18n.sh — i18n-english-default guard
#
# Verifies that no Spanish prose tokens appear in skill and template
# files, and that locale metadata uses English (`lang="en"`,
# `dc:language>en`). Used by CI and as a quick local check.
#
# Exit code 0 = all checks pass
# Exit code 1 = at least one violation found
#
# This script is the bash counterpart to the Go tests in
# internal/compiler/templates_i18n_test.go. Both guard the same
# invariants; the Go test runs in `go test` and the bash script runs
# in CI's lint stage.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

VIOLATIONS=0
BANNED_SPANISH=(
  "Caja de texto"
  "Apariencia"
  "Personalidad"
  "Motivación"
  "Secreto"
  "Involucramiento en Quests"
  "Conexiones"
  "Trasfondo"
  "Alineamiento"
  "Ataque"
  "Daño"
  "Tesoro"
  "Improvisar"
)

# 1. Internal Go source must not contain `lang="es"` or `<dc:language>es`
#    anywhere in non-test files.
echo "→ Checking internal/ for Spanish locale metadata..."
LANGS_ES=$(grep -rE 'lang="es"' "${ROOT}/internal" --include="*.go" 2>/dev/null \
    | grep -v '_test.go' || true)
if [[ -n "${LANGS_ES}" ]]; then
  echo "  ❌ Found lang=\"es\" in non-test source code:"
  echo "${LANGS_ES}" | sed 's/^/    /'
  VIOLATIONS=$((VIOLATIONS + 1))
else
  echo "  ✅ No lang=\"es\" in non-test source"
fi

DC_LANG_ES=$(grep -rE '<dc:language>es</dc:language>' "${ROOT}/internal" \
    --include="*.go" 2>/dev/null | grep -v '_test.go' || true)
if [[ -n "${DC_LANG_ES}" ]]; then
  echo "  ❌ Found <dc:language>es in non-test source code:"
  echo "${DC_LANG_ES}" | sed 's/^/    /'
  VIOLATIONS=$((VIOLATIONS + 1))
else
  echo "  ✅ No <dc:language>es in non-test source"
fi

# 2. Skill files must not contain Spanish prose tokens.
echo "→ Checking skills/ for Spanish prose..."
SKILL_FILES_FOUND=0
for token in "${BANNED_SPANISH[@]}"; do
  HITS=$(grep -rlF "${token}" "${ROOT}/skills/grimorio-"*/SKILL.md 2>/dev/null || true)
  if [[ -n "${HITS}" ]]; then
    echo "  ❌ Found Spanish prose token '${token}' in skills/"
    echo "${HITS}" | sed 's/^/    /'
    SKILL_FILES_FOUND=$((SKILL_FILES_FOUND + 1))
  fi
done
if [[ "${SKILL_FILES_FOUND}" -eq 0 ]]; then
  echo "  ✅ No Spanish prose tokens in skills/"
else
  VIOLATIONS=$((VIOLATIONS + SKILL_FILES_FOUND))
fi

# 3. Template files must not contain Spanish prose tokens. Fenced code
#    blocks are exempt (structural markup).
echo "→ Checking internal/compiler/templates/ for Spanish prose..."
TEMPLATE_FILES_FOUND=0
for token in "${BANNED_SPANISH[@]}"; do
  HITS=$(grep -rlF "${token}" "${ROOT}/internal/compiler/templates/"*.md.tmpl 2>/dev/null || true)
  if [[ -n "${HITS}" ]]; then
    for tmpl in ${HITS}; do
      # Strip code blocks and check
      stripped=$(awk '
        /^```/ { in_block = !in_block; next }
        !in_block { print }
      ' "${tmpl}")
      if echo "${stripped}" | grep -qF "${token}"; then
        echo "  ❌ Found Spanish prose token '${token}' in template ${tmpl}"
        TEMPLATE_FILES_FOUND=$((TEMPLATE_FILES_FOUND + 1))
      fi
    done
  fi
done
if [[ "${TEMPLATE_FILES_FOUND}" -eq 0 ]]; then
  echo "  ✅ No Spanish prose tokens in templates/"
else
  VIOLATIONS=$((VIOLATIONS + TEMPLATE_FILES_FOUND))
fi

echo ""
if [[ "${VIOLATIONS}" -gt 0 ]]; then
  echo "❌ i18n check failed: ${VIOLATIONS} violation(s)"
  exit 1
fi
echo "✅ i18n check passed"
