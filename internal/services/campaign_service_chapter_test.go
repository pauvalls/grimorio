package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

func TestCampaignService_SaveChapter(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	tmpDir := t.TempDir()

	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, tmpDir, "")

	// Create a campaign first
	_, err := service.CreateCampaign("chapter-test", "Chapter Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	chapterContent := `# Capítulo 1: El Comienzo

## Apertura Narrativa

Los héroes llegan al pueblo.

## NPCs en este Capítulo

### Aldeano Mayor
*Neutral humano*

Un viejo líder del pueblo.

## Encuentros

### Encuentro 1: Emboscada
*Dificultad: Medium*

Los bandidos atacan.

## Áreas

### Área 1: Entrada del Pueblo

> Los jugadores ven el pueblo.

La entrada está custodiada.

### Área 2: La Taberna

> Humo y risas.

Un lugar acogedor.

## Consecuencias y Transición

Fin.
`

	err = service.SaveChapter("chapter-test", 1, "El Comienzo", chapterContent)
	if err != nil {
		t.Fatalf("SaveChapter() unexpected error: %v", err)
	}

	// Verify file was created
	chapterPath := filepath.Join(tmpDir, "chapter-test", "chapters", "chapter_01.md")
	if _, err := os.Stat(chapterPath); os.IsNotExist(err) {
		t.Errorf("Chapter file not created at %s", chapterPath)
	}

	// Verify content
	content, err := os.ReadFile(chapterPath)
	if err != nil {
		t.Fatalf("Failed to read chapter file: %v", err)
	}
	if string(content) != chapterContent {
		t.Error("Chapter file content does not match")
	}
}

func TestCampaignService_SaveChapter_InvalidCampaign(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	tmpDir := t.TempDir()

	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, tmpDir, "")

	err := service.SaveChapter("nonexistent", 1, "Title", "content")
	if err == nil {
		t.Fatal("SaveChapter() expected error for nonexistent campaign")
	}
}

func TestCampaignService_SaveChapter_NoAreas(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	tmpDir := t.TempDir()

	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, tmpDir, "")

	// Create a campaign first
	_, err := service.CreateCampaign("no-areas-test", "No Areas Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	// Content without areas
	badContent := `# Capítulo 1: El Comienzo

Sin áreas aquí.
`

	err = service.SaveChapter("no-areas-test", 1, "El Comienzo", badContent)
	if err == nil {
		t.Fatal("SaveChapter() expected error for content without areas")
	}
}

func TestCampaignService_SaveChapter_MultipleChapters(t *testing.T) {
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()
	tmpDir := t.TempDir()

	service := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, tmpDir, "")

	// Create a campaign first
	_, err := service.CreateCampaign("multi-chapter-test", "Multi Chapter Test", "Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	for i := 1; i <= 3; i++ {
		content := fmt.Sprintf("# Capítulo %d\n\n## Áreas\n\n### Área 1: Lugar\n\nDescripción.\n", i)
		err := service.SaveChapter("multi-chapter-test", i, fmt.Sprintf("Capítulo %d", i), content)
		if err != nil {
			t.Fatalf("SaveChapter() chapter %d error: %v", i, err)
		}
	}

	// Verify all 3 files exist
	for i := 1; i <= 3; i++ {
		chapterPath := filepath.Join(tmpDir, "multi-chapter-test", "chapters", fmt.Sprintf("chapter_%02d.md", i))
		if _, err := os.Stat(chapterPath); os.IsNotExist(err) {
			t.Errorf("Chapter %d file not created", i)
		}
	}
}
