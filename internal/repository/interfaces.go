package repository

import (
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// CampaignRepository defines operations for campaign persistence
type CampaignRepository interface {
	Create(campaign *domain.Campaign) error
	Read(name string) (*domain.Campaign, error)
	Update(campaign *domain.Campaign) error
	Delete(name string) error
	List() ([]domain.CampaignSummary, error)
	Exists(name string) bool
}

// ActRepository defines operations for act persistence
type ActRepository interface {
	Save(act *domain.Act) error
	Read(campaignID string, number int) (*domain.Act, error)
	List(campaignID string) ([]domain.Act, error)
	Delete(campaignID string, number int) error
}

// CharacterRepository defines operations for character persistence
type CharacterRepository interface {
	Save(character *domain.Character) error
	Read(campaignID, name string) (*domain.Character, error)
	List(campaignID string) ([]domain.Character, error)
	Delete(campaignID, name string) error
}

// NPCRepository defines operations for NPC persistence
type NPCRepository interface {
	Save(npc *domain.NPC) error
	Read(campaignID, name string) (*domain.NPC, error)
	List(campaignID string) ([]domain.NPC, error)
	Delete(campaignID, name string) error
}

// QuestRepository defines operations for quest persistence
type QuestRepository interface {
	Save(quest *domain.Quest) error
	Read(campaignID, id string) (*domain.Quest, error)
	List(campaignID string) ([]domain.Quest, error)
	ListByCharacter(campaignID, characterID string) ([]domain.Quest, error)
	ListByStatus(campaignID string, status domain.QuestStatus) ([]domain.Quest, error)
	Delete(campaignID, id string) error
}

// MonsterRepository defines operations for monster persistence
type MonsterRepository interface {
	Save(monster *domain.Monster) error
	Read(campaignID, name string) (*domain.Monster, error)
	List(campaignID string) ([]domain.Monster, error)
	Delete(campaignID, name string) error
}

// EncounterRepository defines operations for encounter persistence
type EncounterRepository interface {
	Save(encounter *domain.Encounter) error
	Read(campaignID, name string) (*domain.Encounter, error)
	List(campaignID string) ([]domain.Encounter, error)
	Delete(campaignID, name string) error
}

// MapRepository defines operations for map persistence
type MapRepository interface {
	Save(m *domain.Map) error
	Read(campaignID, name string) (*domain.Map, error)
	List(campaignID string) ([]domain.Map, error)
	Delete(campaignID, name string) error
}

// SessionRepository defines operations for game session persistence
type SessionRepository interface {
	Create(session *domain.Session) error
	Read(id string) (*domain.Session, error)
	Update(session *domain.Session) error
	End(id string) error
	ListByCampaign(campaignID string) ([]*domain.Session, error)
	
	// Player state
	SavePlayerState(sessionID string, player *domain.PlayerState) error
	GetPlayerState(sessionID, characterID string) (*domain.PlayerState, error)
	ListPlayerStates(sessionID string) ([]domain.PlayerState, error)
	
	// Events
	AppendEvent(event *domain.SessionEvent) error
	GetEvents(sessionID string, limit int) ([]*domain.SessionEvent, error)
	GetEventsSince(sessionID string, since time.Time) ([]*domain.SessionEvent, error)
	
	// Combat
	SaveCombatState(sessionID string, combat *domain.CombatState) error
	GetCombatState(sessionID string) (*domain.CombatState, error)
	ClearCombatState(sessionID string) error
}
