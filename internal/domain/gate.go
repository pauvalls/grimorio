package domain

import (
	"fmt"
	"time"
)

// GateStatus represents the state of a consistency gate evaluation
type GateStatus string

const (
	GateStatusPending    GateStatus = "pending"
	GateStatusValidating GateStatus = "validating"
	GateStatusApproved   GateStatus = "approved"
	GateStatusRejected   GateStatus = "rejected"
	GateStatusRetrying   GateStatus = "retrying"
)

// BatchProposal represents a collection of artifacts submitted for gate approval
type BatchProposal struct {
	BatchID    string            `json:"batch_id"`
	CampaignID string            `json:"campaign_id"`
	Artifacts  []ContentProposal `json:"artifacts"`
	Attempt    int               `json:"attempt"` // 1, 2, 3 (max 3)
}

// Validate checks if the batch proposal is valid
func (b *BatchProposal) Validate() error {
	if b.BatchID == "" {
		return NewValidationError("batch_id", "batch ID is required")
	}
	if b.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if len(b.Artifacts) == 0 {
		return NewValidationError("artifacts", "at least one artifact is required")
	}
	if b.Attempt < 1 {
		return NewValidationError("attempt", "attempt must be at least 1")
	}
	return nil
}

// GateResult represents the outcome of a gate evaluation
type GateResult struct {
	Status       GateStatus         `json:"status"`
	BatchID      string             `json:"batch_id"`
	Reports      []ValidationReport `json:"reports"`       // one per artifact
	RetryPrompt  string             `json:"retry_prompt,omitempty"`
	Suggestions  []Suggestion       `json:"suggestions"`
	CanonUpdated bool               `json:"canon_updated"`
	StateUpdated bool               `json:"state_updated"`
}

// Validate checks if the gate result is valid
func (g *GateResult) Validate() error {
	if g.BatchID == "" {
		return NewValidationError("batch_id", "batch ID is required")
	}
	switch g.Status {
	case GateStatusPending, GateStatusValidating, GateStatusApproved, GateStatusRejected, GateStatusRetrying:
		// valid
	default:
		return NewValidationError("status", fmt.Sprintf("invalid gate status: %s", g.Status))
	}
	return nil
}

// LockState tracks in-memory canon lock status
type LockState struct {
	CampaignID string    `json:"campaign_id"`
	LockedBy   string    `json:"locked_by"` // batch_id
	LockedAt   time.Time `json:"locked_at"`
	IsHeld     bool      `json:"is_held"`
}

// Validate checks if the lock state is valid
func (l *LockState) Validate() error {
	if l.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if l.IsHeld && l.LockedBy == "" {
		return NewValidationError("locked_by", "locked_by is required when lock is held")
	}
	return nil
}

// ConsistencyGate represents the gate state for a campaign
type ConsistencyGate struct {
	CampaignID string     `json:"campaign_id"`
	Status     GateStatus `json:"status"`
	BatchID    string     `json:"batch_id,omitempty"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// PipelineCheckpoint represents a saved state for rollback
type PipelineCheckpoint struct {
	CampaignID    string             `json:"campaign_id"`
	BatchID       string             `json:"batch_id"`
	CanonSnapshot *CanonDocument     `json:"canon_snapshot"`
	StateSnapshot *NarrativeState    `json:"state_snapshot"`
	CreatedAt     time.Time          `json:"created_at"`
}

// GateDecision represents a decision made by the gate
type GateDecision struct {
	CampaignID string     `json:"campaign_id"`
	BatchID    string     `json:"batch_id"`
	Decision   GateStatus `json:"decision"`
	Reason     string     `json:"reason"`
	Timestamp  time.Time  `json:"timestamp"`
}
