package compiler

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

//go:embed templates/dnd-style.css
var dndCSS string

//go:embed templates/areas.md.tmpl
var areasTemplate string

//go:embed templates/npc.md.tmpl
var npcTemplate string

//go:embed templates/monster.md.tmpl
var monsterTemplate string

//go:embed templates/encounter.md.tmpl
var encounterTemplate string

//go:embed templates/map.md.tmpl
var mapTemplate string

//go:embed templates/lore.md.tmpl
var loreTemplate string

//go:embed templates/session-zero.md.tmpl
var sessionZeroTemplate string

//go:embed templates/introduction.md.tmpl
var introductionTemplate string

//go:embed templates/setting-guide.md.tmpl
var settingGuideTemplate string

//go:embed templates/appendices.md.tmpl
var appendicesTemplate string

//go:embed templates/session-prep.md.tmpl
var sessionPrepTemplate string

//go:embed templates/character-sheet.md.tmpl
var characterSheetTemplate string

type Compiler struct {
	CampaignDir         string
	PDFEngine           string
	CompilerVersion     int
	seenImages          map[string]bool
	handoutRendererImpl HandoutRenderer
}

func New(campaignDir, pdfEngine string) *Compiler {
	if pdfEngine == "" {
		pdfEngine = "wkhtmltopdf"
	}
	return &Compiler{
		CampaignDir:     campaignDir,
		PDFEngine:       pdfEngine,
		CompilerVersion: 2,
		seenImages:      make(map[string]bool),
	}
}

// NewWithVersion creates a compiler with a specific version (1 or 2)
func NewWithVersion(campaignDir, pdfEngine string, version int) *Compiler {
	c := New(campaignDir, pdfEngine)
	if version == 1 || version == 2 {
		c.CompilerVersion = version
	}
	return c
}

// SetHandoutRenderer sets the handout renderer for v2 compilation
func (c *Compiler) SetHandoutRenderer(renderer HandoutRenderer) {
	c.handoutRendererImpl = renderer
}

func GetTemplate(tmplType string) (string, error) {
	switch tmplType {
	case "areas":
		return areasTemplate, nil
	case "npc":
		return npcTemplate, nil
	case "monster":
		return monsterTemplate, nil
	case "encounter":
		return encounterTemplate, nil
	case "map":
		return mapTemplate, nil
	case "lore":
		return loreTemplate, nil
	case "session-zero":
		return sessionZeroTemplate, nil
	case "introduction":
		return introductionTemplate, nil
	case "setting-guide":
		return settingGuideTemplate, nil
	case "appendices":
		return appendicesTemplate, nil
	case "session-prep":
		return sessionPrepTemplate, nil
	case "character-sheet":
		return characterSheetTemplate, nil
	case "dnd-style":
		return dndCSS, nil
	default:
		return "", fmt.Errorf("unknown template type: %s", tmplType)
	}
}

func (c *Compiler) Compile(ctx context.Context, title string) (string, error) {
	htmlParts, err := c.generateHTML(title)
	if err != nil {
		return "", err
	}

	htmlPath := filepath.Join(c.CampaignDir, "campaign.html")
	if err := os.WriteFile(htmlPath, []byte(strings.Join(htmlParts, "\n")), 0644); err != nil {
		return "", fmt.Errorf("failed to write HTML: %w", err)
	}

	pdfPath := filepath.Join(c.CampaignDir, "campaign.pdf")

	// Retry loop: verify images after PDF generation and retry if missing
	maxRetries := 3
	var lastExpected, lastFound int
	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := c.htmlToPDF(ctx, htmlPath, pdfPath); err != nil {
			return "", fmt.Errorf("failed to convert to PDF (attempt %d): %w", attempt+1, err)
		}

		// Verify images are present in the generated HTML
		expected, found, ok, err := c.verifyImages(htmlPath)
		if err != nil {
			return "", fmt.Errorf("failed to verify images (attempt %d): %w", attempt+1, err)
		}
		lastExpected = expected
		lastFound = found
		if ok {
			return pdfPath, nil
		}

		if attempt < maxRetries-1 {
			// Regenerate HTML (images might need re-embedding)
			// Clear seen images to allow re-embedding
			c.seenImages = make(map[string]bool)
			// Re-read source files and regenerate HTML
			htmlParts, err = c.generateHTML(title)
			if err != nil {
				return "", fmt.Errorf("failed to regenerate HTML (attempt %d): %w", attempt+1, err)
			}
			if err := os.WriteFile(htmlPath, []byte(strings.Join(htmlParts, "\n")), 0644); err != nil {
				return "", fmt.Errorf("failed to write regenerated HTML (attempt %d): %w", attempt+1, err)
			}
		}
	}

	return "", fmt.Errorf("PDF generation failed after %d attempts: images missing (expected %d, found %d)", maxRetries, lastExpected, lastFound)
}

