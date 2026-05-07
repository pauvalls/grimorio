package repository

import (
	"fmt"
	"sync"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryCanonRepository implements CanonRepository in memory
type MemoryCanonRepository struct {
	mu    sync.RWMutex
	canon map[string]*domain.CanonDocument
}

// NewMemoryCanonRepository creates a new in-memory canon repository
func NewMemoryCanonRepository() *MemoryCanonRepository {
	return &MemoryCanonRepository{
		canon: make(map[string]*domain.CanonDocument),
	}
}

func (r *MemoryCanonRepository) Save(campaignID string, doc *domain.CanonDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := doc.Validate(); err != nil {
		return err
	}

	r.canon[campaignID] = doc
	return nil
}

func (r *MemoryCanonRepository) Load(campaignID string) (*domain.CanonDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, exists := r.canon[campaignID]
	if !exists {
		return nil, fmt.Errorf("canon not found for campaign: %s", campaignID)
	}
	return doc, nil
}

func (r *MemoryCanonRepository) Exists(campaignID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.canon[campaignID]
	return exists
}

func (r *MemoryCanonRepository) Delete(campaignID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.canon, campaignID)
}

// Ensure interface is implemented
var _ CanonRepository = (*MemoryCanonRepository)(nil)

// MemoryNarrativeStateRepository implements NarrativeStateRepository in memory
type MemoryNarrativeStateRepository struct {
	mu     sync.RWMutex
	states map[string]*domain.NarrativeState
}

// NewMemoryNarrativeStateRepository creates a new in-memory narrative state repository
func NewMemoryNarrativeStateRepository() *MemoryNarrativeStateRepository {
	return &MemoryNarrativeStateRepository{
		states: make(map[string]*domain.NarrativeState),
	}
}

func (r *MemoryNarrativeStateRepository) Save(campaignID string, state *domain.NarrativeState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := state.Validate(); err != nil {
		return err
	}

	r.states[campaignID] = state
	return nil
}

func (r *MemoryNarrativeStateRepository) Load(campaignID string) (*domain.NarrativeState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, exists := r.states[campaignID]
	if !exists {
		return nil, fmt.Errorf("narrative state not found for campaign: %s", campaignID)
	}
	return state, nil
}

func (r *MemoryNarrativeStateRepository) Exists(campaignID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.states[campaignID]
	return exists
}

// Ensure interface is implemented
var _ NarrativeStateRepository = (*MemoryNarrativeStateRepository)(nil)
