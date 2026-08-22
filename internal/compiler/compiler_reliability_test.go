package compiler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestClassifyTableRows(t *testing.T) {
	longCell := strings.Repeat("visible ", 15)
	longToken := strings.Repeat("x", 32)
	longRow := "| " + strings.Repeat("content ", 20) + " | short |"
	tests := []struct {
		name        string
		rows        []string
		wantComplex bool
		wantColumns int
		wantMedia   bool
		wantCode    bool
	}{
		{
			name:        "compact regular table remains simple",
			rows:        []string{"| Name | Role | Note |", "| --- | --- | --- |", "| Mira | Guide | Short |"},
			wantColumns: 3,
		},
		{
			name:        "four columns are complex",
			rows:        []string{"| A | B | C | D |", "| --- | --- | --- | --- |", "| 1 | 2 | 3 | 4 |"},
			wantComplex: true,
			wantColumns: 4,
		},
		{
			name:        "long visible cell is complex",
			rows:        []string{"| Name | Note |", "| --- | --- |", "| Mira | " + longCell + " |"},
			wantComplex: true,
			wantColumns: 2,
		},
		{
			name:        "indivisible token is complex",
			rows:        []string{"| Name | Note |", "| --- | --- |", "| Mira | " + longToken + " |"},
			wantComplex: true,
			wantColumns: 2,
		},
		{
			name:        "image and inline code are complex",
			rows:        []string{"| Asset | Snippet |", "| --- | --- |", "| ![map](assets/map.png) | `roll 1d6` |"},
			wantComplex: true,
			wantColumns: 2,
			wantMedia:   true,
			wantCode:    true,
		},
		{
			name: "estimated tall table is complex",
			rows: append([]string{"| Name | Note |", "| --- | --- |"}, func() []string {
				rows := make([]string, 0, 20)
				for i := 0; i < 20; i++ {
					rows = append(rows, longRow)
				}
				return rows
			}()...),
			wantComplex: true,
			wantColumns: 2,
		},
		{
			name:        "inconsistent row widths are complex",
			rows:        []string{"| Name | Role |", "| --- | --- |", "| Mira | Guide | Extra |"},
			wantComplex: true,
			wantColumns: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := classifyTableRows(tt.rows)
			if got := profile.columnCount; got != tt.wantColumns {
				t.Fatalf("columnCount = %d, want %d (profile %#v)", got, tt.wantColumns, profile)
			}
			if got := profile.isComplex(); got != tt.wantComplex {
				t.Errorf("isComplex() = %v, want %v (profile %#v)", got, tt.wantComplex, profile)
			}
			if profile.hasMedia != tt.wantMedia {
				t.Errorf("hasMedia = %v, want %v", profile.hasMedia, tt.wantMedia)
			}
			if profile.hasCode != tt.wantCode {
				t.Errorf("hasCode = %v, want %v", profile.hasCode, tt.wantCode)
			}
		})
	}
}

func TestFlushTableAdaptiveHTML(t *testing.T) {
	simple := markdownToHTML("| Name | Role |\n| --- | --- |\n| Mira | Guide |", ".")
	if strings.Contains(simple, "table-island") {
		t.Fatalf("simple table was promoted: %s", simple)
	}
	if got := strings.Count(simple, `<div class="table-wrap">`); got != 1 {
		t.Fatalf("simple table wrapper count = %d, want 1: %s", got, simple)
	}

	complex := markdownToHTML("before\n\n| A | B | C | D |\n| --- | --- | --- | --- |\n| 1 | 2 | 3 | 4 |\n\nafter", ".")
	wrapper := `<div class="table-wrap table-island"><table>`
	if !strings.Contains(complex, wrapper) {
		t.Fatalf("complex table missing island wrapper %q: %s", wrapper, complex)
	}
	boundary := `<div class="table-island-boundary" aria-hidden="true"></div>`
	if !strings.Contains(complex, `</table></div>`+boundary) {
		t.Fatalf("complex table missing immediate sibling boundary %q: %s", boundary, complex)
	}
	if !strings.Contains(complex, "1</td>") || !strings.Contains(complex, "4</td>") {
		t.Fatalf("complex table lost cells: %s", complex)
	}
	if strings.Index(complex, "before") > strings.Index(complex, wrapper) || strings.Index(complex, wrapper) > strings.Index(complex, "after") {
		t.Fatalf("table changed document order: %s", complex)
	}
}

func TestNestedTableDoesNotBecomeIsland(t *testing.T) {
	got := markdownToHTML("> | A | B | C | D |\n> | --- | --- | --- | --- |\n> | 1 | 2 | 3 | 4 |", ".")
	if !strings.Contains(got, `class="read-aloud"`) {
		t.Fatalf("nested table lost its callout: %s", got)
	}
	if strings.Contains(got, "table-island") || strings.Contains(got, "table-island-boundary") {
		t.Fatalf("nested table was promoted outside its callout: %s", got)
	}
}

