package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestDMContextService_loadAreasFromChapters(t *testing.T) {
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

	campaignDir := filepath.Join(tmpDir, "chapter-campaign")
	chaptersDir := filepath.Join(campaignDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)

	// Write chapter_01.md with 2 areas, NPCs, encounters
	chapterContent := `# Capítulo 1: El Comienzo

## NPCs en este Capítulo

### Aldeano Mayor
*Neutral humano*

Un viejo líder del pueblo.

**Estadísticas:** AC 12, HP 15

## Encuentros

### Encuentro 1: Emboscada
*Dificultad: Medium*

**Monstruos:**
- 3x Bandido
- 1x Líder Bandido

## Áreas

### Área 1: Entrada del Pueblo

> Los jugadores ven el pueblo desde la colina.

La entrada está custodiada por guardias.

**Características:**
- **Muralla:** Madera reforzada

### Área 2: La Taberna

> Humo y risas salen de la taberna.

Un lugar acogedor para descansar.
`
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(chapterContent), 0644)

	t.Run("parses chapter markdown into correct AreaContext fields", func(t *testing.T) {
		areas, npcs, monsters, err := svc.loadAreasFromChapters("chapter-campaign")
		assert.NoError(t, err)
		assert.Len(t, areas, 2, "should extract 2 areas")
		assert.Len(t, npcs, 1, "should extract 1 NPC")
		assert.Len(t, monsters, 2, "should extract 2 unique monster names from encounters")

		// Verify first area
		if len(areas) > 0 {
			area1 := areas[0]
			assert.Equal(t, "chapter_01-area-1", area1.ID)
			assert.Equal(t, "chapter_01", area1.ChapterID)
			assert.Equal(t, 1, area1.AreaNumber)
			assert.Equal(t, "Entrada del Pueblo", area1.Title)
			assert.Contains(t, area1.Summary, "La entrada está custodiada")
			assert.Contains(t, area1.PlayerReadAloud, "Los jugadores ven el pueblo")
		}

		// Verify second area
		if len(areas) > 1 {
			area2 := areas[1]
			assert.Equal(t, "chapter_01-area-2", area2.ID)
			assert.Equal(t, "chapter_01", area2.ChapterID)
			assert.Equal(t, 2, area2.AreaNumber)
			assert.Equal(t, "La Taberna", area2.Title)
		}
	})

	t.Run("returns empty for non-existent chapters dir", func(t *testing.T) {
		areas, npcs, monsters, err := svc.loadAreasFromChapters("nonexistent")
		assert.NoError(t, err)
		assert.Empty(t, areas)
		assert.Empty(t, npcs)
		assert.Empty(t, monsters)
	})

	t.Run("handles multiple chapter files", func(t *testing.T) {
		// Add chapter_02.md
		chapter2Content := `# Capítulo 2: La Cueva

## Áreas

### Área 1: Entrada de la Cueva

> Un túnel oscuro se abre ante los PJs.

La cueva está húmeda.
`
		_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_02.md"), []byte(chapter2Content), 0644)

		areas, _, _, err := svc.loadAreasFromChapters("chapter-campaign")
		assert.NoError(t, err)
		assert.Len(t, areas, 3, "should have 3 areas total (2 from ch1 + 1 from ch2)")

		// Verify chapter 2 area
		var caveArea *domain.AreaContext
		for i := range areas {
			if areas[i].ChapterID == "chapter_02" {
				caveArea = &areas[i]
				break
			}
		}
		assert.NotNil(t, caveArea, "should find chapter_02 area")
		assert.Equal(t, "chapter_02-area-1", caveArea.ID)
		assert.Equal(t, "Entrada de la Cueva", caveArea.Title)
		assert.Contains(t, caveArea.PlayerReadAloud, "túnel oscuro")
	})
}

