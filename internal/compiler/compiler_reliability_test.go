package compiler

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

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
