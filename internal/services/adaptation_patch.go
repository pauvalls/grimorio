package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// AdaptationPatchService generates markdown patches from world events
type AdaptationPatchService struct {
	actRepo   repository.ActRepository
	canonRepo repository.CanonRepository
}

// NewAdaptationPatchService creates a new adaptation patch service
func NewAdaptationPatchService(actRepo repository.ActRepository, canonRepo repository.CanonRepository) *AdaptationPatchService {
	return &AdaptationPatchService{
		actRepo:   actRepo,
		canonRepo: canonRepo,
	}
}

// GeneratePatch creates a markdown diff patch for a world event
func (s *AdaptationPatchService) GeneratePatch(ctx context.Context, campaignID string, event domain.WorldEvent) (*domain.AdaptationPatch, error) {
	doc, err := s.canonRepo.Load(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to load canon: %w", err)
	}

	acts, err := s.actRepo.List(campaignID)
	if err != nil {
		return nil, fmt.Errorf("failed to list acts: %w", err)
	}

	patch := &domain.AdaptationPatch{
		CampaignID:     campaignID,
		WorldEvent:     event,
		AffectedActs:   []string{},
		AffectedNPCs:   []string{},
		AffectedQuests: []string{},
		Instructions:   []domain.PatchInstruction{},
		MarkdownDiff:   "",
		IsEmpty:        true,
	}

	// Search for references to the event entity in canon entities and acts
	entityID := event.ID
	triggerType := event.TriggerType

	// Find matching acts
	var diffParts []string
	for _, act := range acts {
		if strings.Contains(strings.ToLower(act.Content), strings.ToLower(entityID)) {
			patch.AffectedActs = append(patch.AffectedActs, fmt.Sprintf("act-%d", act.Number))
			patch.IsEmpty = false

			instruction := domain.PatchInstruction{
				Target:      "act",
				TargetID:    fmt.Sprintf("act-%d", act.Number),
				Action:      "apply",
				Description: fmt.Sprintf("Update act %d to reflect %s: %s", act.Number, triggerType, event.Description),
				OldValue:    act.Content,
				NewValue:    s.suggestUpdate(act.Content, entityID, event),
			}
			patch.Instructions = append(patch.Instructions, instruction)

			diffParts = append(diffParts, fmt.Sprintf("## Act %d: %s\n- %s: %s\n+ %s: %s\n",
				act.Number, act.Title,
				triggerType, act.Content,
				triggerType, instruction.NewValue))
		}
	}

	// Find matching entities
	for _, entity := range doc.Entities {
		if strings.Contains(strings.ToLower(entity.ID), strings.ToLower(entityID)) ||
			strings.Contains(strings.ToLower(entity.Name), strings.ToLower(entityID)) {
			if entity.Type == domain.EntityTypeNPC {
				patch.AffectedNPCs = append(patch.AffectedNPCs, entity.ID)
			}
			patch.IsEmpty = false
		}
	}

	if patch.IsEmpty {
		patch.MarkdownDiff = fmt.Sprintf("# Patch: %s\n\nNo affected content found for event: %s", triggerType, event.Description)
	} else {
		patch.MarkdownDiff = fmt.Sprintf("# Patch: %s\n\n%s", triggerType, strings.Join(diffParts, "\n"))
	}

	return patch, nil
}

func (s *AdaptationPatchService) suggestUpdate(content, entityID string, event domain.WorldEvent) string {
	// Simple suggestion: append a world-change note
	return content + fmt.Sprintf("\n\n**[World Change — Session %d]** %s: %s", event.SessionNum, event.TriggerType, event.Description)
}
