package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// EntityParser extracts structured entities (NPCs, Monsters, Factions) from markdown content
type EntityParser struct {
	npcHeadingRegex    *regexp.Regexp
	npcListItemRegex   *regexp.Regexp
	factionRegex       *regexp.Regexp
	monsterHeadingRegex *regexp.Regexp
	crRegex            *regexp.Regexp
	statRegex          *regexp.Regexp
}

// NewEntityParser creates a new entity parser with pre-compiled regex patterns
func NewEntityParser() *EntityParser {
	return &EntityParser{
		// Matches ## Name or ### Name
		npcHeadingRegex: regexp.MustCompile(`^#{2,3}\s+(.+?)\s*$`),
		// Matches - **Name** — description or - **Name** (CR 1/4) — description
		npcListItemRegex: regexp.MustCompile(`^\s*[-*]\s*\*\*(.+?)\*\*\s*(?:[—-]|\()\s*(.+?)$`),
		// Matches faction sections
		factionRegex: regexp.MustCompile(`(?i)^#{2,3}\s*(?:facciones|factions)\s*$`),
		// Matches monster headings
		monsterHeadingRegex: regexp.MustCompile(`^#{1,2}\s+(.+?)\s*$`),
		// Matches CR in text: CR 1/4, CR 2, CR: 5, etc.
		crRegex: regexp.MustCompile(`(?i)CR[:\s]*([\d/]+)`),
		// Matches stat blocks: AC 15, HP 33, etc.
		statRegex: regexp.MustCompile(`(?i)(?:AC|CA)[:\s]*(\d+)|(?:HP|PG)[:\s]*(\d+)`),
	}
}

// ParseResult contains parsed entities from markdown
type ParseResult struct {
	NPCs       []domain.NPC
	Monsters   []domain.Monster
	Factions   []domain.Faction
	Encounters []domain.Encounter
	Areas      []domain.Area
}

// ParseNPCs extracts NPCs and Factions from markdown content
// Expected format:
// ## NPCs Principales
// ### Thorin Ironforge
// *Legal Good Dwarf Fighter*
// - **Ubicación:** Herrero en el Barrio Bajo
// - **Estadísticas:** AC 15, HP 33
//
// ## Facciones
// ### Orden de la Plata
// - **Líder:** Elara Voss
// - **Objetivo:** Proteger la ciudad
func (p *EntityParser) ParseNPCs(content string, campaignID string) (*ParseResult, error) {
	result := &ParseResult{
		NPCs:     []domain.NPC{},
		Factions: []domain.Faction{},
	}

	lines := strings.Split(content, "\n")
	var currentNPC *domain.NPC
	var currentFaction *domain.Faction
	inFactionSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for faction section
		if p.factionRegex.MatchString(trimmed) {
			inFactionSection = true
			continue
		}

		// Check for NPC/Monster heading (## or ###)
		if matches := p.npcHeadingRegex.FindStringSubmatch(trimmed); matches != nil {
			name := strings.TrimSpace(matches[1])
			
			// Skip generic section headers
			lower := strings.ToLower(name)
			if strings.Contains(lower, "npc") || strings.Contains(lower, "personaje") ||
				strings.Contains(lower, "facción") || strings.Contains(lower, "faccion") {
				continue
			}

			if inFactionSection {
				// Save previous faction if exists
				if currentFaction != nil {
					result.Factions = append(result.Factions, *currentFaction)
				}
				currentFaction = &domain.Faction{
					ID:          sanitizeName(name),
					Name:        name,
					Description: "",
					Agenda:      "",
					ContactNPC:  "",
				}
			} else {
				// Save previous NPC if exists
				if currentNPC != nil {
					result.NPCs = append(result.NPCs, *currentNPC)
				}
				currentNPC = &domain.NPC{
					ID:          sanitizeName(name),
					CampaignID:  campaignID,
					Name:        name,
					Role:        "neutral",
					Description: "",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
			}
			continue
		}

		// Parse list items for current entity
		if matches := p.npcListItemRegex.FindStringSubmatch(trimmed); matches != nil {
			key := strings.TrimSpace(matches[1])
			value := strings.TrimSpace(matches[2])
			// Remove closing parenthesis if present
			value = strings.TrimSuffix(value, ")")

			if inFactionSection && currentFaction != nil {
				p.parseFactionField(currentFaction, key, value)
			} else if currentNPC != nil {
				p.parseNPCField(currentNPC, key, value)
			}
		}

		// Parse stat blocks (AC 15, HP 33, etc.)
		if currentNPC != nil {
			if stats := p.parseStats(trimmed); stats != nil {
				if currentNPC.Stats == nil {
					currentNPC.Stats = stats
				} else {
					// Merge stats
					if stats.AC > 0 {
						currentNPC.Stats.AC = stats.AC
					}
					if stats.HP > 0 {
						currentNPC.Stats.HP = stats.HP
					}
				}
			}
		}
	}

	// Save last entity
	if currentNPC != nil {
		result.NPCs = append(result.NPCs, *currentNPC)
	}
	if currentFaction != nil {
		result.Factions = append(result.Factions, *currentFaction)
	}

	return result, nil
}

