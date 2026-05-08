package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// HandoutGenerator creates player-facing handouts from campaign content
type HandoutGenerator struct {
	CampaignDir string
}

// NewHandoutGenerator creates a new handout generator for a campaign directory
func NewHandoutGenerator(campaignDir string) *HandoutGenerator {
	return &HandoutGenerator{CampaignDir: campaignDir}
}

// GeneratePlayerMap creates a redacted version of a map for players
func (g *HandoutGenerator) GeneratePlayerMap(mapName string) (string, error) {
	mapPath := filepath.Join(g.CampaignDir, "maps", mapName+".md")
	data, err := os.ReadFile(mapPath)
	if err != nil {
		return "", fmt.Errorf("map not found: %w", err)
	}

	content := string(data)

	// Remove secret-related lines
	lines := strings.Split(content, "\n")
	var out []string
	secretPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bsecreta?\b`),
		regexp.MustCompile(`(?i)\btrampa\b`),
		regexp.MustCompile(`(?i)\btesoro\b`),
		regexp.MustCompile(`(?i)\bpista\b`),
		regexp.MustCompile(`(?i)DC\s*\d+`),
		regexp.MustCompile(`(?i)\bocult[oa]\b`),
		regexp.MustCompile(`(?i)\bocultar\b`),
	}

	for _, line := range lines {
		isSecret := false
		for _, re := range secretPatterns {
			if re.MatchString(line) {
				isSecret = true
				break
			}
		}
		if !isSecret {
			out = append(out, line)
		}
	}

	result := strings.Join(out, "\n")
	result = strings.TrimSpace(result)
	if result == "" {
		return "", fmt.Errorf("map is empty after redaction")
	}

	return result, nil
}

// GenerateClueList creates a player-facing list of discovered clues
func (g *HandoutGenerator) GenerateClueList() (string, error) {
	state, err := g.loadNarrativeState()
	if err != nil {
		return "", err
	}

	clues, ok := state["revealed_clues"].([]interface{})
	if !ok || len(clues) == 0 {
		return "## Pistas Descubiertas\n\nAún no habéis descubierto ninguna pista.", nil
	}

	var sb strings.Builder
	sb.WriteString("## Pistas Descubiertas\n\n")
	for _, c := range clues {
		clue, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		desc, _ := clue["description"].(string)
		if desc != "" {
			sb.WriteString("- **Sabéis que** ")
			sb.WriteString(desc)
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// GenerateNPCReference creates a quick-reference sheet for met NPCs
func (g *HandoutGenerator) GenerateNPCReference() (string, error) {
	state, err := g.loadNarrativeState()
	if err != nil {
		return "", err
	}

	metNPCs := make(map[string]bool)
	if met, ok := state["met_npcs"].([]interface{}); ok {
		for _, n := range met {
			if name, ok := n.(string); ok {
				metNPCs[name] = true
			}
		}
	}

	if len(metNPCs) == 0 {
		return "## NPCs Conocidos\n\nAún no habéis conocido a ningún NPC.", nil
	}

	// Read NPCs file
	npcPath := filepath.Join(g.CampaignDir, "npcs", "npcs_and_factions.md")
	data, err := os.ReadFile(npcPath)
	if err != nil {
		return "", fmt.Errorf("npcs file not found: %w", err)
	}

	// Parse NPC blocks
	npcBlocks := parseNPCBlocks(string(data))

	var sb strings.Builder
	sb.WriteString("## NPCs Conocidos\n\n")
	sb.WriteString("| NPC | Raza/Clase | Ubicación | Rol |\n")
	sb.WriteString("|-----|------------|-----------|-----|\n")

	for npcName, block := range npcBlocks {
		if !metNPCs[npcName] {
			continue
		}

		raceClass := extractField(block, "Raza/Clase")
		location := extractField(block, "Ubicación")
		role := extractField(block, "Rol en la historia")

		sb.WriteString(fmt.Sprintf("| **%s** | %s | %s | %s |\n",
			npcName, raceClass, location, role))
	}

	return sb.String(), nil
}

// GenerateSessionRecap creates a brief session recap handout
func (g *HandoutGenerator) GenerateSessionRecap() (string, error) {
	state, err := g.loadNarrativeState()
	if err != nil {
		return "", err
	}

	sessionNum := 0
	if n, ok := state["last_session"].(map[string]interface{}); ok {
		if num, ok := n["session_num"].(float64); ok {
			sessionNum = int(num)
		}
	}

	if sessionNum == 0 {
		return "## Resumen de Sesión\n\nNo hay datos de sesiones previas.", nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Resumen de la Sesión %d\n\n", sessionNum))

	if last, ok := state["last_session"].(map[string]interface{}); ok {
		if areas, ok := last["areas_visited"].([]interface{}); ok {
			sb.WriteString(fmt.Sprintf("En esta sesión explorasteis **%d áreas**: ", len(areas)))
			var names []string
			for _, a := range areas {
				if name, ok := a.(string); ok {
					names = append(names, name)
				}
			}
			sb.WriteString(strings.Join(names, ", "))
			sb.WriteString(".\n\n")
		}

		if combats, ok := last["combats"].(float64); ok {
			sb.WriteString(fmt.Sprintf("Participasteis en **%.0f combates**. ", combats))
		}

		if decisions, ok := last["key_decisions"].([]interface{}); ok && len(decisions) > 0 {
			sb.WriteString("Las decisiones clave de esta sesión fueron:\n\n")
			for _, d := range decisions {
				if desc, ok := d.(string); ok {
					sb.WriteString("- ")
					sb.WriteString(desc)
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String(), nil
}

func (g *HandoutGenerator) loadNarrativeState() (map[string]interface{}, error) {
	statePath := filepath.Join(g.CampaignDir, "narrative_state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		return map[string]interface{}{}, nil // return empty state if file missing
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("invalid narrative state: %w", err)
	}
	return state, nil
}

func parseNPCBlocks(md string) map[string]string {
	blocks := make(map[string]string)
	lines := strings.Split(md, "\n")
	var currentName string
	var currentLines []string

	// Pattern for NPC name: ### Name
	npcHeaderPattern := regexp.MustCompile(`^#{3}\s+(.+)$`)

	for _, line := range lines {
		if m := npcHeaderPattern.FindStringSubmatch(line); m != nil {
			if currentName != "" {
				blocks[currentName] = strings.Join(currentLines, "\n")
			}
			currentName = strings.TrimSpace(m[1])
			currentLines = nil
		} else if currentName != "" {
			currentLines = append(currentLines, line)
		}
	}

	if currentName != "" {
		blocks[currentName] = strings.Join(currentLines, "\n")
	}

	return blocks
}

func extractField(block, fieldName string) string {
	pattern := regexp.MustCompile(`\*\*` + regexp.QuoteMeta(fieldName) + `:\*\*\s*(.+)`)
	if m := pattern.FindStringSubmatch(block); m != nil {
		return strings.TrimSpace(m[1])
	}
	return "—"
}
