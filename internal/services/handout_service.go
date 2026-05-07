package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// HandoutService generates dual-version handouts for players and DMs
type HandoutService struct {
	questRepo repository.QuestRepository
	canonRepo repository.CanonRepository
}

// NewHandoutService creates a new handout service
func NewHandoutService(questRepo repository.QuestRepository, canonRepo repository.CanonRepository) *HandoutService {
	return &HandoutService{
		questRepo: questRepo,
		canonRepo: canonRepo,
	}
}

// GenerateHandout creates a handout with player and/or DM versions
func (s *HandoutService) GenerateHandout(ctx context.Context, campaignID string, handoutType domain.HandoutType, contentRefs []string, version domain.HandoutVersion) (*domain.Handout, error) {
	if len(contentRefs) == 0 {
		return nil, fmt.Errorf("empty_content_refs")
	}
	if !domain.IsValidHandoutType(string(handoutType)) {
		return nil, fmt.Errorf("invalid handout type: %s", handoutType)
	}

	handout := &domain.Handout{
		CampaignID:  campaignID,
		HandoutType: handoutType,
		ContentRefs: contentRefs,
	}

	var parts []string
	for _, ref := range contentRefs {
		content, err := s.resolveContent(ctx, campaignID, handoutType, ref)
		if err != nil {
			return nil, fmt.Errorf("content_ref_not_found: %s", ref)
		}
		parts = append(parts, content)
	}

	fullContent := strings.Join(parts, "\n\n---\n\n")

	switch version {
	case domain.HandoutVersionPlayer:
		handout.PlayerVersion = filterPlayerVersion(fullContent)
	case domain.HandoutVersionDM:
		handout.DMVersion = filterDMVersion(fullContent)
	case domain.HandoutVersionBoth:
		handout.PlayerVersion = filterPlayerVersion(fullContent)
		handout.DMVersion = filterDMVersion(fullContent)
	}

	return handout, nil
}

func (s *HandoutService) resolveContent(ctx context.Context, campaignID string, handoutType domain.HandoutType, ref string) (string, error) {
	switch handoutType {
	case domain.HandoutTypeQuest:
		quest, err := s.questRepo.Read(campaignID, ref)
		if err != nil {
			return "", err
		}
		return quest.Description, nil
	case domain.HandoutTypeFaction:
		doc, err := s.canonRepo.Load(campaignID)
		if err != nil {
			return "", err
		}
		for _, entity := range doc.Entities {
			if entity.ID == ref && entity.Type == domain.EntityTypeFaction {
				return fmt.Sprintf("# %s\n\n%s", entity.Name, entity.Motivation), nil
			}
		}
		return "", fmt.Errorf("faction not found: %s", ref)
	case domain.HandoutTypeLore:
		// For lore, treat ref as a fact ID
		doc, err := s.canonRepo.Load(campaignID)
		if err != nil {
			return "", err
		}
		for _, fact := range doc.Facts {
			if fact.ID == ref {
				return fact.Statement, nil
			}
		}
		return "", fmt.Errorf("fact not found: %s", ref)
	default:
		return ref, nil // fallback for summary/encounter: use ref as raw content
	}
}

func filterPlayerVersion(content string) string {
	lines := strings.Split(content, "\n")
	var out []string
	for _, line := range lines {
		if idx := strings.Index(line, "[DM]"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
			if line != "" {
				out = append(out, line)
			}
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func filterDMVersion(content string) string {
	return content
}
