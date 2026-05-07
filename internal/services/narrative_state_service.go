package services

import (
	"context"
	"fmt"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// NarrativeStateService handles narrative state business logic
type NarrativeStateService struct {
	stateRepo repository.NarrativeStateRepository
	canonRepo repository.CanonRepository
}

// NewNarrativeStateService creates a new narrative state service
func NewNarrativeStateService(stateRepo repository.NarrativeStateRepository, canonRepo repository.CanonRepository) *NarrativeStateService {
	return &NarrativeStateService{
		stateRepo: stateRepo,
		canonRepo: canonRepo,
	}
}

// Load retrieves the narrative state for a campaign
func (s *NarrativeStateService) Load(ctx context.Context, campaignID string) (*domain.NarrativeState, error) {
	return s.stateRepo.Load(campaignID)
}

// Save persists a narrative state
func (s *NarrativeStateService) Save(ctx context.Context, state *domain.NarrativeState) error {
	if state == nil {
		return domain.NewValidationError("state", "narrative state is required")
	}
	state.LastUpdated = time.Now()
	return s.stateRepo.Save(state.CampaignID, state)
}

// Update applies a batch update to the narrative state
func (s *NarrativeStateService) Update(ctx context.Context, campaignID string, update domain.StateUpdate) (*domain.NarrativeState, error) {
	state, err := s.stateRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load narrative state: %w", err)
	}

	// Append revealed clues
	state.RevealedClues = append(state.RevealedClues, update.RevealedClues...)

	// Move completed quests from active to completed
	if len(update.CompletedQuests) > 0 {
		completedSet := make(map[string]bool)
		for _, id := range update.CompletedQuests {
			completedSet[id] = true
		}

		var remainingActive []domain.QuestState
		for _, quest := range state.ActiveQuests {
			if completedSet[quest.ID] {
				quest.Status = "completed"
				state.CompletedQuests = append(state.CompletedQuests, quest)
			} else {
				remainingActive = append(remainingActive, quest)
			}
		}
		state.ActiveQuests = remainingActive
	}

	// Append new quests
	state.ActiveQuests = append(state.ActiveQuests, update.NewQuests...)

	// Append dead NPCs
	state.DeadNPCs = append(state.DeadNPCs, update.DeadNPCs...)

	// Append/update key items
	for _, newItem := range update.KeyItems {
		found := false
		for i := range state.KeyItems {
			if state.KeyItems[i].ID == newItem.ID {
				state.KeyItems[i] = newItem
				found = true
				break
			}
		}
		if !found {
			state.KeyItems = append(state.KeyItems, newItem)
		}
	}

	// Append session log
	if update.SessionNum > 0 || update.SessionSummary != "" {
		record := domain.SessionRecord{
			SessionNum:   update.SessionNum,
			Date:         time.Now(),
			Summary:      update.SessionSummary,
			KeyDecisions: update.KeyDecisions,
			XPAwarded:    update.XPAwarded,
			LootAcquired: update.LootAcquired,
			DMNotes:      update.DMNotes,
		}
		state.SessionLog = append(state.SessionLog, record)
		state.CurrentSession = update.SessionNum
	}

	state.LastUpdated = time.Now()

	if err := s.stateRepo.Save(campaignID, state); err != nil {
		return nil, fmt.Errorf("failed to save narrative state: %w", err)
	}

	return state, nil
}

// GetSessionPrepContext generates preparation context for the next session
func (s *NarrativeStateService) GetSessionPrepContext(ctx context.Context, campaignID string, nextSession int) (*domain.SessionPrepContext, error) {
	state, err := s.stateRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load narrative state: %w", err)
	}

	ctxPrep := &domain.SessionPrepContext{
		ActiveQuests: state.ActiveQuests,
		DMWarnings:   []string{},
	}

	// PreviouslyOn from last session log
	if len(state.SessionLog) > 0 {
		lastLog := state.SessionLog[len(state.SessionLog)-1]
		ctxPrep.PreviouslyOn = lastLog.Summary
	}

	// DM Warnings: dead NPCs
	for _, death := range state.DeadNPCs {
		ctxPrep.DMWarnings = append(ctxPrep.DMWarnings, fmt.Sprintf("%s está muerto", death.Name))
	}

	// Relevant NPCs: quest givers looked up in canon
	if s.canonRepo != nil {
		canon, canonErr := s.canonRepo.Load(campaignID)
		if canonErr == nil {
			giverSet := make(map[string]bool)
			for _, quest := range state.ActiveQuests {
				if quest.GiverNPC != "" {
					giverSet[quest.GiverNPC] = true
				}
			}

			for _, entity := range canon.Entities {
				if giverSet[entity.ID] && entity.CanonState == domain.EntityStateAlive {
					ctxPrep.RelevantNPCs = append(ctxPrep.RelevantNPCs, entity)
				}
			}
		}
	}

	return ctxPrep, nil
}
