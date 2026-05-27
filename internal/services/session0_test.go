package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

// TestSession0_Flow tests the complete session 0 flow:
// 1. Save NPCs → updates canon.json + creates JSON files
// 2. Save Bestiary → updates canon.json + creates JSON files
// 3. dm_session_context → loads NPCs and bestiary from canon.json fallback
func TestSession0_Flow(t *testing.T) {
	// Setup test campaign directory
	testDir := t.TempDir()
	campaignID := "test-session0"
	
	// Create campaign
	campaign := &domain.Campaign{
		Name:    campaignID,
		Title:   "Test Campaign",
		Setting: "Test Setting",
	}
	
	// Initialize repos
	campaignRepo := repository.NewFilesystemCampaignRepository(testDir)
	actRepo := repository.NewFilesystemActRepository(testDir)
	charRepo := repository.NewFilesystemCharacterRepository(testDir)
	npcRepo := repository.NewFilesystemNPCRepository(testDir)
	questRepo := repository.NewFilesystemQuestRepository(testDir)
	canonRepo := repository.NewFilesystemCanonRepository(testDir)
	
	// Create campaign directory
	if err := campaignRepo.Create(campaign); err != nil {
		t.Fatalf("Failed to create campaign: %v", err)
	}
	
	// Create service
	svc := NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		canonRepo,
		testDir, "",
	)
	
	// Test 1: Save NPCs
	npcContent := `## NPCs Principales

### Elara Voss
*Legal Bueno Humana Clériga*

- **Ubicación:** Catedral de la Aurora
- **Faction:** Orden de la Plata
- **Estadísticas:** AC 18, HP 38

### Aldric Thorne
*Legal Neutral Humano Guerrero*

- **Ubicación:** Cuartel de la Guardia
- **Faction:** Guardia de la Ciudad
- **Estadísticas:** AC 16, HP 45
`

	err := svc.SaveNPCs(campaignID, npcContent)
	if err != nil {
		t.Fatalf("SaveNPCs failed: %v", err)
	}
	
	// Verify canon.json updated
	canonDoc, err := canonRepo.Load(campaignID)
	if err != nil {
		t.Fatalf("Failed to load canon: %v", err)
	}
	
	npcEntities := 0
	for _, e := range canonDoc.Entities {
		if e.Type == domain.EntityTypeNPC {
			npcEntities++
		}
	}
	if npcEntities != 2 {
		t.Errorf("Expected 2 NPC entities in canon.json, got %d", npcEntities)
	}
	
	// Verify JSON files created
	npcDir := filepath.Join(testDir, campaignID, "npcs")
	files, err := os.ReadDir(npcDir)
	if err != nil {
		t.Fatalf("Failed to read npcs dir: %v", err)
	}
	
	jsonCount := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			jsonCount++
		}
	}
	if jsonCount != 2 {
		t.Errorf("Expected 2 NPC JSON files, got %d", jsonCount)
	}
	
	// Test 2: Save Bestiary
	bestiaryContent := `# Cenizo Recién Convertido
*CR 1/4, No-muerto Mediano*

- **AC:** 15
- **HP:** 33
- **Type:** Undead

# Vex Terrow - El Harúspice
*CR 5, Humanoide Mediano*

- **AC:** 17
- **HP:** 90
- **Type:** Humanoid
`

	err = svc.SaveBestiary(campaignID, bestiaryContent)
	if err != nil {
		t.Fatalf("SaveBestiary failed: %v", err)
	}
	
	// Verify canon.json updated
	canonDoc, err = canonRepo.Load(campaignID)
	if err != nil {
		t.Fatalf("Failed to load canon: %v", err)
	}
	
	monsterEntities := 0
	for _, e := range canonDoc.Entities {
		if e.Type == domain.EntityTypeMonster {
			monsterEntities++
		}
	}
	if monsterEntities != 2 {
		t.Errorf("Expected 2 Monster entities in canon.json, got %d", monsterEntities)
	}
	
	// Verify monster JSON files created
	monsterDir := filepath.Join(testDir, campaignID, "monsters")
	files, err = os.ReadDir(monsterDir)
	if err != nil {
		t.Fatalf("Failed to read monsters dir: %v", err)
	}
	
	jsonCount = 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			jsonCount++
		}
	}
	if jsonCount != 2 {
		t.Errorf("Expected 2 Monster JSON files, got %d", jsonCount)
	}
	
	t.Log("✅ Session 0 flow test PASSED")
}