// ParseMonsters extracts monsters from markdown content
// Expected format:
// # Cenizo Recién Convertido
// *CR 1/4, No-muerto Mediano*
// - **AC:** 15
// - **HP:** 33
// - **Tácticas:** Atacar en grupo
func (p *EntityParser) ParseMonsters(content string, campaignID string) ([]domain.Monster, error) {
	monsters := []domain.Monster{}

	lines := strings.Split(content, "\n")
	var currentMonster *domain.Monster

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for monster heading (# or ##)
		if matches := p.monsterHeadingRegex.FindStringSubmatch(trimmed); matches != nil {
			name := strings.TrimSpace(matches[1])
			
			// Skip generic section headers
			lower := strings.ToLower(name)
			if strings.Contains(lower, "traits") || strings.Contains(lower, "actions") ||
				strings.Contains(lower, "tácticas") || strings.Contains(lower, "descripción") ||
				strings.Contains(lower, "debilidades") {
				continue
			}

			// Save previous monster if exists
			if currentMonster != nil {
				monsters = append(monsters, *currentMonster)
			}

			currentMonster = &domain.Monster{
				ID:         sanitizeName(name),
				CampaignID: campaignID,
				Name:       name,
				CR:         "",
				Type:       "monster",
				Size:       "Medium",
				Stats:      domain.StatBlock{},
				CreatedAt:  time.Now(),
			}
			continue
		}

		// Parse CR from line
		if currentMonster != nil {
			if crMatches := p.crRegex.FindStringSubmatch(trimmed); len(crMatches) > 1 {
				currentMonster.CR = crMatches[1]
			}

			// Parse stats
			if stats := p.parseStats(trimmed); stats != nil {
				if stats.AC > 0 {
					currentMonster.Stats.AC = stats.AC
				}
				if stats.HP > 0 {
					currentMonster.Stats.HP = stats.HP
				}
			}

			// Parse type
			if strings.Contains(strings.ToLower(trimmed), "no-muerto") {
				currentMonster.Type = "undead"
			} else if strings.Contains(strings.ToLower(trimmed), "humanoide") {
				currentMonster.Type = "humanoid"
			} else if strings.Contains(strings.ToLower(trimmed), "bestia") {
				currentMonster.Type = "beast"
			}
		}
	}

	// Save last monster
	if currentMonster != nil {
		monsters = append(monsters, *currentMonster)
	}

	return monsters, nil
}

// parseNPCField extracts NPC information from key-value pairs
func (p *EntityParser) parseNPCField(npc *domain.NPC, key, value string) {
	keyLower := strings.ToLower(key)
	
	switch {
	case strings.Contains(keyLower, "ubicación") || strings.Contains(keyLower, "location"):
		// Store location in description for now
		npc.Description = value
	case strings.Contains(keyLower, "rol") || strings.Contains(keyLower, "role") ||
		 strings.Contains(keyLower, "alineamiento") || strings.Contains(keyLower, "alignment"):
		npc.Role = value
	case strings.Contains(keyLower, "facción") || strings.Contains(keyLower, "faction"):
		npc.Faction = value
	default:
		// Append to description if not set
		if npc.Description == "" {
			npc.Description = value
		}
	}
}

// parseFactionField extracts Faction information from key-value pairs
func (p *EntityParser) parseFactionField(faction *domain.Faction, key, value string) {
	keyLower := strings.ToLower(key)
	
	switch {
	case strings.Contains(keyLower, "líder") || strings.Contains(keyLower, "leader"):
		// Store leader in description
		faction.Description = fmt.Sprintf("Líder: %s", value)
	case strings.Contains(keyLower, "objetivo") || strings.Contains(keyLower, "goal") ||
		 strings.Contains(keyLower, "propósito") || strings.Contains(keyLower, "purpose"):
		if faction.Description != "" {
			faction.Description += ". " + value
		} else {
			faction.Description = value
		}
	default:
		if faction.Description == "" {
			faction.Description = value
		}
	}
}

