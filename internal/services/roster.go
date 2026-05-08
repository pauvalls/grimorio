package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// AdventureRosterService compiles a master table of campaign entities
type AdventureRosterService struct {
	campaignRepo repository.CampaignRepository
	baseDir      string
}

// NewAdventureRosterService creates a new roster service
func NewAdventureRosterService(campaignRepo repository.CampaignRepository, baseDir string) *AdventureRosterService {
	return &AdventureRosterService{
		campaignRepo: campaignRepo,
		baseDir:      baseDir,
	}
}

// BuildRoster scans the campaign directory and builds a roster
func (s *AdventureRosterService) BuildRoster(ctx context.Context, campaignID string) (*domain.AdventureRoster, error) {
	if !s.campaignRepo.Exists(campaignID) {
		return nil, fmt.Errorf("campaign not found: %s", campaignID)
	}

	campaignDir := filepath.Join(s.baseDir, campaignID)
	roster := &domain.AdventureRoster{
		CampaignID: campaignID,
	}

	// Scan acts for NPCs, monsters, encounters
	actsDir := filepath.Join(campaignDir, "acts")
	if info, err := os.Stat(actsDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(actsDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, err := os.ReadFile(filepath.Join(actsDir, f.Name()))
				if err != nil {
					continue
				}
				actNum := strings.TrimSuffix(f.Name(), ".md")
				npcs, monsters, encounters := s.parseMarkdown(string(content), actNum)
				roster.NPCs = append(roster.NPCs, npcs...)
				roster.Monsters = append(roster.Monsters, monsters...)
				roster.Encounters = append(roster.Encounters, encounters...)
			}
		}
	}

	// Scan dedicated NPC directory
	npcsDir := filepath.Join(campaignDir, "npcs")
	if info, err := os.Stat(npcsDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(npcsDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, err := os.ReadFile(filepath.Join(npcsDir, f.Name()))
				if err != nil {
					continue
				}
				npcs, _, _ := s.parseMarkdown(string(content), "")
				// For dedicated NPC files, set page ref to file name
				for i := range npcs {
					if npcs[i].PageRef == "" {
						npcs[i].PageRef = "TBD"
					}
				}
				roster.NPCs = append(roster.NPCs, npcs...)
			}
		}
	}

	// Scan bestiary
	bestiaryDir := filepath.Join(campaignDir, "bestiary")
	if info, err := os.Stat(bestiaryDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(bestiaryDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, err := os.ReadFile(filepath.Join(bestiaryDir, f.Name()))
				if err != nil {
					continue
				}
				_, monsters, _ := s.parseMarkdown(string(content), "")
				for i := range monsters {
					if monsters[i].PageRef == "" {
						monsters[i].PageRef = "TBD"
					}
				}
				roster.Monsters = append(roster.Monsters, monsters...)
			}
		}
	}

	// Scan encounters
	encountersDir := filepath.Join(campaignDir, "encounters")
	if info, err := os.Stat(encountersDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(encountersDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				content, err := os.ReadFile(filepath.Join(encountersDir, f.Name()))
				if err != nil {
					continue
				}
				_, _, encounters := s.parseMarkdown(string(content), "")
				for i := range encounters {
					if encounters[i].PageRef == "" {
						encounters[i].PageRef = "TBD"
					}
				}
				roster.Encounters = append(roster.Encounters, encounters...)
			}
		}
	}

	// Sort by act then name
	sort.Slice(roster.NPCs, func(i, j int) bool {
		if roster.NPCs[i].Act != roster.NPCs[j].Act {
			return roster.NPCs[i].Act < roster.NPCs[j].Act
		}
		return roster.NPCs[i].Name < roster.NPCs[j].Name
	})
	sort.Slice(roster.Monsters, func(i, j int) bool {
		if roster.Monsters[i].Act != roster.Monsters[j].Act {
			return roster.Monsters[i].Act < roster.Monsters[j].Act
		}
		return roster.Monsters[i].Name < roster.Monsters[j].Name
	})
	sort.Slice(roster.Encounters, func(i, j int) bool {
		if roster.Encounters[i].Act != roster.Encounters[j].Act {
			return roster.Encounters[i].Act < roster.Encounters[j].Act
		}
		return roster.Encounters[i].Name < roster.Encounters[j].Name
	})

	return roster, nil
}

