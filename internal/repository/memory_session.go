package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemorySessionRepository is an in-memory implementation of SessionRepository
type MemorySessionRepository struct {
	mu        sync.RWMutex
	sessions  map[string]*domain.Session
	events    map[string][]*domain.SessionEvent
	combat    map[string]*domain.CombatState
	players   map[string]map[string]*domain.PlayerState // sessionID -> characterID -> PlayerState
}

// NewMemorySessionRepository creates a new memory session repository
func NewMemorySessionRepository() *MemorySessionRepository {
	return &MemorySessionRepository{
		sessions: make(map[string]*domain.Session),
		events:   make(map[string][]*domain.SessionEvent),
		combat:   make(map[string]*domain.CombatState),
		players:  make(map[string]map[string]*domain.PlayerState),
	}
}

// Create stores a new session
func (r *MemorySessionRepository) Create(session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if session.ID == "" {
		session.ID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}
	
	r.sessions[session.ID] = session
	r.players[session.ID] = make(map[string]*domain.PlayerState)
	
	for i := range session.Players {
		player := &session.Players[i]
		r.players[session.ID][player.CharacterID] = player
	}
	
	return nil
}

// Read retrieves a session by ID
func (r *MemorySessionRepository) Read(id string) (*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	session, ok := r.sessions[id]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	
	// Return a copy
	result := *session
	return &result, nil
}

// Update updates an existing session
func (r *MemorySessionRepository) Update(session *domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.sessions[session.ID]; !ok {
		return fmt.Errorf("session not found: %s", session.ID)
	}
	
	r.sessions[session.ID] = session
	
	// Update player states
	r.players[session.ID] = make(map[string]*domain.PlayerState)
	for i := range session.Players {
		player := &session.Players[i]
		r.players[session.ID][player.CharacterID] = player
	}
	
	return nil
}

// End marks a session as ended
func (r *MemorySessionRepository) End(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	session, ok := r.sessions[id]
	if !ok {
		return fmt.Errorf("session not found: %s", id)
	}
	
	now := time.Now()
	session.EndedAt = &now
	return nil
}

// ListByCampaign returns all sessions for a campaign
func (r *MemorySessionRepository) ListByCampaign(campaignID string) ([]*domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*domain.Session
	for _, session := range r.sessions {
		if session.CampaignID == campaignID {
			s := *session
			result = append(result, &s)
		}
	}
	return result, nil
}

// SavePlayerState saves a player's state
func (r *MemorySessionRepository) SavePlayerState(sessionID string, player *domain.PlayerState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, ok := r.players[sessionID]; !ok {
		r.players[sessionID] = make(map[string]*domain.PlayerState)
	}
	
	p := *player
	r.players[sessionID][player.CharacterID] = &p
	return nil
}

// GetPlayerState retrieves a player's state
func (r *MemorySessionRepository) GetPlayerState(sessionID, characterID string) (*domain.PlayerState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	sessionPlayers, ok := r.players[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	player, ok := sessionPlayers[characterID]
	if !ok {
		return nil, fmt.Errorf("player not found: %s", characterID)
	}
	
	p := *player
	return &p, nil
}

// ListPlayerStates returns all player states for a session
func (r *MemorySessionRepository) ListPlayerStates(sessionID string) ([]domain.PlayerState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	sessionPlayers, ok := r.players[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	result := make([]domain.PlayerState, 0, len(sessionPlayers))
	for _, player := range sessionPlayers {
		p := *player
		result = append(result, p)
	}
	return result, nil
}

// AppendEvent adds an event to the session log
func (r *MemorySessionRepository) AppendEvent(event *domain.SessionEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	event.ID = int64(len(r.events[event.SessionID]) + 1)
	event.Timestamp = time.Now()
	
	r.events[event.SessionID] = append(r.events[event.SessionID], event)
	return nil
}

// GetEvents returns events for a session (most recent first)
func (r *MemorySessionRepository) GetEvents(sessionID string, limit int) ([]*domain.SessionEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	events := r.events[sessionID]
	if len(events) == 0 {
		return []*domain.SessionEvent{}, nil
	}
	
	// Return most recent first
	start := len(events) - limit
	if start < 0 {
		start = 0
	}
	
	result := make([]*domain.SessionEvent, 0, len(events)-start)
	for i := len(events) - 1; i >= start; i-- {
		e := *events[i]
		result = append(result, &e)
	}
	
	return result, nil
}

// GetEventsSince returns events since a given time
func (r *MemorySessionRepository) GetEventsSince(sessionID string, since time.Time) ([]*domain.SessionEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var result []*domain.SessionEvent
	for _, event := range r.events[sessionID] {
		if event.Timestamp.After(since) || event.Timestamp.Equal(since) {
			e := *event
			result = append(result, &e)
		}
	}
	return result, nil
}

// SaveCombatState saves the combat state
func (r *MemorySessionRepository) SaveCombatState(sessionID string, combat *domain.CombatState) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	c := *combat
	r.combat[sessionID] = &c
	return nil
}

// GetCombatState retrieves the combat state
func (r *MemorySessionRepository) GetCombatState(sessionID string) (*domain.CombatState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	combat, ok := r.combat[sessionID]
	if !ok {
		return nil, fmt.Errorf("no combat state for session: %s", sessionID)
	}
	
	c := *combat
	return &c, nil
}

// ClearCombatState removes the combat state
func (r *MemorySessionRepository) ClearCombatState(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.combat, sessionID)
	return nil
}

// Compile-time interface check
var _ SessionRepository = (*MemorySessionRepository)(nil)
