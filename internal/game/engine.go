package game

import (
	"fmt"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// Engine is the core game engine that manages game sessions
type Engine struct {
	sessionRepo  repository.SessionRepository
	campaignRepo repository.CampaignRepository
	charRepo     repository.CharacterRepository
	questRepo    repository.QuestRepository
	resolver     *CombatResolver
	llmClient    LLMClient
}

// NewEngine creates a new game engine
func NewEngine(
	sessionRepo repository.SessionRepository,
	campaignRepo repository.CampaignRepository,
	charRepo repository.CharacterRepository,
	questRepo repository.QuestRepository,
	llmClient LLMClient,
) *Engine {
	return &Engine{
		sessionRepo:  sessionRepo,
		campaignRepo: campaignRepo,
		charRepo:     charRepo,
		questRepo:    questRepo,
		resolver:     NewCombatResolver(),
		llmClient:    llmClient,
	}
}

// StartSession creates a new game session for a campaign
func (e *Engine) StartSession(campaignID string, playerNames []string) (*domain.Session, error) {
	// Verify campaign exists
	campaign, err := e.campaignRepo.Read(campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}

	// Load player characters
	players := make([]domain.PlayerState, 0, len(playerNames))
	for _, name := range playerNames {
		char, err := e.charRepo.Read(campaignID, name)
		if err != nil {
			return nil, fmt.Errorf("character not found: %s - %w", name, err)
		}

		maxHP := calculateMaxHP(char)
		ac := calculateAC(char)
		
		players = append(players, domain.PlayerState{
			CharacterID: char.Name,
			CurrentHP:   maxHP,
			MaxHP:       maxHP,
			TempHP:      0,
			AC:          ac,
			Position:    domain.Point{X: 0, Y: 0},
			Initiative:  0,
			IsActive:    false,
			Conditions:  []domain.Condition{},
		})
	}

	session := &domain.Session{
		ID:         generateSessionID(),
		CampaignID: campaignID,
		Players:    players,
		InCombat:   false,
		StartedAt:  time.Now(),
	}

	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	if err := e.sessionRepo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log session start event
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: session.ID,
		Type:      domain.EventTypeSceneChange,
		Actor:     "system",
		Content:   fmt.Sprintf("Session started for campaign: %s", campaign.Title),
	})

	return session, nil
}

// GetSession retrieves a session by ID
func (e *Engine) GetSession(sessionID string) (*domain.Session, error) {
	session, err := e.sessionRepo.Read(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Load current player states
	players, err := e.sessionRepo.ListPlayerStates(sessionID)
	if err == nil {
		session.Players = players
	}

	return session, nil
}

// EndSession ends a game session and returns a summary
func (e *Engine) EndSession(sessionID string) (*domain.SessionSummary, error) {
	session, err := e.sessionRepo.Read(sessionID)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if err := e.sessionRepo.End(sessionID); err != nil {
		return nil, fmt.Errorf("failed to end session: %w", err)
	}

	// Get events count
	events, _ := e.sessionRepo.GetEvents(sessionID, 100000)
	
	// Count alive/dead players
	players, _ := e.sessionRepo.ListPlayerStates(sessionID)
	alive := 0
	dead := 0
	for _, p := range players {
		if p.IsAlive() {
			alive++
		} else {
			dead++
		}
	}

	// Calculate duration
	var duration time.Duration
	if session.EndedAt != nil {
		duration = session.EndedAt.Sub(session.StartedAt)
	}

	endedAt := time.Now()
	if session.EndedAt != nil {
		endedAt = *session.EndedAt
	}
	
	summary := &domain.SessionSummary{
		SessionID:    sessionID,
		CampaignID:   session.CampaignID,
		Duration:     duration.String(),
		RoundsPlayed: 0, // Would track from combat state
		EventsCount:  len(events),
		PlayersAlive: alive,
		PlayersDead:  dead,
		KeyEvents:    extractKeyEvents(events),
		EndedAt:      endedAt,
	}

	return summary, nil
}

// ProcessAction processes a player action
func (e *Engine) ProcessAction(sessionID, characterID, action string) (*domain.ActionResult, error) {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	player := session.GetPlayer(characterID)
	if player == nil {
		return nil, fmt.Errorf("player not found: %s", characterID)
	}

	result := &domain.ActionResult{
		Success:     true,
		Description: fmt.Sprintf("%s performs: %s", characterID, action),
	}

	// Log the action
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeAction,
		Actor:     characterID,
		Content:   action,
		Result:    domain.JSON{"success": true},
	})

	// If LLM client is available, get DM narration
	if e.llmClient != nil {
		narration, err := e.generateNarration(session, characterID, action)
		if err == nil {
			result.Description = narration
			e.sessionRepo.AppendEvent(&domain.SessionEvent{
				SessionID: sessionID,
				Type:      domain.EventTypeNarration,
				Actor:     "dm",
				Content:   narration,
			})
		}
	}

	return result, nil
}

