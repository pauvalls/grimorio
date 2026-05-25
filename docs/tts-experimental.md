# TTS (Text-to-Speech) — Experimental

> **⚠️ Experimental**: This feature requires manual setup and is not enabled by default. Once configured, it works automatically in every session.

## Overview

Grimorio uses **Piper TTS** (local neural voice synthesis) to narrate DM responses aloud. The `grimorio-dm` agent automatically narrates every narrative block when TTS is enabled by the player.

**How it works:**
```
grimorio-dm → setsid narrate "text" → Piper HTTP → WAV → aplay → Speakers
```

**Key features:**
- **Fully local** — No cloud, no API keys, no internet required after setup
- **Automatic** — Agent narrates every response without asking
- **Auto-chunking** — Splits long text into ~150 char sentence chunks
- **Background audio** — Uses `setsid` to survive shell timeouts
- **Session-scoped** — Player enables/disables per session

## Prerequisites

- Piper binary (`~/.local/bin/piper/piper`)
- Voice model (e.g., `es_ES-davefx-medium`)
- Piper HTTP server running on localhost:5000
- `aplay` or `paplay` (audio player)
- `curl`
- `setsid` (usually pre-installed on Linux)

## Installation

### 1. Install Piper

```bash
# Create directories
mkdir -p ~/.local/bin/piper
mkdir -p ~/.local/share/piper

# Download Piper prebuilt binary
wget https://github.com/rhasspy/piper/releases/download/v1.2.0/piper_amd64.tar.gz -O /tmp/piper.tar.gz
tar -xzf /tmp/piper.tar.gz -C ~/.local/bin/piper

# Make it available in PATH
ln -sf ~/.local/bin/piper/piper ~/.local/bin/piper
```

### 2. Download a Voice

