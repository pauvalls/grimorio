package domain

import (
	"fmt"
	"time"
)

// Severity represents the severity level of a validation issue or health finding
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// HealthStatus represents overall campaign health
type HealthStatus string

const (
	HealthStatusExcellent HealthStatus = "excellent" // 0 issues
	HealthStatusGood      HealthStatus = "good"      // info only
	HealthStatusFair      HealthStatus = "fair"      // warnings only
	HealthStatusPoor      HealthStatus = "poor"      // errors present
	HealthStatusCritical  HealthStatus = "critical"  // critical issues present
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
	CampaignID     string             `json:"campaign_id"`
	BatchID        string             `json:"batch_id"`
	CanonSnapshot  *CanonDocument     `json:"canon_snapshot"`
	StateSnapshot  *NarrativeState    `json:"state_snapshot"`
	CreatedAt      time.Time          `json:"created_at"`
	SessionNum     int                `json:"session_num"`           // Session at checkpoint time
	ChapterID      string             `json:"chapter_id,omitempty"`  // Chapter at checkpoint time (V3+)
	CheckpointHash string             `json:"checkpoint_hash"`       // SHA256 of canon+state for integrity
	CheckpointType string             `json:"checkpoint_type"`       // "session" or "chapter"
}

// Validate checks if the checkpoint is valid
func (p *PipelineCheckpoint) Validate() error {
	if p.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if p.BatchID == "" {
		return NewValidationError("batch_id", "batch ID is required")
	}
	if p.CanonSnapshot == nil {
		return NewValidationError("canon_snapshot", "canon snapshot is required")
	}
	if p.StateSnapshot == nil {
		return NewValidationError("state_snapshot", "state snapshot is required")
	}
	if p.CheckpointHash == "" {
		return NewValidationError("checkpoint_hash", "checkpoint hash is required")
	}
	return nil
}

// AuditLogEntry represents a single audit log record
type AuditLogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	CampaignID  string    `json:"campaign_id"`
	BatchID     string    `json:"batch_id"`
	ArtifactIDs []string  `json:"artifact_ids"`
	Decision    string    `json:"decision"` // approved, rejected
	Reason      string    `json:"reason"`
	UserID      string    `json:"user_id,omitempty"` // If available from context
	SessionNum  int       `json:"session_num,omitempty"`
}

// Validate checks if the audit log entry is valid
func (a *AuditLogEntry) Validate() error {
	if a.CampaignID == "" {
		return NewValidationError("campaign_id", "campaign ID is required")
	}
	if a.BatchID == "" {
		return NewValidationError("batch_id", "batch ID is required")
	}
	if a.Decision == "" {
		return NewValidationError("decision", "decision is required")
	}
	return nil
}

// GateDecision represents a decision made by the gate
type GateDecision struct {
	CampaignID string     `json:"campaign_id"`
	BatchID    string     `json:"batch_id"`
	Decision   GateStatus `json:"decision"`
	Reason     string     `json:"reason"`
	Timestamp  time.Time  `json:"timestamp"`
}
