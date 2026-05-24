package dm

import (
	"regexp"
	"strings"
	"unicode"
)

// ChunkType categorizes a piece of narrative text.
type ChunkType string

const (
	ChunkNarration   ChunkType = "narration"
	ChunkNPCDialogue ChunkType = "npc_dialogue"
	ChunkAction      ChunkType = "action"
	ChunkDiceRoll    ChunkType = "dice_roll"
)

// NarrativeChunk represents a single unit of narrative to be spoken.
type NarrativeChunk struct {
	Type     ChunkType `json:"type"`
	Text     string    `json:"text"`
	NPCID    string    `json:"npc_id,omitempty"`
	VoiceID  string    `json:"voice_id"`
	ChunkID  string    `json:"chunk_id,omitempty"`
	Priority int       `json:"priority,omitempty"`
	Preload  bool      `json:"preload,omitempty"`
}

var (
	// dialogueRegex matches quoted text, optionally capturing the speaker.
	dialogueRegex = regexp.MustCompile(`"([^"]+)"`)
	// actionRegex matches *italicized action* text.
	actionRegex = regexp.MustCompile(`\*([^*]+)\*`)
	// diceRegex matches [dice roll descriptions].
	diceRegex = regexp.MustCompile(`\[([^\]]+)\]`)
	// speakerRegex tries to find the NPC name after dialogue.
	speakerRegex = regexp.MustCompile(`"[^"]+"\s*(?:dice|dice\s+el|dice\s+la|dice\s+la|dice\s+el|dice|dijo|dijo\s+el|dijo\s+la|says|said|grita|gritó)\s+([A-ZÁÉÍÓÚÑ][a-záéíóúñ]+(?:\s+[A-ZÁÉÍÓÚÑ][a-záéíóúñ]+)*)`)
)

// TextFilter filters text before chunking.
type TextFilter interface {
	Filter(text string) string
}

// SplitFiltered first applies the filter, then splits into chunks.
// If filter is nil, it behaves identically to SplitIntoChunks.
func SplitFiltered(text string, filter TextFilter, maxLen int) []NarrativeChunk {
	if filter != nil {
		text = filter.Filter(text)
	}
	return SplitIntoChunks(text, maxLen)
}

// SplitIntoChunks parses DM text into typed narrative chunks.
// It first segments by paragraph, then detects dialogue, action, and dice markers.
// Long segments are split on sentence boundaries respecting maxLen.
func SplitIntoChunks(text string, maxLen int) []NarrativeChunk {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	paragraphs := strings.Split(text, "\n\n")
	var chunks []NarrativeChunk

	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		chunks = append(chunks, splitParagraph(para, maxLen)...)
	}

	return chunks
}

func splitParagraph(para string, maxLen int) []NarrativeChunk {
	// Find all special segments with their positions
	segments := extractSegments(para)

	if len(segments) == 0 {
		return splitLongSegment(para, ChunkNarration, "", maxLen)
	}

	var chunks []NarrativeChunk
	lastEnd := 0

	for _, seg := range segments {
		if seg.start > lastEnd {
			narration := strings.TrimSpace(para[lastEnd:seg.start])
			if narration != "" {
				chunks = append(chunks, splitLongSegment(narration, ChunkNarration, "", maxLen)...)
			}
		}

		voiceID := "narrator"
		npcID := ""
		if seg.chunkType == ChunkNPCDialogue {
			npcID = extractSpeaker(para, seg.start, seg.end)
			voiceID = "npc-" + Slugify(npcID)
		}

		chunks = append(chunks, splitLongSegment(seg.text, seg.chunkType, voiceID, maxLen)...)
		if seg.chunkType == ChunkNPCDialogue {
			// Tag the last chunk with NPCID
			lastIdx := len(chunks) - 1
			chunks[lastIdx].NPCID = npcID
			chunks[lastIdx].VoiceID = voiceID
		}

		lastEnd = seg.end
	}

	if lastEnd < len(para) {
		narration := strings.TrimSpace(para[lastEnd:])
		if narration != "" {
			chunks = append(chunks, splitLongSegment(narration, ChunkNarration, "", maxLen)...)
		}
	}

	return chunks
}

type segment struct {
	start     int
	end       int
	text      string
	chunkType ChunkType
}

