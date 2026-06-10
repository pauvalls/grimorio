package services

import (
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
)

// campaignTemplates holds the built-in campaign presets.
var campaignTemplates = map[string]domain.CampaignTemplate{
	"urban fantasy": {
		Name:          "Urban Fantasy",
		Tone:          "grim",
		SettingType:   "urban",
		GameMode:      "sandbox_urbano",
		DefaultThemes: []string{"corruption", "magic", "mystery"},
	},
	"gothic horror": {
		Name:          "Gothic Horror",
		Tone:          "horror",
		SettingType:   "wilderness",
		GameMode:      "investigacion",
		DefaultThemes: []string{"madness", "decay", "superstition"},
	},
	"maritime adventure": {
		Name:          "Maritime Adventure",
		Tone:          "heroic",
		SettingType:   "maritime",
		GameMode:      "viaje",
		DefaultThemes: []string{"exploration", "piracy", "sea monsters"},
	},
	"dungeon crawl": {
		Name:          "Dungeon Crawl",
		Tone:          "heroic",
		SettingType:   "dungeon",
		GameMode:      "dungeon_lineal",
		DefaultThemes: []string{"exploration", "combat", "treasure"},
	},
	"political intrigue": {
		Name:          "Political Intrigue",
		Tone:          "political",
		SettingType:   "urban",
		GameMode:      "intriga",
		DefaultThemes: []string{"espionage", "betrayal", "power"},
	},
}

// GetTemplate retrieves a campaign template by name (case-insensitive).
func GetTemplate(name string) (domain.CampaignTemplate, error) {
	key := strings.ToLower(strings.TrimSpace(name))
	tmpl, ok := campaignTemplates[key]
	if !ok {
		return domain.CampaignTemplate{}, fmt.Errorf("template not found: %s", name)
	}
	return tmpl, nil
}

// ApplyTemplate applies preset defaults to a campaign when fields are empty.
func ApplyTemplate(campaign *domain.Campaign, tmpl domain.CampaignTemplate) {
	if campaign.Status == "" {
		campaign.Status = "active"
	}
}
