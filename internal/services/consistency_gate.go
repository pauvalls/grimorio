package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// ConsistencyGateService validates batches and manages gate lifecycle
type ConsistencyGateService struct {
	canonSvc   *CanonService
	stateSvc   *NarrativeStateService
	validator  *ValidationEngine
	locks      map[string]*domain.LockState
	results    map[string]*domain.GateResult // key: campaignID+batchID
	mu         sync.RWMutex
	maxRetries int
}

// NewConsistencyGateService creates a new consistency gate service
func NewConsistencyGateService(
	canonSvc *CanonService,
	stateSvc *NarrativeStateService,
	validator *ValidationEngine,
) *ConsistencyGateService {
	return &ConsistencyGateService{
		canonSvc:   canonSvc,
		stateSvc:   stateSvc,
		validator:  validator,
		locks:      make(map[string]*domain.LockState),
		results:    make(map[string]*domain.GateResult),
		maxRetries: 2,
	}
}

// ProcessBatch validates a batch proposal and returns a gate result
func (s *ConsistencyGateService) ProcessBatch(
	ctx context.Context,
	proposal domain.BatchProposal,
	fastMode bool,
) (*domain.GateResult, error) {
	if err := proposal.Validate(); err != nil {
		return nil, fmt.Errorf("invalid proposal: %w", err)
	}

	// 1. Acquire canon lock
	if !s.acquireLock(proposal.CampaignID, proposal.BatchID) {
		return nil, fmt.Errorf("canon locked by %s", s.locks[proposal.CampaignID].LockedBy)
	}
	defer s.releaseLock(proposal.CampaignID, proposal.BatchID)

	// 2. Transition to validating
	result := &domain.GateResult{
		BatchID: proposal.BatchID,
		Status:  domain.GateStatusValidating,
	}

	// 3. Validate each artifact; collect reports
	for _, artifact := range proposal.Artifacts {
		report, _ := s.validator.validate(ctx, proposal.CampaignID, artifact)
		if fastMode {
			report = s.filterNonCritical(report)
		}
		result.Reports = append(result.Reports, *report)
	}

	// 4. Aggregate decision
	hasCritical := false
	hasError := false
	for _, r := range result.Reports {
		for _, c := range r.Checks {
			if !c.Passed {
				if c.Severity == "critical" {
					hasCritical = true
				}
				if c.Severity == "error" {
					hasError = true
				}
			}
		}
	}

	switch {
	case hasCritical || hasError:
		result.Status = domain.GateStatusRejected
		result.Suggestions = s.aggregateSuggestions(result.Reports)
		result.RetryPrompt = s.renderRetryPrompt(proposal, result.Suggestions)
		if proposal.Attempt >= s.maxRetries {
			result.RetryPrompt += "\n[Maximum retries exceeded. Human review required.]"
		}
	default:
		result.Status = domain.GateStatusApproved
		// 5. Auto-save canon + state atomically
		if err := s.autoSave(ctx, proposal.CampaignID); err != nil {
			return nil, fmt.Errorf("auto-save failed: %w", err)
		}
		result.CanonUpdated = true
		result.StateUpdated = true
	}

	// Store result for retrieval
	s.mu.Lock()
	s.results[proposal.CampaignID+":"+proposal.BatchID] = result
	s.mu.Unlock()

	return result, nil
}

// GetGateStatus retrieves the status of a gate evaluation
func (s *ConsistencyGateService) GetGateStatus(
	ctx context.Context,
	campaignID string,
	batchID string,
) (*domain.GateResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, exists := s.results[campaignID+":"+batchID]
	if exists {
		return result, nil
	}

	// Fallback to lock state
	lock, exists := s.locks[campaignID]
	if !exists {
		return nil, fmt.Errorf("no gate status found for campaign %s", campaignID)
	}

	// Return current status based on lock state
	status := domain.GateStatusPending
	if lock.IsHeld {
		status = domain.GateStatusValidating
	}

	return &domain.GateResult{
		BatchID: batchID,
		Status:  status,
	}, nil
}

