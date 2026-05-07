package services

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestAdventureRosterService_BuildRoster(t *testing.T) {
	ctx := context.Background()

	// Create a temp campaign directory
	baseDir := t.TempDir()
	campaignID := "test-roster-campaign"
	campaignDir := filepath.Join(baseDir, campaignID)
	_ = os.MkdirAll(campaignDir, 0755)

	// Create subdirectories and markdown files
	actsDir := filepath.Join(campaignDir, "acts")
	_ = os.MkdirAll(actsDir, 0755)
	_ = os.WriteFile(filepath.Join(actsDir, "act-01.md"), []byte(`# Acto 1: El Comienzo

## Escena 1: La Taberna

Eldrin el Sabio les pide ayuda.

### NPCs presentes
- **Eldrin el Sabio** — Mago anciano
- **Thorn el Valiente** — Guerrero mercenario

## Escena 2: El Bosque

Encuentran pistas del artefacto.
`), 0644)

	_ = os.WriteFile(filepath.Join(actsDir, "act-02.md"), []byte(`# Acto 2: La Búsqueda

## La Mazmorra

### Monstruos
- **Goblin** (CR 1/4)
- **Orco Jefe** (CR 2)

### Encuentros
- Emboscada en el bosque
- Defensa del campamento
`), 0644)

	npcsDir := filepath.Join(campaignDir, "npcs")
	_ = os.MkdirAll(npcsDir, 0755)
	_ = os.WriteFile(filepath.Join(npcsDir, "npcs_and_factions.md"), []byte(`# NPCs y Facciones

## Eldrin el Sabio
Mago anciano que guía a los aventureros.

## Thorn el Valiente
Guerrero que busca redención.
`), 0644)

	bestiaryDir := filepath.Join(campaignDir, "bestiary")
	_ = os.MkdirAll(bestiaryDir, 0755)
	_ = os.WriteFile(filepath.Join(bestiaryDir, "bestiary.md"), []byte(`# Bestiario

## Goblin
*Small humanoid*
- **CR:** 1/4

## Orco Jefe
*Large humanoid*
- **CR:** 2
`), 0644)

	encountersDir := filepath.Join(campaignDir, "encounters")
	_ = os.MkdirAll(encountersDir, 0755)
	_ = os.WriteFile(filepath.Join(encountersDir, "encounters.md"), []byte(`# Encuentros

## Emboscada en el bosque
3 goblins acechan desde los árboles.

## Defensa del campamento
Los orcos atacan al amanecer.
`), 0644)

	campaignRepo := repository.NewMemoryCampaignRepository()
	_ = campaignRepo.Create(&domain.Campaign{Name: campaignID, Title: "Test Campaign", Status: "active"})

	svc := NewAdventureRosterService(campaignRepo, baseDir)

	t.Run("full roster with all entities", func(t *testing.T) {
		roster, err := svc.BuildRoster(ctx, campaignID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if roster == nil {
			t.Fatalf("expected roster, got nil")
		}
		if roster.CampaignID != campaignID {
			t.Fatalf("campaign_id = %q, want %q", roster.CampaignID, campaignID)
		}

		// Should have NPCs
		if len(roster.NPCs) == 0 {
			t.Fatalf("expected NPCs in roster")
		}
		foundEldrin := false
		for _, npc := range roster.NPCs {
			if npc.Name == "Eldrin el Sabio" {
				foundEldrin = true
				if npc.Act == "" {
					t.Fatalf("expected NPC to have act reference")
				}
			}
		}
		if !foundEldrin {
			t.Fatalf("expected Eldrin in NPC roster, got %v", roster.NPCs)
		}

		// Should have monsters
		if len(roster.Monsters) == 0 {
			t.Fatalf("expected monsters in roster")
		}
		foundGoblin := false
		for _, m := range roster.Monsters {
			if m.Name == "Goblin" {
				foundGoblin = true
			}
		}
		if !foundGoblin {
			t.Fatalf("expected Goblin in monster roster, got %v", roster.Monsters)
		}

		// Should have encounters
		if len(roster.Encounters) == 0 {
			t.Fatalf("expected encounters in roster")
		}
	})

	t.Run("missing campaign returns error", func(t *testing.T) {
		_, err := svc.BuildRoster(ctx, "nonexistent")
		if err == nil {
			t.Fatalf("expected error for missing campaign")
		}
	})

	t.Run("empty directories return partial roster", func(t *testing.T) {
		emptyDir := t.TempDir()
		emptyCampaign := "empty-campaign"
		emptyCampaignDir := filepath.Join(emptyDir, emptyCampaign)
		_ = os.MkdirAll(emptyCampaignDir, 0755)
		// Create empty subdirs
		_ = os.MkdirAll(filepath.Join(emptyCampaignDir, "acts"), 0755)
		_ = os.MkdirAll(filepath.Join(emptyCampaignDir, "npcs"), 0755)
		_ = os.MkdirAll(filepath.Join(emptyCampaignDir, "bestiary"), 0755)
		_ = os.MkdirAll(filepath.Join(emptyCampaignDir, "encounters"), 0755)

		_ = campaignRepo.Create(&domain.Campaign{Name: emptyCampaign, Title: "Empty", Status: "active"})
		emptySvc := NewAdventureRosterService(campaignRepo, emptyDir)

		roster, err := emptySvc.BuildRoster(ctx, emptyCampaign)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if roster == nil {
			t.Fatalf("expected roster, got nil")
		}
		// Should be empty but valid
		if len(roster.NPCs) != 0 {
			t.Fatalf("expected no NPCs for empty campaign")
		}
	})
}
