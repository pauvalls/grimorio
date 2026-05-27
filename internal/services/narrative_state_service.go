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
	stateRepo   repository.NarrativeStateRepository
	canonRepo   repository.CanonRepository
	characterRepo repository.CharacterRepository
}

// NewNarrativeStateService creates a new narrative state service
func NewNarrativeStateService(stateRepo repository.NarrativeStateRepository, canonRepo repository.CanonRepository) *NarrativeStateService {
	return &NarrativeStateService{
		stateRepo: stateRepo,
		canonRepo: canonRepo,
	}
}

// SetCharacterRepository sets the character repository for HP persistence
func (s *NarrativeStateService) SetCharacterRepository(repo repository.CharacterRepository) {
	s.characterRepo = repo
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

// dedupClues merges incoming clues into existing, skipping duplicates by ID (first wins)
func dedupClues(existing, incoming []domain.RevealedClue) []domain.RevealedClue {
	seen := make(map[string]bool, len(existing)+len(incoming))
	result := make([]domain.RevealedClue, 0, len(existing)+len(incoming))

	for _, c := range existing {
		if !seen[c.ID] {
			seen[c.ID] = true
			result = append(result, c)
		}
	}
	for _, c := range incoming {
		if !seen[c.ID] {
			seen[c.ID] = true
			result = append(result, c)
		}
	}
	return result
}

// dedupDeadNPCs merges incoming dead NPCs into existing, skipping duplicates by NPCID (first wins)
func dedupDeadNPCs(existing, incoming []domain.NPCDeathRecord) []domain.NPCDeathRecord {
	seen := make(map[string]bool, len(existing)+len(incoming))
	result := make([]domain.NPCDeathRecord, 0, len(existing)+len(incoming))

	for _, d := range existing {
		if !seen[d.NPCID] {
			seen[d.NPCID] = true
			result = append(result, d)
		}
	}
	for _, d := range incoming {
		if !seen[d.NPCID] {
			seen[d.NPCID] = true
			result = append(result, d)
		}
	}
	return result
}

// mergeQuests upserts incoming quests into existing by ID, preserving order of existing quests
// and appending truly new quests at the end
func mergeQuests(existing, incoming []domain.QuestState) []domain.QuestState {
	incomingMap := make(map[string]domain.QuestState, len(incoming))
	for _, q := range incoming {
		incomingMap[q.ID] = q
	}

	result := make([]domain.QuestState, 0, len(existing)+len(incoming))
	seen := make(map[string]bool, len(existing)+len(incoming))

	// First, process existing quests, updating in place if present in incoming
	for _, q := range existing {
		if upd, ok := incomingMap[q.ID]; ok {
			result = append(result, upd)
			seen[q.ID] = true
		} else {
			result = append(result, q)
			seen[q.ID] = true
		}
	}

	// Then append any truly new incoming quests
	for _, q := range incoming {
		if !seen[q.ID] {
			result = append(result, q)
			seen[q.ID] = true
		}
	}

	return result
}

// mergeKeyItems upserts incoming key items into existing by ID, preserving order of existing items
// and appending truly new items at the end
func mergeKeyItems(existing, incoming []domain.KeyItem) []domain.KeyItem {
	incomingMap := make(map[string]domain.KeyItem, len(incoming))
	for _, item := range incoming {
		incomingMap[item.ID] = item
	}

	result := make([]domain.KeyItem, 0, len(existing)+len(incoming))
	seen := make(map[string]bool, len(existing)+len(incoming))

	// First, process existing items, updating in place if present in incoming
	for _, item := range existing {
		if upd, ok := incomingMap[item.ID]; ok {
			result = append(result, upd)
			seen[item.ID] = true
		} else {
			result = append(result, item)
			seen[item.ID] = true
		}
	}

	// Then append any truly new incoming items
	for _, item := range incoming {
		if !seen[item.ID] {
			result = append(result, item)
			seen[item.ID] = true
		}
	}

	return result
}

// Update applies a batch update to the narrative state
func (s *NarrativeStateService) Update(ctx context.Context, campaignID string, update domain.StateUpdate) (*domain.NarrativeState, error) {
	state, err := s.stateRepo.Load(campaignID)
	if err != nil {
		// If no state exists, create an initial one
		state = &domain.NarrativeState{
			SchemaVersion:     domain.SchemaVersionV2,
			CampaignID:        campaignID,
			CurrentSession:    0,
			CurrentChapter:    "",
			CompletedChapters: []string{},
			RevealedClues:     []domain.RevealedClue{},
			ActiveQuests:      []domain.QuestState{},
			CompletedQuests:   []domain.QuestState{},
			FailedQuests:      []domain.QuestState{},
			DeadNPCs:          []domain.NPCDeathRecord{},
			KeyItems:          []domain.KeyItem{},
			LootAcquired:      []string{},
			SessionLog:        []domain.SessionRecord{},
			DMOverrides:       []domain.DMOverride{},
			PendingEffects:    []domain.DelayedEffect{},
			LastUpdated:       time.Now(),
		}
		// Save the initial state
		if saveErr := s.stateRepo.Save(campaignID, state); saveErr != nil {
			return nil, fmt.Errorf("failed to create initial narrative state: %w", saveErr)
		}
	}

	// Merge revealed clues with deduplication
	state.RevealedClues = dedupClues(state.RevealedClues, update.RevealedClues)

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

	// Merge quest state (upsert by ID, preserve order)
	if len(update.NewQuests) > 0 {
		state.ActiveQuests = mergeQuests(state.ActiveQuests, update.NewQuests)
	}

	// Merge dead NPCs with deduplication
	state.DeadNPCs = dedupDeadNPCs(state.DeadNPCs, update.DeadNPCs)

	// Merge key items state (upsert by ID, preserve order)
	if len(update.KeyItems) > 0 {
		state.KeyItems = mergeKeyItems(state.KeyItems, update.KeyItems)
	}

	// Update current location
	if update.CurrentLocation != "" {
		state.CurrentLocation = update.CurrentLocation
	}

	// Update PC statuses
	if len(update.PCStatuses) > 0 {
		state.PCStatuses = update.PCStatuses
	}

	// Update chapter tracking
	if update.CurrentChapterID != "" {
		state.CurrentChapter = update.CurrentChapterID
	}
	if len(update.CompletedChapters) > 0 {
		// Merge completed chapters (dedup)
		existingSet := make(map[string]bool)
		for _, ch := range state.CompletedChapters {
			existingSet[ch] = true
		}
		for _, ch := range update.CompletedChapters {
			if !existingSet[ch] {
				state.CompletedChapters = append(state.CompletedChapters, ch)
			}
		}
	}

	// Save session metadata to root state for easy access
	state.DMNotes = update.DMNotes
	state.LootAcquired = update.LootAcquired

	// Session log: upsert instead of append
	sessionNum := update.SessionNum
	if sessionNum < 0 {
		return nil, domain.NewValidationError("session_num", "session_num cannot be negative")
	}
	if sessionNum == 0 {
		sessionNum = state.CurrentSession + 1
	}

	// Update XP tracking (after sessionNum is defined)
	if update.XPAwarded > 0 {
		// Add to total
		state.XPTotal += update.XPAwarded
		
		// Calculate party level from XP
		state.PartyLevel = CalculatePartyLevel(state.XPTotal)
		
		// Add to ledger
		xpReason := update.XPReason
		if xpReason == "" {
			xpReason = "session"
		}
		state.XPLedger = append(state.XPLedger, domain.XPEntry{
			SessionNum: sessionNum,
			Amount:     update.XPAwarded,
			Reason:     xpReason,
			Timestamp:  time.Now(),
		})
	}

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

		// Upsert: replace existing session log entry if session_num already exists
		found := false
		for i := range state.SessionLog {
			if state.SessionLog[i].SessionNum == sessionNum {
				state.SessionLog[i] = record
				found = true
				break
			}
		}
		if !found {
			state.SessionLog = append(state.SessionLog, record)
		}
		state.CurrentSession = sessionNum
	}

	// Normalize nil arrays to empty arrays for clean JSON serialization
	if state.RevealedClues == nil {
		state.RevealedClues = []domain.RevealedClue{}
	}
	if state.ActiveQuests == nil {
		state.ActiveQuests = []domain.QuestState{}
	}
	if state.CompletedQuests == nil {
		state.CompletedQuests = []domain.QuestState{}
	}
	if state.FailedQuests == nil {
		state.FailedQuests = []domain.QuestState{}
	}
	if state.DeadNPCs == nil {
		state.DeadNPCs = []domain.NPCDeathRecord{}
	}
	if state.KeyItems == nil {
		state.KeyItems = []domain.KeyItem{}
	}
	if state.LootAcquired == nil {
		state.LootAcquired = []string{}
	}
	if state.SessionLog == nil {
		state.SessionLog = []domain.SessionRecord{}
	}
	if state.DMOverrides == nil {
		state.DMOverrides = []domain.DMOverride{}
	}
	if state.PCStatuses == nil {
		state.PCStatuses = []domain.PCStatus{}
	}
	if state.PendingEffects == nil {
		state.PendingEffects = []domain.DelayedEffect{}
	}
	if state.CompletedChapters == nil {
		state.CompletedChapters = []string{}
	}
	if state.XPLedger == nil {
		state.XPLedger = []domain.XPEntry{}
	}

	state.LastUpdated = time.Now()

	if err := s.stateRepo.Save(campaignID, state); err != nil {
		return nil, fmt.Errorf("failed to save narrative state: %w", err)
	}

	return state, nil
}