// generateHTML reads all markdown sources and generates the full HTML document.
func (c *Compiler) generateHTML(title string) ([]string, error) {
	sections := []struct {
		name  string
		path  string
		isDir bool
	}{
		{"Introduction", filepath.Join(c.CampaignDir, "introduction.md"), false},
		{"Lore y Ambientación", filepath.Join(c.CampaignDir, "lore.md"), false},
		{"Chapters (Areas)", filepath.Join(c.CampaignDir, "areas"), true},
		{"Setting Guide", filepath.Join(c.CampaignDir, "setting-guide.md"), false},
		{"Apéndice A: NPCs y Facciones", filepath.Join(c.CampaignDir, "npcs"), true},
		{"Apéndice B: Bestiario", filepath.Join(c.CampaignDir, "bestiary"), true},
		{"Apéndice C: Encuentros", filepath.Join(c.CampaignDir, "encounters"), true},
		{"Apéndice D: Mapas de Referencia", filepath.Join(c.CampaignDir, "maps"), true},
		{"Appendices", filepath.Join(c.CampaignDir, "appendices.md"), false},
		{"Apéndice G: Character Sheets", filepath.Join(c.CampaignDir, "characters"), true},
		{"Apéndice H: Quests", filepath.Join(c.CampaignDir, "quests"), true},
	}

	var htmlParts []string
	headingCounter := 0

	htmlParts = append(htmlParts, "<!DOCTYPE html>")
	htmlParts = append(htmlParts, "<html><head><meta charset='utf-8'>")
	htmlParts = append(htmlParts, "<style>"+dndCSS+"</style>")
	htmlParts = append(htmlParts, "</head><body>")

	// Cover page — single page, no split
	htmlParts = append(htmlParts, fmt.Sprintf(`<div class="cover-wrapper" id="cover"><h1>%s</h1><p class="subtitle">Generated by Grimorio</p>`, html.EscapeString(title)))

	// Insert cover art if exists — search for cover.* or cover-*.*
	coverPath := findCoverImage(c.CampaignDir)
	if coverPath != "" {
		htmlParts = append(htmlParts, fmt.Sprintf(`<div class="cover-image">%s</div>`, embedImage(coverPath, "Cover Art", c.CampaignDir, c.seenImages)))
	}

	htmlParts = append(htmlParts, `</div>`)

	// Session Zero guidance (if available)
	sessionZeroHTML := c.generateSessionZero()
	if sessionZeroHTML != "" {
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
		htmlParts = append(htmlParts, sessionZeroHTML)
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
	}

	// Flowchart SVG (if available)
	flowchartHTML := c.generateFlowchartEmbed()
	if flowchartHTML != "" {
		htmlParts = append(htmlParts, flowchartHTML)
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
	}

	// TOC
	htmlParts = append(htmlParts, c.generateTOC(sections))
	htmlParts = append(htmlParts, `<div class="section-break"></div>`)

	for _, sec := range sections {
		if _, err := os.Stat(sec.path); err != nil {
			continue
		}

		if sec.isDir {
			files, err := os.ReadDir(sec.path)
			if err != nil {
				continue
			}
			hasContent := false
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".md") {
					content, err := os.ReadFile(filepath.Join(sec.path, f.Name()))
					if err != nil {
						continue
					}
					sectionID := "sec-" + sanitizeID(sec.name+"-"+f.Name())
					htmlResult := c.markdownToHTMLWithID(string(content), c.CampaignDir, sectionID, &headingCounter, c.seenImages)
					htmlResult = postProcessHTML(htmlResult, c.CompilerVersion)
					if strings.TrimSpace(htmlResult) != "" {
						htmlParts = append(htmlParts, htmlResult)
						hasContent = true
					}
				}
			}
			if hasContent {
				htmlParts = append(htmlParts, `<div class="section-break"></div>`)
			}
		} else {
			content, err := os.ReadFile(sec.path)
			if err != nil {
				continue
			}
			sectionID := "sec-" + sanitizeID(sec.name)
			htmlResult := c.markdownToHTMLWithID(string(content), c.CampaignDir, sectionID, &headingCounter, c.seenImages)
			htmlResult = postProcessHTML(htmlResult, c.CompilerVersion)
			if strings.TrimSpace(htmlResult) != "" {
				htmlParts = append(htmlParts, htmlResult)
			}
			htmlParts = append(htmlParts, `<div class="section-break"></div>`)
		}
	}

	// Apéndice E: Faction Tracker
	trackerHTML := c.generateFactionTracker()
	if trackerHTML != "" {
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
		htmlParts = append(htmlParts, trackerHTML)
	}

	// Apéndice F: Adventure Roster
	rosterHTML := c.generateAdventureRoster()
	if rosterHTML != "" {
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
		htmlParts = append(htmlParts, rosterHTML)
	}

	// Handout pages (v2 only)
	if c.CompilerVersion == 2 {
		handoutHTML, err := c.generateHandouts()
		if err == nil && handoutHTML != "" {
			htmlParts = append(htmlParts, `<div class="section-break"></div>`)
			htmlParts = append(htmlParts, handoutHTML)
		}
	}

	htmlParts = append(htmlParts, "</body></html>")
	return htmlParts, nil
}

// generateTOC creates a hierarchical table of contents
func (c *Compiler) generateTOC(sections []struct {
	name  string
	path  string
	isDir bool
}) string {
	var b strings.Builder
	b.WriteString(`<div class="toc" id="toc"><h2>Table of Contents</h2><ul>`)

	for _, sec := range sections {
		if _, err := os.Stat(sec.path); err != nil {
			continue
		}

		id := "sec-" + sanitizeID(sec.name)
		fmt.Fprintf(&b, `<li><a href="#%s">%s</a><span class="page-ref"></span></li>`, id, html.EscapeString(sec.name))

		// In v2, extract areas from act files for hierarchical TOC
		if c.CompilerVersion == 2 && sec.isDir && (strings.Contains(strings.ToLower(sec.name), "chapter") || strings.Contains(strings.ToLower(sec.name), "area")) {
			areas := c.extractAreasFromDir(sec.path)
			if len(areas) > 0 {
				b.WriteString(`<ul class="toc-areas">`)
				for _, area := range areas {
					fmt.Fprintf(&b, `<li><a href="#%s">%s</a><span class="page-ref"></span></li>`, area.ID, html.EscapeString(area.Name))
				}
				b.WriteString(`</ul>`)
			}
		}
	}

	b.WriteString(`</ul></div>`)
	return b.String()
}

