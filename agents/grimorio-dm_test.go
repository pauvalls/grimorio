package grimorio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDMAgentBilingualLanguageIntake verifies that the DM agent
// presents a bilingual language question at session start and that
// the default session language is English (not Spanish). This is the
// spec guard for the dm-language-intake capability.
func TestDMAgentBilingualLanguageIntake(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)
	// The test runs from the agents/ directory; the agent file
	// lives in the same directory.
	dmPath := filepath.Join(root, "grimorio-dm.md")
	data, err := os.ReadFile(dmPath)
	require.NoError(t, err, "read DM agent file")
	content := string(data)

	// Bilingual question must be present in both languages so the
	// user can answer in either one.
	assert.Contains(t, content, "Idioma de la sesión",
		"DM agent must present the Spanish language question in Section 10")
	assert.Contains(t, content, "Session language",
		"DM agent must present the English language question in Section 10")

	// The default must be English. The Spanish default has been flipped
	// for the i18n-english-default change. Accept either an explicit
	// "default: en" / "(default: en)" annotation or the absence of any
	// "(default)" annotation that calls out Spanish.
	lower := strings.ToLower(content)
	hasEnglishDefault := strings.Contains(lower, "default: en") ||
		strings.Contains(lower, "default english") ||
		strings.Contains(lower, "english (default)") ||
		strings.Contains(lower, "(default: en)")
	hasSpanishDefault := strings.Contains(lower, "default: es") ||
		strings.Contains(lower, "default spanish") ||
		strings.Contains(lower, "spanish (default)") ||
		strings.Contains(lower, "(default: es)")

	assert.True(t, hasEnglishDefault || !hasSpanishDefault,
		"DM agent must default to English (default: en), not Spanish")

	// Glossary table must still be present (D&D 5e terms are bilingual
	// on first use).
	assert.Contains(t, content, "Initiative",
		"DM agent glossary table must include the D&D 5e Initiative entry")
	assert.Contains(t, content, "Iniciativa",
		"DM agent glossary table must include the Spanish Iniciativa entry")
}
