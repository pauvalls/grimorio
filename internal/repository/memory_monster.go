package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryMonsterRepository implements MonsterRepository in memory
type MemoryMonsterRepository struct {
	mu       sync.RWMutex
	monsters map[string][]*domain.Monster
}

// NewMemoryMonsterRepository creates a new in-memory monster repository
func NewMemoryMonsterRepository() *MemoryMonsterRepository {
	return &MemoryMonsterRepository{
		monsters: make(map[string][]*domain.Monster),
	}
}

func (r *MemoryMonsterRepository) Save(monster *domain.Monster) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if monster.CampaignID == "" {
		return domain.NewValidationError("campaign_id", "campaign ID is required")
	}
	if monster.Name == "" {
		return domain.NewValidationError("name", "monster name is required")
	}

	monster.UpdatedAt = time.Now()
	if monster.CreatedAt.IsZero() {
		monster.CreatedAt = time.Now()
	}

	monsters := r.monsters[monster.CampaignID]
	for i, m := range monsters {
		if m.Name == monster.Name {
			monsters = append(monsters[:i], monsters[i+1:]...)
			break
		}
	}

	r.monsters[monster.CampaignID] = append(monsters, monster)
	return nil
}

func (r *MemoryMonsterRepository) Read(campaignID, name string) (*domain.Monster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, monster := range r.monsters[campaignID] {
		if monster.Name == name {
			return monster, nil
		}
	}

	return nil, fmt.Errorf("monster not found: %s", name)
}

func (r *MemoryMonsterRepository) List(campaignID string) ([]domain.Monster, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	monsters := r.monsters[campaignID]
	result := make([]domain.Monster, len(monsters))
	for i, m := range monsters {
		result[i] = *m
	}
	return result, nil
}

func (r *MemoryMonsterRepository) Delete(ctx context.Context, campaignID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	monsters := r.monsters[campaignID]
	for i, m := range monsters {
		if m.Name == name {
			r.monsters[campaignID] = append(monsters[:i], monsters[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("monster not found: %s", name)
}
