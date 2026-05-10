package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HandoutRenderer generates player-facing handout content
type HandoutRenderer interface {
	GeneratePlayerMap(mapName string) (string, error)
	GenerateClueList() (string, error)
	GenerateNPCReference() (string, error)
	GenerateSessionRecap() (string, error)
}

// generateHandouts creates handout HTML pages for the compiled PDF
func (c *Compiler) generateHandouts() (string, error) {
	if c.CompilerVersion != 2 {
		return "", nil // handouts only in v2
	}

	var parts []string

	// Handout header
	parts = append(parts, `<div class="handout-page" id="handouts">`)
	parts = append(parts, `<h1>Player Handouts</h1>`)
	parts = append(parts, `<!-- <p class="handout-note">Estas páginas están diseñadas para imprimir y entregar a los jugadores.</p> -->`)

	// Try to generate each handout type
	handoutGen := c.handoutRenderer()
	if handoutGen != nil {
		// Clue list
		if clues, err := handoutGen.GenerateClueList(); err == nil && clues != "" {
			parts = append(parts, `<div class="handout-section">`)
			parts = append(parts, markdownToHTML(clues, c.CampaignDir))
			parts = append(parts, `</div>`)
		}

		// NPC reference
		if npcRef, err := handoutGen.GenerateNPCReference(); err == nil && npcRef != "" {
			parts = append(parts, `<div class="handout-section">`)
			parts = append(parts, markdownToHTML(npcRef, c.CampaignDir))
			parts = append(parts, `</div>`)
		}

		// Session recap
		if recap, err := handoutGen.GenerateSessionRecap(); err == nil && recap != "" {
			parts = append(parts, `<div class="handout-section">`)
			parts = append(parts, markdownToHTML(recap, c.CampaignDir))
			parts = append(parts, `</div>`)
		}
	}

	// Player maps
	mapsDir := filepath.Join(c.CampaignDir, "maps")
	if info, err := os.Stat(mapsDir); err == nil && info.IsDir() {
		files, _ := os.ReadDir(mapsDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".md") {
				mapName := strings.TrimSuffix(f.Name(), ".md")
				if handoutGen != nil {
					if redacted, err := handoutGen.GeneratePlayerMap(mapName); err == nil && redacted != "" {
						parts = append(parts, `<div class="handout-section handout-map">`)
						parts = append(parts, fmt.Sprintf(`<h2>Mapa: %s</h2>`, mapName))
						parts = append(parts, markdownToHTML(redacted, c.CampaignDir))
						parts = append(parts, `</div>`)
					}
				}
			}
		}
	}

	parts = append(parts, `</div>`)

	return strings.Join(parts, "\n"), nil
}

// handoutRenderer returns a HandoutRenderer for the current campaign
func (c *Compiler) handoutRenderer() HandoutRenderer {
	// Import the services package — we can't import directly due to cycle,
	// so we use an interface approach. The actual renderer is created by
	// the caller and set on the Compiler.
	return c.handoutRendererImpl
}
