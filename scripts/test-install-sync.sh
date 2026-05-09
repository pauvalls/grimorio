#!/bin/bash
# Test script for install.sh command sync feature

set -e

echo "=== Testing install.sh Command Sync ==="

# Test 1: Check sync_command_from_agent function exists
echo -n "Test 1: sync_command_from_agent function exists... "
if grep -q "sync_command_from_agent()" /home/pau/Grimorio/install.sh; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi

# Test 2: Check Question 3 in hardcoded template
echo -n "Test 2: Question 3 (brief description) in template... "
if grep -q "Campaign idea / brief description" /home/pau/Grimorio/install.sh; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi

# Test 3: Check sync function is called
echo -n "Test 3: sync_command_from_agent called in configure_opencode_command... "
if grep -q 'sync_command_from_agent "$OPENCODE_CONFIG"' /home/pau/Grimorio/install.sh; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi

# Test 4: Check function location (should be before configure_opencode_command)
echo -n "Test 4: Function defined before first call... "
SYNC_FUNC_LINE=$(grep -n "sync_command_from_agent()" /home/pau/Grimorio/install.sh | cut -d: -f1)
SYNC_CALL_LINE=$(grep -n 'sync_command_from_agent "$OPENCODE_CONFIG"' /home/pau/Grimorio/install.sh | cut -d: -f1)
if [ "$SYNC_FUNC_LINE" -lt "$SYNC_CALL_LINE" ]; then
    echo "✅ PASS (function at line $SYNC_FUNC_LINE, call at line $SYNC_CALL_LINE)"
else
    echo "❌ FAIL (function at line $SYNC_FUNC_LINE, call at line $SYNC_CALL_LINE)"
    exit 1
fi

# Test 5: Check grimorio-architect.md has Phase 1 section
echo -n "Test 5: grimorio-architect.md has Phase 1 section... "
if grep -q "### Phase 1: Gather Requirements" /home/pau/Grimorio/agents/grimorio-architect.md; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi

# Test 6: Check Phase 1 includes Question 3
echo -n "Test 6: Phase 1 includes Question 3 (brief description)... "
if grep -q "Campaign idea / brief description" /home/pau/Grimorio/agents/grimorio-architect.md; then
    echo "✅ PASS"
else
    echo "❌ FAIL"
    exit 1
fi

echo ""
echo "=== All Tests Passed ==="
echo ""
echo "Summary:"
echo "- sync_command_from_agent function: ✅"
echo "- Question 3 in hardcoded template: ✅"
echo "- Function integration: ✅"
echo "- Function ordering: ✅"
echo "- Agent file Phase 1: ✅"
echo "- Agent file Question 3: ✅"
echo ""
echo "Next: Run ./install.sh to apply changes to opencode.json"