// parseStats extracts AC and HP from text
func (p *EntityParser) parseStats(line string) *domain.StatBlock {
	stats := &domain.StatBlock{}

	matches := p.statRegex.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		// match[0] is full match, match[1] is AC group, match[2] is HP group
		if len(match) > 1 && match[1] != "" {
			num := parseInt(match[1])
			if num > 0 {
				stats.AC = num
			}
		}
		if len(match) > 2 && match[2] != "" {
			num := parseInt(match[2])
			if num > 0 {
				stats.HP = num
			}
		}
	}

	// Return nil if no stats found
	if stats.AC == 0 && stats.HP == 0 {
		return nil
	}

	return stats
}

// sanitizeName converts a name to a safe identifier (package-private version)
func sanitizeName(name string) string {
	// Convert to lowercase
	id := strings.ToLower(name)
	// Replace spaces with hyphens
	id = strings.ReplaceAll(id, " ", "-")
	// Remove special characters
	id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")
	// Remove consecutive hyphens
	id = regexp.MustCompile(`-+`).ReplaceAllString(id, "-")
	// Trim hyphens from ends
	id = strings.Trim(id, "-")
	return id
}

// parseInt attempts to parse an integer from string
func parseInt(s string) int {
	var result int
	_, _ = fmt.Sscanf(s, "%d", &result)
	return result
}

// ParseEncounters extracts encounters from markdown content
// Expected format:
// ## Encuentro 1: Emboscada en el Callejón
// *Dificultad: Medium*
// - **Ubicación:** Barrio Bajo
// - **Monstruos:** 3x Cenizo, 1x Seguidor
// - **Recompensa:** 100 XP, 50 gold
func (p *EntityParser) ParseEncounters(content string, campaignID string) ([]domain.Encounter, error) {
	encounters := []domain.Encounter{}

	lines := strings.Split(content, "\n")
	var currentEncounter *domain.Encounter

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for encounter heading (## Encuentro X: Nombre)
		if matches := regexp.MustCompile(`^#{2}\s+(?:Encuentro\s+\d+[:\s]*)?(.+?)\s*$`).FindStringSubmatch(trimmed); matches != nil {
			name := strings.TrimSpace(matches[1])
			// Remove "Encuentro X:" prefix if present
			name = regexp.MustCompile(`^Encuentro\s+\d+[:\s]*`).ReplaceAllString(name, "")
			name = strings.TrimSpace(name)

			if name == "" {
				continue
			}

			// Save previous encounter
			if currentEncounter != nil {
				encounters = append(encounters, *currentEncounter)
			}

			currentEncounter = &domain.Encounter{
				ID:         sanitizeName(name),
				CampaignID: campaignID,
				Name:       name,
				Difficulty: "medium",
				Location:   "",
				Monsters:   []domain.MonsterRef{},
				Rewards:    []domain.Reward{},
				CreatedAt:  time.Now(),
			}
			continue
		}

		// Parse encounter details
		if currentEncounter != nil {
			lower := strings.ToLower(trimmed)
			
			// Parse difficulty
			if strings.Contains(lower, "dificultad") || strings.Contains(lower, "difficulty") {
				if strings.Contains(lower, "easy") || strings.Contains(lower, "fácil") {
					currentEncounter.Difficulty = "easy"
				} else if strings.Contains(lower, "hard") || strings.Contains(lower, "difícil") {
					currentEncounter.Difficulty = "hard"
				} else if strings.Contains(lower, "deadly") || strings.Contains(lower, "mortal") {
					currentEncounter.Difficulty = "deadly"
				}
			}

			// Parse location
			if strings.Contains(lower, "ubicación") || strings.Contains(lower, "location") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentEncounter.Location = strings.TrimSpace(parts[1])
				}
			}

			// Parse monsters (format: "3x Cenizo, 1x Seguidor")
			if strings.Contains(lower, "monstruos") || strings.Contains(lower, "monsters") ||
				strings.Contains(lower, "criaturas") || strings.Contains(lower, "creatures") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					monsterStr := strings.TrimSpace(parts[1])
					// Parse each monster: "3x Cenizo"
					monsterParts := strings.Split(monsterStr, ",")
					for _, mp := range monsterParts {
						mp = strings.TrimSpace(mp)
						if mp == "" {
							continue
						}
						// Extract quantity and name
						matches := regexp.MustCompile(`(\d+)?\s*x?\s*(.+)`).FindStringSubmatch(mp)
						if len(matches) > 2 {
							qty := 1
							if matches[1] != "" {
								qty = parseInt(matches[1])
							}
							name := strings.TrimSpace(matches[2])
							currentEncounter.Monsters = append(currentEncounter.Monsters, domain.MonsterRef{
								Name:     name,
								Quantity: qty,
							})
						}
					}
				}
			}

			// Parse rewards
			if strings.Contains(lower, "recompensa") || strings.Contains(lower, "reward") ||
				strings.Contains(lower, "xp") || strings.Contains(lower, "gold") {
				// Extract XP
				xpMatch := regexp.MustCompile(`(\d+)\s*XP`).FindStringSubmatch(trimmed)
				if len(xpMatch) > 1 {
					currentEncounter.Rewards = append(currentEncounter.Rewards, domain.Reward{
						Type:        "xp",
						Description: fmt.Sprintf("%s XP", xpMatch[1]),
						Value:       xpMatch[1],
					})
				}
				// Extract gold
				goldMatch := regexp.MustCompile(`(\d+)\s*gold`).FindStringSubmatch(trimmed)
				if len(goldMatch) > 1 {
					currentEncounter.Rewards = append(currentEncounter.Rewards, domain.Reward{
						Type:        "gold",
						Description: fmt.Sprintf("%s gold", goldMatch[1]),
						Value:       goldMatch[1],
					})
				}
			}
		}
	}

	// Save last encounter
	if currentEncounter != nil {
		encounters = append(encounters, *currentEncounter)
	}

	return encounters, nil
}

