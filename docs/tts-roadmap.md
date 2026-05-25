# TTS Roadmap — Piper Integration for Grimorio DM

## Overview

Grimorio now supports **local Text-to-Speech (TTS)** narration for DM responses using the **Piper** neural TTS engine. This replaces the previous remote WebSocket-based TTS architecture with a fully local, offline-capable pipeline.

## Architecture

```
┌──────────┐   texto crudo   ┌──────────┐   texto limpio   ┌──────────┐
│  LLM     │ ───────────────▶│  Filter  │ ────────────────▶│ Chunker  │
│  Output  │                 │ (tablas+ │                  │ (≤150ch) │
└──────────┘                 │ thinking)│                  └────┬─────┘
                               └──────────┘                       │
                                                                  ▼
                    ┌───────────────────────────────────────────┘
                    │
                    ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    PIPELINE TTS (2 goroutines)                           │
│  ┌─────────────┐     WAV      ┌─────────────┐     play cmd    ┌────────┐ │
│  │ Synthesizer │ ────────────▶│  Audio Queue│ ──────────────▶ │ Player │ │
│  │   Worker    │  (precarga)  │   (FIFO)    │   (aplay/...)  │ Worker │ │
│  └──────┬──────┘              └─────────────┘                └────────┘ │
│         │                                                                 │
│         │  Mientras C1 se reproduce, C2 se envía a Piper HTTP            │
│         │  POST localhost:5000/synthesize → WAV bytes → encolar          │
└──────────────────────────────────────────────────────────────────────────┘
```

### Components

| Component | File | Purpose |
|-----------|------|---------|
| **TextFilter** | `internal/tts/piper/filter.go` | Removes markdown tables and `<thinking>` blocks |
| **Chunker** | `internal/dm/chunker.go` | Splits text into ≤150 char sentence-aligned chunks |
| **PiperClient** | `internal/tts/piper/client.go` | HTTP client to localhost Piper server |
| **LifecycleManager** | `internal/tts/piper/lifecycle.go` | Starts/stops/monitors Piper process |
| **AudioPlayer** | `internal/tts/piper/player.go` | FIFO queue + system player (aplay/paplay/ffplay) |
| **NarratorPipeline** | `internal/tts/piper/narrator.go` | Orchestrates filter→chunk→synthesize→play with preload |
| **TTSService** | `internal/services/tts_service.go` | High-level service: mode switching, voice registry, control |
| **TTSHandlers** | `internal/mcp/handlers/tts.go` | MCP tools: `set_dm_mode`, `assign_npc_voice`, `tts_control`, `list_tts_voices`, `get_tts_status` |

## Installation

### Prerequisites

- **Linux** (primary platform)
- **Go 1.21+**
- **Piper** binary in `PATH`
- One of: `aplay` (ALSA), `paplay` (PulseAudio), or `ffplay` (ffmpeg)

### Install Piper

```bash
# Download prebuilt binary from https://github.com/rhasspy/piper/releases
# Example for x86_64 Linux:
wget https://github.com/rhasspy/piper/releases/download/2023.11.14-2/piper_linux_x86_64.tar.gz
tar -xzf piper_linux_x86_64.tar.gz -C ~/.local/bin

# Download a Spanish voice model (e.g., es_ES-davefx-medium)
wget https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx
wget https://huggingface.co/rhasspy/piper-voices/resolve/v1.0.0/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx.json
```

### Configure Grimorio

Create or edit `~/.config/grimorio/config.json`:

```json
{
  "output_dir": "~/campaigns",
  "tts": {
    "enabled": true,
    "piper": {
      "model_path": "/home/user/.local/share/piper/es_ES-davefx-medium.onnx",
      "config_path": "/home/user/.local/share/piper/es_ES-davefx-medium.onnx.json",
      "port": 5000,
      "host": "127.0.0.1",
      "length_scale": 1.0,
      "volume": 0.8
    },
    "chunker": {
      "max_chunk_size": 150
    },
    "audio": {
      "player": "auto",
      "preload_buffer": 1
    }
  }
}
```

Or use environment variables:

```bash
export GRIMORIO_TTS_ENABLED=true
export PIPER_MODEL_PATH=/home/user/.local/share/piper/es_ES-davefx-medium.onnx
export PIPER_CONFIG_PATH=/home/user/.local/share/piper/es_ES-davefx-medium.onnx.json
export AUDIO_PLAYER=auto
```

## MCP Tools

### `set_dm_mode`

Switch between `written` (text-only) and `tts` (spoken narration) modes.

```json
{
  "mode": "tts"
}
```

### `assign_npc_voice`

Assign a voice description to an NPC for consistent dialogue rendering.

```json
{
  "campaign_id": "mi-campaña",
  "npc_name": "Goblin Chief",
  "voice_prompt": "gruff and low, speaks in short sentences"
}
```

### `tts_control`

Control playback: `skip`, `stop`, `pause`, `resume`.

```json
{
  "action": "pause"
}
```

### `list_tts_voices`

List all assigned NPC voices for a campaign.

```json
{
  "campaign_id": "mi-campaña"
}
```

### `get_tts_status`

Get current TTS status (enabled, mode, available, playing).

## Configuration Reference

| Variable | Description | Default |
|----------|-------------|---------|
| `GRIMORIO_TTS_ENABLED` | Enable TTS globally | `false` |
| `PIPER_MODEL_PATH` | Path to `.onnx` model | — |
| `PIPER_CONFIG_PATH` | Path to `.onnx.json` config | (auto) |
| `PIPER_PORT` | Piper HTTP server port | `5000` |
| `PIPER_HOST` | Bind address (localhost only) | `127.0.0.1` |
| `PIPER_LENGTH_SCALE` | Speech speed (<1 faster, >1 slower) | `1.0` |
| `PIPER_VOLUME` | Playback volume (0.0–1.0) | `0.8` |
| `PIPER_MAX_RESTARTS` | Auto-restart attempts on crash | `3` |
| `CHUNKER_MAX_SIZE` | Max characters per chunk | `150` |
| `AUDIO_PLAYER` | Preferred player: `auto`, `aplay`, `paplay`, `ffplay` | `auto` |
| `AUDIO_DEVICE` | Optional audio device string | — |

## Troubleshooting

### "Piper TTS not installed"

- Ensure `piper` binary is in `PATH`: `which piper`
- Verify binary is executable: `chmod +x ~/.local/bin/piper`

### "Piper TTS failed to start"

- Check `PIPER_MODEL_PATH` points to a valid `.onnx` file
- Check port 5000 is not in use: `lsof -i :5000`
- Try a different port via `PIPER_PORT`

### No audio output

- Verify a player is installed: `which aplay` or `which paplay` or `which ffplay`
- Check ALSA/PulseAudio is working: `aplay -l`
- Set `AUDIO_PLAYER` explicitly if auto-detection fails

### Choppy playback or gaps

- Reduce `CHUNKER_MAX_SIZE` to create smaller chunks (faster synthesis)
- Ensure `AUDIO_PRELOAD_BUFFER` is at least `1`
- Check CPU usage during synthesis — Piper may need a modern CPU

### TTS mode fails to activate

- `get_tts_status` will show `available: false` if Piper is not running
- The DM agent will still function in `written` mode automatically

## Migration from Remote TTS

If you used the previous `feat/tts-server` WebSocket-based TTS:

1. Remove `tts_server_url` from config
2. Install Piper locally
3. Set `PIPER_MODEL_PATH` to your `.onnx` model
4. Keep `GRIMORIO_TTS_ENABLED=true`

The MCP tool API (`set_dm_mode`, `assign_npc_voice`, etc.) remains unchanged.
