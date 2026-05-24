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
		// If no state exists, create an initial one
		state = &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     campaignID,
			CurrentSession: 0,
			RevealedClues:  []domain.RevealedClue{},
			ActiveQuests:   []domain.QuestState{},
			CompletedQuests: []domain.QuestState{},
			FailedQuests:   []domain.QuestState{},
			DeadNPCs:       []domain.NPCDeathRecord{},
			KeyItems:       []domain.KeyItem{},
			SessionLog:     []domain.SessionRecord{},
			DMOverrides:    []domain.DMOverride{},
			LastUpdated:    time.Now(),
		}
		// Save the initial state
		if saveErr := s.stateRepo.Save(campaignID, state); saveErr != nil {
			return nil, fmt.Errorf("failed to create initial narrative state: %w", saveErr)
		}
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

	// Replace quest state (active_quests represents current state, not delta)
	if len(update.NewQuests) > 0 {
		state.ActiveQuests = update.NewQuests
		state.QuestNames = nil
		for _, q := range update.NewQuests {
			if q.Name != "" {
				state.QuestNames = append(state.QuestNames, q.Name)
			}
		}
	}

	// Append dead NPCs
	state.DeadNPCs = append(state.DeadNPCs, update.DeadNPCs...)

	// Replace key items state (key_items represents current state, not delta)
	if len(update.KeyItems) > 0 {
		state.KeyItems = update.KeyItems
		state.ItemNames = nil
		for _, item := range update.KeyItems {
			if item.Name != "" {
				state.ItemNames = append(state.ItemNames, item.Name)
			}
		}
	}

	// Append session log
	sessionNum := update.SessionNum
	if sessionNum < 0 {
		return nil, domain.NewValidationError("session_num", "session_num cannot be negative")
	}
	if sessionNum == 0 {
		sessionNum = state.CurrentSession + 1
	}
	// Save session metadata to root state for easy access
	state.DMNotes = update.DMNotes
	state.LootAcquired = update.LootAcquired

	if sessionNum > 0 || update.SessionSummary != "" {
		record := domain.SessionRecord{
			SessionNum:   sessionNum,
			Date:         time.Now(),
			Summary:      update.SessionSummary,
			KeyDecisions: update.KeyDecisions,
			XPAwarded:    update.XPAwarded,
			LootAcquired: update.LootAcquired,
			DMNotes:      update.DMNotes,
		}
		state.SessionLog = append(state.SessionLog, record)
		state.CurrentSession = sessionNum
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