// ParseChapter extracts inline NPCs, encounters, and areas from a self-contained chapter markdown.
// Expected format:
// # Capítulo N: Title
// ## NPCs en este Capítulo
// ### NPC Name
// ## Encuentros
// ### Encuentro N: Name
// ## Áreas
// ### Área N: Name
func (p *EntityParser) ParseChapter(content, campaignID string, chapterNum int) (*ParseResult, error) {
	result := &ParseResult{
		NPCs:       []domain.NPC{},
		Encounters: []domain.Encounter{},
		Areas:      []domain.Area{},
	}

	lines := strings.Split(content, "\n")
	var currentNPC *domain.NPC
	var currentEncounter *domain.Encounter
	var currentArea *domain.Area

	inNPCSection := false
	inEncounterSection := false
	inAreaSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)

		// Detect section boundaries
		if strings.HasPrefix(lower, "## ") {
			inNPCSection = false
			inEncounterSection = false
			inAreaSection = false

			if strings.Contains(lower, "npc") {
				inNPCSection = true
			} else if strings.Contains(lower, "encuentro") {
				inEncounterSection = true
			} else if strings.Contains(lower, "área") || strings.Contains(lower, "area") {
				inAreaSection = true
			}
			continue
		}

		// Parse area heading: ### Área N: Name
		if inAreaSection {
			if matches := regexp.MustCompile(`^#{3}\s+(?:Área|Area)\s+(\d+)[:\s]*(.+?)\s*$`).FindStringSubmatch(trimmed); matches != nil {
				areaNum := parseInt(matches[1])
				name := strings.TrimSpace(matches[2])

				if currentArea != nil {
					result.Areas = append(result.Areas, *currentArea)
				}

				chapterID := fmt.Sprintf("chapter-%d", chapterNum)
				currentArea = &domain.Area{
					ID:         fmt.Sprintf("%s-area-%d", chapterID, areaNum),
					ChapterID:  chapterID,
					AreaNumber: areaNum,
					Title:      name,
				}
				continue
			}

			if currentArea != nil {
				// Parse blockquote as PlayerReadAloud
				if strings.HasPrefix(trimmed, ">") {
					readAloud := strings.TrimSpace(strings.TrimPrefix(trimmed, ">"))
					if readAloud != "" {
						if currentArea.PlayerReadAloud != "" {
							currentArea.PlayerReadAloud += " "
						}
						currentArea.PlayerReadAloud += readAloud
					}
					continue
				}
				// Accumulate description (skip headings and list markers)
				if !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "*") {
					if currentArea.Description != "" {
						currentArea.Description += " "
					}
					currentArea.Description += trimmed
				}
			}
		}

		// Parse encounter heading: ### Encuentro N: Name
		if inEncounterSection {
			if matches := regexp.MustCompile(`^#{3}\s+Encuentro\s+(\d+)[:\s]*(.+?)\s*$`).FindStringSubmatch(trimmed); matches != nil {
				name := strings.TrimSpace(matches[2])

				if currentEncounter != nil {
					result.Encounters = append(result.Encounters, *currentEncounter)
				}

				currentEncounter = &domain.Encounter{
					ID:         sanitizeName(name),
					CampaignID: campaignID,
					Name:       name,
					Difficulty: "medium",
					Monsters:   []domain.MonsterRef{},
					Rewards:    []domain.Reward{},
				}
				continue
			}

			if currentEncounter != nil {
				lower := strings.ToLower(trimmed)

				// Parse difficulty
				if strings.Contains(lower, "dificultad") || strings.Contains(lower, "difficulty") {
					if strings.Contains(lower, "easy") || strings.Contains(lower, "fácil") {
						currentEncounter.Difficulty = "easy"
					} else if strings.Contains(lower, "hard") || strings.Contains(lower, "difícil") {
						currentEncounter.Difficulty = "hard"
					} else if strings.Contains(lower, "deadly") || strings.Contains(lower, "mortal") {
						currentEncounter.Difficulty = "deadly"
					}
				}

				// Parse location
				if strings.Contains(lower, "ubicación") || strings.Contains(lower, "location") {
					parts := strings.SplitN(trimmed, ":", 2)
					if len(parts) == 2 {
						currentEncounter.Location = strings.TrimSpace(parts[1])
					}
				}

			// Parse monsters from inline format: "3x Cenizo, 1x Seguidor"
			if strings.Contains(lower, "monstruos") || strings.Contains(lower, "monsters") ||
				strings.Contains(lower, "criaturas") || strings.Contains(lower, "creatures") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					monsterStr := strings.TrimSpace(parts[1])
					monsterStr = strings.TrimPrefix(monsterStr, "**")
					monsterStr = strings.TrimSuffix(monsterStr, "**")
					monsterStr = strings.TrimSpace(monsterStr)
					if monsterStr != "" {
						monsterParts := strings.Split(monsterStr, ",")
						for _, mp := range monsterParts {
							mp = strings.TrimSpace(mp)
							if mp == "" {
								continue
							}
							matches := regexp.MustCompile(`(\d+)?\s*x?\s*(.+)`).FindStringSubmatch(mp)
							if len(matches) > 2 {
								qty := 1
								if matches[1] != "" {
									qty = parseInt(matches[1])
								}
								name := strings.TrimSpace(matches[2])
								currentEncounter.Monsters = append(currentEncounter.Monsters, domain.MonsterRef{
									Name:     name,
									Quantity: qty,
								})
							}
						}
					}
				}
			}

			// Parse monsters from list items (e.g., "- 3x Bandido")
			if currentEncounter != nil && (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")) {
				listItem := strings.TrimPrefix(trimmed, "- ")
				listItem = strings.TrimPrefix(listItem, "* ")
				listItem = strings.TrimSpace(listItem)
				// Check if it looks like a monster quantity reference
				matches := regexp.MustCompile(`^(\d+)?\s*x?\s*(.+)`).FindStringSubmatch(listItem)
				if len(matches) > 2 {
					potentialName := strings.TrimSpace(matches[2])
					// Heuristic: if the line is short and doesn't look like a reward or property
					if potentialName != "" && !strings.Contains(strings.ToLower(potentialName), "xp") &&
						!strings.Contains(strings.ToLower(potentialName), "gold") {
						qty := 1
						if matches[1] != "" {
							qty = parseInt(matches[1])
						}
						currentEncounter.Monsters = append(currentEncounter.Monsters, domain.MonsterRef{
							Name:     potentialName,
							Quantity: qty,
						})
					}
				}
			}

				// Parse rewards
				if strings.Contains(lower, "recompensa") || strings.Contains(lower, "reward") ||
					strings.Contains(lower, "xp") || strings.Contains(lower, "gold") {
					xpMatch := regexp.MustCompile(`(\d+)\s*XP`).FindStringSubmatch(trimmed)
					if len(xpMatch) > 1 {
						currentEncounter.Rewards = append(currentEncounter.Rewards, domain.Reward{
							Type:        "xp",
							Description: fmt.Sprintf("%s XP", xpMatch[1]),
							Value:       xpMatch[1],
						})
					}
					goldMatch := regexp.MustCompile(`(\d+)\s*gold`).FindStringSubmatch(trimmed)
					if len(goldMatch) > 1 {
						currentEncounter.Rewards = append(currentEncounter.Rewards, domain.Reward{
							Type:        "gold",
							Description: fmt.Sprintf("%s gold", goldMatch[1]),
							Value:       goldMatch[1],
						})
					}
				}
			}
		}

		// Parse NPC heading: ### NPC Name
		if inNPCSection {
			if matches := regexp.MustCompile(`^#{3}\s+(.+?)\s*$`).FindStringSubmatch(trimmed); matches != nil {
				name := strings.TrimSpace(matches[1])

				// Skip generic headers
				lowerName := strings.ToLower(name)
				if strings.Contains(lowerName, "npc") || strings.Contains(lowerName, "personaje") {
					continue
				}

				if currentNPC != nil {
					result.NPCs = append(result.NPCs, *currentNPC)
				}

				currentNPC = &domain.NPC{
					ID:         sanitizeName(name),
					CampaignID: campaignID,
					Name:       name,
					Role:       "chapter-inline",
				}
				continue
			}

			if currentNPC != nil {
				// Parse stats
				if stats := p.parseStats(trimmed); stats != nil {
					if currentNPC.Stats == nil {
						currentNPC.Stats = stats
					} else {
						if stats.AC > 0 {
							currentNPC.Stats.AC = stats.AC
						}
						if stats.HP > 0 {
							currentNPC.Stats.HP = stats.HP
						}
					}
				}

			// Parse role from italic line into description, keep Role as chapter-inline
			if strings.HasPrefix(trimmed, "*") && strings.HasSuffix(trimmed, "*") {
				role := strings.Trim(trimmed, "*")
				if currentNPC.Description != "" {
					currentNPC.Description += " "
				}
				currentNPC.Description += role
			}

				// Accumulate description
				if !strings.HasPrefix(trimmed, "*") && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "#") {
					if currentNPC.Description != "" {
						currentNPC.Description += " "
					}
					currentNPC.Description += trimmed
				}
			}
		}
	}

	// Save last entities
	if currentArea != nil {
		result.Areas = append(result.Areas, *currentArea)
	}
	if currentEncounter != nil {
		result.Encounters = append(result.Encounters, *currentEncounter)
	}
	if currentNPC != nil {
		result.NPCs = append(result.NPCs, *currentNPC)
	}

	// Extract unique monster names from all encounters into result.Monsters
	monsterSet := make(map[string]bool)
	for _, enc := range result.Encounters {
		for _, m := range enc.Monsters {
			monsterSet[m.Name] = true
		}
	}
	for name := range monsterSet {
		result.Monsters = append(result.Monsters, domain.Monster{
			ID:         sanitizeName(name),
			CampaignID: campaignID,
			Name:       name,
		})
	}

	if len(result.Areas) == 0 {
		return nil, fmt.Errorf("no areas found in chapter - expected format: ### Área X: Nombre")
	}

	return result, nil
}