type tocArea struct {
	ID   string
	Name string
}

func (c *Compiler) extractAreasFromDir(dirPath string) []tocArea {
	var areas []tocArea
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return areas
	}

	areaPattern := regexp.MustCompile(`(?m)^#{3,4}\s+[Áa]rea\s+(\d+)(?:\s*:\s*(.+))?$`)

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dirPath, f.Name()))
		if err != nil {
			continue
		}

		matches := areaPattern.FindAllStringSubmatch(string(content), -1)
		for _, m := range matches {
			num := m[1]
			name := "Área " + num
			if len(m) > 2 && m[2] != "" {
				name = name + ": " + strings.TrimSpace(m[2])
			}
			areas = append(areas, tocArea{
				ID:   "area-" + num,
				Name: name,
			})
		}
	}

	return areas
}

// generateFactionTracker reads the reputation matrix and generates an HTML appendix.
func (c *Compiler) generateFactionTracker() string {
	matrixPath := filepath.Join(c.CampaignDir, "factions", "reputation_matrix.json")
	data, err := os.ReadFile(matrixPath)
	if err != nil {
		return "" // no factions, no tracker
	}

	var matrix struct {
		CampaignID string `json:"campaign_id"`
		Entries    []struct {
			FactionID string `json:"faction_id"`
			PartyID   string `json:"party_id"`
			Score     int8   `json:"score"`
			Status    string `json:"status"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(data, &matrix); err != nil {
		return ""
	}
	if len(matrix.Entries) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`<h2 id="sec-faction-tracker">Apéndice E: Faction Tracker</h2>`)
	b.WriteString(`<table><thead><tr><th>Faction</th><th>Party</th><th>Score</th><th>Status</th></tr></thead><tbody>`)
	for _, e := range matrix.Entries {
		statusClass := "status-" + strings.ToLower(e.Status)
		fmt.Fprintf(&b, `<tr class="%s"><td>%s</td><td>%s</td><td>%d</td><td>%s</td></tr>`,
			statusClass, html.EscapeString(e.FactionID), html.EscapeString(e.PartyID), e.Score, html.EscapeString(e.Status))
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

// generateSessionZero reads session-zero.md and converts to HTML
func (c *Compiler) generateSessionZero() string {
	path := filepath.Join(c.CampaignDir, "session-zero.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "" // no session zero, skip
	}

	htmlResult := markdownToHTMLWithID(string(data), c.CampaignDir, "sec-session-zero", new(int), c.seenImages, c.CompilerVersion)
	if strings.TrimSpace(htmlResult) == "" {
		return ""
	}

	return htmlResult // markdown already contains the heading with proper id
}

// generateFlowchartEmbed embeds the flowchart SVG if available
func (c *Compiler) generateFlowchartEmbed() string {
	path := filepath.Join(c.CampaignDir, "flowchart.svg")
	data, err := os.ReadFile(path)
	if err != nil {
		return "" // no flowchart, skip
	}

	return fmt.Sprintf(`<h2 id="sec-flowchart">Campaign Flowchart</h2><div class="flowchart">%s</div>`, string(data))
}

// generateDMSidebar generates a DM-only sidebar with tips and secrets for an area.
// nolint:unused // reserved for future use
func (c *Compiler) generateDMSidebar(areaID string, tip string, secret string) string {
	if tip == "" && secret == "" {
		return ""
	}

	var b strings.Builder
	b.WriteString(`<div class="dm-sidebar">`)

	if tip != "" {
		b.WriteString(`<h5>DM Tip</h5>`)
		fmt.Fprintf(&b, `<p>%s</p>`, html.EscapeString(tip))
	}

	if secret != "" {
		b.WriteString(`<h5>Secreto</h5>`)
		fmt.Fprintf(&b, `<p>%s</p>`, html.EscapeString(secret))
	}

	b.WriteString(`</div>`)
	return b.String()
}

// generateSessionPrepHTML generates HTML for session preparation content.
// nolint:unused // reserved for future use
func (c *Compiler) generateSessionPrepHTML(sessionNum int) string {
	// Look for session prep markdown file
	path := filepath.Join(c.CampaignDir, "session-prep.md")
	data, err := os.ReadFile(path)
	if err != nil {
		// Try numbered session prep file
		path = filepath.Join(c.CampaignDir, fmt.Sprintf("session-prep-%d.md", sessionNum))
		data, err = os.ReadFile(path)
		if err != nil {
			return ""
		}
	}

	htmlResult := markdownToHTMLWithID(string(data), c.CampaignDir, fmt.Sprintf("sec-session-prep-%d", sessionNum), new(int), c.seenImages, c.CompilerVersion)
	if strings.TrimSpace(htmlResult) == "" {
		return ""
	}

	return htmlResult
}

// generateCharacterSheetHTML generates HTML for a character sheet.
// nolint:unused // reserved for future use
func (c *Compiler) generateCharacterSheetHTML(characterID string) string {
	// Look for character sheet markdown file
	path := filepath.Join(c.CampaignDir, "characters", fmt.Sprintf("%s.md", characterID))
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	htmlResult := markdownToHTMLWithID(string(data), c.CampaignDir, fmt.Sprintf("sec-character-%s", characterID), new(int), c.seenImages, c.CompilerVersion)
	if strings.TrimSpace(htmlResult) == "" {
		return ""
	}

	return htmlResult
}

// generateShockPointsHTML generates HTML for shock points content warnings.
// nolint:unused // reserved for future use
func (c *Compiler) generateShockPointsHTML(shockPoints []struct {
	Type        string
	Severity    string
	Description string
	SafetyTools []string
}) string {
	if len(shockPoints) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`<h3 id="sec-shock-points">Puntos de Shock y Advertencias de Contenido</h3>`)

	for _, sp := range shockPoints {
		fmt.Fprintf(&b, `<div class="shock-point %s">`, sp.Severity)
		fmt.Fprintf(&b, `<span class="severity-badge">%s</span>`, sp.Severity)
		fmt.Fprintf(&b, `<strong>%s</strong>: %s`, html.EscapeString(sp.Type), html.EscapeString(sp.Description))

		if len(sp.SafetyTools) > 0 {
			b.WriteString(`<p><strong>Herramientas de seguridad:</strong> `)
			for i, tool := range sp.SafetyTools {
				if i > 0 {
					b.WriteString(", ")
				}
				b.WriteString(html.EscapeString(tool))
			}
			b.WriteString(`</p>`)
		}

		b.WriteString(`</div>`)
	}

	return b.String()
}

// generateAdventureRoster builds Apéndice F from scanned markdown files
func (c *Compiler) generateAdventureRoster() string {
	var npcs, monsters, encounters []string

	// Scan acts for entities
	actsDir := filepath.Join(c.CampaignDir, "areas")
	if info, err := os.Stat(actsDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(actsDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, _ := os.ReadFile(filepath.Join(actsDir, f.Name()))
				n, m, e := extractRosterEntries(string(content))
				npcs = append(npcs, n...)
				monsters = append(monsters, m...)
				encounters = append(encounters, e...)
			}
		}
	}

	// Scan npcs dir
	npcsDir := filepath.Join(c.CampaignDir, "npcs")
	if info, err := os.Stat(npcsDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(npcsDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, _ := os.ReadFile(filepath.Join(npcsDir, f.Name()))
				n, _, _ := extractRosterEntries(string(content))
				npcs = append(npcs, n...)
			}
		}
	}

	// Scan bestiary
	bestiaryDir := filepath.Join(c.CampaignDir, "bestiary")
	if info, err := os.Stat(bestiaryDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(bestiaryDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, _ := os.ReadFile(filepath.Join(bestiaryDir, f.Name()))
				_, m, _ := extractRosterEntries(string(content))
				monsters = append(monsters, m...)
			}
		}
	}

	// Scan encounters
	encountersDir := filepath.Join(c.CampaignDir, "encounters")
	if info, err := os.Stat(encountersDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(encountersDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, _ := os.ReadFile(filepath.Join(encountersDir, f.Name()))
				_, _, e := extractRosterEntries(string(content))
				encounters = append(encounters, e...)
			}
		}
	}

	if len(npcs) == 0 && len(monsters) == 0 && len(encounters) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`<h2 id="sec-adventure-roster">Apéndice F: Adventure Roster</h2>`)

	if len(npcs) > 0 {
		b.WriteString(`<h3>NPCs</h3><table><thead><tr><th>Nombre</th><th>Rol</th></tr></thead><tbody>`)
		for _, n := range npcs {
			parts := strings.SplitN(n, "|", 2)
			name := parts[0]
			role := ""
			if len(parts) > 1 {
				role = parts[1]
			}
			fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td></tr>`, html.EscapeString(name), html.EscapeString(role))
		}
		b.WriteString(`</tbody></table>`)
	}

	if len(monsters) > 0 {
		b.WriteString(`<h3>Monstruos</h3><table><thead><tr><th>Nombre</th><th>CR</th></tr></thead><tbody>`)
		for _, m := range monsters {
			parts := strings.SplitN(m, "|", 2)
			name := parts[0]
			cr := ""
			if len(parts) > 1 {
				cr = parts[1]
			}
			fmt.Fprintf(&b, `<tr><td>%s</td><td>%s</td></tr>`, html.EscapeString(name), html.EscapeString(cr))
		}
		b.WriteString(`</tbody></table>`)
	}

	if len(encounters) > 0 {
		b.WriteString(`<h3>Encuentros</h3><table><thead><tr><th>Nombre</th></tr></thead><tbody>`)
		for _, e := range encounters {
			fmt.Fprintf(&b, `<tr><td>%s</td></tr>`, html.EscapeString(e))
		}
		b.WriteString(`</tbody></table>`)
	}

	return b.String()
}

