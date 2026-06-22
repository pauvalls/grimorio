package monster

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pauvalls/grimorio/internal/monster/rules/parser"
)

// AuditReport summarizes the CR drift for an entire campaign.
type AuditReport struct {
	CampaignID string             `json:"campaign_id"`
	Monsters   []*ValidationResult `json:"monsters"`
	Summary    AuditSummary       `json:"summary"`
}

// AuditSummary counts the monsters in each severity bucket.
type AuditSummary struct {
	Total int `json:"total"`
	OK    int `json:"ok"`
	Minor int `json:"minor"`
	Major int `json:"major"`
}

// MonsterAuditor validates every monster in a campaign's bestiary.
type MonsterAuditor struct {
	BaseDir string
}

// NewMonsterAuditor returns a new auditor. baseDir is the campaigns
// root (defaults to $HOME/campaigns when empty).
func NewMonsterAuditor(baseDir string) *MonsterAuditor {
	if baseDir == "" {
		baseDir = os.Getenv("HOME") + "/campaigns"
	}
	return &MonsterAuditor{BaseDir: baseDir}
}

// ErrNotFound is returned when a campaign or its bestiary does not
// exist.
var ErrNotFound = fmt.Errorf("campaign not found")

// AuditCampaign reads the bestiary.md of the given campaign and
// validates every monster. Non-blocking: a malformed monster is
// reported as a Major finding, not an error.
func (a *MonsterAuditor) AuditCampaign(campaignID string) (*AuditReport, error) {
	bestiaryPath := filepath.Join(a.BaseDir, campaignID, "bestiary", "bestiary.md")
	if _, err := os.Stat(bestiaryPath); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err.Error())
	}
	bytes, err := os.ReadFile(bestiaryPath)
	if err != nil {
		return nil, fmt.Errorf("read bestiary: %w", err)
	}
	content := string(bytes)
	monsters := splitMonsters(content)
	report := &AuditReport{
		CampaignID: campaignID,
		Monsters:   []*ValidationResult{},
	}
	v := NewMonsterValidator()
	for _, blk := range monsters {
		m, perr := parser.ParseStatBlock(blk)
		if perr != nil {
			// Skip unparseable sections — they aren't strict WotC stat
			// blocks. Count them as Major findings to keep the audit
			// informative.
			report.Summary.Major++
			report.Summary.Total++
			continue
		}
		r := v.Validate(m)
		report.Monsters = append(report.Monsters, r)
		report.Summary.Total++
		switch r.Severity {
		case SeverityOK:
			report.Summary.OK++
		case SeverityMinor:
			report.Summary.Minor++
		case SeverityMajor:
			report.Summary.Major++
		}
	}
	return report, nil
}

// splitMonsters splits a bestiary markdown file into per-monster
// markdown blocks. The heuristic is: each top-level H2 (`## Name`)
// starts a new block. Blocks are returned as a slice of strings.
func splitMonsters(content string) []string {
	lines := strings.Split(content, "\n")
	var blocks []string
	current := strings.Builder{}
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "## ") && !strings.HasPrefix(t, "### ") {
			if current.Len() > 0 {
				blocks = append(blocks, current.String())
			}
			current.Reset()
			current.WriteString(l)
			current.WriteString("\n")
			continue
		}
		// Skip top-level H1.
		if strings.HasPrefix(t, "# ") && !strings.HasPrefix(t, "## ") {
			continue
		}
		current.WriteString(l)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		blocks = append(blocks, current.String())
	}
	// Filter out empty blocks and ones that look like a TOC / index.
	var out []string
	for _, b := range blocks {
		stripped := strings.TrimSpace(b)
		// Skip the "## Índice de Criaturas" / TOC section.
		if strings.Contains(strings.ToLower(stripped), "índice de criaturas") ||
			strings.Contains(strings.ToLower(stripped), "indice de criaturas") {
			continue
		}
		// Skip blocks that don't have any stat-block fields.
		if !looksLikeStatBlock(stripped) {
			continue
		}
		out = append(out, b)
	}
	return out
}

// looksLikeStatBlock returns true if the markdown block contains at
// least one of the canonical stat-block field markers.
var statBlockMarker = regexp.MustCompile(`(?i)\*\*\s*(?:Armor Class|Class\s*de\s*Armadura|Hit Points|Puntos\s*de\s*Golpe|Speed|Velocidad|Senses|Sentidos|Languages|Idiomas|Challenge|Desafío|CR)\s*\*\*`)

func looksLikeStatBlock(s string) bool {
	return statBlockMarker.MatchString(s)
}

// FindMonsterMarkdown returns the raw markdown block for the named
// monster in the campaign's bestiary. The lookup matches the H2
// heading (case-insensitive, trimmed). Returns ErrNotFound if no
// bestiary exists or no monster with that name is found.
func (a *MonsterAuditor) FindMonsterMarkdown(campaignID, monsterName string) (string, error) {
	bestiaryPath := filepath.Join(a.BaseDir, campaignID, "bestiary", "bestiary.md")
	bytes, err := os.ReadFile(bestiaryPath)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrNotFound, err.Error())
	}
	content := string(bytes)
	blocks := splitMonsters(content)
	target := strings.ToLower(strings.TrimSpace(monsterName))
	for _, blk := range blocks {
		// Extract the H2 heading.
		for _, l := range strings.Split(blk, "\n") {
			t := strings.TrimSpace(l)
			if strings.HasPrefix(t, "## ") && !strings.HasPrefix(t, "### ") {
				heading := strings.TrimSpace(strings.TrimPrefix(t, "## "))
				if strings.ToLower(heading) == target {
					return blk, nil
				}
				break
			}
		}
	}
	return "", fmt.Errorf("%w: monster %q not in %s", ErrNotFound, monsterName, bestiaryPath)
}
