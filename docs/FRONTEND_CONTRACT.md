# Grimorio Game Engine — Frontend Contract

## Overview

This document defines the contract between the **Grimorio Game Engine Backend** and the **Frontend Application**. The backend exposes a WebSocket API and REST API for real-time game session management.

## Base URL

```
http://localhost:8080
```

## API Endpoints

### REST API

#### Health Check
```
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "mode": "game_engine"
}
```

#### List Campaigns
```
GET /api/campaigns
```

**Response:**
```json
[
  {
    "name": "ciudad-sumergida",
    "title": "Ciudad Sumergida",
    "setting": "An underwater city...",
    "acts_count": 3,
    "npcs_count": 5,
    "characters_count": 2
  }
]
```

#### Get Campaign
```
GET /api/campaigns/:name
```

**Response:** Campaign object with full details.

#### Create Session
```
POST /api/sessions
Content-Type: application/json

{
  "campaign_id": "ciudad-sumergida",
  "players": ["Eldric", "Lira"]
}
```

**Response:**
```json
{
  "id": "session-1234567890",
  "campaign_id": "ciudad-sumergida",
  "players": [
    {
      "character_id": "Eldric",
      "current_hp": 11,
      "max_hp": 11,
      "temp_hp": 0,
      "ac": 12,
      "position": {"x": 0, "y": 0},
      "initiative": 0,
      "is_active": false,
      "conditions": []
    }
  ],
  "in_combat": false,
  "started_at": "2026-05-06T10:00:00Z"
}
```

#### Get Session State
```
GET /api/sessions/:id/state
```

**Response:**
```json
{
  "session_id": "session-1234567890",
  "campaign_id": "ciudad-sumergida",
  "in_combat": false,
  "current_scene": null,
  "players": [...],
  "combat": null,
  "active_actor": "",
  "round": 0
}
```

#### Get Session Events
```
GET /api/sessions/:id/events?limit=50
```

**Response:** Array of session events (most recent first).

#### End Session
```
POST /api/sessions/:id/end
```

**Response:** Session summary object.

---

## WebSocket API

### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
```

### Message Protocol

All messages are JSON with the following structure:

```typescript
interface WSMessage {
  type: string;
  payload: Record<string, any>;
  timestamp: string; // ISO 8601
}
```

### Client → Server Messages

#### Join Session
```json
{
  "type": "join_session",
  "payload": {
    "session_id": "session-1234567890",
    "character_id": "Eldric"
  }
}
```

#### Player Action
```json
{
  "type": "player_action",
  "payload": {
    "action": "I attack the goblin with my sword"
  }
}
```

#### Roll Dice
```json
{
  "type": "roll_dice",
  "payload": {
    "dice": "2d6+3"
  }
}
```

Default is `"d20"` if not specified.

#### Move Token
```json
{
  "type": "move_token",
  "payload": {
    "x": 5,
    "y": 10
  }
}
```

#### Start Combat
```json
{
  "type": "start_combat",
  "payload": {
    "enemies": [
      {
        "character_id": "goblin-1",
        "current_hp": 7,
        "max_hp": 7,
        "ac": 12
      }
    ]
  }
}
```

#### End Combat
```json
{
  "type": "end_combat",
  "payload": {}
}
```

#### Next Turn
```json
{
  "type": "next_turn",
  "payload": {}
}
```

#### Attack
```json
{
  "type": "attack",
  "payload": {
    "target_id": "goblin-1",
    "attack_roll": 15
  }
}
```

#### Get State
```json
{
  "type": "get_state",
  "payload": {}
}
```

### Server → Client Messages

#### Connected
```json
{
  "type": "connected",
  "payload": {
    "session_id": "session-1234567890"
  },
  "timestamp": "2026-05-06T10:00:00Z"
}
```

#### Narration
```json
{
  "type": "narration",
  "payload": {
    "text": "Eldric swings his sword at the goblin...",
    "actor": "Eldric",
    "success": true
  },
  "timestamp": "2026-05-06T10:00:01Z"
}
```

#### Dice Result
```json
{
  "type": "dice_result",
  "payload": {
    "dice": "2d6+3",
    "results": [4, 5],
    "total": 12,
    "actor": "Eldric"
  },
  "timestamp": "2026-05-06T10:00:02Z"
}
```

#### State Update
```json
{
  "type": "state_update",
  "payload": {
    "session_id": "session-1234567890",
    "campaign_id": "ciudad-sumergida",
    "in_combat": true,
    "current_scene": null,
    "players": [
      {
        "character_id": "Eldric",
        "current_hp": 11,
        "max_hp": 11,
        "temp_hp": 0,
        "ac": 12,
        "position": {"x": 0, "y": 0},
        "initiative": 0,
        "is_active": true,
        "conditions": []
      }
    ],
    "combat": {
      "session_id": "session-1234567890",
      "round": 1,
      "initiative_order": ["Eldric", "goblin-1"],
      "active_index": 0,
      "map_id": ""
    },
    "active_actor": "Eldric",
    "round": 1
  },
  "timestamp": "2026-05-06T10:00:03Z"
}
```

#### Event
```json
{
  "type": "event",
  "payload": {
    "type": "attack",
    "actor": "Eldric",
    "target": "goblin-1",
    "hit": true,
    "damage": 6,
    "critical": false
  },
  "timestamp": "2026-05-06T10:00:04Z"
}
```

#### Combat Turn
```json
{
  "type": "combat_turn",
  "payload": {
    "active_character": "goblin-1",
    "round": 1
  },
  "timestamp": "2026-05-06T10:00:05Z"
}
```

#### Error
```json
{
  "type": "error",
  "payload": {
    "message": "Not your turn"
  },
  "timestamp": "2026-05-06T10:00:06Z"
}
```

#### Disconnected
```json
{
  "type": "disconnected",
  "payload": {},
  "timestamp": "2026-05-06T10:00:07Z"
}
```

---

## Data Models

### PlayerState
```typescript
interface PlayerState {
  character_id: string;
  current_hp: number;
  max_hp: number;
  temp_hp: number;
  ac: number;
  position: { x: number; y: number };
  initiative: number;
  is_active: boolean;
  conditions: Condition[];
}
```

### Condition
```typescript
interface Condition {
  type: string; // "poisoned", "stunned", "frightened", etc.
  duration: number; // rounds, 0 = permanent
  applied_at: string; // ISO 8601
  source: string;
}
```

### CombatState
```typescript
interface CombatState {
  session_id: string;
  round: number;
  initiative_order: string[];
  active_index: number;
  map_id: string;
}
```

### SessionEvent
```typescript
interface SessionEvent {
  id: number;
  session_id: string;
  type: string; // "action", "dice_roll", "narration", "combat_start", etc.
  actor: string;
  content: string;
  result: Record<string, any>;
  timestamp: string;
}
```

---

## Game Flow

### 1. Lobby (Before Session)
1. Frontend lists campaigns via `GET /api/campaigns`
2. User selects campaign and characters
3. Frontend creates session via `POST /api/sessions`

### 2. Session Start
1. Frontend connects WebSocket: `ws://localhost:8080/ws`
2. Frontend sends `join_session` message
3. Server responds with `connected` + `state_update`

