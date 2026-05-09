package domain

// HandoutType represents the type of handout
type HandoutType string

const (
	HandoutTypeSummary   HandoutType = "summary"
	HandoutTypeEncounter HandoutType = "encounter"
	HandoutTypeQuest     HandoutType = "quest"
	HandoutTypeLore      HandoutType = "lore"
	HandoutTypeFaction   HandoutType = "faction"
	// V3 enhanced handout types
	HandoutTypeLetter     HandoutType = "letter"
	HandoutTypeMap        HandoutType = "map"
	HandoutTypeClue       HandoutType = "clue"
	HandoutTypeDocument   HandoutType = "document"
	HandoutTypeJournal    HandoutType = "journal"
	HandoutTypeProclamation HandoutType = "proclamation"
	HandoutTypeArtifact   HandoutType = "artifact"
)

// HandoutFormat represents the format of a handout (V3)
type HandoutFormat string

const (
	FormatText  HandoutFormat = "text"
	FormatImage HandoutFormat = "image"
	FormatMixed HandoutFormat = "mixed"
)

// HandoutStyle represents the stylistic approach for a handout (V3)
type HandoutStyle string

const (
	StyleFormal     HandoutStyle = "formal"
	StyleInformal   HandoutStyle = "informal"
	StyleAncient    HandoutStyle = "ancient"
	StyleUrgent     HandoutStyle = "urgent"
	StyleMysterious HandoutStyle = "mysterious"
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
	// V3 enhanced fields
	ID               string        `json:"id,omitempty"`
	Type             HandoutType   `json:"type,omitempty"`
	Title            string        `json:"title,omitempty"`
	Content          string        `json:"content,omitempty"`
	DMNotes          string        `json:"dm_notes,omitempty"`
	QuestRefs        []string      `json:"quest_refs,omitempty"`
	AreaRefs         []string      `json:"area_refs,omitempty"`
	Format           HandoutFormat `json:"format,omitempty"`
	Style            HandoutStyle  `json:"style,omitempty"`
	RevealConditions []string      `json:"reveal_conditions,omitempty"`
}

// IsValidHandoutType checks if a string is a valid handout type
func IsValidHandoutType(t string) bool {
	switch HandoutType(t) {
	case HandoutTypeSummary, HandoutTypeEncounter, HandoutTypeQuest, HandoutTypeLore, HandoutTypeFaction,
		HandoutTypeLetter, HandoutTypeMap, HandoutTypeClue, HandoutTypeDocument,
		HandoutTypeJournal, HandoutTypeProclamation, HandoutTypeArtifact:
		return true
	default:
		return false
	}
}

// IsValidHandoutFormat checks if a format is valid (V3)
func IsValidHandoutFormat(f HandoutFormat) bool {
	switch f {
	case FormatText, FormatImage, FormatMixed:
		return true
	default:
		return false
	}
}

// IsValidHandoutStyle checks if a style is valid (V3)
func IsValidHandoutStyle(s HandoutStyle) bool {
	switch s {
	case StyleFormal, StyleInformal, StyleAncient, StyleUrgent, StyleMysterious:
		return true
	default:
		return false
	}
}

// Validate checks handout validity (V3)
func (h *Handout) Validate() error {
	if h.Type == "" {
		return nil // Use legacy validation
	}
	if !IsValidHandoutType(string(h.Type)) {
		return NewValidationError("type", "invalid handout type: "+string(h.Type))
	}
	if h.Title == "" {
		return NewValidationError("title", "title is required")
	}
	if h.Content == "" {
		return NewValidationError("content", "content is required")
	}
	// Clue handouts must have reveal conditions
	if h.Type == HandoutTypeClue && len(h.RevealConditions) == 0 {
		return NewValidationError("reveal_conditions", "clue handouts must have reveal conditions")
	}
	return nil
}

// HasQuestRef checks if handout references a specific quest (V3)
func (h *Handout) HasQuestRef(questID string) bool {
	for _, ref := range h.QuestRefs {
		if ref == questID {
			return true
		}
	}
	return false
}

// HasAreaRef checks if handout references a specific area (V3)
func (h *Handout) HasAreaRef(areaID string) bool {
	for _, ref := range h.AreaRefs {
		if ref == areaID {
			return true
		}
	}
	return false
}
