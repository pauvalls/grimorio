package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryCampaignRepository implements CampaignRepository in memory
type MemoryCampaignRepository struct {
	mu        sync.RWMutex
	campaigns map[string]*domain.Campaign
}

// NewMemoryCampaignRepository creates a new in-memory campaign repository
func NewMemoryCampaignRepository() *MemoryCampaignRepository {
	return &MemoryCampaignRepository{
		campaigns: make(map[string]*domain.Campaign),
	}
}

func (r *MemoryCampaignRepository) Create(campaign *domain.Campaign) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := campaign.Validate(); err != nil {
		return err
	}

	if _, exists := r.campaigns[campaign.Name]; exists {
		return domain.NewValidationError("name", "campaign already exists")
	}

	campaign.CreatedAt = time.Now()
	campaign.UpdatedAt = time.Now()
	r.campaigns[campaign.Name] = campaign
	return nil
}

func (r *MemoryCampaignRepository) Read(name string) (*domain.Campaign, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	campaign, exists := r.campaigns[name]
	if !exists {
		return nil, fmt.Errorf("campaign not found: %s", name)
	}
	return campaign, nil
}

func (r *MemoryCampaignRepository) Update(campaign *domain.Campaign) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.campaigns[campaign.Name]; !exists {
		return fmt.Errorf("campaign not found: %s", campaign.Name)
	}

	campaign.UpdatedAt = time.Now()
	r.campaigns[campaign.Name] = campaign
	return nil
}

func (r *MemoryCampaignRepository) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.campaigns, name)
	return nil
}

func (r *MemoryCampaignRepository) List() ([]domain.CampaignSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var summaries []domain.CampaignSummary
	for _, campaign := range r.campaigns {
		summaries = append(summaries, domain.CampaignSummary{
			Name:      campaign.Name,
			Title:     campaign.Title,
			Setting:   campaign.Setting,
			Status:    campaign.Status,
			UpdatedAt: campaign.UpdatedAt,
		})
	}
	return summaries, nil
}

func (r *MemoryCampaignRepository) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.campaigns[name]
	return exists
}

// Ensure interface is implemented
var _ CampaignRepository = (*MemoryCampaignRepository)(nil)