func TestDMContextService_loadCampaignAreas(t *testing.T) {
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

	t.Run("prefers chapters when both dirs exist", func(t *testing.T) {
		campaignDir := filepath.Join(tmpDir, "both-campaign")
		chaptersDir := filepath.Join(campaignDir, "chapters")
		areasDir := filepath.Join(campaignDir, "areas")
		_ = os.MkdirAll(chaptersDir, 0755)
		_ = os.MkdirAll(areasDir, 0755)

		// Write a chapter file
		_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(`# Capítulo 1

## Áreas

### Área 1: Desde Capítulo

> Texto para leer.

Descripción del área desde capítulo.
`), 0644)

		// Write a legacy area file
		_ = os.WriteFile(filepath.Join(areasDir, "chapter_01.md"), []byte(`## Area 1: Desde Areas

Resumen del área desde areas.
`), 0644)

		areas, _, _, warnings, err := svc.loadCampaignAreas("both-campaign")
		assert.NoError(t, err)
		assert.Len(t, areas, 1)
		assert.Equal(t, "Desde Capítulo", areas[0].Title, "should prefer chapters/ content")
		assert.Empty(t, warnings)
	})

	t.Run("falls back to areas when no chapters dir", func(t *testing.T) {
		campaignDir := filepath.Join(tmpDir, "legacy-campaign")
		areasDir := filepath.Join(campaignDir, "areas")
		_ = os.MkdirAll(areasDir, 0755)

		_ = os.WriteFile(filepath.Join(areasDir, "chapter_01.md"), []byte(`## Area 1: Desde Areas Legacy

Resumen del área.
`), 0644)

		areas, _, _, warnings, err := svc.loadCampaignAreas("legacy-campaign")
		assert.NoError(t, err)
		assert.Len(t, areas, 1)
		assert.Equal(t, "Desde Areas Legacy", areas[0].Title, "should fallback to areas/ content")
		assert.Empty(t, warnings)
	})

	t.Run("returns empty when neither dir exists", func(t *testing.T) {
		areas, _, _, warnings, err := svc.loadCampaignAreas("empty-campaign")
		assert.NoError(t, err)
		assert.Empty(t, areas)
		assert.Empty(t, warnings)
	})
}

