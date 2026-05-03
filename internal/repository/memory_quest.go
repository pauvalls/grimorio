package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryQuestRepository implements QuestRepository in memory
type MemoryQuestRepository struct {
	mu     sync.RWMutex
	quests map[string][]*domain.Quest
}

// NewMemoryQuestRepository creates a new in-memory quest repository
func NewMemoryQuestRepository() *MemoryQuestRepository {
	return &MemoryQuestRepository{
		quests: make(map[string][]*domain.Quest),
	}
}

func (r *MemoryQuestRepository) Save(quest *domain.Quest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := quest.Validate(); err != nil {
		return err
	}

	quest.UpdatedAt = time.Now()
	if quest.CreatedAt.IsZero() {
		quest.CreatedAt = time.Now()
	}
	if quest.ID == "" {
		quest.ID = fmt.Sprintf("quest_%d", time.Now().UnixNano())
	}

	quests := r.quests[quest.CampaignID]
	for i, q := range quests {
		if q.ID == quest.ID {
			quests = append(quests[:i], quests[i+1:]...)
			break
		}
	}

	r.quests[quest.CampaignID] = append(quests, quest)
	return nil
}

func (r *MemoryQuestRepository) Read(campaignID, id string) (*domain.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, quest := range r.quests[campaignID] {
		if quest.ID == id {
			return quest, nil
		}
	}
	return nil, fmt.Errorf("quest not found: %s", id)
}

func (r *MemoryQuestRepository) List(campaignID string) ([]domain.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Quest
	for _, quest := range r.quests[campaignID] {
		result = append(result, *quest)
	}
	return result, nil
}

func (r *MemoryQuestRepository) ListByCharacter(campaignID, characterID string) ([]domain.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Quest
	for _, quest := range r.quests[campaignID] {
		if quest.CharacterID != nil && *quest.CharacterID == characterID {
			result = append(result, *quest)
		}
	}
	return result, nil
}

func (r *MemoryQuestRepository) ListByStatus(campaignID string, status domain.QuestStatus) ([]domain.Quest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Quest
	for _, quest := range r.quests[campaignID] {
		if quest.Status == status {
			result = append(result, *quest)
		}
	}
	return result, nil
}

func (r *MemoryQuestRepository) Delete(campaignID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	quests := r.quests[campaignID]
	for i, quest := range quests {
		if quest.ID == id {
			r.quests[campaignID] = append(quests[:i], quests[i+1:]...)
			return nil
		}
	}
	return nil
}

// Ensure interface is implemented
var _ QuestRepository = (*MemoryQuestRepository)(nil)
