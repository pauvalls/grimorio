#!/bin/bash
# TTS Narrate — Helper script for grimorio-dm voice narration
# Automatically splits text into chunks and narrates sequentially
# Usage: narrate "Texto completo. Se divide en chunks automáticamente."

set -e

TEXT="$1"
if [ -z "$TEXT" ]; then
    echo "Usage: narrate 'Texto a narrar'"
    exit 1
fi

# Default values
PIPER_HOST="${PIPER_HOST:-127.0.0.1}"
PIPER_PORT="${PIPER_PORT:-5000}"
AUDIO_DEVICE="${AUDIO_DEVICE:-}"
MAX_CHUNK_LENGTH="${MAX_CHUNK_LENGTH:-150}"

# Detect audio player
if command -v aplay &> /dev/null; then
    PLAYER="aplay"
    DEVICE_ARG=""
    if [ -n "$AUDIO_DEVICE" ]; then
        DEVICE_ARG="-D $AUDIO_DEVICE"
    fi
elif command -v paplay &> /dev/null; then
    PLAYER="paplay"
    DEVICE_ARG=""
elif command -v ffplay &> /dev/null; then
    PLAYER="ffplay"
    DEVICE_ARG="-nodisp -autoexit"
else
    echo "ERROR: No audio player found (aplay, paplay, ffplay)"
    exit 1
fi

# Function to split text into chunks
split_into_chunks() {
    local text="$1"
    local max_len="$2"
    local chunks=()
    local current=""
    
    # Split by sentence endings (., !, ?)
    while IFS= read -r sentence; do
        sentence=$(echo "$sentence" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
        if [ -z "$sentence" ]; then continue; fi
        
        # If adding this sentence exceeds max length, save current chunk
        local test_chunk="$current $sentence"
        test_chunk=$(echo "$test_chunk" | sed 's/^[[:space:]]*//')
        
        if [ ${#test_chunk} -gt "$max_len" ] && [ -n "$current" ]; then
            chunks+=("$current")
            current="$sentence"
        else
            current="$test_chunk"
        fi
    done <<< "$(echo "$text" | sed 's/\([.!?]\)/\1\n/g')"
    
    # Don't forget the last chunk
    if [ -n "$current" ]; then
        chunks+=("$current")
    fi
    
    printf '%s\n' "${chunks[@]}"
}

# Function to play a chunk
play_chunk() {
    local chunk="$1"
    local chunk_num="$2"
    local total="$3"
    
    echo "  [Chunk $chunk_num/$total] Narrating..." >&2
    
    if [ "$PLAYER" = "aplay" ]; then
        echo "$chunk" | curl -s -X POST "http://$PIPER_HOST:$PIPER_PORT" --data-binary @- | aplay $DEVICE_ARG -
    elif [ "$PLAYER" = "paplay" ]; then
        echo "$chunk" | curl -s -X POST "http://$PIPER_HOST:$PIPER_PORT" --data-binary @- | paplay --raw -
    elif [ "$PLAYER" = "ffplay" ]; then
        echo "$chunk" | curl -s -X POST "http://$PIPER_HOST:$PIPER_PORT" --data-binary @- > /tmp/tts-$$-$chunk_num.wav
        ffplay $DEVICE_ARG /tmp/tts-$$-$chunk_num.wav
        rm -f /tmp/tts-$$-$chunk_num.wav
    fi
}

# Main: split and narrate
echo "🎙️  TTS Narration starting..." >&2

# Read chunks into array
mapfile -t chunks <<< "$(split_into_chunks "$TEXT" "$MAX_CHUNK_LENGTH")"
total=${#chunks[@]}

if [ "$total" -eq 0 ]; then
    echo "ERROR: No text to narrate"
    exit 1
fi

echo "   Split into $total chunk(s)" >&2

# Play each chunk sequentially
for i in "${!chunks[@]}"; do
    chunk_num=$((i + 1))
    chunk="${chunks[$i]}"
    play_chunk "$chunk" "$chunk_num" "$total"
done

echo "✅ Narration complete" >&2
