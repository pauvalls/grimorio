package dm

import (
	"strings"
	"testing"
)

func TestChunkTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		ct       ChunkType
		expected string
	}{
		{"narration", ChunkNarration, "narration"},
		{"npc_dialogue", ChunkNPCDialogue, "npc_dialogue"},
		{"action", ChunkAction, "action"},
		{"dice_roll", ChunkDiceRoll, "dice_roll"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.ct) != tt.expected {
				t.Errorf("ChunkType = %q, want %q", tt.ct, tt.expected)
			}
		})
	}
}

func TestSplitIntoChunks_EmptyText(t *testing.T) {
	chunks := SplitIntoChunks("", 500)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty text, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_SimpleNarration(t *testing.T) {
	text := "The adventurers enter the dark cave."
	chunks := SplitIntoChunks(text, 500)

	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	if chunks[0].Type != ChunkNarration {
		t.Errorf("expected type narration, got %q", chunks[0].Type)
	}

	if chunks[0].Text != text {
		t.Errorf("expected text %q, got %q", text, chunks[0].Text)
	}

	if chunks[0].VoiceID != "narrator" {
		t.Errorf("expected voice_id narrator, got %q", chunks[0].VoiceID)
	}
}

func TestSplitIntoChunks_NPCDialogue(t *testing.T) {
	text := `The tavern is noisy. "Welcome, travelers!" says the innkeeper. You look around.`
	chunks := SplitIntoChunks(text, 500)

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}

	foundDialogue := false
	for _, c := range chunks {
		if c.Type == ChunkNPCDialogue {
			foundDialogue = true
			if !strings.Contains(c.Text, "Welcome, travelers!") {
				t.Errorf("dialogue chunk missing expected text: %q", c.Text)
			}
			if c.VoiceID == "" {
				t.Error("dialogue chunk should have a voice_id")
			}
		}
	}

	if !foundDialogue {
		t.Error("expected at least one npc_dialogue chunk")
	}
}

func TestSplitIntoChunks_ActionMarkers(t *testing.T) {
	text := `The dragon roars. *The ground shakes beneath your feet.* You brace yourselves.`
	chunks := SplitIntoChunks(text, 500)

	foundAction := false
	for _, c := range chunks {
		if c.Type == ChunkAction {
			foundAction = true
			if !strings.Contains(c.Text, "ground shakes") {
				t.Errorf("action chunk missing expected text: %q", c.Text)
			}
		}
	}

	if !foundAction {
		t.Error("expected at least one action chunk")
	}
}

func TestSplitIntoChunks_DiceRoll(t *testing.T) {
	text := `The trap triggers. [Dexterity saving throw, DC 15] You dive aside.`
	chunks := SplitIntoChunks(text, 500)

	foundDice := false
	for _, c := range chunks {
		if c.Type == ChunkDiceRoll {
			foundDice = true
			if !strings.Contains(c.Text, "Dexterity saving throw") {
				t.Errorf("dice chunk missing expected text: %q", c.Text)
			}
		}
	}

	if !foundDice {
		t.Error("expected at least one dice_roll chunk")
	}
}

func TestSplitIntoChunks_MaxLengthSplit(t *testing.T) {
	// Create a text longer than maxLen without clear sentence boundaries
	text := strings.Repeat("a", 600)
	chunks := SplitIntoChunks(text, 500)

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}

	for _, c := range chunks {
		if len(c.Text) > 500 {
			t.Errorf("chunk text length %d exceeds maxLen 500", len(c.Text))
		}
	}
}

func TestSplitIntoChunks_SpanishPunctuation(t *testing.T) {
	text := `¿Dónde está el tesoro? ¡Está aquí! Los aventureros miran.`
	chunks := SplitIntoChunks(text, 500)

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}

	// Should split on Spanish punctuation boundaries
	foundSpanish := false
	for _, c := range chunks {
		if strings.Contains(c.Text, "¿") || strings.Contains(c.Text, "¡") {
			foundSpanish = true
		}
	}

	if !foundSpanish {
		t.Error("expected chunks to preserve Spanish punctuation")
	}
}

func TestSplitIntoChunks_MixedContent(t *testing.T) {
	text := `The heroes arrive at the gates. "Halt! Who goes there?" shouts the guard. *The wind howls through the battlements.* [Perception check, DC 12] You sense danger.`
	chunks := SplitIntoChunks(text, 500)

	if len(chunks) == 0 {
		t.Fatal("expected chunks, got none")
	}

	typeCounts := make(map[ChunkType]int)
	for _, c := range chunks {
		typeCounts[c.Type]++
	}

	if typeCounts[ChunkNarration] == 0 {
		t.Error("expected at least one narration chunk")
	}
	if typeCounts[ChunkNPCDialogue] == 0 {
		t.Error("expected at least one npc_dialogue chunk")
	}
	if typeCounts[ChunkAction] == 0 {
		t.Error("expected at least one action chunk")
	}
	if typeCounts[ChunkDiceRoll] == 0 {
		t.Error("expected at least one dice_roll chunk")
	}
}

func TestSplitIntoChunks_LongTextParagraphSplit(t *testing.T) {
	// Two paragraphs, each under 500 chars
	para1 := "The first paragraph describes the scene in detail. It has multiple sentences."
	para2 := "The second paragraph continues the story. Another sentence here."
	text := para1 + "\n\n" + para2
	chunks := SplitIntoChunks(text, 500)

	if len(chunks) != 2 {
		t.Errorf("expected 2 chunks for two paragraphs, got %d", len(chunks))
	}
}

func TestSplitIntoChunks_NPCIDExtraction(t *testing.T) {
	text := `"We must hurry," says Elara. The group nods.`
	chunks := SplitIntoChunks(text, 500)

	foundElara := false
	for _, c := range chunks {
		if c.Type == ChunkNPCDialogue {
			foundElara = true
			if c.NPCID == "" {
				t.Error("expected NPCID to be set for dialogue chunk")
			}
		}
	}

	if !foundElara {
		t.Error("expected npc_dialogue chunk")
	}
}