func extractSegments(para string) []segment {
	var segments []segment

	// Find dialogue
	for _, m := range dialogueRegex.FindAllStringIndex(para, -1) {
		segments = append(segments, segment{
			start:     m[0],
			end:       m[1],
			text:      para[m[0]:m[1]],
			chunkType: ChunkNPCDialogue,
		})
	}

	// Find action
	for _, m := range actionRegex.FindAllStringIndex(para, -1) {
		segments = append(segments, segment{
			start:     m[0],
			end:       m[1],
			text:      para[m[0]:m[1]],
			chunkType: ChunkAction,
		})
	}

	// Find dice
	for _, m := range diceRegex.FindAllStringIndex(para, -1) {
		segments = append(segments, segment{
			start:     m[0],
			end:       m[1],
			text:      para[m[0]:m[1]],
			chunkType: ChunkDiceRoll,
		})
	}

	// Sort by start position
	for i := 0; i < len(segments); i++ {
		for j := i + 1; j < len(segments); j++ {
			if segments[j].start < segments[i].start {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}

	// Remove overlapping segments (first one wins)
	var filtered []segment
	var lastEnd int = -1
	for _, seg := range segments {
		if seg.start >= lastEnd {
			filtered = append(filtered, seg)
			lastEnd = seg.end
		}
	}

	return filtered
}

func extractSpeaker(para string, segStart, segEnd int) string {
	// Look in a window after the dialogue for speaker attribution
	windowEnd := segEnd + 60
	if windowEnd > len(para) {
		windowEnd = len(para)
	}
	window := para[segEnd:windowEnd]

	matches := speakerRegex.FindStringSubmatch(window)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return "unknown"
}

func splitLongSegment(text string, chunkType ChunkType, voiceID string, maxLen int) []NarrativeChunk {
	if len(text) <= maxLen {
		v := voiceID
		if v == "" {
			v = "narrator"
		}
		return []NarrativeChunk{{
			Type:    chunkType,
			Text:    text,
			VoiceID: v,
		}}
	}

	// Find sentence boundary split points
	breaks := []string{". ", "! ", "? ", "\n"}
	var splitPoints []int
	for _, b := range breaks {
		for i := 0; i <= len(text)-len(b); i++ {
			if text[i:i+len(b)] == b {
				splitPoints = append(splitPoints, i+len(b))
			}
		}
	}
	// Sort split points
	for i := 0; i < len(splitPoints); i++ {
		for j := i + 1; j < len(splitPoints); j++ {
			if splitPoints[j] < splitPoints[i] {
				splitPoints[i], splitPoints[j] = splitPoints[j], splitPoints[i]
			}
		}
	}

	var chunks []NarrativeChunk
	start := 0

	for _, sp := range splitPoints {
		if sp <= start {
			continue
		}
		if sp-start <= maxLen {
			chunks = appendChunk(chunks, text, start, sp, chunkType, voiceID)
			start = sp
		} else {
			// This split point is too far; need to split earlier
			breakAt := findSplitPoint(text, start, maxLen)
			chunks = appendChunk(chunks, text, start, breakAt, chunkType, voiceID)
			start = breakAt
		}
	}

	// Handle remaining text
	for start < len(text) {
		breakAt := findSplitPoint(text, start, maxLen)
		chunks = appendChunk(chunks, text, start, breakAt, chunkType, voiceID)
		start = breakAt
	}

	return chunks
}

func appendChunk(chunks []NarrativeChunk, text string, start, end int, chunkType ChunkType, voiceID string) []NarrativeChunk {
	v := voiceID
	if v == "" {
		v = "narrator"
	}
	return append(chunks, NarrativeChunk{
		Type:    chunkType,
		Text:    strings.TrimSpace(text[start:end]),
		VoiceID: v,
	})
}

func findSplitPoint(text string, start, maxLen int) int {
	end := start + maxLen
	if end >= len(text) {
		return len(text)
	}
	// Look for a space to break on, searching backwards from end
	for i := end; i > start; i-- {
		if text[i-1] == ' ' {
			return i
		}
	}
	// No space found, hard split
	return end
}

// Slugify converts a name to a URL-friendly slug.
func Slugify(name string) string {
	name = strings.ToLower(name)
	var result []rune
	for _, r := range name {
		switch r {
		case 'á', 'à', 'â', 'ä':
			result = append(result, 'a')
		case 'é', 'è', 'ê', 'ë':
			result = append(result, 'e')
		case 'í', 'ì', 'î', 'ï':
			result = append(result, 'i')
		case 'ó', 'ò', 'ô', 'ö':
			result = append(result, 'o')
		case 'ú', 'ù', 'û', 'ü':
			result = append(result, 'u')
		case 'ñ':
			result = append(result, 'n')
		case ' ':
			result = append(result, '-')
		default:
			if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' {
				result = append(result, r)
			}
		}
	}
	return string(result)
}