var (
	// Matches ## Name or ### Name heading
	headingRegex = regexp.MustCompile(`^(#{2,3})\s+(.+)$`)
	// Matches bold list item like - **Name** — description or - **Name** (CR 1/4)
	listItemRegex = regexp.MustCompile(`^\s*[-*]\s*\*\*([^*]+)\*\*\s*(?:[-—]|\()(.*)$`)
	// Matches CR in text
	crRegex = regexp.MustCompile(`(?i)CR[:\s]*([\d/]+)`)
)

func (s *AdventureRosterService) parseMarkdown(md string, actNum string) ([]domain.RosterNPC, []domain.RosterMonster, []domain.RosterEncounter) {
	var npcs []domain.RosterNPC
	var monsters []domain.RosterMonster
	var encounters []domain.RosterEncounter

	lines := strings.Split(md, "\n")
	var currentSection string
	var currentArea string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Track headings
		if m := headingRegex.FindStringSubmatch(trimmed); m != nil {
			level := len(m[1])
			text := strings.TrimSpace(m[2])
			if level == 2 {
				currentSection = text
				currentArea = ""
				// In dedicated encounter files, ## headings are encounter names
				if actNum == "" {
					lower := strings.ToLower(text)
					if strings.Contains(lower, "encuentro") {
						// This is the main heading, skip
					} else if strings.Contains(lower, "npc") || strings.Contains(lower, "personaje") ||
						strings.Contains(lower, "monstruo") || strings.Contains(lower, "criatura") {
						// Entity section heading
					} else {
						// Could be an encounter or entity name in a dedicated file
						encounters = append(encounters, domain.RosterEncounter{
							ID:      sanitizeID(text),
							Name:    text,
							Act:     actNum,
							Area:    currentArea,
							PageRef: "TBD",
						})
					}
				}
			} else if level == 3 {
				currentArea = text
			}
			continue
		}

		// Extract list items with bold names
		if m := listItemRegex.FindStringSubmatch(trimmed); m != nil {
			name := strings.TrimSpace(m[1])
			desc := strings.TrimSpace(m[2])

			// Categorize by section or area context
			ctx := strings.ToLower(currentSection + " " + currentArea)
			if strings.Contains(ctx, "npc") || strings.Contains(ctx, "personaje") {
				npcs = append(npcs, domain.RosterNPC{
					ID:       sanitizeID(name),
					Name:     name,
					Role:     desc,
					Act:      actNum,
					Location: currentArea,
					PageRef:  "TBD",
				})
			} else if strings.Contains(ctx, "monstruo") || strings.Contains(ctx, "criatura") || strings.Contains(ctx, "bestia") {
				cr := extractCR(trimmed)
				monsters = append(monsters, domain.RosterMonster{
					ID:        sanitizeID(name),
					Name:      name,
					CR:        cr,
					Act:       actNum,
					Locations: []string{currentArea},
					PageRef:   "TBD",
				})
			} else if strings.Contains(ctx, "encuentro") {
				encounters = append(encounters, domain.RosterEncounter{
					ID:      sanitizeID(name),
					Name:    name,
					Act:     actNum,
					Area:    currentArea,
					PageRef: "TBD",
				})
			}
			continue
		}

		// Extract plain list items in encounter sections (e.g., "- Emboscada en el bosque")
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			name := strings.TrimPrefix(trimmed, "- ")
			if strings.HasPrefix(trimmed, "* ") {
				name = strings.TrimPrefix(trimmed, "* ")
			}
			name = strings.TrimSpace(name)
			ctx := strings.ToLower(currentSection + " " + currentArea)
			if strings.Contains(ctx, "encuentro") && name != "" {
				encounters = append(encounters, domain.RosterEncounter{
					ID:      sanitizeID(name),
					Name:    name,
					Act:     actNum,
					Area:    currentArea,
					PageRef: "TBD",
				})
			}
		}

		// Extract standalone ## headings as entities in dedicated files
		if currentSection != "" && actNum == "" {
			lowerSection := strings.ToLower(currentSection)
			if !strings.Contains(lowerSection, "acto") && !strings.Contains(lowerSection, "intro") && !strings.Contains(lowerSection, "encuentro") {
				if cr := extractCR(trimmed); cr != "" {
					monsters = append(monsters, domain.RosterMonster{
						ID:        sanitizeID(currentSection),
						Name:      currentSection,
						CR:        cr,
						Locations: []string{},
						PageRef:   "TBD",
					})
				}
			}
			// Note: ## headings in encounter files are encounter names
			// but we already captured them via currentSection tracking
		}
	}

	return npcs, monsters, encounters
}

func extractCR(text string) string {
	if m := crRegex.FindStringSubmatch(text); m != nil {
		return m[1]
	}
	return ""
}

func sanitizeID(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	var result []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	return string(result)
}