func TestRegexes(t *testing.T) {
	tests := []struct {
		name     string
		re       *regexp.Regexp
		input    string
		want     bool
		captures map[string]string
	}{
		{
			name:     "link internal",
			re:       linkRegex,
			input:    "[Background](#adventure-background)",
			want:     true,
			captures: map[string]string{"text": "Background", "href": "#adventure-background"},
		},
		{
			name:     "link cross-file",
			re:       linkRegex,
			input:    "[Appendix A](appendices.md#appendix-a-magic-items)",
			want:     true,
			captures: map[string]string{"text": "Appendix A", "href": "appendices.md#appendix-a-magic-items"},
		},
		{
			name:  "link no match",
			re:    linkRegex,
			input: "not a link",
			want:  false,
		},
		{
			name:  "dm sidebar prefix heading",
			re:    dmSidebarPrefixRe,
			input: "##### DM Sidebar: Title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix bold",
			re:    dmSidebarPrefixRe,
			input: "**DM Sidebar:** Title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix case insensitive",
			re:    dmSidebarPrefixRe,
			input: "##### dm sidebar: Title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix no colon bare",
			re:    dmSidebarPrefixRe,
			input: "DM Sidebar Title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix no colon heading",
			re:    dmSidebarPrefixRe,
			input: "##### DM Sidebar Title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix no colon bolded",
			re:    dmSidebarPrefixRe,
			input: "**DM Sidebar** Title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix no colon mixed case",
			re:    dmSidebarPrefixRe,
			input: "##### Dm Sidebar title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix no colon lower",
			re:    dmSidebarPrefixRe,
			input: "##### dm sidebar: title",
			want:  true,
		},
		{
			name:  "dm sidebar prefix no colon leading whitespace",
			re:    dmSidebarPrefixRe,
			input: "  ##### DM Sidebar Title",
			want:  false,
		},
		{
			name:  "read-aloud text prefix",
			re:    readAloudPrefixRe,
			input: "**Read-Aloud Text:** description",
			want:  true,
		},
		{
			name:  "read-aloud prefix",
			re:    readAloudPrefixRe,
			input: "**Read-Aloud:** description",
			want:  true,
		},
		{
			name:  "para leer en voz alta prefix",
			re:    readAloudPrefixRe,
			input: "**Para Leer en Voz Alta:** descripción",
			want:  true,
		},
		{
			name:  "worksheet div",
			re:    worksheetDivRe,
			input: `<div class="character-worksheet">`,
			want:  true,
		},
		{
			name:  "worksheet div with extra attrs",
			re:    worksheetDivRe,
			input: `<div class="character-worksheet" data-form="creation">`,
			want:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tt.re.FindStringSubmatch(tt.input)
			got := matches != nil
			if got != tt.want {
				t.Fatalf("%s: match = %v, want %v", tt.name, got, tt.want)
			}
			if !tt.want {
				return
			}
			for name, want := range tt.captures {
				idx := tt.re.SubexpIndex(name)
				if idx < 0 {
					t.Fatalf("%s: capture group %q not found", tt.name, name)
				}
				if idx >= len(matches) {
					t.Fatalf("%s: capture group %q out of range", tt.name, name)
				}
				if matches[idx] != want {
					t.Errorf("%s: capture %q = %q, want %q", tt.name, name, matches[idx], want)
				}
			}
		})
	}
}

func TestStripCharacterWorksheets(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantNotContains string
		wantContains    string
	}{
		{
			name:            "removes simple worksheet div",
			input:           `<div class="character-worksheet">content</div>`,
			wantNotContains: `<div class="character-worksheet">`,
			wantContains:    "",
		},
		{
			name: "removes nested worksheet div",
			input: `<div class="character-worksheet">
<div class="worksheet-section">
<h4>Prompt</h4>
<div class="prompt-box">Question?</div>
</div>
</div>`,
			wantNotContains: `<div class="character-worksheet">`,
			wantContains:    "",
		},
		{
			name:            "preserves non-worksheet divs",
			input:           `<div class="dm-sidebar">tip</div><div class="character-worksheet">ws</div>`,
			wantNotContains: `<div class="character-worksheet">`,
			wantContains:    `<div class="dm-sidebar">`,
		},
		{
			name:            "preserves text around worksheet",
			input:           "Before\n<div class=\"character-worksheet\">WS</div>\nAfter",
			wantNotContains: `<div class="character-worksheet">`,
			wantContains:    "After",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripCharacterWorksheets(tt.input)
			if tt.wantNotContains != "" && strings.Contains(got, tt.wantNotContains) {
				t.Errorf("stripCharacterWorksheets() should not contain %q\nGot: %s", tt.wantNotContains, got)
			}
			if tt.wantContains != "" && !strings.Contains(got, tt.wantContains) {
				t.Errorf("stripCharacterWorksheets() should contain %q\nGot: %s", tt.wantContains, got)
			}
		})
	}
}

func TestMarkdownToHTMLWithID_StripsWorksheetsV2(t *testing.T) {
	md := `### Worksheet de Creación

<div class="character-worksheet">
<div class="worksheet-section">
<h4>Prompt</h4>
</div>
</div>

After worksheet.
`
	result := markdownToHTMLWithID(nil, md, "/tmp", "sec-test", new(int), make(map[string]bool), 2, nil, "")
	if strings.Contains(result, `<div class="character-worksheet">`) {
		t.Errorf("v2 should strip character-worksheet divs, got: %s", result)
	}
	if !strings.Contains(result, "After worksheet.") {
		t.Errorf("v2 should preserve surrounding text, got: %s", result)
	}
}

func TestMarkdownToHTMLWithID_KeepsWorksheetsV1(t *testing.T) {
	md := `<div class="character-worksheet">
<div class="worksheet-section">
<h4>Prompt</h4>
</div>
</div>
`
	result := markdownToHTMLWithID(nil, md, "/tmp", "sec-test", new(int), make(map[string]bool), 1, nil, "")
	if !strings.Contains(result, `<div class="character-worksheet">`) {
		t.Errorf("v1 should keep character-worksheet divs, got: %s", result)
	}
}

