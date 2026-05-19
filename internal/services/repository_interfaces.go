package services

import (
	"context"
	"github.com/pauvalls/grimorio/internal/domain"
)

// This file contains repository interfaces for V3 services.
// In-memory implementations are in inmemory/ package.

// MilestoneXPRepository is already defined in milestone_service.go

// MagicItemRepository is already defined in item_service.go

// Tactics repositories are already defined in tactics_service.go

// PlayerMapRepository is already defined in player_map_service.go

// SessionZeroRepository is already defined in session_zero_service.go

// ConsequenceTableRepository is already defined in consequence_service.go

// AreaRepositoryV3 is already defined in area_service.go

// QuestRepositoryV3 defines repository interface for V3 quests.
type QuestRepositoryV3 interface {
	Create(ctx context.Context, campaignID string, quest *domain.Quest) error
	Read(ctx context.Context, campaignID string, questID string) (*domain.Quest, error)
	Update(ctx context.Context, campaignID string, quest *domain.Quest) error
	Delete(ctx context.Context, campaignID string, questID string) error
	GetByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.Quest, error)
	GetByNPC(ctx context.Context, campaignID string, npcID string) ([]*domain.Quest, error)
	UpdateStatus(ctx context.Context, campaignID string, questID string, status domain.QuestStatus, notes string) error
}

// CharacterRepositoryV3 defines repository interface for V3 characters.
type CharacterRepositoryV3 interface {
	Create(ctx context.Context, campaignID string, character *domain.PregenCharacter) error
	Read(ctx context.Context, campaignID string, characterID string) (*domain.PregenCharacter, error)
	Update(ctx context.Context, campaignID string, character *domain.PregenCharacter) error
	Delete(ctx context.Context, campaignID string, characterID string) error
	GetByCampaign(ctx context.Context, campaignID string) ([]*domain.PregenCharacter, error)
}

// TacticsRepository defines repository interface for tactics.
type TacticsRepository interface {
	Create(ctx context.Context, campaignID string, tactics *domain.Tactics) error
	Read(ctx context.Context, campaignID string, tacticsID string) (*domain.Tactics, error)
	ListByEncounter(ctx context.Context, campaignID string, encounterID string) ([]*domain.Tactics, error)
}

// HandoutRepositoryV3 defines repository interface for V3 handouts.
type HandoutRepositoryV3 interface {
	Create(ctx context.Context, campaignID string, handout *domain.Handout) error
	Read(ctx context.Context, campaignID string, handoutID string) (*domain.Handout, error)
	Update(ctx context.Context, campaignID string, handout *domain.Handout) error
	Delete(ctx context.Context, campaignID string, handoutID string) error
	GetByQuest(ctx context.Context, campaignID string, questID string) ([]*domain.Handout, error)
	GetByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.Handout, error)
}
