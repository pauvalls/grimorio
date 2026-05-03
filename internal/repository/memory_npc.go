package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryNPCRepository implements NPCRepository in memory
type MemoryNPCRepository struct {
	mu   sync.RWMutex
	npcs map[string][]*domain.NPC
}

// NewMemoryNPCRepository creates a new in-memory NPC repository
func NewMemoryNPCRepository() *MemoryNPCRepository {
	return &MemoryNPCRepository{
		npcs: make(map[string][]*domain.NPC),
	}
}

func (r *MemoryNPCRepository) Save(npc *domain.NPC) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if npc.CampaignID == "" {
		return domain.NewValidationError("campaign_id", "campaign ID is required")
	}
	if npc.Name == "" {
		return domain.NewValidationError("name", "NPC name is required")
	}

	npc.UpdatedAt = time.Now()
	if npc.CreatedAt.IsZero() {
		npc.CreatedAt = time.Now()
	}

	npcs := r.npcs[npc.CampaignID]
	for i, n := range npcs {
		if n.Name == npc.Name {
			npcs = append(npcs[:i], npcs[i+1:]...)
			break
		}
	}

	r.npcs[npc.CampaignID] = append(npcs, npc)
	return nil
}

func (r *MemoryNPCRepository) Read(campaignID, name string) (*domain.NPC, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, npc := range r.npcs[campaignID] {
		if npc.Name == name {
			return npc, nil
		}
	}
	return nil, fmt.Errorf("NPC not found: %s", name)
}

func (r *MemoryNPCRepository) List(campaignID string) ([]domain.NPC, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.NPC
	for _, npc := range r.npcs[campaignID] {
		result = append(result, *npc)
	}
	return result, nil
}

func (r *MemoryNPCRepository) Delete(campaignID, name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	npcs := r.npcs[campaignID]
	for i, npc := range npcs {
		if npc.Name == name {
			r.npcs[campaignID] = append(npcs[:i], npcs[i+1:]...)
			return nil
		}
	}
	return nil
}

// Ensure interface is implemented
var _ NPCRepository = (*MemoryNPCRepository)(nil)