func TestClassifyBlockquote(t *testing.T) {
	tests := []struct {
		name      string
		sectionID string
		lines     []string
		wantClass blockquoteClass
		wantLines []string
	}{
		{
			name:      "read-aloud",
			sectionID: "sec-chapter-test",
			lines:     []string{"**Read-Aloud:** The room is dark."},
			wantClass: bqReadAloud,
			wantLines: []string{"The room is dark."},
		},
		{
			name:      "read-aloud text",
			sectionID: "sec-chapter-test",
			lines:     []string{"**Read-Aloud Text:** The room is dark."},
			wantClass: bqReadAloud,
			wantLines: []string{"The room is dark."},
		},
		{
			name:      "para leer en voz alta",
			sectionID: "sec-chapter-test",
			lines:     []string{"**Para Leer en Voz Alta:** La habitación es oscura."},
			wantClass: bqReadAloud,
			wantLines: []string{"La habitación es oscura."},
		},
		{
			name:      "dm sidebar heading",
			sectionID: "sec-chapter-test",
			lines:     []string{"##### DM Sidebar: Traps", "The floor is trapped."},
			wantClass: bqDMSidebar,
			wantLines: []string{"Traps", "The floor is trapped."},
		},
		{
			name:      "dm sidebar bold",
			sectionID: "sec-chapter-test",
			lines:     []string{"**DM Sidebar:** Secrets", "A hidden door."},
			wantClass: bqDMSidebar,
			wantLines: []string{"Secrets", "A hidden door."},
		},
		{
			name:      "introduction sidebar",
			sectionID: "sec-introduction",
			lines:     []string{"##### Optional Rule: Bonds", "Bonds connect characters."},
			wantClass: bqIntroductionSidebar,
			wantLines: []string{"##### Optional Rule: Bonds", "Bonds connect characters."},
		},
		{
			name:      "chapter summary",
			sectionID: "sec-chapters",
			lines:     []string{"- Level 1-2", "- 2-3 hours"},
			wantClass: bqChapterSummary,
			wantLines: []string{"- Level 1-2", "- 2-3 hours"},
		},
		{
			name:      "chapter summary with mixed text falls back to read-aloud",
			sectionID: "sec-chapters",
			lines:     []string{"Level 1-2", "Duration: 2 hours"},
			wantClass: bqReadAloud,
			wantLines: []string{"Level 1-2", "Duration: 2 hours"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClass, gotLines := classifyBlockquote(tt.lines, tt.sectionID, "", nil)
			if gotClass != tt.wantClass {
				t.Errorf("classifyBlockquote() class = %v, want %v", gotClass, tt.wantClass)
			}
			if len(gotLines) != len(tt.wantLines) {
				t.Fatalf("classifyBlockquote() lines = %v, want %v", gotLines, tt.wantLines)
			}
			for i := range gotLines {
				if gotLines[i] != tt.wantLines[i] {
					t.Errorf("classifyBlockquote() line[%d] = %q, want %q", i, gotLines[i], tt.wantLines[i])
				}
			}
		})
	}
}

