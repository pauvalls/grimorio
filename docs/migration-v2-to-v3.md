# Migration Guide: v2.6.0 to v3.0.0

This guide covers the migration process from Grimorio v2.6.0 to v3.0.0.

## Breaking Changes

### Area Format

**v2.6.0**: Areas in acts without sequential numbering
```json
{
  "acts": [{
    "areas": [
      {"title": "The Courtyard", "description": "..."},
      {"title": "The Throne Room", "description": "..."}
    ]
  }]
}
```

**v3.0.0**: Unified WotC format with sequential numbering (1-15 per chapter)
```json
{
  "acts": [{
    "areas": [
      {"area_number": 1, "title": "The Courtyard", "description": "..."},
      {"area_number": 2, "title": "The Throne Room", "description": "..."}
    ]
  }]
}
```

### Quest Format

**v2.6.0**: Basic quest structure
```json
{
  "title": "Save the Village",
  "type": "main",
  "hook": "Villagers ask for help"
}
```

**v3.0.0**: Enhanced with approaches and failure states
```json
{
  "title": "Save the Village",
  "type": "main",
  "tier": "chapter",
  "hook": "Villagers ask for help",
  "approaches": [
    {"type": "combat", "title": "Fight the Raiders", "steps": [...]},
    {"type": "social", "title": "Negotiate Peace", "steps": [...]},
    {"type": "stealth", "title": "Sabotage the Camp", "steps": [...]}
  ],
  "failure_states": [
    {"type": "soft", "trigger": "Party retreats", "consequences": ["Village burned"]},
    {"type": "hard", "trigger": "Party defeated", "consequences": ["Campaign ends"]}
  ]
}
```

## Migration Steps

### Automated Migration (Recommended)

```bash
# Navigate to your campaign directory
cd campaigns/my-campaign

# Run migration script
go run ../../scripts/migrate_v2_to_v3.go .

# Verify migration
cat VERSION  # Should show "3.0.0"
```

The migration script will:
1. Create a backup of your campaign
2. Detect areas in old format
3. Add sequential numbering (1-15 per chapter)
4. Update version marker to 3.0.0

### Manual Migration

If you prefer manual migration:

1. **Backup your campaign**:
   ```bash
   cp -r my-campaign my-campaign.backup
   ```

2. **Add area numbers** to each area in `canon.json`:
   ```json
   {
     "area_number": 1,
     "title": "The Courtyard",
     ...
   }
   ```

3. **Update version marker**:
   ```bash
   echo "3.0.0" > VERSION
   ```

## Rollback Instructions

If you need to rollback to v2.6.0:

```bash
# If you used the migration script:
cp -r my-campaign.backup.* my-campaign-restored

# Or manually revert changes:
git checkout HEAD -- canon.json
rm VERSION
```

## New Features in v3.0.0

### Milestone XP Tracking
- Per-chapter XP tables following PHB thresholds
- Party level progression tracking
- Session-by-session milestone tracking

### Enhanced Magic Items
- Full WotC stat block format
- Rarity validation (common to artifact)
- Cursed items with removal methods
- Attunement requirements

### Combat Tactics
- Intelligence-based tactics (instinctive to strategic)
- Environmental tactics from area features
- Target priorities and retreat conditions
- Pack behavior for social monsters

### Player-Facing Maps
- Automatic secret feature redaction
- Multiple export formats (SVG, PDF, PNG)
- Blind and fog-of-war variants

### Session Zero Guide
- Campaign-specific content warnings
- Safety tools with activation instructions
- Character creation worksheets
- Timed agenda (90-180 minutes)

### Structured Quests
- 3+ distinct approaches (combat, social, stealth)
- Failure states (soft, hard, complication)
- Quest clues with discovery conditions
- Campaign integration (NPCs, factions, locations)

### Consequence Tables
- Act transition tracking
- Faction reputation propagation
- NPC status changes
- New opportunities and locked content

## FAQ

### Q: Will my existing campaigns still work?
**A**: Yes! v3.0.0 is backward compatible. Existing campaigns will work without modification, but won't have new features until migrated.

### Q: Do I need to migrate my campaigns?
**A**: No, migration is optional. However, migrating unlocks new features like milestone XP tables, structured quests, and player maps.

### Q: Can I undo the migration?
**A**: Yes, the migration script creates a backup. You can restore from backup at any time.

### Q: What if migration fails?
**A**: The migration script validates each step. If it fails, your original campaign remains unchanged. Check the error message and try again.

### Q: Are there any manual steps after migration?
**A**: After automated migration, you may want to:
- Review area numbers for logical flow
- Add failure states to existing quests
- Generate player maps for key areas
- Create Session Zero guide for your campaign

## Support

For migration issues or questions:
- Open an issue on GitHub
- Check existing issues for similar problems
- Provide your campaign structure (without sensitive content)
