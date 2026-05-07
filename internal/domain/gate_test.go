package domain

import (
	"testing"
	"time"
)

func TestGateStatus_ValidValues(t *testing.T) {
	validStatuses := []GateStatus{
		GateStatusPending,
		GateStatusValidating,
		GateStatusApproved,
		GateStatusRejected,
		GateStatusRetrying,
	}
	for _, status := range validStatuses {
		if status == "" {
			t.Fatalf("GateStatus should not be empty")
		}
	}
}

func TestBatchProposal_Validate(t *testing.T) {
	tests := []struct {
		name    string
		proposal BatchProposal
		wantErr bool
	}{
		{
			name: "valid batch proposal",
			proposal: BatchProposal{
				BatchID:    "batch-001",
				CampaignID: "test-campaign",
				Artifacts: []ContentProposal{
					{ID: "npc-001", Type: "npc", Content: "Test NPC"},
				},
				Attempt: 1,
			},
			wantErr: false,
		},
		{
			name: "missing batch ID",
			proposal: BatchProposal{
				CampaignID: "test-campaign",
				Artifacts:  []ContentProposal{{ID: "npc-001", Type: "npc", Content: "Test NPC"}},
				Attempt:    1,
			},
			wantErr: true,
		},
		{
			name: "missing campaign ID",
			proposal: BatchProposal{
				BatchID:   "batch-001",
				Artifacts: []ContentProposal{{ID: "npc-001", Type: "npc", Content: "Test NPC"}},
				Attempt:   1,
			},
			wantErr: true,
		},
		{
			name: "empty artifacts",
			proposal: BatchProposal{
				BatchID:    "batch-001",
				CampaignID: "test-campaign",
				Artifacts:  []ContentProposal{},
				Attempt:    1,
			},
			wantErr: true,
		},
		{
			name: "invalid attempt number",
			proposal: BatchProposal{
				BatchID:    "batch-001",
				CampaignID: "test-campaign",
				Artifacts:  []ContentProposal{{ID: "npc-001", Type: "npc", Content: "Test NPC"}},
				Attempt:    0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.proposal.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGateResult_Validate(t *testing.T) {
	tests := []struct {
		name    string
		result  GateResult
		wantErr bool
	}{
		{
			name: "valid approved result",
			result: GateResult{
				Status:  GateStatusApproved,
				BatchID: "batch-001",
				Reports: []ValidationReport{},
			},
			wantErr: false,
		},
		{
			name: "valid rejected result with suggestions",
			result: GateResult{
				Status:      GateStatusRejected,
				BatchID:     "batch-001",
				Reports:     []ValidationReport{},
				Suggestions: []Suggestion{{Problem: "NPC is dead", Fix: "Replace", Rationale: "Canon"}},
			},
			wantErr: false,
		},
		{
			name: "missing batch ID",
			result: GateResult{
				Status:  GateStatusApproved,
				Reports: []ValidationReport{},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			result: GateResult{
				Status:  "unknown",
				BatchID: "batch-001",
				Reports: []ValidationReport{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.result.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestLockState_Validate(t *testing.T) {
	tests := []struct {
		name     string
		lock     LockState
		wantErr  bool
	}{
		{
			name: "valid held lock",
			lock: LockState{
				CampaignID: "test-campaign",
				LockedBy:   "batch-001",
				LockedAt:   time.Now(),
				IsHeld:     true,
			},
			wantErr: false,
		},
		{
			name: "valid free lock",
			lock: LockState{
				CampaignID: "test-campaign",
				IsHeld:     false,
			},
			wantErr: false,
		},
		{
			name: "missing campaign ID",
			lock: LockState{
				LockedBy: "batch-001",
				LockedAt: time.Now(),
				IsHeld:   true,
			},
			wantErr: true,
		},
		{
			name: "held lock without locked_by",
			lock: LockState{
				CampaignID: "test-campaign",
				LockedAt:   time.Now(),
				IsHeld:     true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.lock.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestConsistencyGate_StateTransitions(t *testing.T) {
	gate := &ConsistencyGate{
		CampaignID: "test-campaign",
		Status:     GateStatusPending,
	}

	// Transition to validating
	gate.Status = GateStatusValidating
	if gate.Status != GateStatusValidating {
		t.Fatalf("expected status validating, got %s", gate.Status)
	}

	// Transition to approved
	gate.Status = GateStatusApproved
	if gate.Status != GateStatusApproved {
		t.Fatalf("expected status approved, got %s", gate.Status)
	}

	// Reset to pending
	gate.Status = GateStatusPending
	if gate.Status != GateStatusPending {
		t.Fatalf("expected status pending after reset, got %s", gate.Status)
	}
}

func TestContentProposal_BatchID(t *testing.T) {
	proposal := ContentProposal{
		ID:      "test-001",
		Type:    "npc",
		Content: "Test content",
		BatchID: "batch-001",
	}
	if proposal.BatchID != "batch-001" {
		t.Fatalf("expected batch_id 'batch-001', got %s", proposal.BatchID)
	}
}

func TestValidationReport_GateStatus(t *testing.T) {
	report := ValidationReport{
		ArtifactID:    "test-001",
		ArtifactType:  "npc",
		OverallStatus: "approved",
		GateStatus:    GateStatusApproved,
	}
	if report.GateStatus != GateStatusApproved {
		t.Fatalf("expected gate_status %s, got %s", GateStatusApproved, report.GateStatus)
	}
}
