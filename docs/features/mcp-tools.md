# MCP Tools Reference

Grimorio exposes 30+ MCP tools over the `mcp` stdio server. Each tool is
documented in its corresponding agent prompt under `agents/grimorio-*.md`
and in its handler file under `internal/mcp/handlers/`.

## Tool Categories

### Campaign Generation (`grimorio-architect`)
- `grimorio_create_campaign` — initialize a new campaign directory
- `grimorio_generate_adventure_bible` — canon + bible in one batch
- `grimorio_save_lore`, `grimorio_save_npcs`, `grimorio_save_bestiary`,
  `grimorio_save_encounters`, `grimorio_save_introduction`,
  `grimorio_save_setting_guide`, `grimorio_save_appendices`,
  `grimorio_save_chapter`, `grimorio_save_areas`, `grimorio_save_maps`,
  `grimorio_save_characters`, `grimorio_save_quests`
- `grimorio_create_personal_quest` — character-driven quest hooks
- `grimorio_generate_character` / `list_characters` / `get_character`

### Visual & Layout
- `grimorio_generate_image` — AI cover art / illustrations
- `grimorio_generate_map` / `grimorio_generate_divider` / `grimorio_generate_flowchart`
- `grimorio_generate_player_map` — secrets redacted
- `grimorio_generate_handouts` / `grimorio_export_handout`

### Validation & Consistency
- `grimorio_check_consistency` — full / lore_only / structure scope
- `grimorio_process_consistency_gate` — proposal-level gate
- `grimorio_validate_canon` — `domain.ValidationReport`
- `grimorio_visualize_relationship_graph` — D3 entity graph

### Live Session (`grimorio-dm`)
- `grimorio_dm_session_context` — aggregate campaign payload
- `grimorio_update_narrative_state` — append session log
- `grimorio_generate_session_prep` — pre-session checklist
- `grimorio_generate_tactics` / `grimorio_get_tactics`
- `grimorio_update_faction_reputation` / `grimorio_faction_reputation_dashboard`
- `grimorio_session_timeline` / `grimorio_list_quests` / `grimorio_update_quest_status`
- `grimorio_evaluate_consequences` — deferred-effect resolution

### TTS / Narration (optional)
- `grimorio_tts_speak`, `grimorio_tts_control`, `grimorio_set_dm_mode`,
  `grimorio_list_tts_voices`, `grimorio_assign_npc_voice`,
  `grimorio_get_tts_status`

## Where to Find Per-Tool Schemas

The tool schemas (input/output JSON shape) live alongside the handlers:

```
internal/mcp/handlers/
  campaign.go      →  save_lore / save_npcs / save_areas / …
  canon.go         →  check_consistency / validate_canon
  dm_context.go    →  dm_session_context
  faction.go       →  update_faction_reputation
  handout.go       →  generate_handouts / export_handout
  image.go         →  generate_image / generate_map / generate_divider
  narrative_state.go → update_narrative_state / session_timeline
  prologue.go      →  generate_prologue
  quest.go         →  save_quests / list_quests / update_quest_status
  random_table.go  →  generate_random_tables
  session_prep.go  →  generate_session_prep
  tactics.go       →  generate_tactics / get_tactics
  tts.go           →  tts_speak / tts_control / set_dm_mode
```

## How Tools Are Wired

`internal/mcp/server.go` registers every tool. Each handler returns
`mcp.NewToolResultText(string)` or `mcp.NewToolResultError(string)` and
must remain JSON-serialisable. New tools should be added:

1. Handler in `internal/mcp/handlers/<name>.go`.
2. Registration in `internal/mcp/server.go` (or the appropriate
   sub-server file for `grimorio-dm`).
3. Reference in the relevant agent prompt under `agents/`.

See also: [Developer Guide](developer-guide.md) for the handler test
pattern and [Campaign Consistency](campaign-consistency.md) for the
gate pipeline.
