package compiler

import (
	"bytes"
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

//go:embed templates/chapter.md.tmpl
var chapterTemplate string

//go:embed templates/session-prep.md.tmpl
var sessionPrepTemplate string

//go:embed templates/character-sheet.md.tmpl
var characterSheetTemplate string

//go:embed templates/prologue.md.tmpl
var prologueTemplate string

type Compiler struct {
	CampaignDir         string
	PDFEngine           string
	CompilerVersion     int
	seenImages          map[string]bool
	handoutRendererImpl HandoutRenderer
	// warnedFilesDebounce tracks files that have already emitted a
	// "no-colon DM Sidebar" warning in this compile session (REQ-1.5).
	warnedFilesDebounce map[string]struct{}
}

// blockquoteClass classifies the semantic role of a markdown blockquote.
type blockquoteClass int

const (
	bqReadAloud blockquoteClass = iota
	bqDMSidebar
	bqChapterSummary
	bqIntroductionSidebar
)

// anchorRegistry maps a requested link fragment to the element ID emitted in the HTML.
type anchorRegistry map[string]string

// pdfEnginePriority defines the preferred order of PDF engines.
var pdfEnginePriority = []string{
	"chromium",
	"chrome",
	"google-chrome",
	"google-chrome-stable",
	"wkhtmltopdf",
}

// detectPDFEngine searches for an available PDF engine in PATH.
// It prefers Chromium/Chrome headless over legacy wkhtmltopdf.
func detectPDFEngine() string {
	for _, engine := range pdfEnginePriority {
		if _, err := exec.LookPath(engine); err == nil {
			return engine
		}
	}
	return "chromium" // fallback: will fail gracefully at runtime if truly missing
}

// IsPDFEngineAvailable reports whether any supported PDF engine is installed.
func IsPDFEngineAvailable() bool {
	for _, engine := range pdfEnginePriority {
		if _, err := exec.LookPath(engine); err == nil {
			return true
		}
	}
	return false
}

// SupportedEngines returns the list of supported PDF engine names.
func SupportedEngines() []string {
	return append([]string(nil), pdfEnginePriority...)
}

func New(campaignDir, pdfEngine string) *Compiler {
	if pdfEngine == "" {
		pdfEngine = detectPDFEngine()
	}
	return &Compiler{
		CampaignDir:         campaignDir,
		PDFEngine:           pdfEngine,
		CompilerVersion:     2,
		seenImages:          make(map[string]bool),
		warnedFilesDebounce: make(map[string]struct{}),
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
	case "chapter":
		return chapterTemplate, nil
	case "session-prep":
		return sessionPrepTemplate, nil
	case "character-sheet":
		return characterSheetTemplate, nil
	case "prologue":
		return prologueTemplate, nil
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

	// Single-pass: generate PDF, then verify images (advisory only)
	if err := c.htmlToPDF(ctx, htmlPath, pdfPath); err != nil {
		return "", fmt.Errorf("PDF generation failed: %w", err)
	}

	// Verify images are present in the generated HTML (advisory — warnings only)
	expected, found, warnings, err := c.verifyImages(htmlPath)
	if err != nil {
		// I/O error during verification — log but don't discard the PDF
		fmt.Fprintf(os.Stderr, "grimorio: image verification error: %v\n", err)
	} else if len(warnings) > 0 {
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "grimorio: warning: %s\n", w)
		}
		_ = expected
		_ = found
	}

	return pdfPath, nil
}

// generateHTML reads all markdown sources and generates the full HTML document.
type campaignSection struct {
	name  string
	path  string
	isDir bool
}

// campaignSections returns the ordered list of markdown sources for a campaign.
// It mirrors the source order used by generateHTML.
func campaignSections(campaignDir string) []campaignSection {
	chapterDir := filepath.Join(campaignDir, "chapters")

	sections := []campaignSection{
		{"Introduction", filepath.Join(campaignDir, "introduction.md"), false},
		{"Lore y Ambientación", filepath.Join(campaignDir, "lore.md"), false},
	}

	prologueChapterPath := filepath.Join(chapterDir, "chapter_00.md")
	if data, err := os.ReadFile(prologueChapterPath); err == nil {
		if hasPrologueFrontmatter(string(data)) {
			sections = append(sections, campaignSection{"Prologue", prologueChapterPath, false})
		}
	}

	sections = append(sections, []campaignSection{
		{"Chapters", chapterDir, true},
		{"Setting Guide", filepath.Join(campaignDir, "setting-guide.md"), false},
		{"Apéndice A: NPCs y Facciones", filepath.Join(campaignDir, "npcs"), true},
		{"Apéndice B: Bestiario", filepath.Join(campaignDir, "bestiary"), true},
		{"Apéndice C: Encuentros", filepath.Join(campaignDir, "encounters"), true},
		{"Apéndice D: Mapas de Referencia", filepath.Join(campaignDir, "maps"), true},
		{"Appendices", filepath.Join(campaignDir, "appendices.md"), false},
		{"Apéndice G: Character Sheets", filepath.Join(campaignDir, "characters"), true},
		{"Apéndice H: Quests", filepath.Join(campaignDir, "quests"), true},
	}...)

	return sections
}

// buildAnchorRegistry scans all campaign markdown sources and maps link fragments
// to the emitted element IDs used in the generated HTML.
func buildAnchorRegistry(campaignDir string) anchorRegistry {
	reg := make(anchorRegistry)
	sections := campaignSections(campaignDir)

	for _, sec := range sections {
		paths := sourcePathsForSection(sec)
		for _, path := range paths {
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			md := string(content)

			// Compute sectionID the same way generateHTML does.
			sectionID := "sec-" + sanitizeID(sec.name)
			if sec.isDir {
				sectionID = "sec-" + sanitizeID(sec.name+"-"+filepath.Base(path))
			}

			// Explicit anchors take precedence.
			for _, m := range explicitAnchorPattern.FindAllStringSubmatch(md, -1) {
				if len(m) >= 2 && m[1] != "" {
					reg[m[1]] = m[1]
				}
			}

			// Area headings.
			for _, m := range areaHeadingLinePattern.FindAllStringSubmatch(md, -1) {
				if len(m) >= 2 {
					areaID := "area-" + m[1]
					reg[areaID] = areaID
				}
			}

			// Other headings: register slug -> sectionID-slug for the first occurrence.
			for _, m := range markdownHeadingPattern.FindAllStringSubmatch(md, -1) {
				if len(m) >= 3 {
					text := strings.TrimSpace(m[2])
					slug := slugify(text)
					if slug == "" {
						continue
					}
					emittedID := sectionID + "-" + slug
					if _, exists := reg[slug]; !exists {
						reg[slug] = emittedID
					}
				}
			}
		}
	}

	return reg
}

// sourcePathsForSection returns the markdown file paths for a campaign section.
func sourcePathsForSection(sec campaignSection) []string {
	info, err := os.Stat(sec.path)
	if err != nil {
		return nil
	}
	if !info.IsDir() {
		return []string{sec.path}
	}

	entries, err := os.ReadDir(sec.path)
	if err != nil {
		return nil
	}
	var paths []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			paths = append(paths, filepath.Join(sec.path, e.Name()))
		}
	}
	return paths
}

