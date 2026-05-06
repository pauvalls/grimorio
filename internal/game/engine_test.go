package game

import (
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestGameEngine_StartSession(t *testing.T) {
	// Setup
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	// Create a test campaign
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	// Create test characters
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1, Stats: domain.Stats{STR: 14, DEX: 12, CON: 13, INT: 10, WIS: 10, CHA: 8}}
	char2 := &domain.Character{CampaignID: "test-campaign", Name: "Hero2", Class: "mago", Level: 1, Stats: domain.Stats{STR: 8, DEX: 10, CON: 10, INT: 16, WIS: 14, CHA: 10}}
	charRepo.Save(char1)
	charRepo.Save(char2)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	
	// Test start session
	session, err := engine.StartSession("test-campaign", []string{"Hero1", "Hero2"})
	if err != nil {
		t.Fatalf("StartSession() error: %v", err)
	}
	
	if session.ID == "" {
		t.Error("StartSession() session ID should not be empty")
	}
	if session.CampaignID != "test-campaign" {
		t.Errorf("StartSession() campaign = %v, want test-campaign", session.CampaignID)
	}
	if len(session.Players) != 2 {
		t.Errorf("StartSession() players = %d, want 2", len(session.Players))
	}
	
	// Verify players have correct HP
	player1 := session.GetPlayer("Hero1")
	if player1 == nil {
		t.Fatal("StartSession() Hero1 not found")
	}
	if player1.MaxHP != 11 { // Level 1 fighter: 10 base + 1 CON modifier
		t.Errorf("StartSession() Hero1 max HP = %d, want 11", player1.MaxHP)
	}
	
	// Verify session is persisted
	stored, err := sessionRepo.Read(session.ID)
	if err != nil {
		t.Fatalf("StartSession() session not persisted: %v", err)
	}
	if stored.ID != session.ID {
		t.Error("StartSession() persisted session ID mismatch")
	}
}

func TestGameEngine_ProcessAction(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1, Stats: domain.Stats{STR: 14, DEX: 12, CON: 13, INT: 10, WIS: 10, CHA: 8}}
	charRepo.Save(char1)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	session, _ := engine.StartSession("test-campaign", []string{"Hero1"})
	
	// Test exploration action
	result, err := engine.ProcessAction(session.ID, "Hero1", "explore the cave")
	if err != nil {
		t.Fatalf("ProcessAction() error: %v", err)
	}
	if result == nil {
		t.Fatal("ProcessAction() result is nil")
	}
	
	// Verify event was logged
	events, _ := sessionRepo.GetEvents(session.ID, 10)
	if len(events) == 0 {
		t.Error("ProcessAction() no events logged")
	}
}

func TestGameEngine_RollDice(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1}
	charRepo.Save(char1)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	session, _ := engine.StartSession("test-campaign", []string{"Hero1"})
	
	// Test rolling dice
	result, err := engine.RollDice(session.ID, "Hero1", "2d6+3")
	if err != nil {
		t.Fatalf("RollDice() error: %v", err)
	}
	if result.Total < 5 || result.Total > 15 {
		t.Errorf("RollDice() total = %d, want between 5 and 15", result.Total)
	}
	
	// Verify event was logged
	events, _ := sessionRepo.GetEvents(session.ID, 10)
	foundDiceEvent := false
	for _, e := range events {
		if e.Type == domain.EventTypeDiceRoll {
			foundDiceEvent = true
			break
		}
	}
	if !foundDiceEvent {
		t.Error("RollDice() no dice roll event logged")
	}
}

