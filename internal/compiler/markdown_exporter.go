package compiler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MarkdownExporter concatenates campaign markdown files in canonical order.
type MarkdownExporter struct{}

// NewMarkdownExporter creates a new markdown exporter.
func NewMarkdownExporter() *MarkdownExporter {
	return &MarkdownExporter{}
}

// Format returns the exporter format identifier.
func (e *MarkdownExporter) Format() string {
	return "markdown"
}

// Export walks the campaign directory in canonical order and joins markdown with \n---\n.
func (e *MarkdownExporter) Export(ctx context.Context, campaignDir, title string) (string, error) {
	if _, err := os.Stat(campaignDir); err != nil {
		return "", fmt.Errorf("campaign directory not found: %w", err)
	}

	var sections []sectionRef
	sections = append(sections, sectionRef{name: "Introduction", path: filepath.Join(campaignDir, "introduction.md")})
	sections = append(sections, sectionRef{name: "Lore", path: filepath.Join(campaignDir, "lore.md")})
	sections = append(sections, sectionRef{name: "Setting Guide", path: filepath.Join(campaignDir, "setting-guide.md")})

	// Chapters directory
	chaptersDir := filepath.Join(campaignDir, "chapters")
	if info, err := os.Stat(chaptersDir); err == nil && info.IsDir() {
		files, err := os.ReadDir(chaptersDir)
		if err == nil {
			var chapterFiles []string
			for _, f := range files {
				if strings.HasSuffix(f.Name(), ".md") {
					chapterFiles = append(chapterFiles, f.Name())
				}
			}
			sort.Strings(chapterFiles)
			for _, cf := range chapterFiles {
				sections = append(sections, sectionRef{name: "Chapter: " + cf, path: filepath.Join(chaptersDir, cf)})
			}
		}
	}

	sections = append(sections, sectionRef{name: "NPCs", path: filepath.Join(campaignDir, "npcs", "npcs_and_factions.md")})
	sections = append(sections, sectionRef{name: "Bestiary", path: filepath.Join(campaignDir, "bestiary", "bestiary.md")})
	sections = append(sections, sectionRef{name: "Encounters", path: filepath.Join(campaignDir, "encounters", "encounters.md")})
	sections = append(sections, sectionRef{name: "Maps", path: filepath.Join(campaignDir, "maps", "maps.md")})
	sections = append(sections, sectionRef{name: "Quests", path: filepath.Join(campaignDir, "quests", "quests.md")})
	sections = append(sections, sectionRef{name: "Appendices", path: filepath.Join(campaignDir, "appendices.md")})

	var parts []string
	for _, sec := range sections {
		content, err := os.ReadFile(sec.path)
		if err != nil {
			continue // skip missing sections
		}
		if strings.TrimSpace(string(content)) != "" {
			parts = append(parts, strings.TrimSpace(string(content)))
		}
	}

	if len(parts) == 0 {
		return "", fmt.Errorf("no markdown content found in campaign directory")
	}

	outputPath := filepath.Join(campaignDir, "campaign.md")
	output := strings.Join(parts, "\n\n---\n\n")
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("failed to write markdown export: %w", err)
	}

	return outputPath, nil
}

type sectionRef struct {
	name string
	path string
}
