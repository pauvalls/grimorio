package domain

import (
	"testing"
	"time"
)

func TestSessionValidation(t *testing.T) {
	tests := []struct {
		name     string
		session  Session
		wantErr  bool
		errField string
	}{
		{
			name: "valid session",
			session: Session{
				CampaignID: "my-campaign",
				Players: []PlayerState{
					{CharacterID: "char-1", CurrentHP: 10, MaxHP: 10},
				},
			},
			wantErr: false,
		},
		{
			name: "missing campaign",
			session: Session{
				Players: []PlayerState{
					{CharacterID: "char-1", CurrentHP: 10, MaxHP: 10},
				},
			},
			wantErr:  true,
			errField: "campaign_id",
		},
		{
			name: "no players",
			session: Session{
				CampaignID: "my-campaign",
				Players:    []PlayerState{},
			},
			wantErr:  true,
			errField: "players",
		},
		{
			name: "invalid player - negative HP",
			session: Session{
				CampaignID: "my-campaign",
				Players: []PlayerState{
					{CharacterID: "char-1", CurrentHP: -5, MaxHP: 10},
				},
			},
			wantErr:  true,
			errField: "players",
		},
		{
			name: "invalid player - current HP > max HP",
			session: Session{
				CampaignID: "my-campaign",
				Players: []PlayerState{
					{CharacterID: "char-1", CurrentHP: 15, MaxHP: 10},
				},
			},
			wantErr:  true,
			errField: "players",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				if ve, ok := err.(ValidationError); ok {
					if ve.Field != tt.errField {
						t.Errorf("Validate() error field = %v, want %v", ve.Field, tt.errField)
					}
				} else {
					t.Errorf("Validate() expected ValidationError but got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSessionEventValidation(t *testing.T) {
	tests := []struct {
		name    string
		event   SessionEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: SessionEvent{
				SessionID: "session-1",
				Type:      EventTypeAction,
				Actor:     "player-1",
				Content:   "I attack the goblin",
			},
			wantErr: false,
		},
		{
			name: "missing session ID",
			event: SessionEvent{
				Type:    EventTypeAction,
				Actor:   "player-1",
				Content: "I attack",
			},
			wantErr: true,
		},
		{
			name: "invalid event type",
			event: SessionEvent{
				SessionID: "session-1",
				Type:      "invalid_type",
				Actor:     "player-1",
				Content:   "I attack",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				return
			}
			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestCombatStateValidation(t *testing.T) {
	tests := []struct {
		name    string
		combat  CombatState
		wantErr bool
	}{
		{
			name: "valid combat",
			combat: CombatState{
				SessionID:       "session-1",
				Round:           1,
				InitiativeOrder: []string{"char-1", "goblin-1"},
				ActiveIndex:     0,
			},
			wantErr: false,
		},
		{
			name: "empty initiative",
			combat: CombatState{
				SessionID:       "session-1",
				Round:           1,
				InitiativeOrder: []string{},
				ActiveIndex:     0,
			},
			wantErr: true,
		},
		{
			name: "active index out of bounds",
			combat: CombatState{
				SessionID:       "session-1",
				Round:           1,
				InitiativeOrder: []string{"char-1"},
				ActiveIndex:     5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.combat.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error but got nil")
					return
				}
				return
			}
			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestSession_GetActivePlayer(t *testing.T) {
	session := Session{
		ID:         "session-1",
		CampaignID: "campaign-1",
		Players: []PlayerState{
			{CharacterID: "char-1", CurrentHP: 10, MaxHP: 10, IsActive: true},
			{CharacterID: "char-2", CurrentHP: 8, MaxHP: 8, IsActive: false},
		},
	}

	active := session.GetActivePlayer()
	if active == nil {
		t.Fatal("GetActivePlayer() returned nil")
	}
	if active.CharacterID != "char-1" {
		t.Errorf("GetActivePlayer() = %v, want char-1", active.CharacterID)
	}
}

func TestSession_IsPlayerAlive(t *testing.T) {
	session := Session{
		ID:         "session-1",
		CampaignID: "campaign-1",
		Players: []PlayerState{
			{CharacterID: "char-1", CurrentHP: 10, MaxHP: 10},
			{CharacterID: "char-2", CurrentHP: 0, MaxHP: 8},
			{CharacterID: "char-3", CurrentHP: -2, MaxHP: 8},
		},
	}

	if !session.IsPlayerAlive("char-1") {
		t.Error("IsPlayerAlive(char-1) should be true")
	}
	if session.IsPlayerAlive("char-2") {
		t.Error("IsPlayerAlive(char-2) should be false (0 HP)")
	}
	if session.IsPlayerAlive("char-3") {
		t.Error("IsPlayerAlive(char-3) should be false (negative HP)")
	}
	if session.IsPlayerAlive("char-nonexistent") {
		t.Error("IsPlayerAlive(char-nonexistent) should be false")
	}
}

func TestCombatState_GetActiveActor(t *testing.T) {
	combat := CombatState{
		SessionID:       "session-1",
		Round:           1,
		InitiativeOrder: []string{"char-1", "goblin-1", "char-2"},
		ActiveIndex:     1,
	}

	active := combat.GetActiveActor()
	if active != "goblin-1" {
		t.Errorf("GetActiveActor() = %v, want goblin-1", active)
	}
}

func TestCombatState_NextTurn(t *testing.T) {
	combat := CombatState{
		SessionID:       "session-1",
		Round:           1,
		InitiativeOrder: []string{"char-1", "goblin-1", "char-2"},
		ActiveIndex:     0,
	}

	// First turn: char-1
	if combat.GetActiveActor() != "char-1" {
		t.Errorf("Initial active = %v, want char-1", combat.GetActiveActor())
	}

	// Next turn: goblin-1
	combat.NextTurn()
	if combat.GetActiveActor() != "goblin-1" {
		t.Errorf("After NextTurn() active = %v, want goblin-1", combat.GetActiveActor())
	}
	if combat.Round != 1 {
		t.Errorf("Round should still be 1, got %d", combat.Round)
	}

	// Next turn: char-2
	combat.NextTurn()
	if combat.GetActiveActor() != "char-2" {
		t.Errorf("After second NextTurn() active = %v, want char-2", combat.GetActiveActor())
	}

	// Next turn: back to char-1, round increments
	combat.NextTurn()
	if combat.GetActiveActor() != "char-1" {
		t.Errorf("After third NextTurn() active = %v, want char-1", combat.GetActiveActor())
	}
	if combat.Round != 2 {
		t.Errorf("Round should be 2, got %d", combat.Round)
	}
}

func TestPlayerState_ApplyDamage(t *testing.T) {
	tests := []struct {
		name         string
		player       PlayerState
		damage       int
		wantCurrent  int
		wantTemp     int
		wantAlive    bool
	}{
		{
			name:        "normal damage",
			player:      PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 0},
			damage:      3,
			wantCurrent: 7,
			wantTemp:    0,
			wantAlive:   true,
		},
		{
			name:        "damage with temp HP",
			player:      PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 5},
			damage:      3,
			wantCurrent: 10,
			wantTemp:    2,
			wantAlive:   true,
		},
		{
			name:        "damage exceeds temp HP",
			player:      PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 2},
			damage:      5,
			wantCurrent: 7,
			wantTemp:    0,
			wantAlive:   true,
		},
		{
			name:        "lethal damage",
			player:      PlayerState{CurrentHP: 5, MaxHP: 10, TempHP: 0},
			damage:      10,
			wantCurrent: 0,
			wantTemp:    0,
			wantAlive:   false,
		},
		{
			name:        "massive damage",
			player:      PlayerState{CurrentHP: 10, MaxHP: 10, TempHP: 5},
			damage:      50,
			wantCurrent: 0,
			wantTemp:    0,
			wantAlive:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.player
			player.ApplyDamage(tt.damage)
			if player.CurrentHP != tt.wantCurrent {
				t.Errorf("ApplyDamage() current HP = %d, want %d", player.CurrentHP, tt.wantCurrent)
			}
			if player.TempHP != tt.wantTemp {
				t.Errorf("ApplyDamage() temp HP = %d, want %d", player.TempHP, tt.wantTemp)
			}
			if player.IsAlive() != tt.wantAlive {
				t.Errorf("ApplyDamage() alive = %v, want %v", player.IsAlive(), tt.wantAlive)
			}
		})
	}
}

func TestPlayerState_Heal(t *testing.T) {
	tests := []struct {
		name        string
		player      PlayerState
		amount      int
		wantCurrent int
	}{
		{
			name:        "normal heal",
			player:      PlayerState{CurrentHP: 5, MaxHP: 10},
			amount:      3,
			wantCurrent: 8,
		},
		{
			name:        "heal to full",
			player:      PlayerState{CurrentHP: 7, MaxHP: 10},
			amount:      5,
			wantCurrent: 10,
		},
		{
			name:        "heal when at full",
			player:      PlayerState{CurrentHP: 10, MaxHP: 10},
			amount:      5,
			wantCurrent: 10,
		},
		{
			name:        "heal from 0",
			player:      PlayerState{CurrentHP: 0, MaxHP: 10},
			amount:      3,
			wantCurrent: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.player
			player.Heal(tt.amount)
			if player.CurrentHP != tt.wantCurrent {
				t.Errorf("Heal() current HP = %d, want %d", player.CurrentHP, tt.wantCurrent)
			}
		})
	}
}

func TestCondition_IsExpired(t *testing.T) {
	now := time.Now()

	active := Condition{
		Type:      ConditionPoisoned,
		Duration:  3,
		AppliedAt: now,
	}

	if active.IsExpired(0) {
		t.Error("Condition should not be expired at round 0")
	}
	if active.IsExpired(2) {
		t.Error("Condition should not be expired at round 2")
	}
	if !active.IsExpired(3) {
		t.Error("Condition should be expired at round 3")
	}
	if !active.IsExpired(5) {
		t.Error("Condition should be expired at round 5")
	}
}