func TestDMContextService_GetContext_ChapterCampaign(t *testing.T) {
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

	// Seed canon with an NPC entity and a monster entity for enrichment
	canonDoc := &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "chapter-campaign",
		Entities: []domain.CanonEntity{
			{
				ID:         "npc-aldeano",
				Name:       "Aldeano Mayor",
				Type:       domain.EntityTypeNPC,
				CanonState: domain.EntityStateAlive,
				Motivation: "proteger el pueblo",
				Secret:     "conoce la ubicación de la cueva",
				Properties: map[string]any{
					"dialogue_voice":     "Habla con voz cansada",
					"personality_traits": []string{"prudente", "sabio"},
					"tactics":            "Busca ayuda de los guardias",
				},
			},
			{
				ID:   "monster-bandido",
				Name: "Bandido",
				Type: domain.EntityTypeMonster,
				Properties: map[string]any{
					"cr":    "1/4",
					"hp":    11,
					"ac":    12,
					"tactics": "Atacar en grupo desde la cubierta",
					"descriptive_cues": map[string]any{
						"full_hp":  "El bandido te apunta con su arco.",
						"half_hp":  "El bandido sangra pero sigue disparando.",
						"low_hp":   "El bandido intenta huir.",
						"defeated": "El bandido cae al suelo.",
					},
				},
			},
		},
	}
	_ = canonRepo.Save("chapter-campaign", canonDoc)
	_ = stateRepo.Save("chapter-campaign", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "chapter-campaign",
		CurrentSession: 1,
	})

	// Create chapters directory with content
	campaignDir := filepath.Join(tmpDir, "chapter-campaign")
	chaptersDir := filepath.Join(campaignDir, "chapters")
	_ = os.MkdirAll(chaptersDir, 0755)

	chapterContent := `# Capítulo 1: El Comienzo

## NPCs en este Capítulo

### Aldeano Mayor
*Neutral humano*

Un viejo líder del pueblo.

**Estadísticas:** AC 12, HP 15

## Encuentros

### Encuentro 1: Emboscada
*Dificultad: Medium*

**Monstruos:**
- 3x Bandido
- 1x Líder Bandido

## Áreas

### Área 1: Entrada del Pueblo

> Los jugadores ven el pueblo desde la colina.

La entrada está custodiada por guardias.

### Área 2: La Taberna

> Humo y risas salen de la taberna.

Un lugar acogedor.
`
	_ = os.WriteFile(filepath.Join(chaptersDir, "chapter_01.md"), []byte(chapterContent), 0644)

	t.Run("full chapter campaign returns areas from chapters", func(t *testing.T) {
		payload, _, err := svc.GetContext(ctx, "chapter-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		// Should have 2 areas from chapter
		assert.Len(t, payload.Areas, 2, "should load areas from chapters/")

		area1, ok := payload.Areas["chapter_01-area-1"]
		assert.True(t, ok, "should have area chapter_01-area-1")
		assert.Equal(t, "Entrada del Pueblo", area1.Title)
		assert.Equal(t, "chapter_01", area1.ChapterID)
		assert.Contains(t, area1.PlayerReadAloud, "Los jugadores ven el pueblo")

		area2, ok := payload.Areas["chapter_01-area-2"]
		assert.True(t, ok, "should have area chapter_01-area-2")
		assert.Equal(t, "La Taberna", area2.Title)
	})

	t.Run("chapter NPCs merge with canon enrichment", func(t *testing.T) {
		payload, _, err := svc.GetContext(ctx, "chapter-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		// Should have NPC from chapter, enriched from canon
		npc, ok := payload.NPCs["Aldeano Mayor"]
		assert.True(t, ok, "should have Aldeano Mayor in NPCs")
		assert.Equal(t, "Aldeano Mayor", npc.Name)
		assert.Contains(t, npc.Description, "Un viejo líder del pueblo.", "chapter description should be used")
		assert.Equal(t, "proteger el pueblo", npc.Motivation, "canon motivation should enrich")
		assert.Equal(t, "conoce la ubicación de la cueva", npc.Secret, "canon secret should enrich")
		assert.Equal(t, "Habla con voz cansada", npc.DialogueVoice, "canon voice should enrich")
		assert.Equal(t, []string{"prudente", "sabio"}, npc.Personality, "canon personality should enrich")
		assert.Equal(t, "Busca ayuda de los guardias", npc.Tactics, "canon tactics should enrich")
		assert.Equal(t, 12, npc.Stats.AC, "chapter stats AC")
		assert.Equal(t, 15, npc.Stats.HP, "chapter stats HP")
	})

	t.Run("encounter monsters added to bestiary and enriched from canon", func(t *testing.T) {
		payload, warnings, err := svc.GetContext(ctx, "chapter-campaign", 0, false, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)

		// Bandido should be in bestiary from encounter ref, enriched from canon
		bandido, ok := payload.Bestiary["Bandido"]
		assert.True(t, ok, "should have Bandido in bestiary from encounter refs")
		assert.Equal(t, "Bandido", bandido.Name)
		assert.Equal(t, "1/4", bandido.CR, "canon CR should enrich")
		assert.Equal(t, 12, bandido.AC, "canon AC should enrich")
		assert.Equal(t, 11, bandido.HP, "canon HP should enrich")
		assert.Equal(t, "Atacar en grupo desde la cubierta", bandido.Tactics, "canon tactics should enrich")
		assert.NotNil(t, bandido.DescriptiveCues)
		assert.Contains(t, bandido.DescriptiveCues, "full_hp")
		assert.Contains(t, bandido.DescriptiveCues, "defeated")

		// Líder Bandido should be in bestiary as stub (no canon entity)
		lider, ok := payload.Bestiary["Líder Bandido"]
		assert.True(t, ok, "should have Líder Bandido in bestiary")
		assert.Equal(t, "Líder Bandido", lider.Name)
		assert.Empty(t, lider.CR, "no canon entity means empty CR")

		// Should have warning for monster without CR
		var hasCRWarning bool
		for _, w := range warnings {
			if strings.Contains(w, "Líder Bandido") && strings.Contains(w, "CR") {
				hasCRWarning = true
				break
			}
		}
		assert.True(t, hasCRWarning, "should warn about monster without CR")
	})
}

func TestDMContextService_GetContext_ChapterAndLegacyFallback(t *testing.T) {
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

	_ = canonRepo.Save("legacy-only", &domain.CanonDocument{
		SchemaVersion: domain.SchemaVersionV2,
		CampaignID:    "legacy-only",
	})
	_ = stateRepo.Save("legacy-only", &domain.NarrativeState{
		SchemaVersion:  domain.SchemaVersionV2,
		CampaignID:     "legacy-only",
		CurrentSession: 1,
	})

	// Create only areas/ (no chapters/)
	campaignDir := filepath.Join(tmpDir, "legacy-only")
	areasDir := filepath.Join(campaignDir, "areas")
	_ = os.MkdirAll(areasDir, 0755)

	_ = os.WriteFile(filepath.Join(areasDir, "chapter_01.md"), []byte(`## Area 1: Entrada Legacy

Resumen del área legacy.
`), 0644)

	t.Run("legacy areas/ fallback still works", func(t *testing.T) {
		payload, _, err := svc.GetContext(ctx, "legacy-only", 0, false, false, false, 5)
		assert.NoError(t, err)
		assert.NotNil(t, payload)
		assert.Len(t, payload.Areas, 1, "should fallback to areas/ when no chapters/")

		area, ok := payload.Areas["chapter_01-area-1"]
		assert.True(t, ok)
		assert.Equal(t, "Entrada Legacy", area.Title)
	})
}

