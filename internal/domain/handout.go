package domain

// HandoutType represents the type of handout
type HandoutType string

const (
	HandoutTypeSummary   HandoutType = "summary"
	HandoutTypeEncounter HandoutType = "encounter"
	HandoutTypeQuest     HandoutType = "quest"
	HandoutTypeLore      HandoutType = "lore"
	HandoutTypeFaction   HandoutType = "faction"
)

// HandoutVersion represents which version of a handout to generate
type HandoutVersion string

const (
	HandoutVersionPlayer HandoutVersion = "player"
	HandoutVersionDM     HandoutVersion = "dm"
	HandoutVersionBoth   HandoutVersion = "both"
)

// Handout represents a generated handout with dual versions
type Handout struct {
	CampaignID   string       `json:"campaign_id"`
	HandoutType  HandoutType  `json:"handout_type"`
	ContentRefs  []string     `json:"content_refs"`
	PlayerVersion string      `json:"player_version"`
	DMVersion     string      `json:"dm_version"`
	MarkdownPath  string      `json:"markdown_path,omitempty"`
	SVGPath       string      `json:"svg_path,omitempty"`
}

// IsValidHandoutType checks if a string is a valid handout type
func IsValidHandoutType(t string) bool {
	switch HandoutType(t) {
	case HandoutTypeSummary, HandoutTypeEncounter, HandoutTypeQuest, HandoutTypeLore, HandoutTypeFaction:
		return true
	default:
		return false
	}
}
