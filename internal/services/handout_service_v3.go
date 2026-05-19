package services

import (
	"context"
	"fmt"
	"github.com/pauvalls/grimorio/internal/domain"
)

// HandoutServiceV3 generates enhanced handouts for V3.
type HandoutServiceV3 struct {
	handoutRepo HandoutRepositoryV3
}

// NewHandoutServiceV3 creates a new HandoutServiceV3.
func NewHandoutServiceV3(handoutRepo HandoutRepositoryV3) *HandoutServiceV3 {
	return &HandoutServiceV3{handoutRepo: handoutRepo}
}

// GenerateHandout generates a handout by type.
func (s *HandoutServiceV3) GenerateHandout(ctx context.Context, campaignID string, handoutType domain.HandoutType, context map[string]interface{}) (*domain.Handout, error) {
	handout := &domain.Handout{
		ID:         fmt.Sprintf("handout_%s_%d", campaignID, len(context)),
		CampaignID: campaignID,
		Type:       handoutType,
		Title:      generateHandoutTitle(handoutType),
		Content:    generateHandoutContent(handoutType, context),
		DMNotes:    generateDMNotes(handoutType, context),
		Format:     domain.FormatText,
		Style:      domain.StyleFormal,
	}

	if err := handout.Validate(); err != nil {
		return nil, fmt.Errorf("generated handout validation failed: %w", err)
	}

	return handout, nil
}

// GenerateLetter generates a letter handout.
func (s *HandoutServiceV3) GenerateLetter(ctx context.Context, campaignID, sender, recipient, purpose string) (*domain.Handout, error) {
	context := map[string]interface{}{
		"sender":    sender,
		"recipient": recipient,
		"purpose":   purpose,
	}
	return s.GenerateHandout(ctx, campaignID, domain.HandoutTypeLetter, context)
}

// GenerateClue generates a clue handout.
func (s *HandoutServiceV3) GenerateClue(ctx context.Context, campaignID, questID, clueType string) (*domain.Handout, error) {
	context := map[string]interface{}{
		"quest_id":  questID,
		"clue_type": clueType,
	}
	handout, err := s.GenerateHandout(ctx, campaignID, domain.HandoutTypeClue, context)
	if err != nil {
		return nil, err
	}
	handout.RevealConditions = []string{"Found during investigation", "DC 15 Perception check"}
	handout.QuestRefs = []string{questID}
	return handout, nil
}

// GetHandoutsByQuest retrieves handouts for a quest.
func (s *HandoutServiceV3) GetHandoutsByQuest(ctx context.Context, campaignID, questID string) ([]*domain.Handout, error) {
	return s.handoutRepo.GetByQuest(ctx, campaignID, questID)
}

// ExportHandout exports a handout in the specified format.
func (s *HandoutServiceV3) ExportHandout(ctx context.Context, campaignID, handoutID, format string) ([]byte, error) {
	if format != "text" {
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	handout, err := s.handoutRepo.Read(ctx, campaignID, handoutID)
	if err != nil {
		return nil, fmt.Errorf("failed to load handout: %w", err)
	}

	return []byte(handout.Content), nil
}

func generateHandoutTitle(t domain.HandoutType) string {
	switch t {
	case domain.HandoutTypeLetter:
		return "Mysterious Letter"
	case domain.HandoutTypeClue:
		return "Cryptic Clue"
	case domain.HandoutTypeDocument:
		return "Ancient Document"
	case domain.HandoutTypeJournal:
		return "Traveler's Journal"
	default:
		return "Handout"
	}
}

func generateHandoutContent(t domain.HandoutType, context map[string]interface{}) string {
	switch t {
	case domain.HandoutTypeLetter:
		sender, _ := context["sender"].(string)
		recipient, _ := context["recipient"].(string)
		return fmt.Sprintf("To %s,\n\nFrom %s.\n\nThe time has come...", recipient, sender)
	case domain.HandoutTypeClue:
		return "A scrap of paper with strange markings..."
	default:
		return "Content goes here..."
	}
}

func generateDMNotes(t domain.HandoutType, context map[string]interface{}) string {
	switch t {
	case domain.HandoutTypeLetter:
		return "Give this to players when they meet the quest giver"
	case domain.HandoutTypeClue:
		return "Hide this in area #3, behind a loose stone"
	default:
		return "DM context and usage notes"
	}
}