Browse voices at [huggingface.co/rhasspy/piper-voices](https://huggingface.co/rhasspy/piper-voices/tree/main)

**Spanish (Spain) — default:**
```bash
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx -P ~/.local/share/piper/
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_ES/davefx/medium/es_ES-davefx-medium.onnx.json -P ~/.local/share/piper/
```

**Spanish (Mexico):**
```bash
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_MX/ald/medium/es_MX-ald-medium.onnx -P ~/.local/share/piper/
wget https://huggingface.co/rhasspy/piper-voices/resolve/main/es/es_MX/ald/medium/es_MX-ald-medium.onnx.json -P ~/.local/share/piper/
```

### 3. Create the Narrate Script

Create `~/.local/bin/narrate`:

```bash
cat > ~/.local/bin/narrate << 'EOF'
#!/bin/bash
# narrate — TTS for Grimorio DM (streaming, non-blocking usage)
# Usage: narrate "Text to narrate"
# Intended to be run with setsid: setsid narrate "text" > /dev/null 2>&1

TEXT="$1"
if [ -z "$TEXT" ]; then
    TEXT=$(cat)
fi
[ -z "$TEXT" ] && exit 0

PIPER_HOST="${PIPER_HOST:-127.0.0.1}"
PIPER_PORT="${PIPER_PORT:-5000}"

# Detect audio player
if command -v aplay &> /dev/null; then
    PLAYER="aplay"
elif command -v paplay &> /dev/null; then
    PLAYER="paplay"
elif command -v ffplay &> /dev/null; then
    PLAYER="ffplay"
else
    echo "No audio player found" >&2
    exit 1
fi

# Play one chunk via Piper HTTP -> audio player
play_chunk() {
    local text="$1"
    case "$PLAYER" in
        aplay)
            echo "$text" | curl -s -X POST "http://$PIPER_HOST:$PIPER_PORT" --data-binary @- | aplay -q -
            ;;
        paplay)
            echo "$text" | curl -s -X POST "http://$PIPER_HOST:$PIPER_PORT" --data-binary @- | paplay --raw -
            ;;
        ffplay)
            local wav="/tmp/tts-$$.wav"
            echo "$text" | curl -s -X POST "http://$PIPER_HOST:$PIPER_PORT" --data-binary @- > "$wav"
            ffplay -nodisp -autoexit "$wav" 2>/dev/null
            rm -f "$wav"
            ;;
    esac
}

# Split into chunks by sentence breaks, max ~150 chars
split_and_play() {
    local text="$1"
    local max_len=150
    local total_len=${#text}
    local pos=0
    
    while [ "$pos" -lt "$total_len" ]; do
        local end=$((pos + max_len))
        [ "$end" -gt "$total_len" ] && end=$total_len
        
        local break_at=$end
        local i=$end
        while [ "$i" -gt "$pos" ]; do
            local c="${text:$i:1}"
            if [ "$c" = "." ] || [ "$c" = "!" ] || [ "$c" = "?" ]; then
                break_at=$((i + 1))
                break
            fi
            i=$((i - 1))
        done
        
        local chunk="${text:$pos:$((break_at - pos))}"
        chunk="${chunk#"${chunk%%[![:space:]]*}"}"  # trim leading
        [ -n "$chunk" ] && play_chunk "$chunk"
        
        pos=$break_at
        # skip whitespace
        while [ "$pos" -lt "$total_len" ] && [ "${text:$pos:1}" = " " ]; do
            pos=$((pos + 1))
        done
    done
}

split_and_play "$TEXT"
EOF

chmod +x ~/.local/bin/narrate
```

### 4. Start Piper HTTP Server

Create a startup script `~/.local/bin/piper-http-server`:

```bash
cat > ~/.local/bin/piper-http-server << 'EOF'
#!/usr/bin/env python3
import http.server, socketserver, subprocess, sys, os

PORT = int(os.environ.get("PIPER_PORT", "5000"))
MODEL = os.environ.get("PIPER_MODEL_PATH", os.path.expanduser("~/.local/share/piper/es_ES-davefx-medium.onnx"))
CONFIG = os.environ.get("PIPER_CONFIG_PATH", MODEL + ".json")
PIPER_BIN = os.path.expanduser("~/.local/bin/piper/piper")

class Handler(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        length = int(self.headers.get('Content-Length', 0))
        text = self.rfile.read(length).decode('utf-8')
        self.send_response(200)
        self.send_header('Content-Type', 'audio/wav')
        self.end_headers()
        
        cmd = [PIPER_BIN, "--model", MODEL, "--config", CONFIG, "--output_file", "-"]
        proc = subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=self.wfile)
        proc.stdin.write(text.encode('utf-8'))
        proc.stdin.close()
        proc.wait()

    def log_message(self, format, *args):
        pass  # Silence logs

with socketserver.TCPServer(("127.0.0.1", PORT), Handler) as httpd:
    print(f"Piper HTTP server on 127.0.0.1:{PORT}")
    httpd.serve_forever()
EOF

chmod +x ~/.local/bin/piper-http-server
```

Start the server (keep it running in a separate terminal):
```bash
piper-http-server
```

Or with systemd/user service for auto-start.

### 5. Configure Environment

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
# Piper TTS Configuration
export PATH="$HOME/.local/bin:$PATH"
export PIPER_MODEL_PATH="$HOME/.local/share/piper/es_ES-davefx-medium.onnx"
export PIPER_CONFIG_PATH="$HOME/.local/share/piper/es_ES-davefx-medium.onnx.json"
export PIPER_PORT=5000
export PIPER_HOST="127.0.0.1"
export AUDIO_PLAYER=auto
```

Reload:
```bash
source ~/.bashrc  # or ~/.zshrc
```

### 6. Apply to OpenCode

```bash
grimorio update commands
```

This adds TTS environment variables to `~/.config/opencode/opencode.json`.

## Usage in Sessions

1. **Start a session** with `grimorio-dm`
2. **Agent asks:** "🔊 TTS: Available. Activate? Yes/No"
3. **Say Yes** — TTS is enabled for this session
4. **Every narrative response** is automatically narrated:
   - Agent writes text on screen
   - Shows "🎙️ Narrando..."
   - Executes: `setsid narrate "text" > /dev/null 2>&1`
5. **Audio plays** in background while you continue reading/playing

**To disable during session:** Say "detener voz" or "stop TTS"

## Testing

```bash
# Test Piper server is running
curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:5000
# Should return: 200

# Test narrate script
echo "Hola, esto es una prueba." | narrate

# Test with setsid (production mode)
setsid narrate "Hola desde Grimorio. La voz funciona perfectamente." > /dev/null 2>&1

# List audio devices
aplay -l
```

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `curl: connection refused` | Piper HTTP server not running | Start `piper-http-server` |
| No audio output | Wrong audio device | Check `aplay -l`, set `AUDIO_DEVICE` |
| `narrate: command not found` | Script not in PATH | Ensure `~/.local/bin` is in PATH |
| Audio cuts off mid-sentence | Shell timeout kills background process | Must use `setsid`, not `&` |
| TTS not running automatically | Agent not executing bash tool | Check agent instructions updated |
| "🎙️ Narrando..." but no sound | Piper server down or audio issue | Check `curl http://127.0.0.1:5000` |

## Changing Voice

1. Download a new voice from [Piper Voices](https://huggingface.co/rhasspy/piper-voices/tree/main)
2. Update `PIPER_MODEL_PATH` and `PIPER_CONFIG_PATH`
3. Restart `piper-http-server`

```bash
# Example: Switch to Mexican Spanish
export PIPER_MODEL_PATH="$HOME/.local/share/piper/es_MX-ald-medium.onnx"
export PIPER_CONFIG_PATH="$HOME/.local/share/piper/es_MX-ald-medium.onnx.json"
# Restart piper-http-server
```

## Uninstall

```bash
rm -rf ~/.local/bin/piper ~/.local/bin/piper-http-server ~/.local/bin/narrate ~/.local/share/piper
# Remove env vars from ~/.bashrc or ~/.zshrc
```

## Known Limitations

- **Experimental** — Requires manual setup, not automatic
- **One voice per session** — Cannot switch voices mid-session
- **Sentence chunking** — Small gaps between chunks (not seamless)
- **Shell timeout** — Expected timeout message, audio continues (this is normal)
- **No NPC voice differentiation** — All text uses the same voice

## Future Improvements

- [ ] Auto-installer for Piper + voices
- [ ] Voice selection per NPC
- [ ] Seamless streaming without gaps
- [ ] macOS/Windows audio support beyond aplay
- [ ] Pause/resume functionality