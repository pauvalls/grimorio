package domain

import (
	"time"
)

// Checkpoint represents a saved session or chapter state for recovery
type Checkpoint struct {
	ID             string          `json:"id"`
	CampaignID     string          `json:"campaign_id"`
	CheckpointType string          `json:"checkpoint_type"` // "session_end" or "chapter_complete"
	SessionNum     int             `json:"session_num"`
	ChapterID      string          `json:"chapter_id,omitempty"`
	State          *NarrativeState `json:"state"`
	CanonHash      string          `json:"canon_hash"` // SHA256 hash of canon at checkpoint time
	Metadata       map[string]any  `json:"metadata,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// Validate checks if the checkpoint is valid
func (c *Checkpoint) Validate() error {
	if c.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if c.CheckpointType == "" {
		return NewValidationError("checkpoint_type", "checkpoint type is required")
	}
	if c.SessionNum < 0 {
		return NewValidationError("session_num", "session number cannot be negative")
	}
	if c.State == nil {
		return NewValidationError("state", "state is required")
	}
	if c.CanonHash == "" {
		return NewValidationError("canon_hash", "canon hash is required")
	}
	return nil
}