// RollDice rolls dice for a player
func (e *Engine) RollDice(sessionID, characterID, dice string) (*domain.DiceResult, error) {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.GetPlayer(characterID) == nil {
		return nil, fmt.Errorf("player not found: %s", characterID)
	}

	spec, err := ParseDice(dice)
	if err != nil {
		return nil, fmt.Errorf("invalid dice notation: %w", err)
	}

	result := Roll(spec)

	// Log the roll
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeDiceRoll,
		Actor:     characterID,
		Content:   fmt.Sprintf("Rolled %s = %d", dice, result.Total),
		Result: domain.JSON{
			"dice":    dice,
			"results": result.Results,
			"total":   result.Total,
		},
	})

	return result, nil
}

// MoveToken moves a player's token on the map
func (e *Engine) MoveToken(sessionID, characterID string, x, y int) error {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return err
	}

	player := session.GetPlayer(characterID)
	if player == nil {
		return fmt.Errorf("player not found: %s", characterID)
	}

	player.Position = domain.Point{X: x, Y: y}

	if err := e.sessionRepo.SavePlayerState(sessionID, player); err != nil {
		return fmt.Errorf("failed to save player state: %w", err)
	}

	// Log movement
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeAction,
		Actor:     characterID,
		Content:   fmt.Sprintf("Moved to position (%d, %d)", x, y),
		Result: domain.JSON{
			"position": domain.Point{X: x, Y: y},
		},
	})

	return nil
}

// StartCombat begins combat in a session
func (e *Engine) StartCombat(sessionID string, enemies []domain.PlayerState) error {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return err
	}

	// Combine players and enemies
	allActors := make([]domain.PlayerState, 0, len(session.Players)+len(enemies))
	
	// Load current player states
	players, _ := e.sessionRepo.ListPlayerStates(sessionID)
	allActors = append(allActors, players...)
	allActors = append(allActors, enemies...)

	// Calculate initiative
	initiativeActors := make([]InitiativeActor, len(allActors))
	for i, actor := range allActors {
		initiativeActors[i] = InitiativeActor{
			CharacterID: actor.CharacterID,
			DEXModifier: domain.CalculateModifier(0), // TODO: Get actual DEX from character
		}
	}
	initiativeOrder := e.resolver.CalculateInitiativeOrder(initiativeActors)

	combat := &domain.CombatState{
		SessionID:       sessionID,
		Round:           1,
		InitiativeOrder: initiativeOrder,
		ActiveIndex:     0,
	}

	if err := combat.Validate(); err != nil {
		return fmt.Errorf("invalid combat state: %w", err)
	}

	if err := e.sessionRepo.SaveCombatState(sessionID, combat); err != nil {
		return fmt.Errorf("failed to save combat state: %w", err)
	}

	// Mark enemies as players in session (simplified)
	for _, enemy := range enemies {
		e.sessionRepo.SavePlayerState(sessionID, &enemy)
	}

	session.InCombat = true
	if err := e.sessionRepo.Update(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Log combat start
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeCombatStart,
		Actor:     "system",
		Content:   fmt.Sprintf("Combat started! Round 1. Initiative order: %v", initiativeOrder),
		Result: domain.JSON{
			"initiative_order": initiativeOrder,
			"round":            1,
		},
	})

	return nil
}

// Attack resolves an attack from one actor to another
func (e *Engine) Attack(sessionID, attackerID, targetID string, attackRoll int) (*AttackResult, error) {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if !session.InCombat {
		return nil, fmt.Errorf("not in combat")
	}

	attacker := session.GetPlayer(attackerID)
	if attacker == nil {
		return nil, fmt.Errorf("attacker not found: %s", attackerID)
	}

	target := session.GetPlayer(targetID)
	if target == nil {
		return nil, fmt.Errorf("target not found: %s", targetID)
	}

	// Get combat state
	combat, err := e.sessionRepo.GetCombatState(sessionID)
	if err != nil {
		return nil, fmt.Errorf("combat state not found: %w", err)
	}

	// Verify it's the attacker's turn
	if combat.GetActiveActor() != attackerID {
		return nil, fmt.Errorf("not %s's turn", attackerID)
	}

	// Resolve attack
	// TODO: Get actual attack bonus from character stats
	attackBonus := 0
	result := e.resolver.ResolveAttack(attacker, target, attackRoll, attackBonus)

	if result.Hit {
		// Roll damage (simplified: 1d8 for now)
		damage, err := e.resolver.CalculateDamage("1d8", result.CriticalHit)
		if err == nil {
			result.Damage = damage
			e.resolver.ApplyDamageToPlayer(target, damage)
			
			// Save updated target state
			e.sessionRepo.SavePlayerState(sessionID, target)
		}
	}

	// Log attack
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeAction,
		Actor:     attackerID,
		Content:   fmt.Sprintf("Attacked %s: %s", targetID, result.Description),
		Result: domain.JSON{
			"attack_roll":   attackRoll,
			"hit":           result.Hit,
			"critical_hit":  result.CriticalHit,
			"damage":        result.Damage,
			"target_hp":     target.CurrentHP,
		},
	})

	return result, nil
}

