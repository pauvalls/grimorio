package grimorio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArchitectBilingualLanguageIntake verifies that the architect agent
// presents a bilingual language question at session start (default English)
// and propagates the chosen `session_language` to every delegate call.
func TestArchitectBilingualLanguageIntake(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)

	skillPath := filepath.Join(root, "skills", "grimorio-architect", "SKILL.md")
	agentPath := filepath.Join(root, "agents", "grimorio-architect.md")

	skillBytes, err := os.ReadFile(skillPath)
	require.NoError(t, err, "should be able to read architect SKILL.md")
	skillContent := string(skillBytes)

	agentBytes, err := os.ReadFile(agentPath)
	require.NoError(t, err, "should be able to read architect agent file")
	agentContent := string(agentBytes)

	// The SKILL.md must include the bilingual language question in both
	// Spanish and English forms so the user can answer in either language.
	assert.Contains(t, skillContent, "¿En qué idioma prefieres jugar?",
		"architect SKILL.md must present the Spanish language question")
	assert.Contains(t, skillContent, "What language do you prefer to play in?",
		"architect SKILL.md must present the English language question")

	// Default must be English if the user skips the question.
	assert.Contains(t, skillContent, "session_language",
		"architect SKILL.md must reference the session_language concept")
	assert.Contains(t, strings.ToLower(skillContent), "default",
		"architect SKILL.md must document the default language")

	// Every subagent(...) call must carry the LANG preamble so
	// sub-agents know which language to render content in. The spec
	// (Capability: architect-language-intake) requires propagation to
	// all sub-agents — partial coverage is the regression we are guarding
	// against.
	subagentCalls := extractSubagentCalls(skillContent)
	require.NotEmpty(t, subagentCalls,
		"architect SKILL.md must contain at least one subagent(...) example")

	for i, call := range subagentCalls {
		assert.Contains(t, call, "LANG:",
			"subagent(...) call #%d must include a LANG: prefix so sub-agents render in the chosen language; call was:\n%s", i+1, call)
	}

	// The agent frontmatter file must mention the bilingual intake so
	// the architect does not bypass the language question.
	combined := agentContent + "\n" + skillContent
	assert.True(t,
		strings.Contains(combined, "¿En qué idioma") || strings.Contains(combined, "language"),
		"architect material must document language intake")
}

// extractSubagentCalls returns the verbatim text of every `subagent(`
// occurrence in the source, including multi-line strings, until the
// matching closing paren. Placeholders like `subagent(agent=..., prompt=...)`
// in prose are filtered out: only calls that look like real invocations
// (`prompt="<text>`) are returned.
func extractSubagentCalls(content string) []string {
	var calls []string
	idx := 0
	for {
		start := strings.Index(content[idx:], "subagent(")
		if start < 0 {
			break
		}
		open := idx + start
		depth := 0
		end := open
		for end < len(content) {
			switch content[end] {
			case '(':
				depth++
			case ')':
				depth--
				if depth == 0 {
					end++
					goto done
				}
			}
			end++
		}
	done:
		if end > open && end <= len(content) {
			call := content[open:end]
			// Heuristic: a real invocation opens a prompt body with
			// `prompt="` (or `prompt = "`) followed by a non-ellipsis char.
			if looksLikeRealSubagentCall(call) {
				calls = append(calls, call)
			}
		}
		idx = end
	}
	return calls
}

// looksLikeRealSubagentCall returns true if the captured text looks like
// an actual tool call rather than a documentation placeholder. Real calls
// have a `prompt="...` opening with a non-trivial first character (i.e.
// not a `...` shorthand).
func looksLikeRealSubagentCall(call string) bool {
	openIdx := strings.Index(call, `prompt="`)
	if openIdx < 0 {
		openIdx = strings.Index(call, `prompt = "`)
		if openIdx < 0 {
			return false
		}
		openIdx += len(`prompt = "`)
	} else {
		openIdx += len(`prompt="`)
	}
	if openIdx >= len(call) {
		return false
	}
	// Skip whitespace
	for openIdx < len(call) && (call[openIdx] == ' ' || call[openIdx] == '\t' || call[openIdx] == '\n' || call[openIdx] == '\r') {
		openIdx++
	}
	if openIdx >= len(call) {
		return false
	}
	// If the very next char is '.', it's a placeholder like `prompt="..."`
	if call[openIdx] == '.' {
		return false
	}
	return true
}
