package main

import (
	"fmt"
	"time"
)

// ChangelogEntry represents a changelog entry.
type ChangelogEntry struct {
	Description string
	CommitHash  string
	IssueRefs   []string
	Author      string
}

// ChangelogVersion represents a version in the changelog.
type ChangelogVersion struct {
	Version   string
	Date      string
	Summary   string
	Added     []ChangelogEntry
	Changed   []ChangelogEntry
	Deprecated []ChangelogEntry
	Removed   []ChangelogEntry
	Fixed     []ChangelogEntry
	Security  []ChangelogEntry
}

// Changelog represents the full changelog.
type Changelog struct {
	ProjectName   string
	RepositoryURL string
	Versions      []ChangelogVersion
}

func main() {
	fromVersion := "v2.6.0"
	toVersion := "v3.0.0"

	changelog := generateChangelog(fromVersion, toVersion)
	markdown := formatChangelogMarkdown(changelog)

	fmt.Println(markdown)
}

func generateChangelog(fromVersion, toVersion string) Changelog {
	return Changelog{
		ProjectName:   "Grimorio",
		RepositoryURL: "https://github.com/pauvalls/grimorio",
		Versions: []ChangelogVersion{
			{
				Version: "3.0.0",
				Date:    time.Now().Format("2006-01-02"),
				Summary: "WotC Professional Quality - Major version bump with unified area format, milestone XP, enhanced handouts, and E2E testing",
				Added: []ChangelogEntry{
					{Description: "MilestoneXP domain model with PHB threshold tracking", CommitHash: "c7173a9"},
					{Description: "MagicItem domain model with rarity, attunement, and curse support", CommitHash: "c7173a9"},
					{Description: "Tactics domain model with intelligence tiers and environmental tactics", CommitHash: "c7173a9"},
					{Description: "PlayerMap domain model for player-facing maps with secret redaction", CommitHash: "c7173a9"},
					{Description: "SessionZeroGuide domain model with content warnings and safety tools", CommitHash: "c7173a9"},
					{Description: "ConsequenceTable domain model for act transition tracking", CommitHash: "c7173a9"},
					{Description: "Area unified WotC format with sequential numbering 1-15", CommitHash: "c7173a9"},
					{Description: "PregenCharacter with campaign-specific bonds/ideals/flaws", CommitHash: "c7173a9"},
					{Description: "MilestoneService for XP table generation and party level tracking", CommitHash: "06c22d4"},
					{Description: "ItemService for magic item generation with rarity validation", CommitHash: "06c22d4"},
					{Description: "TacticsService for enemy AI and combat guidance", CommitHash: "06c22d4"},
					{Description: "AreaService for unified WotC area generation", CommitHash: "06c22d4"},
					{Description: "MCP handlers for XP, items, tactics, and areas", CommitHash: "23e4559"},
					{Description: "WotC templates for milestone XP and areas", CommitHash: "23e4559"},
					{Description: "Validators for area and quest quality checks", CommitHash: "23e4559"},
					{Description: "E2E test suite with 7 comprehensive tests", CommitHash: "pending"},
					{Description: "Changelog automation script", CommitHash: "pending"},
					{Description: "Migration script v2 to v3", CommitHash: "pending"},
				},
				Changed: []ChangelogEntry{
					{Description: "Extended domain.Handout with V3 types (letter, clue, document, journal, etc.)", CommitHash: "c7173a9"},
					{Description: "Extended domain.Quest with QuestApproach, QuestFailure, QuestClue structures", CommitHash: "c7173a9"},
					{Description: "Exported domain.IsValidRarity for service use", CommitHash: "06c22d4"},
				},
			},
		},
	}
}

func formatChangelogMarkdown(changelog Changelog) string {
	md := "# Changelog\n\n"
	md += "All notable changes to this project will be documented in this file.\n\n"
	md += "The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),\n"
	md += "and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n"

	for _, version := range changelog.Versions {
		md += fmt.Sprintf("## [%s] - %s\n\n", version.Version, version.Date)
		
		if version.Summary != "" {
			md += fmt.Sprintf("**%s**\n\n", version.Summary)
		}

		if len(version.Added) > 0 {
			md += "### Added\n\n"
			for _, entry := range version.Added {
				md += formatEntry(entry, changelog.RepositoryURL)
			}
			md += "\n"
		}

		if len(version.Changed) > 0 {
			md += "### Changed\n\n"
			for _, entry := range version.Changed {
				md += formatEntry(entry, changelog.RepositoryURL)
			}
			md += "\n"
		}

		if len(version.Deprecated) > 0 {
			md += "### Deprecated\n\n"
			for _, entry := range version.Deprecated {
				md += formatEntry(entry, changelog.RepositoryURL)
			}
			md += "\n"
		}

		if len(version.Removed) > 0 {
			md += "### Removed\n\n"
			for _, entry := range version.Removed {
				md += formatEntry(entry, changelog.RepositoryURL)
			}
			md += "\n"
		}

		if len(version.Fixed) > 0 {
			md += "### Fixed\n\n"
			for _, entry := range version.Fixed {
				md += formatEntry(entry, changelog.RepositoryURL)
			}
			md += "\n"
		}

		if len(version.Security) > 0 {
			md += "### Security\n\n"
			for _, entry := range version.Security {
				md += formatEntry(entry, changelog.RepositoryURL)
			}
			md += "\n"
		}
	}

	return md
}

func formatEntry(entry ChangelogEntry, repoURL string) string {
	line := fmt.Sprintf("- %s", entry.Description)
	
	if entry.CommitHash != "" && entry.CommitHash != "pending" {
		line += fmt.Sprintf(" ([`%s`](%s/commit/%s))", entry.CommitHash[:7], repoURL, entry.CommitHash)
	}
	
	if len(entry.IssueRefs) > 0 {
		for _, ref := range entry.IssueRefs {
			line += fmt.Sprintf(" (#%s)", ref)
		}
	}
	
	line += "\n"
	return line
}