// ResetGate resets the gate state for a campaign
func (s *ConsistencyGateService) ResetGate(
	ctx context.Context,
	campaignID string,
	batchID string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, exists := s.locks[campaignID]
	if exists && lock.IsHeld && lock.LockedBy == batchID {
		lock.IsHeld = false
		lock.LockedBy = ""
	}

	return nil
}

// acquireLock attempts to acquire the canon lock for a campaign
func (s *ConsistencyGateService) acquireLock(campaignID, batchID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, exists := s.locks[campaignID]
	if exists && lock.IsHeld {
		return false
	}

	s.locks[campaignID] = &domain.LockState{
		CampaignID: campaignID,
		LockedBy:   batchID,
		LockedAt:   time.Now(),
		IsHeld:     true,
	}
	return true
}

// releaseLock releases the canon lock for a campaign
func (s *ConsistencyGateService) releaseLock(campaignID, batchID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, exists := s.locks[campaignID]
	if exists && lock.LockedBy == batchID {
		lock.IsHeld = false
		lock.LockedBy = ""
	}
}

// isLocked checks if a campaign is currently locked
func (s *ConsistencyGateService) isLocked(campaignID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lock, exists := s.locks[campaignID]
	return exists && lock.IsHeld
}

// autoSave persists canon and narrative state
func (s *ConsistencyGateService) autoSave(ctx context.Context, campaignID string) error {
	doc, err := s.canonSvc.LoadCanon(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to load canon for auto-save: %w", err)
	}

	if err := s.canonSvc.SaveCanon(ctx, doc); err != nil {
		return fmt.Errorf("failed to save canon: %w", err)
	}

	state, err := s.stateSvc.Load(ctx, campaignID)
	if err != nil {
		return fmt.Errorf("failed to load state for auto-save: %w", err)
	}

	if err := s.stateSvc.Save(ctx, state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	return nil
}

// filterNonCritical removes non-critical checks from a report
func (s *ConsistencyGateService) filterNonCritical(report *domain.ValidationReport) *domain.ValidationReport {
	filtered := &domain.ValidationReport{
		ArtifactID:    report.ArtifactID,
		ArtifactType:  report.ArtifactType,
		OverallStatus: report.OverallStatus,
		Checks:        []domain.CheckResult{},
		Suggestions:   report.Suggestions,
		GateStatus:    report.GateStatus,
		CanonUpdates:  report.CanonUpdates,
	}

	for _, check := range report.Checks {
		if check.Severity == "critical" || check.Severity == "error" {
			filtered.Checks = append(filtered.Checks, check)
		}
	}

	filtered.ComputeOverallStatus()
	return filtered
}

// aggregateSuggestions collects all suggestions from reports
func (s *ConsistencyGateService) aggregateSuggestions(reports []domain.ValidationReport) []domain.Suggestion {
	var suggestions []domain.Suggestion
	seen := make(map[string]bool)

	for _, report := range reports {
		for _, suggestion := range report.Suggestions {
			key := suggestion.Problem + suggestion.Fix
			if !seen[key] {
				seen[key] = true
				suggestions = append(suggestions, suggestion)
			}
		}
	}

	return suggestions
}

// renderRetryPrompt generates a natural-language retry prompt
func (s *ConsistencyGateService) renderRetryPrompt(proposal domain.BatchProposal, suggestions []domain.Suggestion) string {
	var builder strings.Builder

	fmt.Fprintf(&builder, "Batch %s (attempt %d) was rejected by the consistency gate.\n\n", proposal.BatchID, proposal.Attempt)
	builder.WriteString("Issues found:\n")

	for i, suggestion := range suggestions {
		fmt.Fprintf(&builder, "%d. %s\n", i+1, suggestion.Problem)
		fmt.Fprintf(&builder, "   Fix: %s\n", suggestion.Fix)
		fmt.Fprintf(&builder, "   Rationale: %s\n", suggestion.Rationale)
	}

	builder.WriteString("\nPlease fix these issues and resubmit the batch.")

	return builder.String()
}
