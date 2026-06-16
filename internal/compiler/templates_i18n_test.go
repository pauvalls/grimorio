package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// templateProhibitedSpanish are Spanish prose tokens that MUST NOT appear
// in the body of any template. Template variables ({{...}}) and fenced
// code blocks (```...```) are stripped before checking so structural
// markup is exempt. These tokens cover the locked glossary from the
// i18n-english-default design doc.
var templateProhibitedSpanish = []string{
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

// TestTemplatesNoSpanishProse verifies that no template body contains
// Spanish prose tokens from the locked glossary. Template variables
// ({{...}}) and fenced code blocks are stripped before checking so
// that structural markup is exempt.
func TestTemplatesNoSpanishProse(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)
	templatesDir := filepath.Join(root, "templates")

	entries, err := os.ReadDir(templatesDir)
	require.NoError(t, err, "should be able to read templates directory")

	var templateFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md.tmpl") {
			continue
		}
		templateFiles = append(templateFiles, entry.Name())
	}

	require.NotEmpty(t, templateFiles,
		"no .md.tmpl files found in internal/compiler/templates/ — guard would silently pass")

	for _, filename := range templateFiles {
		t.Run(filename, func(t *testing.T) {
			path := filepath.Join(templatesDir, filename)
			data, err := os.ReadFile(path)
			require.NoError(t, err, "read template %s", filename)
			stripped := stripCodeBlocks(string(data))
			stripped = stripTemplateVars(stripped)
			for _, banned := range templateProhibitedSpanish {
				assert.NotContains(t, stripped, banned,
					"template %s still contains Spanish prose %q — i18n-english-default regression", filename, banned)
			}
		})
	}
}

// TestAreasTemplatePreservesWotCGrepTargets asserts that the WotC
// structural anchors in areas.md.tmpl are preserved byte-identical.
// These are the markers validate-campaign.sh greps for to compute
// thresholds (boxed text, character hooks, developments, sidebars,
// running guidance).
func TestAreasTemplatePreservesWotCGrepTargets(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)
	areasPath := filepath.Join(root, "templates", "areas.md.tmpl")
	data, err := os.ReadFile(areasPath)
	require.NoError(t, err)
	content := string(data)

	// Heading anchor — the canonical areas.md.tmpl has a sample
	// "## Area 1: ..." section that documents the format acts must
	// follow. The grep target "^## Area" must still match.
	hasAreaHeading := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## Area") {
			hasAreaHeading = true
			break
		}
	}
	assert.True(t, hasAreaHeading,
		"areas.md.tmpl must contain a '## Area' heading anchor for validate-campaign.sh's area counter")

	// Running the Scene heading is required by the WotC validation
	// awk filter. The Spanish form "^### Cómo Dirigir esta Escena"
	// is no longer produced by the template; the English form is.
	assert.Contains(t, content, "### Running the Scene",
		"areas.md.tmpl must contain the English '### Running the Scene' heading so validate-campaign.sh's awk filter still matches")

	// Sidebar marker (^> #####) — must be present.
	hasSidebarMarker := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "> #####") {
			hasSidebarMarker = true
			break
		}
	}
	assert.True(t, hasSidebarMarker,
		"areas.md.tmpl must contain a '> #####' sidebar marker line for validate-campaign.sh's sidebar check")

	// Boxed text marker (^>>) — at least one read-aloud sample.
	hasBoxedMarker := false
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimRight(line, " \t"), ">>") {
			hasBoxedMarker = true
			break
		}
	}
	assert.True(t, hasBoxedMarker,
		"areas.md.tmpl must contain at least one '>>' boxed-text marker line for validate-campaign.sh's boxed-text check")
}

// TestTemplatesHaveEnglishTopLevelHeading asserts that each template
// has a top-level English heading (i.e. a `# Title` line that is not
// in Spanish). This is a smoke check that the file is structurally
// an English markdown template.
func TestTemplatesHaveEnglishTopLevelHeading(t *testing.T) {
	root, err := os.Getwd()
	require.NoError(t, err)
	templatesDir := filepath.Join(root, "templates")

	entries, err := os.ReadDir(templatesDir)
	require.NoError(t, err)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md.tmpl") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			path := filepath.Join(templatesDir, entry.Name())
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			content := string(data)
			stripped := stripCodeBlocks(content)
			stripped = stripTemplateVars(stripped)
			hasTopHeading := false
			for _, line := range strings.Split(stripped, "\n") {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "<h1") || strings.HasPrefix(trimmed, "<h2") {
					hasTopHeading = true
					break
				}
			}
			assert.True(t, hasTopHeading,
				"template %s must have a top-level '# Title' or '<h1>/<h2>' heading", entry.Name())
		})
	}
}

// stripCodeBlocks removes all fenced code blocks (```...```) from the
// template content. Code blocks are structural and exempt from the
// i18n check.
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