// ParseAreas extracts areas from markdown content (WotC format)
// Expected format:
// ### Área 1: Puerta del Norte
// **Número de Área:** 1
// **Ubicación:** Norte de Velmorath
// **Resumen:** Los PJs llegan a la ciudad
func (p *EntityParser) ParseAreas(content string, campaignID string, chapterID string) ([]domain.Area, error) {
	areas := []domain.Area{}

	lines := strings.Split(content, "\n")
	var currentArea *domain.Area

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Check for area heading (### Área X: Nombre or ### Area X: Name)
		if matches := regexp.MustCompile(`^#{3}\s+(?:Área|Area)\s+(\d+)[:\s]*(.+?)\s*$`).FindStringSubmatch(trimmed); matches != nil {
			areaNum := parseInt(matches[1])
			name := strings.TrimSpace(matches[2])

			// Save previous area
			if currentArea != nil {
				areas = append(areas, *currentArea)
			}

			currentArea = &domain.Area{
				ID:         fmt.Sprintf("%s-area-%d", chapterID, areaNum),
				ChapterID:  chapterID,
				AreaNumber: areaNum,
				Title:      name,
				Summary:    "",
			}
			continue
		}

		// Parse area details
		if currentArea != nil {
			lower := strings.ToLower(trimmed)
			
			// Parse summary
			if strings.Contains(lower, "resumen") || strings.Contains(lower, "summary") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentArea.Summary = strings.TrimSpace(parts[1])
				}
			}

			// Parse description
			if strings.Contains(lower, "descripción") || strings.Contains(lower, "description") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					currentArea.Description = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// Save last area
	if currentArea != nil {
		areas = append(areas, *currentArea)
	}

	return areas, nil
}