// extractRosterEntries parses markdown for roster entities
func extractRosterEntries(md string) (npcs, monsters, encounters []string) {
	lines := strings.Split(md, "\n")
	var inNPCs, inMonsters, inEncounters bool
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "## ") || strings.HasPrefix(lower, "### ") {
			inNPCs = strings.Contains(lower, "npc") || strings.Contains(lower, "personaje")
			inMonsters = strings.Contains(lower, "monstruo") || strings.Contains(lower, "criatura") || strings.Contains(lower, "bestia")
			inEncounters = strings.Contains(lower, "encuentro")
			continue
		}

		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			item := strings.TrimPrefix(trimmed, "- ")
			if strings.HasPrefix(trimmed, "* ") {
				item = strings.TrimPrefix(trimmed, "* ")
			}
			item = strings.TrimSpace(item)
			// Extract bold name
			if m := boldRegex.FindStringSubmatch(item); m != nil {
				name := m[1]
				rest := strings.TrimSpace(strings.TrimPrefix(item, m[0]))
				rest = strings.TrimPrefix(rest, "—")
				rest = strings.TrimPrefix(rest, "-")
				rest = strings.TrimSpace(rest)
				if inNPCs {
					npcs = append(npcs, name+"|"+rest)
				} else if inMonsters {
					monsters = append(monsters, name+"|"+rest)
				} else if inEncounters {
					encounters = append(encounters, name)
				}
			} else if inEncounters && item != "" {
				encounters = append(encounters, item)
			}
		}
	}
	return
}