func TestGameEngine_Combat(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1, Stats: domain.Stats{STR: 14, DEX: 12, CON: 13, INT: 10, WIS: 10, CHA: 8}}
	char2 := &domain.Character{CampaignID: "test-campaign", Name: "Goblin", Class: "guerrero", Level: 1, Stats: domain.Stats{STR: 8, DEX: 14, CON: 10, INT: 6, WIS: 8, CHA: 6}}
	charRepo.Save(char1)
	charRepo.Save(char2)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	session, _ := engine.StartSession("test-campaign", []string{"Hero1", "Goblin"})
	
	// Start combat
	enemies := []domain.PlayerState{
		{CharacterID: "Goblin", CurrentHP: 7, MaxHP: 7, AC: 12},
	}
	err := engine.StartCombat(session.ID, enemies)
	if err != nil {
		t.Fatalf("StartCombat() error: %v", err)
	}
	
	// Get updated session
	session, _ = engine.GetSession(session.ID)
	if !session.InCombat {
		t.Error("StartCombat() session should be in combat")
	}
	
	// Get combat state
	combat, _ := sessionRepo.GetCombatState(session.ID)
	if combat == nil {
		t.Fatal("StartCombat() combat state not created")
	}
	if combat.Round != 1 {
		t.Errorf("StartCombat() round = %d, want 1", combat.Round)
	}
	
	// Test attack (use the active actor)
	activeActor := combat.GetActiveActor()
	var target string
	if activeActor == "Hero1" {
		target = "Goblin"
	} else {
		target = "Hero1"
	}
	attackResult, err := engine.Attack(session.ID, activeActor, target, 15)
	if err != nil {
		t.Fatalf("Attack() error: %v", err)
	}
	if attackResult == nil {
		t.Fatal("Attack() result is nil")
	}
	
	// Test next turn
	err = engine.NextTurn(session.ID)
	if err != nil {
		t.Fatalf("NextTurn() error: %v", err)
	}
	
	// End combat
	err = engine.EndCombat(session.ID)
	if err != nil {
		t.Fatalf("EndCombat() error: %v", err)
	}
	
	session, _ = engine.GetSession(session.ID)
	if session.InCombat {
		t.Error("EndCombat() session should not be in combat")
	}
}

func TestGameEngine_GetState(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1}
	charRepo.Save(char1)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	session, _ := engine.StartSession("test-campaign", []string{"Hero1"})
	
	state, err := engine.GetState(session.ID)
	if err != nil {
		t.Fatalf("GetState() error: %v", err)
	}
	if state.SessionID != session.ID {
		t.Errorf("GetState() session ID = %v, want %v", state.SessionID, session.ID)
	}
	if state.CampaignID != "test-campaign" {
		t.Errorf("GetState() campaign = %v, want test-campaign", state.CampaignID)
	}
}

func TestGameEngine_MoveToken(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1}
	charRepo.Save(char1)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	session, _ := engine.StartSession("test-campaign", []string{"Hero1"})
	
	err := engine.MoveToken(session.ID, "Hero1", 5, 10)
	if err != nil {
		t.Fatalf("MoveToken() error: %v", err)
	}
	
	// Verify position
	session, _ = engine.GetSession(session.ID)
	player := session.GetPlayer("Hero1")
	if player.Position.X != 5 || player.Position.Y != 10 {
		t.Errorf("MoveToken() position = (%d, %d), want (5, 10)", player.Position.X, player.Position.Y)
	}
}

func TestGameEngine_EndSession(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	char1 := &domain.Character{CampaignID: "test-campaign", Name: "Hero1", Class: "guerrero", Level: 1}
	charRepo.Save(char1)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	session, _ := engine.StartSession("test-campaign", []string{"Hero1"})
	
	// Add some events
	engine.ProcessAction(session.ID, "Hero1", "enter the dungeon")
	engine.RollDice(session.ID, "Hero1", "d20")
	
	// End session
	summary, err := engine.EndSession(session.ID)
	if err != nil {
		t.Fatalf("EndSession() error: %v", err)
	}
	if summary == nil {
		t.Fatal("EndSession() summary is nil")
	}
	if summary.SessionID != session.ID {
		t.Errorf("EndSession() session ID = %v, want %v", summary.SessionID, session.ID)
	}
	if summary.Duration == "" {
		t.Error("EndSession() duration should not be empty")
	}
	if summary.EventsCount == 0 {
		t.Error("EndSession() events count should be > 0")
	}
}

func TestGameEngine_InvalidCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	
	_, err := engine.StartSession("nonexistent", []string{"Hero1"})
	if err == nil {
		t.Error("StartSession() should error for nonexistent campaign")
	}
}

func TestGameEngine_InvalidCharacter(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	questRepo := repository.NewMemoryQuestRepository()
	sessionRepo := repository.NewMemorySessionRepository()
	
	campaign := &domain.Campaign{Name: "test-campaign", Title: "Test"}
	campaignRepo.Create(campaign)
	
	engine := NewEngine(sessionRepo, campaignRepo, charRepo, questRepo, nil)
	
	_, err := engine.StartSession("test-campaign", []string{"NonExistent"})
	if err == nil {
		t.Error("StartSession() should error for nonexistent character")
	}
}
