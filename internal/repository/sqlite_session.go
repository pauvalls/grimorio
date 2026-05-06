package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	_ "modernc.org/sqlite"
)

// SQLiteSessionRepository is a SQLite implementation of SessionRepository
type SQLiteSessionRepository struct {
	db *sql.DB
}

// NewSQLiteSessionRepository creates a new SQLite session repository
func NewSQLiteSessionRepository(dbPath string) (*SQLiteSessionRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	repo := &SQLiteSessionRepository{db: db}
	if err := repo.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return repo, nil
}

// migrate creates the database schema
func (r *SQLiteSessionRepository) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		campaign_id TEXT NOT NULL,
		in_combat INTEGER DEFAULT 0,
		current_scene TEXT,
		started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ended_at DATETIME
	);

	CREATE TABLE IF NOT EXISTS player_states (
		session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
		character_id TEXT NOT NULL,
		current_hp INTEGER NOT NULL,
		max_hp INTEGER NOT NULL,
		temp_hp INTEGER DEFAULT 0,
		ac INTEGER NOT NULL,
		position_x INTEGER DEFAULT 0,
		position_y INTEGER DEFAULT 0,
		initiative INTEGER DEFAULT 0,
		is_active INTEGER DEFAULT 0,
		conditions TEXT DEFAULT '[]',
		PRIMARY KEY (session_id, character_id)
	);

	CREATE TABLE IF NOT EXISTS session_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT REFERENCES sessions(id) ON DELETE CASCADE,
		type TEXT NOT NULL,
		actor TEXT NOT NULL,
		content TEXT NOT NULL,
		result TEXT DEFAULT '{}',
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS combat_state (
		session_id TEXT PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
		round INTEGER DEFAULT 1,
		initiative_order TEXT NOT NULL,
		active_index INTEGER DEFAULT 0,
		map_id TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_events_session ON session_events(session_id);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON session_events(timestamp);
	`

	_, err := r.db.Exec(schema)
	return err
}

// Create stores a new session
func (r *SQLiteSessionRepository) Create(session *domain.Session) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert session
	_, err = tx.Exec(
		"INSERT INTO sessions (id, campaign_id, in_combat, current_scene, started_at) VALUES (?, ?, ?, ?, ?)",
		session.ID, session.CampaignID, boolToInt(session.InCombat), 
		session.CurrentScene, session.StartedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert session: %w", err)
	}

	// Insert player states
	for _, player := range session.Players {
		if err := r.insertPlayerState(tx, session.ID, &player); err != nil {
			return fmt.Errorf("failed to insert player state: %w", err)
		}
	}

	return tx.Commit()
}

// Read retrieves a session by ID
func (r *SQLiteSessionRepository) Read(id string) (*domain.Session, error) {
	row := r.db.QueryRow(
		"SELECT id, campaign_id, in_combat, current_scene, started_at, ended_at FROM sessions WHERE id = ?",
		id,
	)

	session := &domain.Session{}
	var endedAt sql.NullTime
	var inCombat int
	var currentScene sql.NullString

	err := row.Scan(&session.ID, &session.CampaignID, &inCombat,
		&currentScene, &session.StartedAt, &endedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	session.InCombat = inCombat == 1
	if currentScene.Valid {
		sceneName := currentScene.String
		session.CurrentScene = &domain.Scene{Name: sceneName}
	}
	if endedAt.Valid {
		session.EndedAt = &endedAt.Time
	}

	// Load player states
	players, err := r.ListPlayerStates(id)
	if err == nil {
		session.Players = players
	}

	return session, nil
}

// Update updates an existing session
func (r *SQLiteSessionRepository) Update(session *domain.Session) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update session
	_, err = tx.Exec(
		"UPDATE sessions SET campaign_id = ?, in_combat = ?, current_scene = ? WHERE id = ?",
		session.CampaignID, boolToInt(session.InCombat), session.CurrentScene, session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Update player states
	for _, player := range session.Players {
		if err := r.insertPlayerState(tx, session.ID, &player); err != nil {
			return fmt.Errorf("failed to update player state: %w", err)
		}
	}

	return tx.Commit()
}

// End marks a session as ended
func (r *SQLiteSessionRepository) End(id string) error {
	_, err := r.db.Exec(
		"UPDATE sessions SET ended_at = ? WHERE id = ?",
		time.Now(), id,
	)
	if err != nil {
		return fmt.Errorf("failed to end session: %w", err)
	}
	return nil
}

// ListByCampaign returns all sessions for a campaign
func (r *SQLiteSessionRepository) ListByCampaign(campaignID string) ([]*domain.Session, error) {
	rows, err := r.db.Query(
		"SELECT id, campaign_id, in_combat, current_scene, started_at, ended_at FROM sessions WHERE campaign_id = ?",
		campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session := &domain.Session{}
		var endedAt sql.NullTime
		var inCombat int
		var currentScene sql.NullString

		err := rows.Scan(
			&session.ID, &session.CampaignID, &inCombat,
			&currentScene, &session.StartedAt, &endedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		session.InCombat = inCombat == 1
		if currentScene.Valid {
			sceneName := currentScene.String
			session.CurrentScene = &domain.Scene{Name: sceneName}
		}
		if endedAt.Valid {
			session.EndedAt = &endedAt.Time
		}

		sessions = append(sessions, session)
	}

	return sessions, nil
}

// SavePlayerState saves a player's state
func (r *SQLiteSessionRepository) SavePlayerState(sessionID string, player *domain.PlayerState) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO player_states 
		(session_id, character_id, current_hp, max_hp, temp_hp, ac, position_x, position_y, initiative, is_active, conditions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, player.CharacterID, player.CurrentHP, player.MaxHP, player.TempHP,
		player.AC, player.Position.X, player.Position.Y, player.Initiative,
		boolToInt(player.IsActive), "[]", // TODO: Serialize conditions
	)
	if err != nil {
		return fmt.Errorf("failed to save player state: %w", err)
	}
	return nil
}

// GetPlayerState retrieves a player's state
func (r *SQLiteSessionRepository) GetPlayerState(sessionID, characterID string) (*domain.PlayerState, error) {
	row := r.db.QueryRow(
		`SELECT character_id, current_hp, max_hp, temp_hp, ac, position_x, position_y, initiative, is_active 
		FROM player_states WHERE session_id = ? AND character_id = ?`,
		sessionID, characterID,
	)

	player := &domain.PlayerState{}
	var isActive int

	err := row.Scan(
		&player.CharacterID, &player.CurrentHP, &player.MaxHP, &player.TempHP,
		&player.AC, &player.Position.X, &player.Position.Y, &player.Initiative, &isActive,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("player not found: %s", characterID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read player state: %w", err)
	}

	player.IsActive = isActive == 1
	return player, nil
}

// ListPlayerStates returns all player states for a session
func (r *SQLiteSessionRepository) ListPlayerStates(sessionID string) ([]domain.PlayerState, error) {
	rows, err := r.db.Query(
		`SELECT character_id, current_hp, max_hp, temp_hp, ac, position_x, position_y, initiative, is_active 
		FROM player_states WHERE session_id = ?`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list player states: %w", err)
	}
	defer rows.Close()

	var players []domain.PlayerState
	for rows.Next() {
		player := domain.PlayerState{}
		var isActive int

		err := rows.Scan(
			&player.CharacterID, &player.CurrentHP, &player.MaxHP, &player.TempHP,
			&player.AC, &player.Position.X, &player.Position.Y, &player.Initiative, &isActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player state: %w", err)
		}

		player.IsActive = isActive == 1
		players = append(players, player)
	}

	return players, nil
}

// AppendEvent adds an event to the session log
func (r *SQLiteSessionRepository) AppendEvent(event *domain.SessionEvent) error {
	resultJSON := "{}"
	if event.Result != nil {
		// TODO: Serialize result to JSON
	}

	_, err := r.db.Exec(
		"INSERT INTO session_events (session_id, type, actor, content, result) VALUES (?, ?, ?, ?, ?)",
		event.SessionID, event.Type, event.Actor, event.Content, resultJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}
	return nil
}

// GetEvents returns events for a session (most recent first)
func (r *SQLiteSessionRepository) GetEvents(sessionID string, limit int) ([]*domain.SessionEvent, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, type, actor, content, result, timestamp 
		FROM session_events WHERE session_id = ? ORDER BY timestamp DESC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []*domain.SessionEvent
	for rows.Next() {
		event := &domain.SessionEvent{}
		var resultStr string

		err := rows.Scan(
			&event.ID, &event.SessionID, &event.Type, &event.Actor,
			&event.Content, &resultStr, &event.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// TODO: Deserialize result from JSON
		events = append(events, event)
	}

	return events, nil
}

// GetEventsSince returns events since a given time
func (r *SQLiteSessionRepository) GetEventsSince(sessionID string, since time.Time) ([]*domain.SessionEvent, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, type, actor, content, result, timestamp 
		FROM session_events WHERE session_id = ? AND timestamp >= ? ORDER BY timestamp`,
		sessionID, since,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer rows.Close()

	var events []*domain.SessionEvent
	for rows.Next() {
		event := &domain.SessionEvent{}
		var resultStr string

		err := rows.Scan(
			&event.ID, &event.SessionID, &event.Type, &event.Actor,
			&event.Content, &resultStr, &event.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		events = append(events, event)
	}

	return events, nil
}

// SaveCombatState saves the combat state
func (r *SQLiteSessionRepository) SaveCombatState(sessionID string, combat *domain.CombatState) error {
	// Serialize initiative order
	initiativeOrder := ""
	for i, actor := range combat.InitiativeOrder {
		if i > 0 {
			initiativeOrder += ","
		}
		initiativeOrder += actor
	}

	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO combat_state (session_id, round, initiative_order, active_index, map_id)
		VALUES (?, ?, ?, ?, ?)`,
		sessionID, combat.Round, initiativeOrder, combat.ActiveIndex, combat.MapID,
	)
	if err != nil {
		return fmt.Errorf("failed to save combat state: %w", err)
	}
	return nil
}

// GetCombatState retrieves the combat state
func (r *SQLiteSessionRepository) GetCombatState(sessionID string) (*domain.CombatState, error) {
	row := r.db.QueryRow(
		"SELECT session_id, round, initiative_order, active_index, map_id FROM combat_state WHERE session_id = ?",
		sessionID,
	)

	combat := &domain.CombatState{}
	var initiativeOrderStr string

	err := row.Scan(
		&combat.SessionID, &combat.Round, &initiativeOrderStr,
		&combat.ActiveIndex, &combat.MapID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no combat state for session: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read combat state: %w", err)
	}

	// Deserialize initiative order
	if initiativeOrderStr != "" {
		combat.InitiativeOrder = splitInitiativeOrder(initiativeOrderStr)
	}

	return combat, nil
}

// ClearCombatState removes the combat state
func (r *SQLiteSessionRepository) ClearCombatState(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM combat_state WHERE session_id = ?", sessionID)
	if err != nil {
		return fmt.Errorf("failed to clear combat state: %w", err)
	}
	return nil
}

// Helper functions

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (r *SQLiteSessionRepository) insertPlayerState(tx *sql.Tx, sessionID string, player *domain.PlayerState) error {
	_, err := tx.Exec(
		`INSERT OR REPLACE INTO player_states 
		(session_id, character_id, current_hp, max_hp, temp_hp, ac, position_x, position_y, initiative, is_active, conditions)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, player.CharacterID, player.CurrentHP, player.MaxHP, player.TempHP,
		player.AC, player.Position.X, player.Position.Y, player.Initiative,
		boolToInt(player.IsActive), "[]",
	)
	return err
}

func splitInitiativeOrder(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// Compile-time interface check
var _ SessionRepository = (*SQLiteSessionRepository)(nil)
