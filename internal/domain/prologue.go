package domain

import "time"

// ProloguePart represents a single section of a narrative prologue.
type ProloguePart struct {
	Order       int    `json:"order"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsReadAloud bool   `json:"is_read_aloud"`
}

// Prologue represents a structured 4-part narrative introduction for a campaign.
type Prologue struct {
	CampaignID  string         `json:"campaign_id"`
	Tone        string         `json:"tone"`
	Parts       []ProloguePart `json:"parts"`
	GeneratedAt time.Time      `json:"generated_at"`
}

// Validate returns true if the prologue has exactly 4 parts with sequential order.
func (p *Prologue) Validate() bool {
	if len(p.Parts) != 4 {
		return false
	}
	for i, part := range p.Parts {
		if part.Order != i+1 {
			return false
		}
	}
	return true
}