// NextTurn advances to the next turn in combat
func (e *Engine) NextTurn(sessionID string) error {
	combat, err := e.sessionRepo.GetCombatState(sessionID)
	if err != nil {
		return fmt.Errorf("combat state not found: %w", err)
	}

	previousActor := combat.GetActiveActor()
	combat.NextTurn()

	if err := e.sessionRepo.SaveCombatState(sessionID, combat); err != nil {
		return fmt.Errorf("failed to save combat state: %w", err)
	}

	// Update active player
	session, _ := e.GetSession(sessionID)
	for i := range session.Players {
		session.Players[i].IsActive = session.Players[i].CharacterID == combat.GetActiveActor()
	}
	e.sessionRepo.Update(session)

	// Log turn change
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeTurnChange,
		Actor:     "system",
		Content:   fmt.Sprintf("Turn changed from %s to %s (Round %d)", previousActor, combat.GetActiveActor(), combat.Round),
		Result: domain.JSON{
			"previous_actor": previousActor,
			"new_actor":      combat.GetActiveActor(),
			"round":          combat.Round,
		},
	})

	return nil
}

// EndCombat ends combat in a session
func (e *Engine) EndCombat(sessionID string) error {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return err
	}

	if !session.InCombat {
		return fmt.Errorf("not in combat")
	}

	if err := e.sessionRepo.ClearCombatState(sessionID); err != nil {
		return fmt.Errorf("failed to clear combat state: %w", err)
	}

	session.InCombat = false
	for i := range session.Players {
		session.Players[i].IsActive = false
	}

	if err := e.sessionRepo.Update(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Log combat end
	e.sessionRepo.AppendEvent(&domain.SessionEvent{
		SessionID: sessionID,
		Type:      domain.EventTypeCombatEnd,
		Actor:     "system",
		Content:   "Combat ended",
	})

	return nil
}

// GetState returns the current game state
func (e *Engine) GetState(sessionID string) (*domain.GameState, error) {
	session, err := e.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	state := &domain.GameState{
		SessionID:    sessionID,
		CampaignID:   session.CampaignID,
		InCombat:     session.InCombat,
		CurrentScene: session.CurrentScene,
		Players:      session.Players,
	}

	if session.InCombat {
		combat, err := e.sessionRepo.GetCombatState(sessionID)
		if err == nil {
			state.Combat = combat
			state.ActiveActor = combat.GetActiveActor()
			state.Round = combat.Round
		}
	}

	return state, nil
}

// GetRecentEvents returns recent events for a session
func (e *Engine) GetRecentEvents(sessionID string, limit int) ([]*domain.SessionEvent, error) {
	return e.sessionRepo.GetEvents(sessionID, limit)
}

// Helper functions

func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

func calculateMaxHP(char *domain.Character) int {
	// Simplified HP calculation
	// In a real implementation, this would use class hit dice + CON modifier * level
	baseHP := 10
	if char.Stats.CON > 0 {
		conMod := domain.CalculateModifier(char.Stats.CON)
		baseHP += conMod * char.Level
	}
	return baseHP
}

func calculateAC(char *domain.Character) int {
	// Simplified AC calculation
	// Base 10 + DEX modifier (capped at 2 for medium armor)
	ac := 10
	if char.Stats.DEX > 0 {
		dexMod := domain.CalculateModifier(char.Stats.DEX)
		if dexMod > 2 {
			dexMod = 2
		}
		ac += dexMod
	}
	return ac
}

func extractKeyEvents(events []*domain.SessionEvent) []string {
	var keyEvents []string
	for _, event := range events {
		if event.Type == domain.EventTypeCombatStart || 
		   event.Type == domain.EventTypeCombatEnd ||
		   event.Type == domain.EventTypeSceneChange {
			keyEvents = append(keyEvents, event.Content)
		}
	}
	return keyEvents
}

func (e *Engine) generateNarration(session *domain.Session, characterID, action string) (string, error) {
	// TODO: Implement DM narration via LLM
	return fmt.Sprintf("%s %s...", characterID, action), nil
}
