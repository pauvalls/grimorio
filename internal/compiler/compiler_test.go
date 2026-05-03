package compiler

import (
	"strings"
	"testing"
)

func TestMarkdownToHTML_ProcessScenePlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantContains string
		wantNotContains string
	}{
		{
			name:     "scene placeholder becomes descriptive text",
			input:    "*[SCENE: A dark dungeon with flickering torches]*",
			wantContains: "A dark dungeon with flickering torches",
			wantNotContains: "[SCENE:",
		},
		{
			name:     "scene placeholder in standalone line",
			input:    "[SCENE: Epic battle between heroes and dragon]",
			wantContains: "Epic battle between heroes and dragon",
			wantNotContains: "[SCENE:",
		},
		{
			name:     "multiple scene placeholders",
			input:    "[SCENE: First scene]\n\nSome text\n\n[SCENE: Second scene]",
			wantContains: "First scene",
			wantNotContains: "[SCENE:",
		},
		{
			name:     "regular markdown still works",
			input:    "# Heading\n\nSome **bold** text",
			wantContains: "<h1",
			wantNotContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markdownToHTML(tt.input, "/tmp")
			
			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("markdownToHTML() result does not contain %q\nGot: %s", tt.wantContains, result)
			}
			
			if tt.wantNotContains != "" && strings.Contains(result, tt.wantNotContains) {
				t.Errorf("markdownToHTML() result should not contain %q\nGot: %s", tt.wantNotContains, result)
			}
		})
	}
}

func TestMarkdownToHTML_ScenePlaceholder_Formatting(t *testing.T) {
	input := "[SCENE: A mystical forest at dawn]"
	result := markdownToHTML(input, "/tmp")
	
	// Should be wrapped in a styled div or paragraph, not just plain text
	if !strings.Contains(result, "scene-description") && !strings.Contains(result, "<p>") {
		t.Errorf("scene placeholder should be wrapped in HTML element, got: %s", result)
	}
}
