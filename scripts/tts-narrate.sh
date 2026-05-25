#!/bin/bash
# narrate — TTS for Grimorio DM (streaming, non-blocking usage)
# Usage: narrate "Texto a narrar"
#        echo "Texto" | narrate
# Use in background: (narrate "texto" &) 2>/dev/null

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
        # Look for last sentence boundary within max_len from pos
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
