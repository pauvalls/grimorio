package domain

// AdaptationPatch represents a markdown diff generated from a world event
type AdaptationPatch struct {
	CampaignID      string              `json:"campaign_id"`
	WorldEvent      WorldEvent          `json:"world_event"`
	AffectedActs    []string            `json:"affected_acts"`
	AffectedNPCs    []string            `json:"affected_npcs"`
	AffectedQuests  []string            `json:"affected_quests"`
	Instructions    []PatchInstruction  `json:"instructions"`
	MarkdownDiff    string              `json:"markdown_diff"`
	IsEmpty         bool                `json:"is_empty"`
}

// PatchInstruction describes a single apply/discard instruction
type PatchInstruction struct {
	Target      string `json:"target"`
	TargetID    string `json:"target_id"`
	Action      string `json:"action"`      // "apply" or "discard"
	Description string `json:"description"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
}
