package grimorio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skillProhibitedSpanish are Spanish prose tokens that MUST NOT appear
// in the body of any skill file. Code blocks (```...```) and
// template variables ({{...}}) are stripped before checking so that
// structural markup is exempt. These tokens cover the locked glossary
// from the i18n-english-default design doc.
var skillProhibitedSpanish = []string{
	"Caja de texto",
	"Apariencia",
	"Personalidad",
	"Motivación",
	"Secreto",
	"Involucramiento en Quests",
	"Conexiones",
	"Trasfondo",
	"Alineamiento",
	"Ataque",
	"Daño",
	"Tesoro",
	"Improvisar",
}

// TestSkillsNoSpanishProse verifies that no skill file in
// skills/grimorio-*/SKILL.md contains Spanish prose tokens from the
// locked glossary. Code blocks and template variables are stripped
// before checking so that structural markup is exempt.
func TestSkillsNoSpanishProse(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)
	skillsRoot := filepath.Join(root, "skills")

	entries, err := os.ReadDir(skillsRoot)
	require.NoError(t, err, "should be able to read skills directory")

	var skillDirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if !strings.HasPrefix(entry.Name(), "grimorio-") {
			continue
		}
		skillDirs = append(skillDirs, entry.Name())
	}

	// grimorio-areas was removed in v5.0.2 WU7 — see the scope-change
	// decision in engram topic `sdd/v5.0.2-wotc-fidelity-pdf-render/scope-change-delete-areas`.
	require.GreaterOrEqual(t, len(skillDirs), 16,
		"expected at least 16 grimorio-* skill directories, found %d", len(skillDirs))

	for _, dir := range skillDirs {
		t.Run(dir, func(t *testing.T) {
			skillPath := filepath.Join(skillsRoot, dir, "SKILL.md")
			data, err := os.ReadFile(skillPath)
			require.NoError(t, err, "read %s", skillPath)
			stripped := stripCodeBlocks(string(data))
			stripped = stripTemplateVars(stripped)
			for _, banned := range skillProhibitedSpanish {
				assert.NotContains(t, stripped, banned,
					"skill %s still contains Spanish prose %q — i18n-english-default regression", dir, banned)
			}
		})
	}
}

// TestArchitectSkillHasLanguageIntake asserts the architect skill
// preserves the bilingual language intake (Work-Unit 1) and the
// LANG: preamble on every delegate(...) call. This is a guard for
// the spec: every delegate must carry the language preamble.
func TestArchitectSkillHasLanguageIntake(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)
	skillPath := filepath.Join(root, "skills", "grimorio-architect", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	require.NoError(t, err, "read architect SKILL.md")
	content := string(data)

	assert.Contains(t, content, "¿En qué idioma prefieres jugar?",
		"architect skill must present the Spanish language question")
	assert.Contains(t, content, "What language do you prefer to play in?",
		"architect skill must present the English language question")
	assert.Contains(t, content, "LANG:",
		"architect skill must use the LANG: preamble for sub-agent prompts")
}

// stripCodeBlocks removes all fenced code blocks (```...```) from the
// content. Code blocks are structural and exempt from the i18n check.
func stripCodeBlocks(content string) string {
	var out strings.Builder
	inBlock := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inBlock = !inBlock
			continue
		}
		if inBlock {
			continue
		}
		out.WriteString(line)
		out.WriteString("\n")
	}
	return out.String()
}

// stripTemplateVars replaces every `{{...}}` template variable with
// a single space so the i18n check does not see Spanish hints inside
// the variable names. Brace pairs are matched by simple counting.
func stripTemplateVars(content string) string {
	var out strings.Builder
	idx := 0
	for idx < len(content) {
		if idx+1 < len(content) && content[idx] == '{' && content[idx+1] == '{' {
			depth := 2
			j := idx + 2
			for j < len(content)-1 && depth > 0 {
				if content[j] == '{' && content[j+1] == '{' {
					depth += 2
					j += 2
					continue
				}
				if content[j] == '}' && content[j+1] == '}' {
					depth -= 2
					j += 2
					continue
				}
				j++
			}
			out.WriteByte(' ')
			idx = j
			continue
		}
		out.WriteByte(content[idx])
		idx++
	}
	return out.String()
}
