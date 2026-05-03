package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryActRepository implements ActRepository in memory
type MemoryActRepository struct {
	mu   sync.RWMutex
	acts map[string][]*domain.Act
}

// NewMemoryActRepository creates a new in-memory act repository
func NewMemoryActRepository() *MemoryActRepository {
	return &MemoryActRepository{
		acts: make(map[string][]*domain.Act),
	}
}

func (r *MemoryActRepository) Save(act *domain.Act) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := act.Validate(); err != nil {
		return err
	}

	act.UpdatedAt = time.Now()
	if act.CreatedAt.IsZero() {
		act.CreatedAt = time.Now()
	}

	acts := r.acts[act.CampaignID]
	for i, a := range acts {
		if a.Number == act.Number {
			acts = append(acts[:i], acts[i+1:]...)
			break
		}
	}

	r.acts[act.CampaignID] = append(acts, act)
	return nil
}

func (r *MemoryActRepository) Read(campaignID string, number int) (*domain.Act, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, act := range r.acts[campaignID] {
		if act.Number == number {
			return act, nil
		}
	}
	return nil, fmt.Errorf("act %d not found in campaign %s", number, campaignID)
}

func (r *MemoryActRepository) List(campaignID string) ([]domain.Act, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Act
	for _, act := range r.acts[campaignID] {
		result = append(result, *act)
	}
	return result, nil
}

func (r *MemoryActRepository) Delete(campaignID string, number int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	acts := r.acts[campaignID]
	for i, act := range acts {
		if act.Number == number {
			r.acts[campaignID] = append(acts[:i], acts[i+1:]...)
			return nil
		}
	}
	return nil
}

// Ensure interface is implemented
var _ ActRepository = (*MemoryActRepository)(nil)