// SyncStateToCanon propagates narrative state changes to the canon document.
// It updates entity states for dead NPCs and appends timeline events for completed quests.
// Returns warnings for non-fatal issues (missing entities); errors only for repo failures.
func (s *NarrativeStateService) SyncStateToCanon(ctx context.Context, campaignID string, update domain.StateUpdate) ([]string, error) {
	if s.canonRepo == nil {
		return []string{"canon repository not available"}, nil
	}

	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	var warnings []string
	state, _ := s.stateRepo.Load(campaignID)

	// Build quest name lookup from narrative state
	questNames := make(map[string]string)
	if state != nil {
		for _, q := range state.CompletedQuests {
			questNames[q.ID] = q.Name
		}
		for _, q := range state.ActiveQuests {
			if _, ok := questNames[q.ID]; !ok {
				questNames[q.ID] = q.Name
			}
		}
	}

	// Update entity states for dead NPCs (alive -> dead only)
	for _, death := range update.DeadNPCs {
		found := false
		for i := range doc.Entities {
			if doc.Entities[i].ID == death.NPCID {
				found = true
				if doc.Entities[i].CanonState == domain.EntityStateAlive {
					doc.Entities[i].CanonState = domain.EntityStateDead
				}
				break
			}
		}
		if !found {
			warnings = append(warnings, fmt.Sprintf("entity %s not found in canon (NPC death not propagated)", death.NPCID))
		}
	}

	// Append timeline events for completed quests
	for _, questID := range update.CompletedQuests {
		questName := questID
		if name, ok := questNames[questID]; ok && name != "" {
			questName = name
		}
		event := domain.CanonTimelineEvent{
			ID:          fmt.Sprintf("evt-quest-%s-completed", questID),
			Timestamp:   time.Now().Format(time.RFC3339),
			Description: fmt.Sprintf("Quest %s completed", questName),
			Involved:    []string{questID},
			IsRevealed:  true,
		}
		doc.Timeline = append(doc.Timeline, event)
	}

	// Persist PC HP changes to character files
	if s.characterRepo != nil && len(update.PCStatuses) > 0 {
		for _, pcStatus := range update.PCStatuses {
			character, err := s.characterRepo.Read(campaignID, pcStatus.Name)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("character '%s' not found (HP update skipped)", pcStatus.Name))
				continue
			}
			
			// Update HP if changed
			if pcStatus.HPCurrent != character.HP.Current || pcStatus.HPMax != character.HP.Maximum {
				character.HP.Current = pcStatus.HPCurrent
				character.HP.Maximum = pcStatus.HPMax
				character.UpdatedAt = time.Now()
				if saveErr := s.characterRepo.Save(character); saveErr != nil {
					warnings = append(warnings, fmt.Sprintf("failed to save HP for '%s': %v", pcStatus.Name, saveErr))
				}
			}
		}
	}

	// Propagate chapter changes to canon timeline
	if update.CurrentChapterID != "" && state != nil {
		// Check if this is a chapter transition
		if state.CurrentChapter != "" && state.CurrentChapter != update.CurrentChapterID {
			event := domain.CanonTimelineEvent{
				ID:          fmt.Sprintf("evt-chapter-%s-started", update.CurrentChapterID),
				Timestamp:   time.Now().Format(time.RFC3339),
				Description: fmt.Sprintf("Chapter %s started", update.CurrentChapterID),
				Involved:    []string{update.CurrentChapterID},
				IsRevealed:  true,
			}
			doc.Timeline = append(doc.Timeline, event)
		}
	}

	// Propagate XP changes to canon timeline
	if update.XPAwarded > 0 && state != nil {
		xpReason := update.XPReason
		if xpReason == "" {
			xpReason = "session"
		}
		event := domain.CanonTimelineEvent{
			ID:          fmt.Sprintf("evt-xp-session-%d", update.SessionNum),
			Timestamp:   time.Now().Format(time.RFC3339),
			Description: fmt.Sprintf("Party awarded %d XP (%s) - Total: %d, Level: %d", update.XPAwarded, xpReason, state.XPTotal, state.PartyLevel),
			Involved:    []string{},
			IsRevealed:  true,
		}
		doc.Timeline = append(doc.Timeline, event)
	}

	// Update party state in canon
	if state != nil && doc.PartyState != nil {
		doc.PartyState.CurrentLevel = state.PartyLevel
		doc.PartyState.XPTotal = state.XPTotal
		doc.PartyState.XPLedger = state.XPLedger
	}

	doc.UpdatedAt = time.Now()
	if err := s.canonRepo.Save(campaignID, doc); err != nil {
		return nil, fmt.Errorf("failed to save canon: %w", err)
	}

	return warnings, nil
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