// slugify converts heading text to a stable URL-friendly identifier.
func slugify(text string) string {
	text = strings.ToLower(text)
	var result strings.Builder
	for _, r := range text {
		switch {
		case r >= 'a' && r <= 'z':
			result.WriteRune(r)
		case r >= '0' && r <= '9':
			result.WriteRune(r)
		default:
			if result.Len() > 0 && result.String()[result.Len()-1] != '-' {
				result.WriteRune('-')
			}
		}
	}
	s := strings.Trim(result.String(), "-")
	return s
}

func (c *Compiler) generateHTML(title string) ([]string, error) {
	sections := campaignSections(c.CampaignDir)

	// Two-pass anchor registry for v2: first pass maps fragments to emitted IDs.
	var reg anchorRegistry
	if c.CompilerVersion == 2 {
		reg = buildAnchorRegistry(c.CampaignDir)
	}

	var htmlParts []string
	headingCounter := 0

	htmlParts = append(htmlParts, "<!DOCTYPE html>")
	htmlParts = append(htmlParts, "<html><head><meta charset='utf-8'>")
	htmlParts = append(htmlParts, "<style>"+dndCSS+"</style>")
	htmlParts = append(htmlParts, "</head><body>")

	// Cover page — single page, no split
	htmlParts = append(htmlParts, fmt.Sprintf(`<div class="cover-wrapper" id="cover"><h1>%s</h1>`, html.EscapeString(title)))

	// Insert cover art if exists — search for cover.* or cover-*.*
	coverPath := findCoverImage(c.CampaignDir)
	if coverPath != "" {
		htmlParts = append(htmlParts, fmt.Sprintf(`<div class="cover-image">%s</div>`, embedImage(coverPath, "Cover Art", c.CampaignDir, c.seenImages)))
	}

	htmlParts = append(htmlParts, `<p class="cover-footer">Generated by Grimorio</p></div>`)

	// Open the two-column wrapper. The cover is a sibling of this div, so
	// it is naturally outside the multi-column flow (avoids the
	// column-span: all page-break split bug).
	htmlParts = append(htmlParts, `<div class="campaign-body">`)

	// Session Zero guidance (if available)
	sessionZeroHTML := c.generateSessionZero()
	if sessionZeroHTML != "" {
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
		htmlParts = append(htmlParts, sessionZeroHTML)
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
	}

	// Prologue narrative (if available)
	prologueHTML := c.generatePrologue()
	if prologueHTML != "" {
		htmlParts = append(htmlParts, `<div class="section-break"></div>`)
		htmlParts = append(htmlParts, prologueHTML)
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
					if f.Name() == "chapter_00.md" && hasPrologueFrontmatter(string(content)) {
						continue
					}
					sectionID := "sec-" + sanitizeID(sec.name+"-"+f.Name())
					htmlResult := c.markdownToHTMLWithID(string(content), c.CampaignDir, sectionID, &headingCounter, c.seenImages, reg, filepath.Join(sec.path, f.Name()))
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
			htmlResult := c.markdownToHTMLWithID(string(content), c.CampaignDir, sectionID, &headingCounter, c.seenImages, reg, sec.path)
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

	htmlParts = append(htmlParts, `</div>`) // close .campaign-body
	htmlParts = append(htmlParts, "</body></html>")
	return htmlParts, nil
}

// generateTOC creates a hierarchical table of contents
func (c *Compiler) generateTOC(sections []campaignSection) string {
	var b strings.Builder
	b.WriteString(`<div class="toc" id="toc"><h2>Table of Contents</h2><ul>`)

	for _, sec := range sections {
		if _, err := os.Stat(sec.path); err != nil {
			continue
		}

		id := "sec-" + sanitizeID(sec.name)
		fmt.Fprintf(&b, `<li><a href="#%s">%s</a><span class="page-ref"></span></li>`, id, html.EscapeString(sec.name))
	}

	b.WriteString(`</ul></div>`)
	return b.String()
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

	htmlResult := markdownToHTMLWithID(c, string(data), c.CampaignDir, "sec-session-zero", new(int), c.seenImages, c.CompilerVersion, nil, path)
	if strings.TrimSpace(htmlResult) == "" {
		return ""
	}

	return htmlResult // markdown already contains the heading with proper id
}

// hasPrologueFrontmatter checks if a markdown file has is_prologue: true in YAML frontmatter
func hasPrologueFrontmatter(content string) bool {
	return strings.Contains(content, "is_prologue: true")
}

// generatePrologue reads prologue.md and converts to HTML
func (c *Compiler) generatePrologue() string {
	path := filepath.Join(c.CampaignDir, "prologue.md")
	data, err := os.ReadFile(path)
	if err != nil {
		return "" // no prologue, skip
	}

	htmlResult := markdownToHTMLWithID(c, string(data), c.CampaignDir, "sec-prologue", new(int), c.seenImages, c.CompilerVersion, nil, path)
	if strings.TrimSpace(htmlResult) == "" {
		return ""
	}

	return htmlResult
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

	htmlResult := markdownToHTMLWithID(c, string(data), c.CampaignDir, fmt.Sprintf("sec-session-prep-%d", sessionNum), new(int), c.seenImages, c.CompilerVersion, nil, path)
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

	htmlResult := markdownToHTMLWithID(c, string(data), c.CampaignDir, fmt.Sprintf("sec-character-%s", characterID), new(int), c.seenImages, c.CompilerVersion, nil, path)
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

	// chapters/ is the only chapter source — grimorio-areas (legacy
	// areas/ directory) was removed in v5.0.2 WU7.
	chaptersDir := filepath.Join(c.CampaignDir, "chapters")
	if info, err := os.Stat(chaptersDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(chaptersDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, _ := os.ReadFile(filepath.Join(chaptersDir, f.Name()))
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

	return `<div class="roster-wrap">` + b.String() + `</div>`
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
// Returns (expected, found, warnings, error).
// warnings is non-empty when found < expected (advisory — not a blocking error).
// err is only returned for I/O failures.
func (c *Compiler) verifyImages(htmlPath string) (int, int, []string, error) {
	expected, err := c.countImagesInMarkdownSources()
	if err != nil {
		return 0, 0, nil, err
	}

	found, err := countImagesInHTML(htmlPath)
	if err != nil {
		return expected, 0, nil, err
	}

	var warnings []string
	if found < expected {
		warnings = append(warnings, fmt.Sprintf("image mismatch: expected %d, found %d", expected, found))
	}
	return expected, found, warnings, nil
}

// countImagesInMarkdownSources walks all markdown files in the campaign directory
// and counts UNIQUE image paths (deduplicated by resolved absolute path).
// This aligns with embedImage()'s dedup behavior via seenImages.
func (c *Compiler) countImagesInMarkdownSources() (int, error) {
	uniquePaths := make(map[string]bool)

	// grimorio-areas (legacy areas/ dir) was removed in v5.0.2 WU7.
	sections := []string{
		filepath.Join(c.CampaignDir, "session-zero.md"),
		filepath.Join(c.CampaignDir, "introduction.md"),
		filepath.Join(c.CampaignDir, "lore.md"),
		filepath.Join(c.CampaignDir, "setting-guide.md"),
		filepath.Join(c.CampaignDir, "appendices.md"),
		filepath.Join(c.CampaignDir, "chapters"),
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
					for _, p := range extractImagePaths(string(content)) {
						uniquePaths[resolveImagePath(p, c.CampaignDir)] = true
					}
				}
			}
		} else {
			content, err := os.ReadFile(secPath)
			if err != nil {
				continue
			}
			for _, p := range extractImagePaths(string(content)) {
				uniquePaths[resolveImagePath(p, c.CampaignDir)] = true
			}
		}
	}

	return len(uniquePaths), nil
}

// extractImagePaths extracts all image path strings from a markdown document.
// Returns raw paths (not resolved); the caller is responsible for resolution and dedup.
func extractImagePaths(md string) []string {
	var paths []string
	// Markdown images: ![alt](path) — group 2 is the path
	for _, m := range imageRegex.FindAllStringSubmatch(md, -1) {
		if len(m) >= 3 {
			paths = append(paths, m[2])
		}
	}
	// Raw <img> tags: extract src attribute
	for _, tag := range imgTagRegex.FindAllString(md, -1) {
		if src := imgSrcRegex.FindStringSubmatch(tag); len(src) >= 2 {
			paths = append(paths, src[1])
		}
	}
	// Code asset refs: `assets/filename.ext` — group 1 is the filename
	for _, m := range codeAssetRegex.FindAllStringSubmatch(md, -1) {
		if len(m) >= 2 {
			paths = append(paths, "assets/"+m[1])
		}
	}
	return paths
}

// resolveImagePath resolves a relative image path to an absolute path,
// matching the logic used by processImages/embedImage.
func resolveImagePath(path, baseDir string) string {
	for strings.HasPrefix(path, "../") {
		path = strings.TrimPrefix(path, "../")
	}
	for strings.HasPrefix(path, "./") {
		path = strings.TrimPrefix(path, "./")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
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

// isChromiumEngine reports whether the engine is a Chromium/Chrome variant.
func isChromiumEngine(engine string) bool {
	switch engine {
	case "chromium", "chrome", "google-chrome", "google-chrome-stable",
		"chromium-browser", "msedge":
		return true
	default:
		return false
	}
}

func (c *Compiler) htmlToPDF(ctx context.Context, htmlPath, pdfPath string) error {
	if isChromiumEngine(c.PDFEngine) {
		return c.htmlToPDFChromium(ctx, htmlPath, pdfPath)
	}
	return c.htmlToPDFWkhtmltopdf(ctx, htmlPath, pdfPath)
}

// buildChromiumCmd constructs the *exec.Cmd for the Chromium headless print
// invocation. It is extracted from htmlToPDFChromium so tests can inspect the
// argument slice without actually running the engine.
func (c *Compiler) buildChromiumCmd(ctx context.Context, htmlPath, pdfPath string) *exec.Cmd {
	absHTML, err := filepath.Abs(htmlPath)
	if err != nil {
		absHTML = htmlPath
	}
	absPDF, err := filepath.Abs(pdfPath)
	if err != nil {
		absPDF = pdfPath
	}

	// Use file:// URL for local file access
	fileURL := "file://" + absHTML

	return exec.CommandContext(ctx, c.PDFEngine,
		"--headless",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-gpu",
		"--disable-web-security",
		"--allow-file-access-from-files",
		"--run-all-compositor-stages-before-draw",
		"--print-to-pdf="+absPDF,
		"--no-pdf-header-footer",
		"--paper=A4",
		"--virtual-time-budget=10000",
		fileURL,
	)
}

func (c *Compiler) htmlToPDFChromium(ctx context.Context, htmlPath, pdfPath string) error {
	cmd := c.buildChromiumCmd(ctx, htmlPath, pdfPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("chromium headless failed: %w\nOutput: %s", err, string(output))
	}
	return nil
}

func (c *Compiler) htmlToPDFWkhtmltopdf(ctx context.Context, htmlPath, pdfPath string) error {
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
	boldItalicRegex   = regexp.MustCompile(`\*\*\*(.+?)\*\*\*`)
	boldRegex         = regexp.MustCompile(`\*\*(.+?)\*\*`)
	italicRegex       = regexp.MustCompile(`\*(.+?)\*`)
	boldAdjacentRegex = regexp.MustCompile(`</strong>([^\s<:;,\.!?])`) // Exclude punctuation to prevent &thinsp; leak
	htmlCommentRegex  = regexp.MustCompile(`<!--[\s\S]*?-->`)
	imageRegex        = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	imgTagRegex       = regexp.MustCompile(`<img[^>]*(?:/\s*)?>`)
	imgSrcRegex       = regexp.MustCompile(`src="([^"]*)"`)

	htmlBlockPlaceholderRegex = regexp.MustCompile(`^\x00HTMLBLOCK\d+\x00$`)

	// readAloudPrefixRe strips **Read-Aloud Text:**, **Read-Aloud:** or **Para Leer en Voz Alta:** labels from blockquote text
	// (CSS .read-aloud::before pseudo-element provides the visual label, so the inline text would duplicate it)
	readAloudPrefixRe = regexp.MustCompile(`^\*{0,2}(?:Read-Aloud(?:\s+Text)?|Para Leer en Voz Alta)\*{0,2}:\s*\*{0,2}\s*`)

	// dmSidebarPrefixRe strips DM Sidebar labels from blockquote text.
	// The trailing colon is OPTIONAL — both `DM Sidebar:` and `DM Sidebar`
	// (the no-colon variant) must match (REQ-1.4).
	dmSidebarPrefixRe = regexp.MustCompile(`(?i)^\*{0,2}(?:#####\s+)?DM Sidebar:?\s*\*{0,2}\s*`)

	// linkRegex matches markdown links [text](href).
	linkRegex = regexp.MustCompile(`\[(?P<text>[^\]]+)\]\((?P<href>[^)]+)\)`)

	// aTagRegex matches existing HTML anchor tags.
	aTagRegex = regexp.MustCompile(`<a[^>]*>.*?</a>`)

	// worksheetDivRe matches the opening tag of a character-worksheet div block.
	worksheetDivRe = regexp.MustCompile(`<div[^>]*\bclass="character-worksheet"[^>]*>`)

	// rawTagRe strips raw HTML tags (e.g. <div>, <span>) — <img> tags are preserved via stash/restore
	rawTagRe = regexp.MustCompile(`<[^>]+>`)

	blockquoteRe   = regexp.MustCompile(`^>\s*(.*)`)
	codeAssetRegex = regexp.MustCompile("`assets/([\\w\\-]+\\.(svg|png|jpg|jpeg|gif|webp))`")
	sceneRegex     = regexp.MustCompile(`\[SCENE:\s*(.*?)\]`)

	// introductionSidebarMarkerRegex matches an HTML comment used to mark the next blockquote as an introduction sidebar.
	introductionSidebarMarkerRegex = regexp.MustCompile(`^\s*<!--\s*introduction-sidebar\s*-->\s*$`)

	// v2 patterns
	areaHeadingPattern     = regexp.MustCompile(`^[AaÁá]rea\s+(\d+):\s*(.+)$`)
	areaHeadingLinePattern = regexp.MustCompile(`(?m)^###\s+[AaÁá]rea\s+(\d+):\s*(.+)$`)
	areaRefPattern         = regexp.MustCompile(`(?i)\b[Áa]rea\s+(\d+)\b`)
	markdownHeadingPattern = regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	explicitAnchorPattern  = regexp.MustCompile(`<a\s+id=["']([^"']+)["'][^>]*>.*?</a>`)
)

// formatInline processes bold and italic markers, ensuring no word merging after </strong>
func formatInline(text string) string {
	text = boldItalicRegex.ReplaceAllString(text, "<strong><em>$1</em></strong>")
	text = boldRegex.ReplaceAllString(text, "<strong>$1</strong>")
	// Remove &thinsp; - it causes spacing issues like "sol , y" and "volvióazul"
	text = boldAdjacentRegex.ReplaceAllString(text, "</strong>$1")
	return italicRegex.ReplaceAllString(text, "<em>$1</em>")
}

// processInlineText processes markdown images, escapes HTML, formats inline,
// converts markdown links, and preserves existing <img> and <a> tags so they are not escaped.
func processInlineText(text string, baseDir string, seenImages map[string]bool, reg anchorRegistry) string {
	// 1. Process markdown images first
	text = processImages(text, baseDir, seenImages)

	// 2. Convert markdown links to anchors before escaping
	text = processLinks(text, "", reg)

	// 3. Extract and preserve existing <img> and <a> tags
	var imgTags []string
	var linkTags []string
	text = imgTagRegex.ReplaceAllStringFunc(text, func(match string) string {
		imgTags = append(imgTags, match)
		return "\x00IMG\x00"
	})
	text = aTagRegex.ReplaceAllStringFunc(text, func(match string) string {
		linkTags = append(linkTags, match)
		return "\x00LINK\x00"
	})

	// 4. Escape HTML
	text = html.EscapeString(text)

	// 5. Format inline (bold, italic)
	text = formatInline(text)

	// 6. Restore <a> tags
	for _, link := range linkTags {
		text = strings.Replace(text, "\x00LINK\x00", link, 1)
	}

	// 7. Restore <img> tags
	for _, img := range imgTags {
		text = strings.Replace(text, "\x00IMG\x00", img, 1)
	}

	return text
}

// stripCharacterWorksheets removes <div class="character-worksheet"> blocks from markdown.
// It reuses extractBalancedDivs so nested divs are handled correctly, and restores
// any non-worksheet blocks so surrounding HTML is preserved.
func stripCharacterWorksheets(md string) string {
	md, blocks := extractBalancedDivs(md)
	for i, block := range blocks {
		placeholder := fmt.Sprintf("\x00HTMLBLOCK%d\x00", i)
		if worksheetDivRe.MatchString(block) {
			md = strings.Replace(md, placeholder, "", 1)
		} else {
			md = strings.Replace(md, placeholder, block, 1)
		}
	}
	return md
}

// classifyBlockquote determines the semantic role of a blockquote and returns the
// class along with lines that have the class-identifying marker removed.
//
// c and filePath are optional (both may be empty/nil): they enable the
// per-file stderr warning for the no-colon DM Sidebar variant (REQ-1.5).
func classifyBlockquote(lines []string, sectionID, filePath string, c *Compiler) (blockquoteClass, []string) {
	if len(lines) == 0 {
		return bqReadAloud, lines
	}

	first := lines[0]
	cleaned := make([]string, len(lines))
	copy(cleaned, lines)

	// DM Sidebar detection takes precedence.
	if dmSidebarPrefixRe.MatchString(first) {
		// REQ-1.5: warn once per file when the no-colon variant is parsed.
		// The debounce lives on the Compiler so concurrent compiles in the
		// MCP server don't cross-contaminate. Empty filePath / nil Compiler
		// skip the warning (e.g. the unexported markdownToHTMLWithID tests
		// pass nil for the Compiler).
		// Distinguish colon vs no-colon: the line must contain a `:` between
		// "DM Sidebar" and the title content to be the colon variant.
		if c != nil && filePath != "" && !strings.Contains(first, "DM Sidebar:") && !strings.Contains(first, "dm sidebar:") {
			if _, warned := c.warnedFilesDebounce[filePath]; !warned {
				c.warnedFilesDebounce[filePath] = struct{}{}
				fmt.Fprintf(os.Stderr,
					"grimorio: warning: %s: DM Sidebar prefix without ':' — consider adding a colon for consistency\n",
					filePath)
			}
		}
		cleaned[0] = dmSidebarPrefixRe.ReplaceAllString(first, "")
		return bqDMSidebar, cleaned
	}

	// Read-Aloud detection.
	if readAloudPrefixRe.MatchString(first) {
		cleaned[0] = readAloudPrefixRe.ReplaceAllString(first, "")
		return bqReadAloud, cleaned
	}

	// Introduction sidebar: introduction section with a heading-like first line.
	if strings.Contains(sectionID, "introduction") && headingMarkerRe.MatchString(first) {
		return bqIntroductionSidebar, cleaned
	}

	// Chapter summary: chapter section containing list items.
	if strings.Contains(sectionID, "chapter") {
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
				return bqChapterSummary, cleaned
			}
		}
	}

	return bqReadAloud, cleaned
}

// statBlockSizeRegex matches the WotC size word at the start of an italic
// type line that opens a monster stat block. Bilingual, both masculine and
// feminine Spanish forms are covered (the el-exiliado bestiary uses
// Diminuta, Pequeña, Mediana — not just the masculine forms).
var statBlockSizeRegex = regexp.MustCompile(
	`^\*(Tiny|Small|Medium|Large|Huge|Gargantuan|` +
		`Diminut[oa]|Minuscul[oa]|Pequeñ[oa]|Median[oa]|Grand[ea]|Enorme[ns]?|Colosal)\s+\S`)

// statBlockPropertyRegex finds a single **Label** in a stat block line. Go
// regexp does not support lookahead, so the value is extracted separately
// via slice arithmetic in splitPropertyGroups below.
var statBlockPropertyRegex = regexp.MustCompile(`\*\*([^*]+?)\*\*`)

// boldTraitRegex matches a leading "**Ability Name.**" trait header inside a
// stat block (REQ-2.5).
var boldTraitRegex = regexp.MustCompile(`^\*\*[^*]+\.\*\*`)

// splitPropertyGroups splits a stat-block inline line into (label, value)
// pairs by walking the line and finding each `**Label**` boundary. Returns
// the captured groups in source order. A value runs from the end of its
// label to the start of the next label (or end of line), trimmed.
func splitPropertyGroups(line string) [][]string {
	matches := statBlockPropertyRegex.FindAllStringSubmatchIndex(line, -1)
	if len(matches) == 0 {
		return nil
	}
	var groups [][]string
	for i, m := range matches {
		if len(m) < 4 {
			continue
		}
		label := strings.TrimSpace(line[m[2]:m[3]])
		// Value starts after `**label**` and runs to the start of the next label,
		// or to the end of the line.
		valueStart := m[1]
		valueEnd := len(line)
		if i+1 < len(matches) {
			valueEnd = matches[i+1][0]
		}
		value := strings.TrimSpace(line[valueStart:valueEnd])
		groups = append(groups, []string{label, value})
	}
	return groups
}

// isCoreStat reports whether a property label is one of the WotC core stats
// (Armor Class, Hit Points, Speed) that must be rendered as a .stat-line.
func isCoreStat(label string) bool {
	switch label {
	case "Armor Class", "Hit Points", "Speed":
		return true
	}
	return false
}

// detectTraitLine reports whether a stat-block line is a WotC trait header.
//
// The WotC convention is "**Name.** description with optional **inner
// bolds**". The signature is: the FIRST `**…**` group ends in `.` AND at
// least one more `**…**` group exists in the same line. When this matches,
// the line is rendered as a single <p class="trait"> with inner bolds
// preserved inline, instead of being split into one .property-line /
// .stat-line per `**…**` group.
//
// Reference: SRD monster stat block convention ("Damage Vulnerabilities"
// / "Condition Immunities" / "Senses" lines never have a label ending in
// `.`, so a trailing period in the first bold is a reliable trait signal).
func detectTraitLine(line string) bool {
	matches := statBlockPropertyRegex.FindAllStringSubmatchIndex(line, -1)
	if len(matches) < 2 {
		return false
	}
	first := matches[0]
	firstLabel := strings.TrimSpace(line[first[2]:first[3]])
	return strings.HasSuffix(firstLabel, ".")
}

// peekHoistableMonsterImage scans forward from startIdx looking for a
// markdown image whose path's basename starts with `monster-`. If found,
// returns the rendered <img> HTML and the index of the line AFTER the
// image. Otherwise returns ("", 0). This is the convention guard for
// image hoisting: only files explicitly named `monster-*.png` (and
// similar) are hoisted into the just-emitted .stat-block; scene, NPC,
// and cover illustrations keep their normal top-level rendering.
//
// The scan is permissive: it skips blank lines and horizontal rules
// (---, ***, ___, - - -), but it does NOT stop at intermediate text /
// h3 / paragraphs. The real el-exiliado bestiary places the hero image
// AFTER the monster's tactical-phases section (which has its own ---,
// ### heading, and several trait paragraphs before the image). The
// only hard stop is the next `## ` heading — we never cross into the
// next monster's section, even though we scan past any in-section
// content.
//
// REQ (fix-statblock-layout-and-cover-overflow), Decision §Image hoisting.
func peekHoistableMonsterImage(startIdx int, lines []string, baseDir string, seenImages map[string]bool) (string, int) {
	for peek := startIdx; peek < len(lines); peek++ {
		t := strings.TrimSpace(lines[peek])
		// Hard stop: next H2 heading (the next monster's section).
		// We must not cross this boundary, even though we skip blanks
		// and horizontal rules.
		if strings.HasPrefix(t, "## ") {
			return "", 0
		}
		// Skip: blank lines, horizontal rules, and any non-image content.
		// (Tactical phases, h3 sub-headings, and trait paragraphs inside
		// the monster section are all kept as-is in the stat block by
		// parseStatBlock when they appear BEFORE the closing ---; the
		// peek only runs AFTER parseStatBlock returns. Content after
		// the closing --- is conventionally the hero image, possibly
		// interleaved with commentary that the author placed between.)
		if t == "" || t == "---" || t == "***" || t == "___" || t == "- - -" {
			continue
		}
		if imageRegex.MatchString(t) {
			m := imageRegex.FindStringSubmatch(t)
			if len(m) < 3 {
				return "", 0
			}
			imgPath := m[2] // group 1 is alt text, group 2 is the path
			base := filepath.Base(imgPath)
			if strings.HasPrefix(base, "monster-") {
				img := processImages(t, baseDir, seenImages)
				if img == "" {
					// Image was already seen (dedup) or unreadable; advance
					// past it so the main loop doesn't emit a duplicate.
					return "", peek + 1
				}
				return img, peek + 1
			}
		}
		// Non-image, non-skip content. Keep scanning — the image may be
		// a few lines further (the el-exiliado convention is "monster
		// image goes after the LAST --- rule, just before the next
		// ## section, regardless of intermediate commentary").
		continue
	}
	return "", 0
}

// tryStatBlock peeks ahead from a `## ` heading. If the next non-blank line
// matches the WotC size+type italic pattern, it renders a full stat block
// and returns the rendered HTML + the number of lines consumed. Otherwise
// it returns ("", 0) and the caller falls back to the regular <h2> path.
//
// REQ-2.1, 2.2, 2.5, 2.6, 2.7, 2.10, 2.11, 4.4
func tryStatBlock(name string, lines []string, startIdx int, baseDir string, seenImages map[string]bool, reg anchorRegistry) (string, int) {
	// Look for the italic size+type line within the next 3 lines (skipping blanks).
	endIdx := startIdx + 4
	if endIdx > len(lines) {
		endIdx = len(lines)
	}
	typeLineIdx := -1
	for j := startIdx + 1; j < endIdx; j++ {
		t := strings.TrimSpace(lines[j])
		if t == "" {
			continue
		}
		// Must be a full italic line (starts and ends with *), and must start
		// with a recognized size word. The body can contain commas + words.
		if strings.HasPrefix(t, "*") && strings.HasSuffix(t, "*") && statBlockSizeRegex.MatchString(t) {
			typeLineIdx = j
			break
		}
		// First non-blank line wasn't a stat-block opener — bail out.
		break
	}
	if typeLineIdx == -1 {
		return "", 0
	}
	return parseStatBlock(name, lines, typeLineIdx, baseDir, seenImages, reg)
}

// parseStatBlock renders a WotC stat block starting at typeLineIdx (the italic
// size+type line) and consuming up to the next `## ` heading or `---` rule.
// Returns the rendered HTML and the number of lines consumed (including the
// `## ` heading line and the italic type line).
func parseStatBlock(name string, lines []string, typeLineIdx int, baseDir string, seenImages map[string]bool, reg anchorRegistry) (string, int) {
	var b strings.Builder
	fmt.Fprintf(&b, `<div class="stat-block" data-monster="%s">`, html.EscapeString(name))
	fmt.Fprintf(&b, `<h2>%s</h2>`, html.EscapeString(name))

	// Italic type line → <p class="monster-type">
	typeLine := strings.TrimSpace(lines[typeLineIdx])
	typeLine = strings.TrimPrefix(typeLine, "*")
	typeLine = strings.TrimSuffix(typeLine, "*")
	typeLine = strings.TrimSpace(typeLine)
	fmt.Fprintf(&b, `<p class="monster-type">%s</p>`, formatInline(typeLine))

	// Walk subsequent lines until next `## ` or `---` or end of input.
	consumed := typeLineIdx + 1
	for j := typeLineIdx + 1; j < len(lines); j++ {
		raw := lines[j]
		t := strings.TrimSpace(raw)

		// Stop conditions: next H2 (case-insensitive) or horizontal rule.
		if strings.HasPrefix(t, "## ") || t == "---" || t == "***" || t == "___" {
			break
		}

		// Empty line — skip (stat blocks are dense).
		if t == "" {
			consumed = j + 1
			continue
		}

		// Ability-scores table: header row "STR DEX CON INT WIS CHA" with 6 cells
		// AND followed by a separator row. Emit a <table class="ability-scores">.
		if isTableRow(t) && isAbilityScoresTable(j, lines) {
			b.WriteString(renderAbilityScoresTable(lines, &j, baseDir, seenImages, reg))
			consumed = j + 1
			continue
		}

		// Multi-property line: count **Label** groups
		groups := splitPropertyGroups(t)

		// WotC trait detection: a line whose first **…** label ends in "."
		// AND has at least one more **…** group in the same line is a trait
		// description (e.g. "**Radiación Distorsionante (pasiva).** La
		// serpiente tiene **inmunidad** … **ventaja** …"). Render as a
		// SINGLE <p class="trait"> with the whole line preserved (inner
		// bolds kept inline). Action detection (<em> in value) wins when
		// both signatures are present.
		if detectTraitLine(t) {
			rendered := formatInline(t)
			if strings.Contains(rendered, "<em>") {
				fmt.Fprintf(&b, `<p class="action">%s</p>`, rendered)
			} else {
				fmt.Fprintf(&b, `<p class="trait">%s</p>`, rendered)
			}
			consumed = j + 1
			continue
		}

		// Single bold "**Actions**" or "**Legendary Actions**" alone
		// → <h3 class="actions-heading">
		if len(groups) == 1 {
			label := groups[0][0]
			if label == "Actions" || label == "Legendary Actions" {
				fmt.Fprintf(&b, `<h3 class="actions-heading">%s</h3>`, html.EscapeString(label))
				consumed = j + 1
				continue
			}
		}

		// 3 groups on a single line (AC / HP / Speed) → 3 .stat-line divs
		if len(groups) == 3 {
			for _, g := range groups {
				label := g[0]
				value := g[1]
				fmt.Fprintf(&b, `<div class="stat-line"><span class="stat-label">%s</span> <span class="stat-value">%s</span></div>`,
					html.EscapeString(label), formatInline(value))
			}
			consumed = j + 1
			continue
		}

		// 1 group: if the label is one of the WotC core stats (AC/HP/Speed)
		// → .stat-line. Otherwise classify by label/value shape:
		//   - label ends with "." + value has no <em> → .trait (REQ-2.5)
		//   - value has <em>                     → .action (REQ-2.8)
		//   - otherwise                          → .property-line
		// 4+ groups: each → .property-line.
		if len(groups) >= 1 {
			for _, g := range groups {
				label := g[0]
				value := g[1]
				// Strip trailing period from trait-style labels (e.g. "Incorpóreo y luminoso.")
				cleanLabel := strings.TrimSuffix(label, ".")
				if len(groups) == 1 && isCoreStat(cleanLabel) {
					fmt.Fprintf(&b, `<div class="stat-line"><span class="stat-label">%s</span> <span class="stat-value">%s</span></div>`,
						html.EscapeString(cleanLabel), formatInline(value))
				} else if len(groups) == 1 && strings.HasSuffix(label, ".") {
					// Single-group line ending in "." is a trait or action.
					// Distinguish by whether the rendered value contains <em>
					// (italic attack notation like "*Melee Spell Attack:*").
					rendered := formatInline(value)
					if strings.Contains(rendered, "<em>") {
						fmt.Fprintf(&b, `<p class="action"><strong>%s</strong> %s</p>`,
							html.EscapeString(label), rendered)
					} else {
						fmt.Fprintf(&b, `<p class="trait"><strong>%s</strong> %s</p>`,
							html.EscapeString(label), rendered)
					}
				} else {
					fmt.Fprintf(&b, `<p class="property-line"><strong>%s</strong> %s</p>`,
						html.EscapeString(label), formatInline(value))
				}
			}
			consumed = j + 1
			continue
		}

		// No **Label** group at all — render as a plain paragraph (.trait by default).
		// If line has a leading "**Name.**" with a period, emit <p class="trait">.
		// Otherwise treat as narrative.
		if m := boldTraitRegex.FindStringSubmatch(t); m != nil {
			name := strings.TrimSuffix(strings.TrimSpace(m[1]), ".")
			rest := strings.TrimSpace(t[len(m[0]):])
			fmt.Fprintf(&b, `<p class="trait"><strong>%s.</strong> %s</p>`,
				html.EscapeString(name), formatInline(rest))
		} else {
			escaped := processInlineText(t, baseDir, seenImages, reg)
			fmt.Fprintf(&b, `<p>%s</p>`, escaped)
		}
		consumed = j + 1
	}

	b.WriteString(`</div>`)
	return b.String(), consumed
}

// isAbilityScoresTable returns true if the table starting at row j is the
// 6-column ability-scores table (header row contains STR/DEX/CON/INT/WIS/CHA
// and is followed by a separator).
func isAbilityScoresTable(rowIdx int, lines []string) bool {
	if rowIdx+1 >= len(lines) {
		return false
	}
	header := strings.TrimSpace(lines[rowIdx])
	cells := parseTableRow(header)
	if len(cells) != 6 {
		return false
	}
	needles := []string{"STR", "DEX", "CON", "INT", "WIS", "CHA"}
	for i, c := range cells {
		uc := strings.ToUpper(strings.TrimSpace(c))
		if uc != needles[i] {
			return false
		}
	}
	// Next line must be a separator row
	return isTableSeparator(strings.TrimSpace(lines[rowIdx+1]))
}

// renderAbilityScoresTable renders the 6-column ability-scores table starting
// at row j. Returns the HTML and advances j past the table (header + separator
// + data rows). The header row index is at j.
func renderAbilityScoresTable(lines []string, j *int, baseDir string, seenImages map[string]bool, reg anchorRegistry) string {
	var b strings.Builder
	b.WriteString(`<table class="ability-scores"><thead><tr>`)
	headers := parseTableRow(lines[*j])
	for _, h := range headers {
		fmt.Fprintf(&b, `<th>%s</th>`, html.EscapeString(strings.TrimSpace(h)))
	}
	b.WriteString(`</tr></thead><tbody>`)
	*j += 2 // skip header + separator
	for ; *j < len(lines); *j++ {
		t := strings.TrimSpace(lines[*j])
		if !isTableRow(t) {
			break
		}
		row := parseTableRow(t)
		b.WriteString(`<tr>`)
		for _, c := range row {
			fmt.Fprintf(&b, `<td>%s</td>`, processInlineText(strings.TrimSpace(c), baseDir, seenImages, reg))
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	*j-- // outer loop will do *j++
	return b.String()
}
var headingMarkerRe = regexp.MustCompile(`^#{1,5}\s+`)

// stripBlockquotePrefix removes remaining section/heading markers from blockquote
// text so that CSS pseudo-elements can provide the visible label.
func stripBlockquotePrefix(text string, class blockquoteClass) string {
	switch class {
	case bqDMSidebar, bqIntroductionSidebar:
		text = headingMarkerRe.ReplaceAllString(text, "")
	}
	return text
}

// extractBalancedDivs extracts HTML div blocks with proper nesting support.
// It tracks opening and closing div tags to handle nested structures correctly.
// Returns the markdown with div blocks replaced by placeholders, and the extracted blocks.
func extractBalancedDivs(md string) (string, []string) {
	var blocks []string
	var result strings.Builder
	i := 0

	for i < len(md) {
		// Look for <div start
		if i+4 <= len(md) && strings.ToLower(md[i:i+4]) == "<div" {
			// Found div start, now track depth
			start := i
			depth := 0
			j := i

			for j < len(md) {
				// Check for div open
				if j+4 <= len(md) && strings.ToLower(md[j:j+4]) == "<div" {
					depth++
					j += 4
					continue
				}

				// Check for div close
				if j+6 <= len(md) && strings.ToLower(md[j:j+6]) == "</div>" {
					depth--
					j += 6
					if depth == 0 {
						// Complete block found
						block := md[start:j]
						placeholder := fmt.Sprintf("\x00HTMLBLOCK%d\x00", len(blocks))
						result.WriteString(placeholder)
						blocks = append(blocks, block)
						i = j
						break
					}
					continue
				}

				j++
			}

			// If we reached end without closing, auto-close remaining divs
			if depth > 0 {
				block := md[start:j]
				// Append closing tags for remaining depth
				for d := 0; d < depth; d++ {
					block += "</div>"
				}
				placeholder := fmt.Sprintf("\x00HTMLBLOCK%d\x00", len(blocks))
				result.WriteString(placeholder)
				blocks = append(blocks, block)
				i = j
			}
		} else {
			result.WriteByte(md[i])
			i++
		}
	}

	return result.String(), blocks
}

func (c *Compiler) markdownToHTMLWithID(md string, baseDir string, sectionID string, headingCounter *int, seenImages map[string]bool, reg anchorRegistry, filePath string) string {
	return markdownToHTMLWithID(c, md, baseDir, sectionID, headingCounter, seenImages, c.CompilerVersion, reg, filePath)
}

func markdownToHTMLWithID(c *Compiler, md string, baseDir string, sectionID string, headingCounter *int, seenImages map[string]bool, compilerVersion int, reg anchorRegistry, filePath string) string {
	// Strip character-worksheet div blocks before further processing (v2 only)
	if compilerVersion == 2 {
		md = stripCharacterWorksheets(md)
	}

	// Strip HTML comments before processing to prevent artifacts in PDF, but preserve
	// introduction-sidebar markers so they can disambiguate blockquotes.
	md = htmlCommentRegex.ReplaceAllStringFunc(md, func(match string) string {
		inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(match, "<!--"), "-->"))
		if inner == "introduction-sidebar" {
			return match
		}
		return ""
	})

	// Extract HTML blocks (like <div>...</div>) before processing to preserve them
	// Uses depth-tracking to handle nested divs correctly
	md, htmlBlocks := extractBalancedDivs(md)

	// Extract and preserve explicit anchor tags so they survive raw tag stripping.
	anchorPlaceholder := "\x00ANCHOR%d\x00"
	var anchorTags []string
	md = explicitAnchorPattern.ReplaceAllStringFunc(md, func(match string) string {
		placeholder := fmt.Sprintf(anchorPlaceholder, len(anchorTags))
		anchorTags = append(anchorTags, match)
		return placeholder
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

	// Restore explicit anchor tags now that raw tags have been stripped.
	for i, anchor := range anchorTags {
		md = strings.Replace(md, fmt.Sprintf(anchorPlaceholder, i), anchor, 1)
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
	introSidebarPending := false
	var blockquoteClassOverride blockquoteClass = -1

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
		originalLines := blockquoteLines
		blockquoteLines = nil
		text := strings.Join(originalLines, " ")
		if text == "" {
			return
		}

		var className string
		if compilerVersion == 2 {
			class, cleanedLines := classifyBlockquote(originalLines, sectionID, filePath, c)
			if blockquoteClassOverride >= 0 {
				class = blockquoteClassOverride
			}
			text = strings.Join(cleanedLines, " ")
			text = stripBlockquotePrefix(text, class)
			switch class {
			case bqDMSidebar:
				className = "dm-sidebar"
			case bqChapterSummary:
				className = "chapter-summary"
			case bqIntroductionSidebar:
				className = "introduction-sidebar"
			default:
				className = "read-aloud"
			}
		} else {
			// Legacy v1 / helper behavior: all blockquotes are read-aloud.
			text = readAloudPrefixRe.ReplaceAllString(text, "")
			className = "read-aloud"
		}

		escaped := processInlineText(text, baseDir, seenImages, reg)
		out = append(out, fmt.Sprintf(`<div class="%s">%s</div>`, className, escaped))
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
		escaped := processInlineText(text, baseDir, seenImages, reg)
		out = append(out, fmt.Sprintf("<p>%s</p>", escaped))
	}

	flushTable := func() {
		if len(tableRows) < 2 {
			// Not a valid table, treat rows as paragraphs
			for _, row := range tableRows {
				if strings.TrimSpace(row) != "" {
					escaped := processInlineText(strings.TrimSpace(row), baseDir, seenImages, reg)
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
				cellEsc := processInlineText(cell, baseDir, seenImages, reg)
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

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Skip horizontal rules (---, ***, ___, - - -)
		if trimmed == "---" || trimmed == "***" || trimmed == "___" || trimmed == "- - -" {
			continue
		}

		// Preserve introduction-sidebar marker comments for blockquote disambiguation.
		if compilerVersion == 2 && introductionSidebarMarkerRegex.MatchString(trimmed) {
			introSidebarPending = true
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
			if compilerVersion == 2 && introSidebarPending {
				blockquoteClassOverride = bqIntroductionSidebar
				introSidebarPending = false
			}
			blockquoteLines = append(blockquoteLines, bqMatch[1])
			continue
		}

		// If we were in a blockquote but now see non-blockquote content, flush it
		if inBlockquote {
			flushBlockquote()
			inBlockquote = false
			blockquoteClassOverride = -1
		}

		// Handle headings
		if strings.HasPrefix(trimmed, "##### ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "##### ")
			id := headingID(text, sectionID, compilerVersion, headingCounter, reg)
			out = append(out, fmt.Sprintf(`<h5 id="%s">%s</h5>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "#### ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "#### ")
			id := headingID(text, sectionID, compilerVersion, headingCounter, reg)
			out = append(out, fmt.Sprintf(`<h4 id="%s">%s</h4>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "### ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "### ")

			// v2: Area number highlighting and cross-reference IDs
			if compilerVersion == 2 {
				if areaMatch := areaHeadingPattern.FindStringSubmatch(text); areaMatch != nil {
					(*headingCounter)++
					areaNum := areaMatch[1]
					areaName := strings.TrimSpace(areaMatch[2])
					areaID := "area-" + areaNum
					rendered := fmt.Sprintf(`<span class="area-number">Área %s</span> %s`, areaNum, html.EscapeString(areaName))
					out = append(out, fmt.Sprintf(`<h3 id="%s">%s</h3>`, areaID, rendered))
					continue
				}
			}

			id := headingID(text, sectionID, compilerVersion, headingCounter, reg)
			out = append(out, fmt.Sprintf(`<h3 id="%s">%s</h3>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "## ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "## ")

			// NEW (REQ-2.1): peek-ahead for WotC stat block. If the next non-blank
			// line is "*<Size> <Type>*", enter the stat block sub-parser.
			if sb, consumed := tryStatBlock(text, lines, i, baseDir, seenImages, reg); consumed > 0 {
				out = append(out, sb)

				// REQ (fix-statblock-layout-and-cover-overflow): after a
				// stat block, peek for a `monster-{name}.png` hero image
				// and hoist it inside the just-emitted <div class="stat-block">.
				// Convention: any author adding a monster image should name
				// the file `monster-{name}.png` and place it directly after
				// the stat block (typically after the closing --- rule). The
				// convention guard (`monster-` prefix) prevents scene / npc /
				// cover images from being hoisted into the wrong block.
				if imgHTML, newConsumed := peekHoistableMonsterImage(consumed, lines, baseDir, seenImages); imgHTML != "" {
					if len(out) > 0 && strings.HasSuffix(out[len(out)-1], "</div>") {
						out[len(out)-1] = strings.TrimSuffix(out[len(out)-1], "</div>") + imgHTML + "</div>"
						consumed = newConsumed
					}
				}

				// Parser returns the absolute index of the line AFTER the block.
				// Set i = consumed - 1 so the loop's own i++ lands on `consumed`.
				i = consumed - 1
				continue
			}

			id := headingID(text, sectionID, compilerVersion, headingCounter, reg)
			out = append(out, fmt.Sprintf(`<h2 id="%s">%s</h2>`, id, html.EscapeString(text)))
		} else if strings.HasPrefix(trimmed, "# ") {
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			text := strings.TrimPrefix(trimmed, "# ")
			id := headingID(text, sectionID, compilerVersion, headingCounter, reg)
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
			escaped := processInlineText(text, baseDir, seenImages, reg)
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
		} else if htmlBlockPlaceholderRegex.MatchString(trimmed) {
			// HTML block placeholder — flush paragraph and output directly without <p> wrapper
			flushParagraph()
			if inList {
				out = append(out, "</ul>")
				inList = false
			}
			out = append(out, trimmed)
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
	// Re-process inline markdown (`![alt](path)`, `[text](url)`) inside
	// extracted HTML blocks so images and links reach the printer even
	// when the author pre-rendered HTML wraps raw markdown. See
	// visual-issues-pdf PR 2 / REQ-2.1.
	//
	// We call processImages + processLinks directly instead of
	// processInlineText because processInlineText HTML-escapes its input
	// (after stashing <img>/<a>). For pre-rendered HTML blocks the
	// escape step would corrupt attributes like <div class="stat-block">.
	// processImages + processLinks only convert raw markdown patterns
	// to tags; everything else (including pre-rendered <div>, <h2>, <em>,
	// etc.) passes through unchanged.
	result := strings.Join(out, "\n")
	for i, html := range htmlBlocks {
		processed := processImages(html, baseDir, seenImages)
		processed = processLinks(processed, "", reg)
		result = strings.Replace(result, fmt.Sprintf("\x00HTMLBLOCK%d\x00", i), processed, 1)
	}

	return result
}

// headingID returns the emitted id for a heading. In v2 it prefers a stable
// slug-based ID from the registry when it matches the current section; otherwise
// it falls back to the legacy counter-based ID.
func headingID(text, sectionID string, compilerVersion int, headingCounter *int, reg anchorRegistry) string {
	*headingCounter++
	counterID := sectionID + "-h" + strconv.Itoa(*headingCounter)

	if compilerVersion != 2 || reg == nil {
		return counterID
	}

	slug := slugify(text)
	if slug == "" {
		return counterID
	}
	stableID := sectionID + "-" + slug
	if reg[slug] == stableID {
		return stableID
	}
	return counterID
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
	return markdownToHTMLWithID(nil, md, baseDir, "content", &counter, seen, 0, nil, "")
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

func processLinks(text, sectionID string, reg anchorRegistry) string {
	return linkRegex.ReplaceAllStringFunc(text, func(match string) string {
		matches := linkRegex.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		linkText := matches[1]
		href := matches[2]

		resolved := href
		if strings.Contains(href, ".md#") {
			parts := strings.SplitN(href, "#", 2)
			if len(parts) == 2 {
				frag := parts[1]
				if reg != nil && reg[frag] != "" {
					resolved = "#" + reg[frag]
				} else {
					resolved = "#" + frag
				}
			}
		} else if strings.HasPrefix(href, "#") {
			frag := href[1:]
			if reg != nil && reg[frag] != "" {
				resolved = "#" + reg[frag]
			}
		}

		return fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(resolved), html.EscapeString(linkText))
	})
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

	// SVG short-circuit: must come BEFORE magic-byte scan (REQ-1.6).
	// SVG is embedded as a relative path so Chromium can inline-render it.
	if ext == ".svg" {
		relPath, _ := filepath.Rel(baseDir, imgPath)
		if relPath == "" {
			relPath = imgPath
		}
		return fmt.Sprintf(`<img src="%s" alt="%s" class="campaign-image"/>`, html.EscapeString(relPath), html.EscapeString(alt))
	}

	// Detect MIME from magic bytes FIRST. This fixes the bug where `.png`
	// files in assets/ are actually JPEG bytes (REQ-1.1 through 1.5).
	var mimeType string
	if mt, ok := detectMimeType(data); ok {
		mimeType = mt
	} else {
		// Fallback to extension-derived MIME for ambiguous / unknown bytes.
		switch ext {
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
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)
	return fmt.Sprintf(`<img src="%s" alt="%s" class="campaign-image"/>`, dataURI, html.EscapeString(alt))
}

// detectMimeType returns the MIME type inferred from the first 4-12 bytes of data.
// Returns ("", false) when no signature matches; caller must then fall back to
// extension-derived MIME. Recognized signatures: PNG, JPEG, GIF87a/89a, WEBP.
func detectMimeType(data []byte) (mimeType string, detected bool) {
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return "image/png", true
	}
	// JPEG: FF D8 FF
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{0xFF, 0xD8, 0xFF}) {
		return "image/jpeg", true
	}
	// GIF: "GIF87a" or "GIF89a"
	if len(data) >= 6 && (bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a"))) {
		return "image/gif", true
	}
	// WebP: "RIFF" .... "WEBP" (12 bytes covers the 4-byte size field)
	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return "image/webp", true
	}
	return "", false
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
