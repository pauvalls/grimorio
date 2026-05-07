package domain

// TableType represents the type of random table
type TableType string

const (
	TableTypeEncounter TableType = "encounter"
	TableTypeRumor     TableType = "rumor"
	TableTypeWeather   TableType = "weather"
	TableTypeTreasure  TableType = "treasure"
)

// RandomTable represents a generated random table
type RandomTable struct {
	CampaignID string            `json:"campaign_id"`
	TableType  TableType         `json:"table_type"`
	Context    TableContext      `json:"context"`
	Entries    []RandomTableEntry `json:"entries"`
}

// RandomTableEntry is a single entry in a random table
type RandomTableEntry struct {
	Weight      int    `json:"weight"`
	Description string `json:"description"`
	SourceFact  string `json:"source_fact"`
}

// TableContext provides filtering context for table generation
type TableContext struct {
	LevelRange   string `json:"level_range,omitempty"`
	SettingType  string `json:"setting_type,omitempty"`
	PartySize    int    `json:"party_size,omitempty"`
	LocationHint string `json:"location_hint,omitempty"`
}

// IsValidTableType checks if a string is a valid table type
func IsValidTableType(t string) bool {
	switch TableType(t) {
	case TableTypeEncounter, TableTypeRumor, TableTypeWeather, TableTypeTreasure:
		return true
	default:
		return false
	}
}
