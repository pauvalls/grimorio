package compiler

import "testing"

// TestMarkdownToHTML_StripsWotCAnnotation verifies the compiler drops
// <!-- WOTC: ... --> annotations and the literal "WOTC:" token does
// not leak into the rendered HTML. Annotations are author-side notes
// that should not appear in the final PDF.
func TestMarkdownToHTML_StripsWotCAnnotation(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "boxed text annotation",
			input: `## La Mansión Dragomir

> **Read-Aloud:** *Las puertas se abren...*
<!-- WOTC: boxed_text — 142 words, second person present tense, sensory details. -->

Continuación del capítulo.`,
		},
		{
			name: "character hook annotation",
			input: `**Ganchos de Personaje:**
- **Katarina:** Reconoce el escudo Voronova.
<!-- WOTC: character_hook — noble background, secret allegiance. -->

Siguiente hook.`,
		},
		{
			name: "development branch annotation",
			input: `**Desarrollos:**
1. **Los PJs examinan los retratos:** Descubren ancestro del culto.
<!-- WOTC: development_branch — 3 IF-THEN with Consecuencia + Recuperación. -->
`,
		},
		{
			name: "multi-line annotation with newlines",
			input: `# Título

Texto previo.

<!--
WOTC: read_aloud
- second_person: yes
- present_tense: yes
- word_count: 142
-->

Texto posterior.`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := markdownToHTML(tt.input, "")

			if got := containsLiteral(result, "WOTC:"); got {
				t.Errorf("WOTC: literal leaked into HTML: %s", result)
			}
			if got := containsLiteral(result, "<!--"); got {
				t.Errorf("HTML comment marker leaked into HTML: %s", result)
			}
		})
	}
}

// containsLiteral is a tiny helper to keep the test readable.
func containsLiteral(s, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	for i := 0; i+len(needle) <= len(s); i++ {
		if s[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