// verifyImages compares the number of expected images in markdown sources
// with the number of <img> tags in the generated HTML.
// Returns (expected, found, ok, error).
func (c *Compiler) verifyImages(htmlPath string) (int, int, bool, error) {
	expected, err := c.countImagesInMarkdownSources()
	if err != nil {
		return 0, 0, false, err
	}

	found, err := countImagesInHTML(htmlPath)
	if err != nil {
		return expected, 0, false, err
	}

	// ok if found >= expected (some images might be deduplicated)
	return expected, found, found >= expected, nil
}

// countImagesInMarkdownSources walks all markdown files in the campaign directory
// and counts image references: markdown ![alt](path), <img> tags, and `assets/file` refs.
func (c *Compiler) countImagesInMarkdownSources() (int, error) {
	count := 0

	sections := []string{
		filepath.Join(c.CampaignDir, "lore.md"),
		filepath.Join(c.CampaignDir, "areas"),
		filepath.Join(c.CampaignDir, "npcs"),
		filepath.Join(c.CampaignDir, "bestiary"),
		filepath.Join(c.CampaignDir, "encounters"),
		filepath.Join(c.CampaignDir, "maps"),
	}

	for _, secPath := range sections {
		info, err := os.Stat(secPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			files, err := os.ReadDir(secPath)
			if err != nil {
				continue
			}
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".md") {
					content, err := os.ReadFile(filepath.Join(secPath, f.Name()))
					if err != nil {
						continue
					}
					count += countImagesInMarkdown(string(content))
				}
			}
		} else {
			content, err := os.ReadFile(secPath)
			if err != nil {
				continue
			}
			count += countImagesInMarkdown(string(content))
		}
	}

	return count, nil
}

// countImagesInMarkdown counts image references in a single markdown string.
func countImagesInMarkdown(md string) int {
	count := 0
	// Count markdown images: ![alt](path)
	count += len(imageRegex.FindAllString(md, -1))
	// Count raw <img> tags
	count += len(imgTagRegex.FindAllString(md, -1))
	// Count code asset refs: `assets/filename.ext`
	count += len(codeAssetRegex.FindAllString(md, -1))
	return count
}

// countImagesInHTML counts <img> tags in the generated HTML file.
func countImagesInHTML(htmlPath string) (int, error) {
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		return 0, err
	}
	return len(imgTagRegex.FindAllString(string(content), -1)), nil
}

func sanitizeID(s string) string {
	result := []rune(s)
	for i, r := range result {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			result[i] = '-'
		}
	}
	return strings.ToLower(string(result))
}

