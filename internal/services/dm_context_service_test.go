package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
	fsrepo "github.com/pauvalls/grimorio/internal/repository/fs"

	"github.com/stretchr/testify/assert"
)

func TestDMContextService_GetContext(t *testing.T) {
	ctx := context.Background()

	// Create memory repositories
	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()

	// Create filesystem repos for V3 types (using temp dir)
	tmpDir := t.TempDir()
	monsterRepo := fsrepo.NewFilesystemMonsterRepository(tmpDir)
	areaRepo := fsrepo.NewFilesystemAreaRepositoryV3(tmpDir)

	sessionPrepSvc := NewSessionPrepService(canonRepo, stateRepo, nil)

	svc := NewDMContextService(
		canonRepo, stateRepo, charRepo, npcRepo, questRepo,
		monsterRepo, areaRepo, factionRepo, sessionPrepSvc, tmpDir,
	)

	// Seed canon document
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "test-campaign",
		Facts: []domain.CanonFact{
			{ID: "f1", Category: "lore", Statement: "Magic is banned in Thornvale", Source: "canon"},
		},
		Entities: []domain.CanonEntity{
			{ID: "npc-1", Name: "Eldrin", Type: domain.EntityTypeNPC, CanonState: domain.EntityStateAlive, Motivation: "proteger la torre", Properties: map[string]any{"dialogue_voice": "Habla en susurros"}},
			{ID: "faction-1", Name: "Thieves' Guild", Type: domain.EntityTypeFaction, Properties: map[string]any{"attitude": "hostile"}},
			{ID: "monster-1", Name: "Goblin", Type: domain.EntityTypeMonster, Properties: map[string]any{
				"descriptive_cues": map[string]any{
					"full_hp":  "El goblin está alerta y listo para pelear.",
					"half_hp":  "El goblin sangra pero sigue en pie.",
					"low_hp":   "El goblin se tambalea, apenas consciente.",
					"defeated": "El goblin cae al suelo, inmóvil.",
				},
			}},
		},
	}
	_ = canonRepo.Save("test-campaign", canonDoc)

	// Seed narrative state
	state := &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "test-campaign",
		CurrentSession: 2,
		SessionLog: []domain.SessionRecord{
			{SessionNum: 1, Date: time.Now(), Summary: "Started the journey."},
		},
		ActiveQuests: []domain.QuestState{
			{ID: "q1", Name: "Main Quest", Status: "active", SourceAct: "act-1", GiverNPC: "npc-1"},
		},
	}
	_ = stateRepo.Save("test-campaign", state)

	// Seed characters
	_ = charRepo.Save(&domain.Character{
		CampaignID: "test-campaign",
		Name:       "Aric", Race: "humano", Class: "guerrero", Level: 3,
		Background: "soldado", Alignment: "LG",
		HP:    domain.HP{Current: 25, Maximum: 30},
		AC:    16,
		Stats: domain.Stats{STR: 16, DEX: 12, CON: 14, INT: 10, WIS: 13, CHA: 8},
	})

	// Seed NPCs
	_ = npcRepo.Save(&domain.NPC{
		CampaignID:  "test-campaign",
		Name:        "Eldrin",
		Description: "Un mago anciano",
		Faction:     "Thieves' Guild",
		Stats:       &domain.StatBlock{HP: 20, AC: 12},
	})

	// Seed quests
	_ = questRepo.Save(&domain.Quest{
		CampaignID:  "test-campaign",
		ID:          "q1",
		Title:       "Main Quest",
		Status:      domain.QuestStatusActive,
		Type:        domain.QuestTypeMain,
		RelatedNPCs: []string{"Eldrin"},
	})

	// Seed factions
	_ = factionRepo.Save("test-campaign", &domain.FactionReputationMatrix{
		CampaignID: "test-campaign",
		Entries: []domain.ReputationEntry{
			{FactionID: "faction-1", PartyID: "party-1", Score: -30, Status: "hostile"},
		},
	})

	// Seed monsters
	_ = monsterRepo.Save(ctx, "test-campaign", &domain.Monster{
		CampaignID: "test-campaign",
		Name:       "Goblin",
		CR:         "1/4",
		Stats:      domain.StatBlock{HP: 7, AC: 15},
	})

	// Seed areas
	_ = areaRepo.Create(ctx, "test-campaign", &domain.Area{
		ID:         "area_1_1",
		ChapterID:  "chapter_1",
		AreaNumber: 1,
		Title:      "Entrada",
		Summary:    "La entrada a la mazmorra",
		LevelRange: domain.LevelRange{Min: 1, Max: 3},
		Features:   []domain.AreaFeature{{Type: "room", Name: "Entrada", Description: "Una puerta de piedra"}},
		Encounters: []domain.AreaEncounter{{EncounterID: "enc-1", Trigger: "entrar", CRTotal: 0.25, XPValue: 50}},
	})

	// Seed prologue
	_ = os.MkdirAll(filepath.Join(tmpDir, "test-campaign"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "test-campaign", "prologue.md"), []byte("# Prólogo\n\nEl gancho narrativo."), 0644)

	t.Run("complete campaign returns all sections", func(t *testing.T) {
		payload, warnings, err := svc.GetContext(ctx, "test-campaign", 0, true, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		assert.Equal(t, "test-campaign", payload.CampaignID)
		assert.Equal(t, 2, payload.SessionNum)
		assert.NotNil(t, payload.Canon)
		assert.Len(t, payload.Canon.Entities, 3)
		assert.NotNil(t, payload.NarrativeState)
		assert.Equal(t, 2, payload.NarrativeState.CurrentSession)
		assert.NotNil(t, payload.SessionPrep)
		assert.Len(t, payload.Characters, 1)
		assert.Equal(t, "Aric", payload.Characters[0].Name)
		assert.Len(t, payload.NPCs, 1)
		assert.Equal(t, "Eldrin", payload.NPCs["Eldrin"].Name)
		assert.Equal(t, "Habla en susurros", payload.NPCs["Eldrin"].DialogueVoice)
		assert.Len(t, payload.Bestiary, 1)
		assert.Equal(t, "Goblin", payload.Bestiary["Goblin"].Name)
		assert.NotNil(t, payload.Bestiary["Goblin"].DescriptiveCues)
		assert.True(t, payload.Bestiary["Goblin"].HasAllDescriptiveCues())
		assert.Len(t, payload.Areas, 1)
		assert.Equal(t, "Entrada", payload.Areas["area_1_1"].Title)
		assert.Len(t, payload.Factions, 1)
		assert.Equal(t, "Thieves' Guild", payload.Factions["faction-1"].Name)
		assert.Equal(t, int8(-30), payload.Factions["faction-1"].Reputation)
		assert.Len(t, payload.Quests, 1)
		assert.Equal(t, "Main Quest", payload.Quests[0].Title)

		// Ensure no nil slices
		assert.NotNil(t, payload.Characters)
		assert.NotNil(t, payload.Areas)
		assert.NotNil(t, payload.NPCs)
		assert.NotNil(t, payload.Bestiary)
		assert.NotNil(t, payload.Factions)
		assert.NotNil(t, payload.Quests)
		assert.NotNil(t, payload.DMNotes.Warnings)
		assert.NotNil(t, payload.DMNotes.Reminders)

		// Warnings should be minimal for a complete campaign
		assert.Empty(t, warnings)
	})

	t.Run("default session_num from narrative state", func(t *testing.T) {
		payload, _, err := svc.GetContext(ctx, "test-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)
		assert.Equal(t, 2, payload.SessionNum)
	})

	t.Run("missing canon returns error", func(t *testing.T) {
		_, _, err := svc.GetContext(ctx, "nonexistent", 1, false, false, false, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "canon document not found")
	})

	t.Run("minimal campaign with empty data", func(t *testing.T) {
		// Create a minimal campaign with only canon and state
		_ = canonRepo.Save("minimal-campaign", &domain.CanonDocument{
			SchemaVersion: domain.SchemaVersionV2,
			CampaignID:    "minimal-campaign",
		})
		_ = stateRepo.Save("minimal-campaign", &domain.NarrativeState{
			SchemaVersion:  domain.SchemaVersionV2,
			CampaignID:     "minimal-campaign",
			CurrentSession: 1,
		})

		payload, warnings, err := svc.GetContext(ctx, "minimal-campaign", 1, false, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		// All optional sections should be empty, not nil
		assert.Empty(t, payload.Characters)
		assert.NotNil(t, payload.Characters)
		assert.Empty(t, payload.Areas)
		assert.NotNil(t, payload.Areas)
		assert.Empty(t, payload.NPCs)
		assert.NotNil(t, payload.NPCs)
		assert.Empty(t, payload.Bestiary)
		assert.NotNil(t, payload.Bestiary)
		assert.Empty(t, payload.Factions)
		assert.NotNil(t, payload.Factions)
		assert.Empty(t, payload.Quests)
		assert.NotNil(t, payload.Quests)

		// Warnings for missing optional data
		assert.NotEmpty(t, warnings)
	})

	t.Run("payload under size cap", func(t *testing.T) {
		payload, _, err := svc.GetContext(ctx, "test-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)

		jsonBytes, err := json.Marshal(payload)
		assert.NoError(t, err)
		assert.Less(t, len(jsonBytes), 100*1024, "payload should be under 100KB")
	})

	t.Run("descriptive cues populated when bestiary has them", func(t *testing.T) {
		payload, _, err := svc.GetContext(ctx, "test-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)

		goblin, ok := payload.Bestiary["Goblin"]
		assert.True(t, ok)
		assert.NotNil(t, goblin.DescriptiveCues)
		assert.Contains(t, goblin.DescriptiveCues, "full_hp")
		assert.Contains(t, goblin.DescriptiveCues, "half_hp")
		assert.Contains(t, goblin.DescriptiveCues, "low_hp")
		assert.Contains(t, goblin.DescriptiveCues, "defeated")
	})
}

func TestDMContextService_GetContext_LargeCampaign(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	canonRepo := repository.NewMemoryCanonRepository()
	stateRepo := repository.NewMemoryNarrativeStateRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	factionRepo := repository.NewMemoryFactionRepository()
	monsterRepo := fsrepo.NewFilesystemMonsterRepository(tmpDir)
	areaRepo := fsrepo.NewFilesystemAreaRepositoryV3(tmpDir)

	sessionPrepSvc := NewSessionPrepService(canonRepo, stateRepo, nil)
	svc := NewDMContextService(
		canonRepo, stateRepo, charRepo, npcRepo, questRepo,
		monsterRepo, areaRepo, factionRepo, sessionPrepSvc, tmpDir,
	)

	// Seed canon and state
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "large-campaign",
	}
	_ = canonRepo.Save("large-campaign", canonDoc)
	_ = stateRepo.Save("large-campaign", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "large-campaign",
		CurrentSession: 1,
	})

	// Create 20 areas, 30 NPCs, 15 monsters
	for i := 1; i <= 20; i++ {
		_ = areaRepo.Create(ctx, "large-campaign", &domain.Area{
			ID:         fmt.Sprintf("area_%d", i),
			ChapterID:  fmt.Sprintf("chapter_%d", (i-1)/5+1),
			AreaNumber: ((i - 1) % 5) + 1,
			Title:      fmt.Sprintf("Area %d", i),
			Summary:    fmt.Sprintf("Summary for area %d with some descriptive text", i),
			LevelRange: domain.LevelRange{Min: 1, Max: 5},
			Features:   []domain.AreaFeature{{Type: "room", Name: "Room", Description: "A room"}},
			Encounters: []domain.AreaEncounter{{EncounterID: fmt.Sprintf("enc-%d", i), Trigger: "enter", CRTotal: 1, XPValue: 100}},
		})
	}

	for i := 1; i <= 30; i++ {
		_ = npcRepo.Save(&domain.NPC{
			CampaignID:  "large-campaign",
			Name:        fmt.Sprintf("NPC %d", i),
			Description: fmt.Sprintf("Description for NPC %d", i),
		})
	}

	for i := 1; i <= 15; i++ {
		_ = monsterRepo.Save(ctx, "large-campaign", &domain.Monster{
			CampaignID: "large-campaign",
			Name:       fmt.Sprintf("Monster %d", i),
			CR:         "1",
			Stats:      domain.StatBlock{HP: 20, AC: 13},
		})
	}

	t.Run("20-area campaign stays under 100KB", func(t *testing.T) {
		payload, warnings, err := svc.GetContext(ctx, "large-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)
		assert.Len(t, payload.Areas, 20)
		assert.Len(t, payload.NPCs, 30)
		assert.Len(t, payload.Bestiary, 15)

		jsonBytes, err := json.Marshal(payload)
		assert.NoError(t, err)
		assert.Less(t, len(jsonBytes), 100*1024, "payload should be under 100KB for 20-area campaign")

		// Should not have size cap warning
		for _, w := range warnings {
			assert.NotContains(t, w, "exceeds 100KB")
		}
	})
}

