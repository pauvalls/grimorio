package services

import (
	"fmt"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// QuestService handles quest business logic
type QuestService struct {
	repo repository.QuestRepository
}

// NewQuestService creates a new quest service
func NewQuestService(repo repository.QuestRepository) *QuestService {
	return &QuestService{repo: repo}
}

// CreateQuest creates a new quest
func (s *QuestService) CreateQuest(campaignID, title string, questType domain.QuestType, hook, description, stakes string, characterID *string) (*domain.Quest, error) {
	quest := &domain.Quest{
		CampaignID:  campaignID,
		Title:       title,
		Type:        questType,
		Status:      domain.QuestStatusActive,
		Hook:        hook,
		Description: description,
		Stakes:      stakes,
		CharacterID: characterID,
		Objectives:  []domain.Objective{},
		ProgressNotes: []domain.ProgressNote{},
	}

	if err := s.repo.Save(quest); err != nil {
		return nil, fmt.Errorf("failed to create quest: %w", err)
	}

	return quest, nil
}

// GetQuest retrieves a quest by ID
func (s *QuestService) GetQuest(campaignID, id string) (*domain.Quest, error) {
	return s.repo.Read(campaignID, id)
}

// ListQuests returns all quests in a campaign
func (s *QuestService) ListQuests(campaignID string) ([]domain.Quest, error) {
	return s.repo.List(campaignID)
}

// ListQuestsByCharacter returns quests for a specific character
func (s *QuestService) ListQuestsByCharacter(campaignID, characterID string) ([]domain.Quest, error) {
	return s.repo.ListByCharacter(campaignID, characterID)
}

// ListActiveQuests returns active quests
func (s *QuestService) ListActiveQuests(campaignID string) ([]domain.Quest, error) {
	return s.repo.ListByStatus(campaignID, domain.QuestStatusActive)
}

// UpdateQuestStatus updates the status of a quest
func (s *QuestService) UpdateQuestStatus(campaignID, questID string, status domain.QuestStatus, notes string) error {
	quest, err := s.repo.Read(campaignID, questID)
	if err != nil {
		return fmt.Errorf("quest not found: %w", err)
	}

	quest.Status = status
	if notes != "" {
		quest.ProgressNotes = append(quest.ProgressNotes, domain.ProgressNote{
			Date: time.Now(),
			Note: notes,
		})
	}

	return s.repo.Save(quest)
}

// AddObjective adds an objective to a quest
func (s *QuestService) AddObjective(campaignID, questID string, description string) error {
	quest, err := s.repo.Read(campaignID, questID)
	if err != nil {
		return fmt.Errorf("quest not found: %w", err)
	}

	objective := domain.Objective{
		ID:          fmt.Sprintf("obj_%d", len(quest.Objectives)),
		Description: description,
		Completed:   false,
		Order:       len(quest.Objectives),
	}
	quest.Objectives = append(quest.Objectives, objective)

	return s.repo.Save(quest)
}

// CompleteObjective marks an objective as completed
func (s *QuestService) CompleteObjective(campaignID, questID, objectiveID string) error {
	quest, err := s.repo.Read(campaignID, questID)
	if err != nil {
		return fmt.Errorf("quest not found: %w", err)
	}

	for i, obj := range quest.Objectives {
		if obj.ID == objectiveID {
			quest.Objectives[i].Completed = true
			break
		}
	}

	return s.repo.Save(quest)
}

// DeleteQuest deletes a quest
func (s *QuestService) DeleteQuest(campaignID, questID string) error {
	return s.repo.Delete(campaignID, questID)
}
