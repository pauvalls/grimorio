package services

import (
	"fmt"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// CharacterService handles character business logic
type CharacterService struct {
	repo repository.CharacterRepository
}

// NewCharacterService creates a new character service
func NewCharacterService(repo repository.CharacterRepository) *CharacterService {
	return &CharacterService{repo: repo}
}

// CreateCharacter creates a new character
func (s *CharacterService) CreateCharacter(campaignID, name, race, class string, level int, background, alignment string) (*domain.Character, error) {
	if level <= 0 {
		level = 1
	}

	character := &domain.Character{
		CampaignID: campaignID,
		Name:       name,
		Race:       race,
		Class:      class,
		Level:      level,
		Background: background,
		Alignment:  alignment,
		Status:     "alive",
		Stats: domain.Stats{
			STR: 10, DEX: 10, CON: 10, INT: 10, WIS: 10, CHA: 10,
		},
		HP: domain.HP{
			Current: 10,
			Maximum: 10,
		},
		AC:            10,
		Proficiency:   2,
		Skills:        make(map[string]bool),
		Inventory:     []domain.Item{},
		Features:      []domain.Feature{},
		Relationships: []domain.Relationship{},
	}

	if err := s.repo.Save(character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	return character, nil
}

// GetCharacter retrieves a character by name
func (s *CharacterService) GetCharacter(campaignID, name string) (*domain.Character, error) {
	return s.repo.Read(campaignID, name)
}

// ListCharacters returns all characters in a campaign
func (s *CharacterService) ListCharacters(campaignID string) ([]domain.Character, error) {
	return s.repo.List(campaignID)
}

// UpdateCharacter updates a character
func (s *CharacterService) UpdateCharacter(character *domain.Character) error {
	return s.repo.Save(character)
}

// DeleteCharacter deletes a character
func (s *CharacterService) DeleteCharacter(campaignID, name string) error {
	return s.repo.Delete(campaignID, name)
}

// AddRelationship adds a relationship to a character
func (s *CharacterService) AddRelationship(campaignID, characterName string, rel domain.Relationship) error {
	character, err := s.repo.Read(campaignID, characterName)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Update existing or add new
	found := false
	for i, existing := range character.Relationships {
		if existing.EntityID == rel.EntityID {
			character.Relationships[i] = rel
			found = true
			break
		}
	}
	if !found {
		character.Relationships = append(character.Relationships, rel)
	}

	return s.repo.Save(character)
}
