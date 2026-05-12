#!/usr/bin/env python3
"""
Patch opencode.json to use Grimorio skills.
Safe incremental updates with validation after each step.
"""

import json
import subprocess
import sys
from pathlib import Path

OPENCODE_CONFIG = Path.home() / ".config/opencode/opencode.json"
BACKUP_DIR = Path.home() / ".config/opencode"

def test_opencode():
    """Test if opencode config is valid."""
    result = subprocess.run(['opencode', '--version'], capture_output=True, text=True)
    return result.returncode == 0

def patch_config():
    """Apply patches incrementally with validation."""
    
    if not OPENCODE_CONFIG.exists():
        print(f"❌ Config not found: {OPENCODE_CONFIG}")
        return False
    
    with open(OPENCODE_CONFIG, 'r') as f:
        config = json.load(f)
    
    changes = []
    
    # 1. Update grimorio-* agent prompts
    print("1. Updating grimorio-* agent prompts...")
    for key in list(config.get('agent', {}).keys()):
        if key.startswith('grimorio-'):
            skill_name = key.replace('grimorio-', '')
            config['agent'][key]['prompt'] = f"Read ~/.config/opencode/skills/grimorio-{skill_name}/SKILL.md"
            changes.append(f"  - {key}")
    
    with open(OPENCODE_CONFIG, 'w') as f:
        json.dump(config, f, indent=2)
    
    if not test_opencode():
        print("❌ Failed after updating agent prompts")
        return False
    print("✅ Agent prompts updated")
    
    # 2. Add SDD config
    print("2. Adding SDD configuration...")
    config['sdd'] = {
        "delivery_strategy": "exception-ok",
        "chain_strategy": "stacked-to-main",
        "artifact_store": "engram"
    }
    
    with open(OPENCODE_CONFIG, 'w') as f:
        json.dump(config, f, indent=2)
    
    if not test_opencode():
        print("❌ Failed after adding SDD config")
        return False
    print("✅ SDD config added")
    
    # 3. Add agent field to grimorio command
    print("3. Adding agent field to /grimorio command...")
    if 'command' in config and 'grimorio' in config['command']:
        config['command']['grimorio']['agent'] = 'grimorio-architect'
    
    with open(OPENCODE_CONFIG, 'w') as f:
        json.dump(config, f, indent=2)
    
    if not test_opencode():
        print("❌ Failed after adding agent field")
        return False
    print("✅ Agent field added")
    
    # 4. Update grimorio command template
    print("4. Updating /grimorio command template...")
    if 'command' in config and 'grimorio' in config['command']:
        config['command']['grimorio']['template'] = (
            "Load grimorio-architect skill. Ask questions one at a time: "
            "campaign name, language (es/en), one-shot/campaign, brief description, "
            "level range, tone, duration. Create campaign via MCP. "
            "Orchestrate all phases via delegate. Report progress. "
            "Validate WotC before PDF. Compile PDF."
        )
    
    with open(OPENCODE_CONFIG, 'w') as f:
        json.dump(config, f, indent=2)
    
    if not test_opencode():
        print("❌ Failed after updating command template")
        return False
    print("✅ Command template updated")
    
    # 5. Add template reference to grimorio-areas
    print("5. Adding template reference to grimorio-areas...")
    if 'grimorio-areas' in config.get('agent', {}):
        config['agent']['grimorio-areas']['prompt'] += " Read template areas.md.tmpl"
    
    with open(OPENCODE_CONFIG, 'w') as f:
        json.dump(config, f, indent=2)
    
    if not test_opencode():
        print("❌ Failed after adding template reference")
        return False
    print("✅ Template reference added")
    
    # 6. Add WotC validation reference
    print("6. Adding WotC validation reference...")
    if 'command' in config and 'grimorio' in config['command']:
        config['command']['grimorio']['template'] += " Run validate-campaign.sh --check=all"
    
    with open(OPENCODE_CONFIG, 'w') as f:
        json.dump(config, f, indent=2)
    
    if not test_opencode():
        print("❌ Failed after adding WotC validation")
        return False
    print("✅ WotC validation added")
    
    print("\n✅ All patches applied successfully!")
    print(f"\nChanged {len(changes)} agent prompts:")
    for change in changes:
        print(change)
    
    return True

if __name__ == '__main__':
    success = patch_config()
    sys.exit(0 if success else 1)
