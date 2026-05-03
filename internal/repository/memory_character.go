package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryCharacterRepository implements CharacterRepository in memory
type MemoryCharacterRepository struct {
	mu         sync.RWMutex
	characters map[string][]*domain.Character
}

// NewMemoryCharacterRepository creates a new in-memory character repository
func NewMemoryCharacterRepository() *MemoryCharacterRepository {
	return &MemoryCharacterRepository{
		characters: make(map[string][]*domain.Character),
	}
}

func (r *MemoryCharacterRepository) Save(character *domain.Character) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := character.Validate(); err != nil {
		return err
	}

	character.UpdatedAt = time.Now()
	if character.CreatedAt.IsZero() {
		character.CreatedAt = time.Now()
	}

	chars := r.characters[character.CampaignID]
	for i, c := range chars {
		if c.Name == character.Name {
			chars = append(chars[:i], chars[i+1:]...)
			break
		}
	}

	r.characters[character.CampaignID] = append(chars, character)
	return nil
}

func (r *MemoryCharacterRepository) Read(campaignID, name string) (*domain.Character, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, char := range r.characters[campaignID] {
		if char.Name == name {
			return char, nil
		}
	}
	return nil, fmt.Errorf("character not found: %s", name)
}

func (r *MemoryCharacterRepository) List(campaignID string) ([]domain.Character, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Character
	for _, char := range r.characters[campaignID] {
		result = append(result, *char)
	}
	return result, nil
}

func (r *MemoryCharacterRepository) Delete(campaignID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	chars := r.characters[campaignID]
	for i, char := range chars {
		if char.Name == name {
			r.characters[campaignID] = append(chars[:i], chars[i+1:]...)
			return nil
		}
	}
	return nil
}

// Ensure interface is implemented
var _ CharacterRepository = (*MemoryCharacterRepository)(nil)