func (c *Compiler) htmlToPDF(ctx context.Context, htmlPath, pdfPath string) error {
	cmd := exec.CommandContext(ctx, c.PDFEngine,
		"--enable-local-file-access",
		"--page-size", "A4",
		"--margin-top", "15mm",
		"--margin-bottom", "15mm",
		"--margin-left", "15mm",
		"--margin-right", "15mm",
		"--print-media-type",
		"--footer-center", "[page]/[topage]",
		"--footer-font-size", "8",
		htmlPath,
		pdfPath,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wkhtmltopdf failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

var (
	boldRegex         = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex       = regexp.MustCompile(`\*(.+?)\*`)
	boldAdjacentRegex = regexp.MustCompile(`</strong>([^\s<:;,\.!?])`) // Exclude punctuation to prevent &thinsp; leak
	htmlCommentRegex  = regexp.MustCompile(`<!--[\s\S]*?-->`)
	imageRegex        = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	imgTagRegex       = regexp.MustCompile(`<img[^>]*(?:/\s*)?>`)
	htmlBlockRegex    = regexp.MustCompile(`(?s)<div[^>]*>.*?</div>`)

	// readAloudPrefixRe strips **Read-Aloud:** or **Para Leer en Voz Alta:** labels from blockquote text
	// (CSS .read-aloud::before pseudo-element provides the visual label, so the inline text would duplicate it)
	readAloudPrefixRe = regexp.MustCompile(`^\*{2}(?:Read-Aloud|Para Leer en Voz Alta)\*{2}:\s*`)

	// rawTagRe strips raw HTML tags (e.g. <div>, <span>) — <img> tags are preserved via stash/restore
	rawTagRe = regexp.MustCompile(`<[^>]+>`)

	blockquoteRe   = regexp.MustCompile(`^>\s*(.*)`)
	codeAssetRegex = regexp.MustCompile("`assets/([\\w\\-]+\\.(svg|png|jpg|jpeg|gif|webp))`")
	sceneRegex     = regexp.MustCompile(`\[SCENE:\s*(.*?)\]`)

	// v2 patterns
	areaHeadingPattern     = regexp.MustCompile(`^[Áa]rea\s+(\d+):\s*(.+)$`)
	areaRefPattern         = regexp.MustCompile(`(?i)\b[Áa]rea\s+(\d+)\b`)
)

// formatInline processes bold and italic markers, ensuring no word merging after </strong>
func formatInline(text string) string {
	text = boldRegex.ReplaceAllString(text, "<strong>$1</strong>")
	// Remove &thinsp; - it causes spacing issues like "sol , y" and "volvióazul"
	text = boldAdjacentRegex.ReplaceAllString(text, "</strong>$1")
	return italicRegex.ReplaceAllString(text, "<em>$1</em>")
}

// processInlineText processes markdown images, escapes HTML, formats inline,
// and preserves existing <img> tags so they are not escaped.
func processInlineText(text string, baseDir string, seenImages map[string]bool) string {
	// 1. Process markdown images first
	text = processImages(text, baseDir, seenImages)

	// 2. Extract and preserve existing <img> tags
	var imgTags []string
	text = imgTagRegex.ReplaceAllStringFunc(text, func(match string) string {
		imgTags = append(imgTags, match)
		return "\x00IMG\x00"
	})

	// 3. Escape HTML
	text = html.EscapeString(text)

	// 4. Format inline (bold, italic)
	text = formatInline(text)

	// 5. Restore <img> tags
	for _, img := range imgTags {
		text = strings.Replace(text, "\x00IMG\x00", img, 1)
	}

	return text
}

func (c *Compiler) markdownToHTMLWithID(md string, baseDir string, sectionID string, headingCounter *int, seenImages map[string]bool) string {
	return markdownToHTMLWithID(md, baseDir, sectionID, headingCounter, seenImages, c.CompilerVersion)
}

func markdownToHTMLWithID(md string, baseDir string, sectionID string, headingCounter *int, seenImages map[string]bool, compilerVersion int) string {
	// Strip HTML comments before processing to prevent artifacts in PDF
	md = htmlCommentRegex.ReplaceAllString(md, "")

	// Extract HTML blocks (like <div>...</div>) before processing to preserve them
	var htmlBlocks []string
	md = htmlBlockRegex.ReplaceAllStringFunc(md, func(match string) string {
		htmlBlocks = append(htmlBlocks, match)
		return fmt.Sprintf("\x00HTMLBLOCK%d\x00", len(htmlBlocks)-1)
	})

	// Strip raw HTML tags (except <img>) that would be escaped and rendered as visible text.
	// Stash <img> tags, strip all remaining HTML tags, then restore <img> tags.
	imgPlaceholder := "__IMG_TAG_PLACEHOLDER__"
	imgStash := imgTagRegex.FindAllString(md, -1)
	md = imgTagRegex.ReplaceAllString(md, imgPlaceholder)
	md = rawTagRe.ReplaceAllString(md, "")
	for _, imgTag := range imgStash {
		md = strings.Replace(md, imgPlaceholder, imgTag, 1)
	}

	lines := strings.Split(md, "\n")
	var out []string
	inList := false
	var paragraphLines []string

	// Table state
	inTable := false
	var tableRows []string
	var pendingTableRow string // buffered row waiting for separator confirmation

	// Blockquote state (read-aloud)
	inBlockquote := false
	var blockquoteLines []string

	// Code block state
	inCodeBlock := false
	var codeBlockLines []string

	flushCodeBlock := func() {
		if len(codeBlockLines) == 0 {
			return
		}
		code := strings.Join(codeBlockLines, "\n")
		codeBlockLines = nil
		if code == "" {
			return
		}
		escaped := html.EscapeString(code)
		out = append(out, fmt.Sprintf(`<pre class="code-block"><code>%s</code></pre>`, escaped))
	}

	flushBlockquote := func() {
		if len(blockquoteLines) == 0 {
			return
		}
		text := strings.Join(blockquoteLines, " ")
		blockquoteLines = nil
		if text == "" {
			return
		}
		// Strip **Read-Aloud:** or **Para Leer en Voz Alta:** label (CSS ::before provides it)
		text = readAloudPrefixRe.ReplaceAllString(text, "")
		escaped := processInlineText(text, baseDir, seenImages)
		out = append(out, fmt.Sprintf(`<div class="read-aloud">%s</div>`, escaped))
	}

	flushParagraph := func() {
		if len(paragraphLines) == 0 {
			return
		}
		text := strings.Join(paragraphLines, " ")
		paragraphLines = nil
		if text == "" {
			return
		}
		escaped := processInlineText(text, baseDir, seenImages)
		out = append(out, fmt.Sprintf("<p>%s</p>", escaped))
	}

	flushTable := func() {
		if len(tableRows) < 2 {
			// Not a valid table, treat rows as paragraphs
			for _, row := range tableRows {
				if strings.TrimSpace(row) != "" {
					escaped := processInlineText(strings.TrimSpace(row), baseDir, seenImages)
					out = append(out, fmt.Sprintf("<p>%s</p>", escaped))
				}
			}
			tableRows = nil
			return
		}
		headers := parseTableRow(tableRows[0])
		alignments := parseTableAlign(tableRows[1])
		var htmlOut strings.Builder
		htmlOut.WriteString(`<table><thead><tr>`)
		for i, h := range headers {
			align := ""
			if i < len(alignments) {
				align = alignments[i]
			}
			if align != "" {
				fmt.Fprintf(&htmlOut, `<th style="text-align:%s">%s</th>`, align, h)
			} else {
				fmt.Fprintf(&htmlOut, `<th>%s</th>`, h)
			}
		}
		htmlOut.WriteString(`</tr></thead><tbody>`)
		for _, row := range tableRows[2:] {
			cells := parseTableRow(row)
			htmlOut.WriteString(`<tr>`)
			for i, cell := range cells {
				align := ""
				if i < len(alignments) {
					align = alignments[i]
				}
				cellEsc := processInlineText(cell, baseDir, seenImages)
				if align != "" {
					fmt.Fprintf(&htmlOut, `<td style="text-align:%s">%s</td>`, align, cellEsc)
				} else {
					fmt.Fprintf(&htmlOut, `<td>%s</td>`, cellEsc)
				}
			}
			htmlOut.WriteString(`</tr>`)
		}
		htmlOut.WriteString(`</tbody></table>`)
		out = append(out, htmlOut.String())
		tableRows = nil
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip horizontal rules (---, ***, ___, - - -)
		if trimmed == "---" || trimmed == "***" || trimmed == "___" || trimmed == "- - -" {
			continue
		}

		// Skip empty lines inside table
		if inTable && trimmed == "" {
			continue
		}

		// Handle scene placeholders [SCENE: description]
		if sceneMatch := sceneRegex.FindStringSubmatch(trimmed); sceneMatch != nil {
			flushParagraph()
			flushBlockquote()
			flushTable()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			sceneDesc := sceneMatch[1]
			escaped := html.EscapeString(sceneDesc)
			escaped = formatInline(escaped)
			out = append(out, fmt.Sprintf(`<div class="scene-description">🎭 %s</div>`, escaped))
			continue
		}

		// If we have a pending row and current line is a separator, start table
		if pendingTableRow != "" && isTableSeparator(trimmed) {
			inTable = true
			tableRows = append(tableRows, pendingTableRow)
			tableRows = append(tableRows, trimmed)
			pendingTableRow = ""
			continue
		}

		// If we have a pending row but current line is NOT a separator, flush pending as paragraph
		if pendingTableRow != "" && !isTableRow(trimmed) {
			paragraphLines = append(paragraphLines, pendingTableRow)
			pendingTableRow = ""
		}

		// If we're in a table and see another row, add it
		if inTable && isTableRow(trimmed) {
			tableRows = append(tableRows, trimmed)
			continue
		}

		// If we were in a table but now see non-table content, flush table
		if inTable {
			flushTable()
			inTable = false
		}

		// If line looks like a table row, buffer it (might be table start)
		if isTableRow(trimmed) && !inTable {
			pendingTableRow = trimmed
			continue
		}

		// Handle code blocks (```)
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				// End code block
				flushCodeBlock()
				inCodeBlock = false
			} else {
				// Start code block
				flushParagraph()
				flushBlockquote()
				if inList {
					out = append(out, "</ul>")
					inList = false
				}
				inCodeBlock = true
			}
			continue
		}

		// If we're inside a code block, collect lines
		if inCodeBlock {
			codeBlockLines = append(codeBlockLines, line) // Keep original line, not trimmed
			continue
		}

		// Handle blockquotes (read-aloud text)
		if bqMatch := blockquoteRe.FindStringSubmatch(trimmed); bqMatch != nil {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			inBlockquote = true
			blockquoteLines = append(blockquoteLines, bqMatch[1])
			continue
		}

		// If we were in a blockquote but now see non-blockquote content, flush it
		if inBlockquote {
			flushBlockquote()
			inBlockquote = false
		}

		// Handle headings
		if strings.HasPrefix(trimmed, "##### ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "##### ")
			*headingCounter++
			id := sectionID + "-h" + strconv.Itoa(*headingCounter)
			out = append(out, fmt.Sprintf(`<h5 id="%s">%s</h5>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "#### ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "#### ")
			*headingCounter++
			id := sectionID + "-h" + strconv.Itoa(*headingCounter)
			out = append(out, fmt.Sprintf(`<h4 id="%s">%s</h4>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "### ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "### ")
			*headingCounter++
			id := sectionID + "-h" + strconv.Itoa(*headingCounter)

			// v2: Area number highlighting and cross-reference IDs
			if compilerVersion == 2 {
				if areaMatch := areaHeadingPattern.FindStringSubmatch(text); areaMatch != nil {
					areaNum := areaMatch[1]
					areaName := strings.TrimSpace(areaMatch[2])
					areaID := "area-" + areaNum
					rendered := fmt.Sprintf(`<span class="area-number">Área %s</span> %s`, areaNum, html.EscapeString(areaName))
					out = append(out, fmt.Sprintf(`<h3 id="%s">%s</h3>`, areaID, rendered))
					continue
				}
			}

			out = append(out, fmt.Sprintf(`<h3 id="%s">%s</h3>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "## ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "## ")
			*headingCounter++
			id := sectionID + "-h" + strconv.Itoa(*headingCounter)
			out = append(out, fmt.Sprintf(`<h2 id="%s">%s</h2>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "# ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "# ")
			*headingCounter++
			id := sectionID + "-h" + strconv.Itoa(*headingCounter)
			out = append(out, fmt.Sprintf(`<h1 id="%s">%s</h1>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			flushParagraph()
			if !inList {
				out = append(out, "<ul>")
				inList = true
			}
			text := strings.TrimPrefix(trimmed, "- ")
			if strings.HasPrefix(trimmed, "* ") {
				text = strings.TrimPrefix(trimmed, "* ")
			}
			escaped := processInlineText(text, baseDir, seenImages)
			out = append(out, fmt.Sprintf("<li>%s</li>", escaped))
		} else if trimmed == "" {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
		} else if imageRegex.MatchString(trimmed) {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			out = append(out, processImages(trimmed, baseDir, seenImages))
		} else {
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			paragraphLines = append(paragraphLines, trimmed)
		}
	}

	flushParagraph()
	flushTable()
	flushBlockquote()
	flushCodeBlock()
	if inList {
		out = append(out, "</ul>")
	}

	// Restore HTML blocks
	result := strings.Join(out, "\n")
	for i, html := range htmlBlocks {
		result = strings.Replace(result, fmt.Sprintf("\x00HTMLBLOCK%d\x00", i), html, 1)
	}

	return result
}

func isTableRow(line string) bool {
	return strings.HasPrefix(line, "|") && strings.HasSuffix(line, "|") && strings.Count(line, "|") >= 3
}

func isTableSeparator(line string) bool {
	if !isTableRow(line) {
		return false
	}
	inner := strings.Trim(line, "|")
	parts := strings.Split(inner, "|")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if p[0] != '-' && p[0] != ':' {
			return false
		}
	}
	return true
}

func parseTableRow(row string) []string {
	row = strings.TrimSpace(row)
	row = strings.TrimPrefix(row, "|")
	row = strings.TrimSuffix(row, "|")
	parts := strings.Split(row, "|")
	var result []string
	for _, p := range parts {
		result = append(result, strings.TrimSpace(p))
	}
	return result
}

func parseTableAlign(row string) []string {
	cells := parseTableRow(row)
	var alignments []string
	for _, cell := range cells {
		cell = strings.TrimSpace(cell)
		leftDash := strings.HasPrefix(cell, ":")
		rightDash := strings.HasSuffix(cell, ":")
		if leftDash && rightDash {
			alignments = append(alignments, "center")
		} else if rightDash {
			alignments = append(alignments, "right")
		} else {
			alignments = append(alignments, "left")
		}
	}
	return alignments
}

func markdownToHTML(md string, baseDir string) string {
	counter := 0
	seen := make(map[string]bool)
	return markdownToHTMLWithID(md, baseDir, "content", &counter, seen, 0)
}

// postProcessHTML applies v2 cross-reference links and other transformations
func postProcessHTML(html string, compilerVersion int) string {
	if compilerVersion != 2 {
		return html
	}

	// Convert area references to links: "Área 5" → "<a href="#area-5">Área 5</a>"
	html = areaRefPattern.ReplaceAllStringFunc(html, func(match string) string {
		m := areaRefPattern.FindStringSubmatch(match)
		if m == nil {
			return match
		}
		areaNum := m[1]
		return fmt.Sprintf(`<a href="#area-%s">%s</a>`, areaNum, match)
	})

	return html
}

func processImages(text string, baseDir string, seenImages map[string]bool) string {
	text = imageRegex.ReplaceAllStringFunc(text, func(match string) string {
		matches := imageRegex.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		alt := matches[1]
		imgPath := matches[2]

		if !filepath.IsAbs(imgPath) {
			// Strip ../ and ./ prefixes — assets are always under campaign root
			for strings.HasPrefix(imgPath, "../") {
				imgPath = strings.TrimPrefix(imgPath, "../")
			}
			for strings.HasPrefix(imgPath, "./") {
				imgPath = strings.TrimPrefix(imgPath, "./")
			}
			imgPath = filepath.Join(baseDir, imgPath)
		}

		return embedImage(imgPath, alt, baseDir, seenImages)
	})

	// Detect code-formatted asset refs: `assets/filename.ext`
	text = codeAssetRegex.ReplaceAllStringFunc(text, func(match string) string {
		matches := codeAssetRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		filename := matches[1]
		imgPath := filepath.Join(baseDir, "assets", filename)
		return embedImage(imgPath, filename, baseDir, seenImages)
	})

	return text
}

func embedImage(imgPath, alt, baseDir string, seenImages map[string]bool) string {
	// Deduplicate: skip if same image path already embedded
	if seenImages[imgPath] {
		return ""
	}
	seenImages[imgPath] = true

	data, err := os.ReadFile(imgPath)
	if err != nil {
		return fmt.Sprintf(`<span class="image-missing">[Image: %s]</span>`, html.EscapeString(alt))
	}

	ext := strings.ToLower(filepath.Ext(imgPath))
	var mimeType string

	switch ext {
	case ".svg":
		// Use SVG directly - wkhtmltopdf can render SVG files with --enable-local-file-access
		// Return relative path from campaign directory
		relPath, _ := filepath.Rel(baseDir, imgPath)
		if relPath == "" {
			relPath = imgPath
		}
		return fmt.Sprintf(`<img src="%s" alt="%s" class="campaign-image"/>`, html.EscapeString(relPath), html.EscapeString(alt))
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	default:
		mimeType = "image/png"
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
	return fmt.Sprintf(`<img src="%s" alt="%s" class="campaign-image"/>`, dataURI, html.EscapeString(alt))
}

// findCoverImage searches for a cover image in the campaign's assets directory.
// Priority: cover.ext > cover-*.ext (tries png, jpg, jpeg, webp, svg in order)
func findCoverImage(campaignDir string) string {
	extensions := []string{".png", ".jpg", ".jpeg", ".webp", ".svg"}
	baseDir := filepath.Join(campaignDir, "assets")

	for _, ext := range extensions {
		if covers, _ := filepath.Glob(filepath.Join(baseDir, "cover"+ext)); len(covers) > 0 {
			return covers[0]
		}
	}
	for _, ext := range extensions {
		if covers, _ := filepath.Glob(filepath.Join(baseDir, "cover-*"+ext)); len(covers) > 0 {
			return covers[0]
		}
	}
	return ""
}
