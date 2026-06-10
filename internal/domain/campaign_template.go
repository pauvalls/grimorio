package domain

// CampaignTemplate provides preset defaults for campaign creation.
type CampaignTemplate struct {
	Name          string   `json:"name"`
	Tone          string   `json:"tone"`
	SettingType   string   `json:"setting_type"`
	GameMode      string   `json:"game_mode"`
	DefaultThemes []string `json:"default_themes"`
}