func TestImageFirstMonsterBlocksStayCoherent(t *testing.T) {
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"monster-gromerm.png", "monster-error.png", "monster-guardian.png"} {
		if err := os.WriteFile(filepath.Join(assetsDir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	md := `## Gromerm
![Gromerm](assets/monster-gromerm.png)
*Medium beast, unaligned*
**Armor Class** 14

## El Error
![El Error](assets/monster-error.png)
*Small aberration, unaligned*
**Armor Class** 12

## Guardián de Arkanum
![Guardián](assets/monster-guardian.png)
*Large construct, unaligned*
**Armor Class** 16
`
	got := markdownToHTML(md, tmpDir)
	if gotBlocks := strings.Count(got, `class="stat-block"`); gotBlocks != 3 {
		t.Fatalf("image-first monsters should produce three stat blocks, got %d\n%s", gotBlocks, got)
	}
	for _, marker := range []string{"monster-gromerm.png", "monster-error.png", "monster-guardian.png"} {
		if gotImages := strings.Count(got, marker); gotImages != 1 {
			t.Errorf("image %q should occur once in its stat block, got %d", marker, gotImages)
		}
	}
}

func TestImageFirstMonsterStopsAtNonMonsterAndDeduplicates(t *testing.T) {
	tmpDir := t.TempDir()
	assetsDir := filepath.Join(tmpDir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"monster-echo.png", "scene.png"} {
		if err := os.WriteFile(filepath.Join(assetsDir, name), []byte(name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	md := `## Echo
![Echo](assets/monster-echo.png)

*Tiny construct, unaligned*
**Armor Class** 10
---
![Echo duplicate](assets/monster-echo.png)
![Scene](assets/scene.png)
`
	got := markdownToHTML(md, tmpDir)
	if blocks := strings.Count(got, `class="stat-block"`); blocks != 1 {
		t.Fatalf("expected one stat block, got %d\n%s", blocks, got)
	}
	if occurrences := strings.Count(got, "monster-echo.png"); occurrences != 1 {
		t.Fatalf("duplicate monster image should be emitted once, got %d\n%s", occurrences, got)
	}
	if !strings.Contains(got, `alt="Scene"`) {
		t.Fatalf("non-monster image should remain outside the stat block\n%s", got)
	}
}

func TestRosterExtractionExcludesEncounterSolutions(t *testing.T) {
	md := `## Adventure Roster

### NPCs
- **Mira** — Guide

### Encounter: The Bridge
- **Mira** attacks from the west.

### Solution
- **Mira** is secretly the traitor.
`
	npcs, _, encounters := extractRosterEntries(md)
	if len(npcs) != 1 || npcs[0] != "Mira|Guide" {
		t.Fatalf("roster should contain only the bounded NPC row, got %v", npcs)
	}
	if len(encounters) != 0 {
		t.Fatalf("encounter prose and solution rows must not become roster entries, got %v", encounters)
	}
}

func TestParseRosterSectionUsesTypedBoundedRows(t *testing.T) {
	md := `## Adventure Roster

### NPCs
- **Mira** — Guide

### Monstruos
- **Goblin** (CR 1/4)

### Encuentros
- **The Bridge** — Social

### Encounter: prose is not a roster section
- **Mira** attacks from the west.

### Solution
- **Mira** is secretly the traitor.
`
	entries := parseRosterSection(md)
	if len(entries) != 3 {
		t.Fatalf("expected three bounded roster entries, got %d: %#v", len(entries), entries)
	}
	want := []struct {
		category rosterCategory
		name     string
		detail   string
	}{
		{rosterNPC, "Mira", "Guide"},
		{rosterMonster, "Goblin", "(CR 1/4)"},
		{rosterEncounter, "The Bridge", "Social"},
	}
	for i, expected := range want {
		if entries[i].category != expected.category || entries[i].name != expected.name || entries[i].detail != expected.detail {
			t.Errorf("entry %d = %#v, want category=%q name=%q detail=%q", i, entries[i], expected.category, expected.name, expected.detail)
		}
	}
}

func TestParseRosterSectionRejectsUnrecognizedRows(t *testing.T) {
	md := `# Notes
## Encounter: The Bridge
- **Mira** attacks from the west.

## Solution
- **Mira** is secretly the traitor.

| Name | Role |
| --- | --- |
| Fabricated | Row |
`
	if entries := parseRosterSection(md); len(entries) != 0 {
		t.Fatalf("unrecognized prose and arbitrary tables must not create roster rows: %#v", entries)
	}
}

func TestParseRosterSectionAcceptsNameOnlyNPCIdentity(t *testing.T) {
	entries := parseRosterSection("## NPCs\n- **Ivo**\n")
	if len(entries) != 1 || entries[0].category != rosterNPC || entries[0].name != "Ivo" {
		t.Fatalf("name-only NPC identity bullet should be retained, got %#v", entries)
	}
}

func TestDMSidebarWideUsesExplicitWideClass(t *testing.T) {
	headingCounter := 0
	got := markdownToHTMLWithID(nil, "> DM Sidebar Wide: Use the full page for this reference.\n", t.TempDir(), "sec-test", &headingCounter, make(map[string]bool), 2, nil, "")
	if !strings.Contains(got, `<div class="dm-sidebar-wide">`) {
		t.Fatalf("explicit wide DM sidebar should use the wide class, got:\n%s", got)
	}
	if strings.Contains(got, `<div class="dm-sidebar">`) {
		t.Fatalf("wide DM sidebar must not fall back to the narrow class, got:\n%s", got)
	}
}

func TestStripBlockquotePrefix(t *testing.T) {
	tests := []struct {
		name  string
		class blockquoteClass
		text  string
		want  string
	}{
		{
			name:  "read-aloud already clean",
			class: bqReadAloud,
			text:  "The room is dark.",
			want:  "The room is dark.",
		},
		{
			name:  "dm sidebar strips heading marker",
			class: bqDMSidebar,
			text:  "##### Traps The floor is trapped.",
			want:  "Traps The floor is trapped.",
		},
		{
			name:  "introduction sidebar strips heading marker",
			class: bqIntroductionSidebar,
			text:  "##### Optional Rule Bonds connect characters.",
			want:  "Optional Rule Bonds connect characters.",
		},
		{
			name:  "chapter summary unchanged",
			class: bqChapterSummary,
			text:  "- Level 1-2 - 2-3 hours",
			want:  "- Level 1-2 - 2-3 hours",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripBlockquotePrefix(tt.text, tt.class)
			if got != tt.want {
				t.Errorf("stripBlockquotePrefix() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProcessLinks(t *testing.T) {
	tests := []struct {
		name string
		text string
		reg  anchorRegistry
		want string
	}{
		{
			name: "internal anchor",
			text: "See [Background](#adventure-background) for context.",
			reg:  nil,
			want: `See <a href="#adventure-background">Background</a> for context.`,
		},
		{
			name: "cross-file link resolved",
			text: "See [Appendix A](appendices.md#appendix-a-magic-items).",
			reg:  anchorRegistry{"appendix-a-magic-items": "appendix-a-magic-items"},
			want: `See <a href="#appendix-a-magic-items">Appendix A</a>.`,
		},
		{
			name: "cross-file link remapped",
			text: "See [Appendix A](appendices.md#appendix-a-magic-items).",
			reg:  anchorRegistry{"appendix-a-magic-items": "appendix-magic"},
			want: `See <a href="#appendix-magic">Appendix A</a>.`,
		},
		{
			name: "cross-file link unresolved falls back to fragment",
			text: "See [Missing](other.md#missing-thing).",
			reg:  anchorRegistry{},
			want: `See <a href="#missing-thing">Missing</a>.`,
		},
		{
			name: "plain text unchanged",
			text: "No links here.",
			reg:  nil,
			want: "No links here.",
		},
		{
			name: "multiple links",
			text: "[First](#one) and [Second](file.md#two)",
			reg:  anchorRegistry{"two": "section-two"},
			want: `<a href="#one">First</a> and <a href="#section-two">Second</a>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processLinks(tt.text, "", tt.reg)
			if got != tt.want {
				t.Errorf("processLinks() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildAnchorRegistry(t *testing.T) {
	tmpDir := t.TempDir()

	// introduction.md with explicit anchor and heading
	intro := `# Introduction

<a id="adventure-background"></a>
## Adventure Background

Some text.
`
	_ = os.WriteFile(filepath.Join(tmpDir, "introduction.md"), []byte(intro), 0644)

	// chapters/chapter_01.md with area and regular heading
	_ = os.MkdirAll(filepath.Join(tmpDir, "chapters"), 0755)
	chapter := `# Chapter 1

### Area 5: The Crypt

## Treasures

Content.
`
	_ = os.WriteFile(filepath.Join(tmpDir, "chapters", "chapter_01.md"), []byte(chapter), 0644)

	reg := buildAnchorRegistry(tmpDir)

	if got, want := reg["adventure-background"], "adventure-background"; got != want {
		t.Errorf("adventure-background = %q, want %q", got, want)
	}
	if got, want := reg["area-5"], "area-5"; got != want {
		t.Errorf("area-5 = %q, want %q", got, want)
	}
	if got := reg["adventure-background"]; got == "" {
		t.Error("expected adventure-background in registry")
	}
	if got := reg["treasures"]; got == "" {
		t.Error("expected treasures in registry")
	}
}

func TestMarkdownToHTMLWithID_StableHeadingIDs(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		version         int
		wantContains    string
		wantNotContains string
	}{
		{
			name:         "v2 area heading has area id",
			input:        "### Área 5: The Crypt\n\nContent.",
			version:      2,
			wantContains: `<h3 id="area-5">`,
		},
		{
			name:         "v2 regular heading has stable slug id",
			input:        "## Adventure Background\n\nContent.",
			version:      2,
			wantContains: `<h2 id="sec-content-adventure-background">`,
		},
		{
			name:            "v1 regular heading uses counter id",
			input:           "## Adventure Background\n\nContent.",
			version:         1,
			wantContains:    `<h2 id="sec-content-h1">`,
			wantNotContains: `adventure-background`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := anchorRegistry{}
			if tt.version == 2 {
				reg = anchorRegistry{"adventure-background": "sec-content-adventure-background"}
			}
			result := markdownToHTMLWithID(nil, tt.input, "/tmp", "sec-content", new(int), make(map[string]bool), tt.version, reg, "")
			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("markdownToHTMLWithID() should contain %q\nGot: %s", tt.wantContains, result)
			}
			if tt.wantNotContains != "" && strings.Contains(result, tt.wantNotContains) {
				t.Errorf("markdownToHTMLWithID() should not contain %q\nGot: %s", tt.wantNotContains, result)
			}
		})
	}
}

func TestMarkdownToHTMLWithID_CrossFileLink(t *testing.T) {
	input := "See [Appendix A](appendices.md#appendix-a-magic-items) for items."
	reg := anchorRegistry{"appendix-a-magic-items": "appendix-a-magic-items"}
	result := markdownToHTMLWithID(nil, input, "/tmp", "sec-test", new(int), make(map[string]bool), 2, reg, "")
	want := `<a href="#appendix-a-magic-items">Appendix A</a>`
	if !strings.Contains(result, want) {
		t.Errorf("expected %q in result, got: %s", want, result)
	}
}

func TestCountImagesInMarkdownSources_IncludesAllSources(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]string{
		"session-zero.md":  "![Session Zero](assets/sz.png)",
		"introduction.md":  "![Intro](assets/intro.png)",
		"setting-guide.md": "![Setting](assets/setting.png)",
		"appendices.md":    "![Appendix](assets/appendix.png)",
		"lore.md":          "![Lore](assets/lore.png)",
	}

	for relPath, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, relPath), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	_ = os.MkdirAll(filepath.Join(tmpDir, "chapters"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "chapters", "ch1.md"), []byte("![Chapter](assets/chapter.png)"), 0644)

	c := New(tmpDir, "")
	got, err := c.countImagesInMarkdownSources()
	if err != nil {
		t.Fatal(err)
	}
	if got != 6 {
		t.Errorf("countImagesInMarkdownSources() = %d, want 6", got)
	}
}

func TestIntroductionSidebarMarker(t *testing.T) {
	md := `<!-- introduction-sidebar -->
> ##### Optional Rule: Bonds
> Bonds connect characters.
`
	result := markdownToHTMLWithID(nil, md, "/tmp", "sec-introduction", new(int), make(map[string]bool), 2, nil, "")
	if !strings.Contains(result, `<div class="introduction-sidebar">`) {
		t.Errorf("expected introduction-sidebar class, got: %s", result)
	}
	if strings.Contains(result, "introduction-sidebar") && strings.Contains(result, "<!-- introduction-sidebar -->") {
		t.Errorf("marker comment should not appear in output, got: %s", result)
	}
}

func TestGenerateHTML_CoverFooter(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "lore.md"), []byte("# Lore\n\nTest."), 0644)

	c := New(tmpDir, "")
	htmlParts, err := c.generateHTML("Test Campaign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	html := strings.Join(htmlParts, "\n")

	if !strings.Contains(html, `<div class="cover-wrapper" id="cover">`) {
		t.Errorf("expected cover-wrapper div, got: %s", html)
	}
	if !strings.Contains(html, `class="cover-footer"`) {
		t.Errorf("expected cover-footer class, got: %s", html)
	}
	if !strings.Contains(html, "Generated by Grimorio") {
		t.Errorf("expected generated-by footer text, got: %s", html)
	}
}

// TestClassifyBlockquote_NoColonIsDMSidebar asserts the no-colon DM Sidebar
// variant is classified as bqDMSidebar (REQ-1.4). After the regex relaxation
// in WU-5, all three of these inputs must classify as DM Sidebar.
func TestClassifyBlockquote_NoColonIsDMSidebar(t *testing.T) {
	tests := []struct {
		name  string
		input []string
	}{
		{"heading no colon", []string{"##### DM Sidebar Some content"}},
		{"bold no colon", []string{"**DM Sidebar** Some content"}},
		{"bare no colon", []string{"DM Sidebar: Some content"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Compiler{
				seenImages:          make(map[string]bool),
				warnedFilesDebounce: make(map[string]struct{}),
			}
			class, _ := classifyBlockquote(tt.input, "test-section", "test.md", c)
			if class != bqDMSidebar {
				t.Errorf("classifyBlockquote() class = %v, want bqDMSidebar (%v)", class, bqDMSidebar)
			}
		})
	}
}

// TestClassifyBlockquote_StderrDebouncePerFile asserts the stderr warning
// fires exactly once per file, not once per call (REQ-1.5).
// NOT t.Parallel(): shared os.Stderr.
func TestClassifyBlockquote_StderrDebouncePerFile(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	c := &Compiler{
		seenImages:          make(map[string]bool),
		warnedFilesDebounce: make(map[string]struct{}),
	}

	// First file: 3 no-colon lines → exactly 1 warning
	for i := 0; i < 3; i++ {
		_, _ = classifyBlockquote([]string{"##### DM Sidebar line A" + string(rune('0'+i))}, "sec", "file1.md", c)
	}
	// Second file: 2 no-colon lines → exactly 1 more warning
	for i := 0; i < 2; i++ {
		_, _ = classifyBlockquote([]string{"##### DM Sidebar line X" + string(rune('0'+i))}, "sec", "file2.md", c)
	}

	_ = w.Close()
	captured, _ := io.ReadAll(r)

	count := strings.Count(string(captured), "DM Sidebar prefix without")
	if count != 2 {
		t.Errorf("expected 2 warnings (one per file), got %d\noutput:\n%s", count, string(captured))
	}
}

// TestClassifyBlockquote_ColonVariantNoStderr asserts the colon variant
// produces NO warning (REQ-1.5: warning only for no-colon variant).
// NOT t.Parallel(): shared os.Stderr.
func TestClassifyBlockquote_ColonVariantNoStderr(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	c := &Compiler{
		seenImages:          make(map[string]bool),
		warnedFilesDebounce: make(map[string]struct{}),
	}
	for i := 0; i < 5; i++ {
		_, _ = classifyBlockquote([]string{"##### DM Sidebar: colon variant " + string(rune('A'+i))}, "sec", "file1.md", c)
	}
	_ = w.Close()
	captured, _ := io.ReadAll(r)
	if strings.Contains(string(captured), "DM Sidebar prefix without") {
		t.Errorf("colon variant should not produce warnings, got:\n%s", string(captured))
	}
}

// TestProcessInlineText_OnExtractedHTMLBlocks asserts that processInlineText
// converts raw markdown `![alt](path)` and `[text](url)` patterns found
// INSIDE pre-rendered HTML blocks into <img> and <a> tags. This is the
// unit-level proof that REQ-2.1's "re-process inline markdown inside
// extracted HTML blocks" is feasible. The end-to-end assertion is in
// TestMarkdownToHTMLWithID_InlineImageInHTMLBlock below.
//
// Asserts per the spec (REQ-2.1): "the output contains <img (or the
// relative path is embedded as a data URI / assets/... reference per
// the existing embedImage behavior)". embedImage produces a data URI,
// so the realistic check is `<img` + `class="campaign-image"`.
func TestProcessInlineText_OnExtractedHTMLBlocks(t *testing.T) {
	// The asset must exist on disk for processInlineText -> processImages
	// -> embedImage to return a real <img> tag. Without the file, the
	// function returns `<span class="image-missing">[Image: alt]</span>`
	// which still contains the alt but NOT <img or class="campaign-image".
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	// Minimal PNG header (1x1 transparent). embedImage only needs to
	// successfully read the file; the actual PNG content doesn't matter
	// for these substring assertions.
	pngHeader := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}
	for _, name := range []string{"cover.png", "inner.png", "banner.png", "x.png"} {
		if err := os.WriteFile(filepath.Join(assetsDir, name), pngHeader, 0644); err != nil {
			t.Fatalf("write asset %s: %v", name, err)
		}
	}

	tests := []struct {
		name     string
		block    string
		contains []string
	}{
		{
			name:     "image inside single div",
			block:    `<div class="prologue"><h2>Prologue</h2>` + "\n\n" + `![Cover](assets/cover.png)` + "\n\n" + `</div>`,
			contains: []string{`<img`, `class="campaign-image"`},
		},
		{
			name:     "image nested deep",
			block:    `<div class="outer"><div class="inner"><p>![Inner](assets/inner.png)</p></div></div>`,
			contains: []string{`<img`, `class="campaign-image"`},
		},
		{
			name:     "image only block",
			block:    `<div class="banner">![Banner](assets/banner.png)</div>`,
			contains: []string{`<img`, `class="campaign-image"`},
		},
		{
			name:     "image and link",
			block:    `<div class="mixed">![Image](assets/x.png) and [link](https://example.com)</div>`,
			contains: []string{`<img`, `class="campaign-image"`, `href="https://example.com"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := anchorRegistry{}
			seen := make(map[string]bool)
			out := processInlineText(tt.block, dir, seen, reg)
			for _, s := range tt.contains {
				if !strings.Contains(out, s) {
					t.Errorf("expected %q in output, got:\n%s", s, out)
				}
			}
		})
	}
}

// TestProcessInlineText_PureHTMLUnchanged is the idempotency regression
// test for REQ-2.1. processInlineText must be a no-op on pure HTML
// containing only tags it preserves (<img> and <a>). If this test
// fails, the re-processing in the restore loop would strip attributes
// or otherwise corrupt pre-rendered HTML blocks that happen to also
// contain <img> or <a> tags.
//
// NOTE on fixture choice: processInlineText stashes <img> and <a> tags
// via regex (compiler.go:1132-1139) and HTML-escapes everything else.
// A `<div class="stat-block">…</div>` fixture (as suggested in the
// original design) would have its attributes escaped and therefore
// NOT be byte-identical. This test uses the realistic pure-HTML case
// the function was actually designed to preserve: <img>/<a> tags
// passed through unchanged.
func TestProcessInlineText_PureHTMLUnchanged(t *testing.T) {
	block := `<img src="assets/x.png" alt="Cover"> and <a href="https://example.com">link</a>`
	reg := anchorRegistry{}
	seen := make(map[string]bool)
	out := processInlineText(block, t.TempDir(), seen, reg)
	if out != block {
		t.Errorf("pure HTML (img/a only) should be byte-identical\nwant: %s\ngot:  %s", block, out)
	}
}

// TestMarkdownToHTMLWithID_InlineImageInHTMLBlock is the END-TO-END
// assertion for REQ-2.1. A markdown file that mixes pre-rendered HTML
// with raw markdown image syntax inside the <div> must produce an <img>
// tag in the final HTML, not the raw `![alt](path)` text.
//
// This test exercises the full path: extractBalancedDivs stashes the
// block as a \x00HTMLBLOCK<N>\x00 placeholder, the parser emits the
// placeholder as-is, and the restore loop substitutes the block back.
// The fix (WU-1) wraps the restored block in processInlineText so the
// raw `![alt](path)` reaches the image pipeline.
func TestMarkdownToHTMLWithID_InlineImageInHTMLBlock(t *testing.T) {
	dir := t.TempDir()
	// Create the asset so embedImage can read it. The function returns
	// <span class="image-missing"> if the file is absent, which would
	// still contain "assets/cover.png" as alt text — but we want a real
	// <img> tag to prove the conversion happens, so we provide the file.
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	// Minimal valid PNG header (1x1 transparent).
	pngHeader := []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "cover.png"), pngHeader, 0644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	md := `<div class="prologue">

![Cover](assets/cover.png)

</div>
`
	headingCounter := 0
	seen := make(map[string]bool)
	result := markdownToHTMLWithID(nil, md, dir, "sec-prologue", &headingCounter, seen, 2, nil, "prologue.md")

	if strings.Contains(result, `![Cover](assets/cover.png)`) {
		t.Errorf("raw markdown image syntax must NOT survive in output, got:\n%s", result)
	}
	if !strings.Contains(result, `<img`) {
		t.Errorf("expected <img tag in output, got:\n%s", result)
	}
	// Per REQ-2.1, "the final HTML contains an <img> tag for the
	// referenced asset, not the raw `![alt](path)` text". The asset
	// is embedded as a base64 data URI by embedImage (the project's
	// design decision for self-contained PDFs); the alt text retains
	// the original "Cover" name. Asserting `alt="Cover"` is the
	// strongest proof that the image was actually processed (not
	// just escaped) because it shows the alt attribute survived
	// the full extract → restore → processInlineText pipeline.
	if !strings.Contains(result, `alt="Cover"`) {
		t.Errorf("expected alt=\"Cover\" in output (proves image was processed), got:\n%s", result)
	}
}

// TestGenerateAdventureRoster_Wrapper asserts REQ-2.2: the
// generateAdventureRoster output is wrapped in a single
// <div class="roster-wrap">…</div> element that contains the h2
// heading and all three tables (NPCs, Monstruos, Encuentros). The
// wrapper is the outermost element so the h2 inherits the
// column-span context of the .roster-wrap CSS rule (REQ-2.4).
func TestGenerateAdventureRoster_Wrapper(t *testing.T) {
	dir := t.TempDir()

	// Build a minimal campaign with one entry per table.
	// The fixture uses chapters/ because that's the canonical chapter
	// source per the v5.0.2 WU7 removal of the legacy areas/ dir.
	chaptersDir := filepath.Join(dir, "chapters")
	if err := os.MkdirAll(chaptersDir, 0755); err != nil {
		t.Fatalf("mkdir chapters: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chaptersDir, "npcs.md"), []byte(`
## NPCs

- **Eldrin** tavern keeper
`), 0644); err != nil {
		t.Fatalf("write npcs.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chaptersDir, "bestiary.md"), []byte(`
## Monstruos

- **Goblin** Challenge 1
`), 0644); err != nil {
		t.Fatalf("write bestiary.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(chaptersDir, "encounters.md"), []byte(`
## Encuentros

- **Test Encounter 1** A test encounter
`), 0644); err != nil {
		t.Fatalf("write encounters.md: %v", err)
	}

	c := New(dir, "")
	out := c.generateAdventureRoster()

	trimmed := strings.TrimSpace(out)
	if !strings.HasPrefix(trimmed, `<div class="roster-wrap">`) {
		t.Errorf("roster output should start with .roster-wrap, got prefix: %.80q", trimmed)
	}
	if !strings.HasSuffix(trimmed, `</div>`) {
		t.Errorf("roster output should end with </div>, got suffix: %.80q", trimmed[len(trimmed)-80:])
	}
	// Exactly one outer wrap (the function does not nest the wrapper).
	if got := strings.Count(out, `<div class="roster-wrap">`); got != 1 {
		t.Errorf("expected exactly 1 .roster-wrap, got %d. Output:\n%s", got, out)
	}
	if strings.Contains(out, "table-island") {
		t.Error("generated roster tables must not be promoted by the Markdown table classifier")
	}
	// All three h3 headings must be inside the wrapper. The h2
	// ("Apéndice F") and the three h3 headings are the proof that
	// the wrap did not eat the body of the function.
	for _, h3 := range []string{"NPCs", "Monstruos", "Encuentros"} {
		if !strings.Contains(out, h3) {
			t.Errorf("roster missing h3 %q. Output:\n%s", h3, out)
		}
	}
	// The h2 (Apéndice F: Adventure Roster) must also survive inside
	// the wrap.
	if !strings.Contains(out, "Apéndice F") {
		t.Errorf("roster missing h2 'Apéndice F'. Output:\n%s", out)
	}
}

// TestFlushTable_WrapsInTableWrap asserts REQ-3.1: every <table> emitted
// by flushTable is wrapped in a <div class="table-wrap">…</div> so
// Chromium's page-break algorithm can split at row boundaries instead
// of slicing the table across columns (Issue C).
func TestFlushTable_WrapsInTableWrap(t *testing.T) {
	tests := []struct {
		name   string
		md     string
		island bool
	}{
		{
			name: "simple 2-col table",
			md: `| A | B |
|---|---|
| 1 | 2 |
| 3 | 4 |
`,
		},
		{
			name: "3-col table",
			md: `| A | B | C |
|---|---|---|
| 1 | 2 | 3 |
`,
		},
		{
			name: "empty table",
			md: `| H1 | H2 |
|----|----|
`,
		},
		{
			name: "table with bold cell",
			md: `| A | B |
|---|---|
| **bold** | normal |
`,
		},
		{
			name:   "table with code cell",
			md:     "| A | B |\n|---|---|\n| `code` | normal |\n",
			island: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(t.TempDir(), "")
			seen := make(map[string]bool)
			out := c.markdownToHTMLWithID(tt.md, t.TempDir(), "test", new(int), seen, nil, "")
			// Walk every <table> substring in `out` and assert each is
			// immediately preceded by a <div class="table-wrap">.
			rest := out
			for {
				tblIdx := strings.Index(rest, "<table>")
				if tblIdx == -1 {
					break
				}
				// Look backwards from <table> for the nearest "<div" tag.
				divIdx := strings.LastIndex(rest[:tblIdx], "<div")
				if divIdx == -1 {
					t.Errorf("table at offset %d not preceded by <div>, output:\n%s", tblIdx, rest)
					return
				}
				// Verify the div has class="table-wrap" within the next 100 chars.
				probeEnd := divIdx + 100
				if probeEnd > len(rest) {
					probeEnd = len(rest)
				}
				blockStart := rest[divIdx:probeEnd]
				if !strings.Contains(blockStart, `class="table-wrap"`) && !strings.Contains(blockStart, `class="table-wrap table-island"`) {
					t.Errorf("table at offset %d not wrapped in .table-wrap, surrounding div: %q", tblIdx, blockStart)
				}
				if tt.island && !strings.Contains(blockStart, `class="table-wrap table-island"`) {
					t.Errorf("complex table should use an island wrapper, surrounding div: %q", blockStart)
				}
				// Advance past this <table>.
				rest = rest[tblIdx+len("<table>"):]
			}
		})
	}
}

// TestFlushTable_PreservesAllRows asserts REQ-3.5: a 10-row 3-column
// markdown table renders with all 10 <tr> elements intact (header + 10
// data rows = 11 <tr>). This is the unit companion of the gated
// TestCompile_TableIntegrity integration test.
func TestFlushTable_PreservesAllRows(t *testing.T) {
	md := `| Col1 | Col2 | Col3 |
|------|------|------|
` + generate10RowTable()
	c := New(t.TempDir(), "")
	seen := make(map[string]bool)
	out := c.markdownToHTMLWithID(md, t.TempDir(), "test", new(int), seen, nil, "")
	rowCount := strings.Count(out, "<tr>")
	if rowCount < 11 {
		t.Errorf("expected at least 11 <tr> (header + 10 rows), got %d. Output:\n%s", rowCount, out)
	}
}

// generate10RowTable builds a deterministic 10-row 3-column markdown
// table body for TestFlushTable_PreservesAllRows.
func generate10RowTable() string {
	rows := ""
	for i := 1; i <= 10; i++ {
		rows += fmt.Sprintf("| Row %d Col1 | Row %d Col2 | Row %d Col3 |\n", i, i, i)
	}
	return rows
}
