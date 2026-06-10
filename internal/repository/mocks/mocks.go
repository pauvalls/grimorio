package mocks

import (
	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// CanonRepositoryMock is a mock implementation of CanonRepository
type CanonRepositoryMock struct {
	SaveFunc func(campaignID string, doc *domain.CanonDocument) error
	LoadFunc func(campaignID string) (*domain.CanonDocument, error)
	ExistsFunc func(campaignID string) bool
}

// Save implements CanonRepository
func (m *CanonRepositoryMock) Save(campaignID string, doc *domain.CanonDocument) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(campaignID, doc)
	}
	return nil
}

// Load implements CanonRepository
func (m *CanonRepositoryMock) Load(campaignID string) (*domain.CanonDocument, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(campaignID)
	}
	return nil, nil
}

// Exists implements CanonRepository
func (m *CanonRepositoryMock) Exists(campaignID string) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(campaignID)
	}
	return false
}

// CampaignRepositoryMock is a mock implementation of CampaignRepository
type CampaignRepositoryMock struct {
	CreateFunc func(campaign *domain.Campaign) error
	ReadFunc   func(name string) (*domain.Campaign, error)
	UpdateFunc func(campaign *domain.Campaign) error
	DeleteFunc func(name string) error
	ListFunc   func() ([]domain.CampaignSummary, error)
	ExistsFunc func(name string) bool
}

// Create implements CampaignRepository
func (m *CampaignRepositoryMock) Create(campaign *domain.Campaign) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(campaign)
	}
	return nil
}

// Read implements CampaignRepository
func (m *CampaignRepositoryMock) Read(name string) (*domain.Campaign, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(name)
	}
	return nil, nil
}

// Update implements CampaignRepository
func (m *CampaignRepositoryMock) Update(campaign *domain.Campaign) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(campaign)
	}
	return nil
}

// Delete implements CampaignRepository
func (m *CampaignRepositoryMock) Delete(name string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(name)
	}
	return nil
}

// List implements CampaignRepository
func (m *CampaignRepositoryMock) List() ([]domain.CampaignSummary, error) {
	if m.ListFunc != nil {
		return m.ListFunc()
	}
	return nil, nil
}

// Exists implements CampaignRepository
func (m *CampaignRepositoryMock) Exists(name string) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(name)
	}
	return false
}

// NPCRepositoryMock is a mock implementation of NPCRepository
type NPCRepositoryMock struct {
	SaveFunc   func(npc *domain.NPC) error
	ReadFunc   func(campaignID, name string) (*domain.NPC, error)
	ListFunc   func(campaignID string) ([]domain.NPC, error)
	DeleteFunc func(campaignID, name string) error
}

// Save implements NPCRepository
func (m *NPCRepositoryMock) Save(npc *domain.NPC) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(npc)
	}
	return nil
}

// Read implements NPCRepository
func (m *NPCRepositoryMock) Read(campaignID, name string) (*domain.NPC, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(campaignID, name)
	}
	return nil, nil
}

// List implements NPCRepository
func (m *NPCRepositoryMock) List(campaignID string) ([]domain.NPC, error) {
	if m.ListFunc != nil {
		return m.ListFunc(campaignID)
	}
	return nil, nil
}

// Delete implements NPCRepository
func (m *NPCRepositoryMock) Delete(campaignID, name string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(campaignID, name)
	}
	return nil
}

// MonsterRepositoryMock is a mock implementation of MonsterRepository
type MonsterRepositoryMock struct {
	SaveFunc   func(monster *domain.Monster) error
	ReadFunc   func(campaignID, name string) (*domain.Monster, error)
	ListFunc   func(campaignID string) ([]domain.Monster, error)
	DeleteFunc func(campaignID, name string) error
}

// Save implements MonsterRepository
func (m *MonsterRepositoryMock) Save(monster *domain.Monster) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(monster)
	}
	return nil
}

// Read implements MonsterRepository
func (m *MonsterRepositoryMock) Read(campaignID, name string) (*domain.Monster, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(campaignID, name)
	}
	return nil, nil
}

// List implements MonsterRepository
func (m *MonsterRepositoryMock) List(campaignID string) ([]domain.Monster, error) {
	if m.ListFunc != nil {
		return m.ListFunc(campaignID)
	}
	return nil, nil
}

// Delete implements MonsterRepository
func (m *MonsterRepositoryMock) Delete(campaignID, name string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(campaignID, name)
	}
	return nil
}

// NarrativeStateRepositoryMock is a mock implementation of NarrativeStateRepository
type NarrativeStateRepositoryMock struct {
	SaveFunc   func(campaignID string, state *domain.NarrativeState) error
	LoadFunc   func(campaignID string) (*domain.NarrativeState, error)
	ExistsFunc func(campaignID string) bool
}

// Save implements NarrativeStateRepository
func (m *NarrativeStateRepositoryMock) Save(campaignID string, state *domain.NarrativeState) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(campaignID, state)
	}
	return nil
}

// Load implements NarrativeStateRepository
func (m *NarrativeStateRepositoryMock) Load(campaignID string) (*domain.NarrativeState, error) {
	if m.LoadFunc != nil {
		return m.LoadFunc(campaignID)
	}
	return nil, nil
}

// Exists implements NarrativeStateRepository
func (m *NarrativeStateRepositoryMock) Exists(campaignID string) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(campaignID)
	}
	return false
}

// Ensure interfaces are implemented
var (
	_ repository.CanonRepository          = (*CanonRepositoryMock)(nil)
	_ repository.CampaignRepository       = (*CampaignRepositoryMock)(nil)
	_ repository.NPCRepository            = (*NPCRepositoryMock)(nil)
	_ repository.MonsterRepository        = (*MonsterRepositoryMock)(nil)
	_ repository.NarrativeStateRepository = (*NarrativeStateRepositoryMock)(nil)
)