### 3. Gameplay Loop
1. Player performs action → Frontend sends `player_action`
2. Server processes → Broadcasts `narration` + `state_update`
3. Player rolls dice → Frontend sends `roll_dice`
4. Server processes → Broadcasts `dice_result` + `event`

### 4. Combat
1. DM sends `start_combat`
2. Server calculates initiative → Broadcasts `state_update` with combat state
3. Active player sends `attack` or other actions
4. Server resolves → Broadcasts `event` + `state_update`
5. DM sends `next_turn` → Server advances turn → Broadcasts `combat_turn`
6. DM sends `end_combat` when combat ends

### 5. Session End
1. Frontend calls `POST /api/sessions/:id/end`
2. Server returns session summary

---

## Error Handling

The server sends `error` messages for:
- Invalid message types
- Missing required fields
- Game rule violations (attacking out of turn, etc.)
- Session not found
- Character not found

Frontend should display these errors to the user.

---

## CORS

The server allows all origins (`*`) for development. In production, configure allowed origins.

---

## Authentication

Currently no authentication is implemented. All endpoints are open.

---

## Environment Variables

The backend supports these environment variables for LLM configuration:

| Variable | Description | Example |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | `sk-...` |
| `ANTHROPIC_API_KEY` | Anthropic API key | `sk-ant-...` |
| `KIMI_API_KEY` | Kimi (Moonshot AI) API key | `sk-...` |
| `OPENCODE_API_KEY` | OpenCode API key | `sk-...` |
| `OPENAI_MODEL` | OpenAI model | `gpt-4` |
| `ANTHROPIC_MODEL` | Anthropic model | `claude-3-sonnet-20240229` |
| `KIMI_MODEL` | Kimi model | `moonshot-v1-8k` |
| `OPENCODE_MODEL` | OpenCode model | `opencode-default` |

Priority: `KIMI_API_KEY` > `OPENCODE_API_KEY` > `OPENAI_API_KEY` > `ANTHROPIC_API_KEY`

---

## Notes for Frontend Team

1. **WebSocket Reconnection**: Implement automatic reconnection with exponential backoff
2. **State Sync**: On reconnect, send `get_state` to sync current game state
3. **Audio**: Audio pipeline (STT/TTS) will be added in Phase 2. For now, use browser's Web Speech API as a placeholder
4. **Maps**: SVG maps are generated procedurally. Use `GET /api/campaigns/:name` to get map URLs
5. **Character Sheets**: Use `GET /api/characters/:campaign/:name` (to be implemented)
6. **Responsive Design**: The game engine is designed for tablet/desktop. Mobile support is secondary

---

## Development Setup

```bash
# Start the game engine
cd /path/to/grimorio
grimorio serve

# Or with custom port
grimorio serve --port 8080
```

---

## Version

Game Engine API Version: 1.0.0
Last Updated: 2026-05-06
