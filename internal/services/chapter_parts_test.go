package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

func setupChapterPartsTest(t *testing.T) (*CampaignService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	canonRepo := repository.NewMemoryCanonRepository()
	monsterRepo := repository.NewMemoryMonsterRepository()

	svc := NewCampaignService(campaignRepo, actRepo, charRepo, npcRepo, questRepo, canonRepo, monsterRepo, tmpDir, "")

	// Create a test campaign
	_, err := svc.CreateCampaign("test-campaign", "Test Campaign", "Setting", "")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	return svc, tmpDir
}

func TestSaveChapterPart_CreatesDraftDir(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	result, err := svc.SaveChapterPart("test-campaign", 1, "opener", "# Chapter 1\n\nSome opener content.")
	if err != nil {
		t.Fatalf("SaveChapterPart() error: %v", err)
	}

	if result.Status != "ok" {
		t.Errorf("Status = %q, want %q", result.Status, "ok")
	}
	if result.PartsSaved != 1 {
		t.Errorf("PartsSaved = %d, want 1", result.PartsSaved)
	}
	if len(result.PartsReceived) != 1 || result.PartsReceived[0] != "opener" {
		t.Errorf("PartsReceived = %v, want [opener]", result.PartsReceived)
	}
}

func TestSaveChapterPart_AppendsParts(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	// Save first part
	_, err := svc.SaveChapterPart("test-campaign", 1, "opener", "# Chapter 1\n\nOpener.")
	if err != nil {
		t.Fatalf("SaveChapterPart(opener) error: %v", err)
	}

	// Save second part
	result, err := svc.SaveChapterPart("test-campaign", 1, "npcs", "## NPCs\n\n### Test NPC\n")
	if err != nil {
		t.Fatalf("SaveChapterPart(npcs) error: %v", err)
	}

	if result.PartsSaved != 2 {
		t.Errorf("PartsSaved = %d, want 2", result.PartsSaved)
	}
	if len(result.PartsReceived) != 2 {
		t.Errorf("PartsReceived = %v, want 2 parts", result.PartsReceived)
	}
}

func TestSaveChapterPart_InvalidPartName(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	_, err := svc.SaveChapterPart("test-campaign", 1, "invalid_part", "content")
	if err == nil {
		t.Fatal("expected error for invalid part name")
	}
	if !strings.Contains(err.Error(), "invalid part name") {
		t.Errorf("error = %q, want to contain 'invalid part name'", err.Error())
	}
}

func TestSaveChapterPart_GeneralFeatures(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	content := "## General Features\n\n***Ceilings.*** 30 feet high.\n***Light.*** Dim torchlight."
	result, err := svc.SaveChapterPart("test-campaign", 1, "general-features", content)
	if err != nil {
		t.Fatalf("SaveChapterPart(general-features) error: %v", err)
	}
	if result.Status != "ok" {
		t.Errorf("Status = %q, want %q", result.Status, "ok")
	}

	// Verify the file was written with correct order prefix (02)
	draftDir := filepath.Join(result.DraftPath)
	entries, err := os.ReadDir(draftDir)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}
	found := false
	for _, e := range entries {
		if strings.Contains(e.Name(), "general-features") {
			found = true
			if !strings.HasPrefix(e.Name(), "02-") {
				t.Errorf("general-features file = %q, want prefix '02-'", e.Name())
			}
		}
	}
	if !found {
		t.Error("general-features part file not found in draft directory")
	}
}

func TestSaveChapterPart_InvalidCampaign(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	_, err := svc.SaveChapterPart("nonexistent-campaign", 1, "opener", "content")
	if err == nil {
		t.Fatal("expected error for nonexistent campaign")
	}
}

func TestFinalizeChapter_NoDraft(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	_, err := svc.FinalizeChapter("test-campaign", 1, "Test Chapter")
	if err == nil {
		t.Fatal("expected error when no draft exists")
	}
	if !strings.Contains(err.Error(), "no draft found") {
		t.Errorf("error = %q, want to contain 'no draft found'", err.Error())
	}
}

func TestFinalizeChapter_ValidatesContent(t *testing.T) {
	svc, _ := setupChapterPartsTest(t)

	// Save a single part with insufficient content (too few areas)
	_, err := svc.SaveChapterPart("test-campaign", 1, "opener", "# Chapter 1\n\nJust a small opener.")
	if err != nil {
		t.Fatalf("SaveChapterPart() error: %v", err)
	}

	// Finalize should fail validation (too few areas)
	_, err = svc.FinalizeChapter("test-campaign", 1, "Test Chapter")
	if err == nil {
		t.Fatal("expected validation error for chapter with too few areas")
	}
	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("error = %q, want to contain 'validation failed'", err.Error())
	}

	// Draft should still exist (not cleaned up on failure)
	draftDir := filepath.Join(svc.baseDir, "test-campaign", "chapters", ".draft", "chapter_01")
	if _, err := os.Stat(draftDir); os.IsNotExist(err) {
		t.Error("draft directory should still exist after validation failure")
	}
}

func TestFinalizeChapter_Success(t *testing.T) {
	tests := []struct {
		name       string
		chapterNum int
		title      string
		areaCount  int
	}{
		{
			name:       "chapter 1 with 8 areas",
			chapterNum: 1,
			title:      "The Dark Cavern",
			areaCount:  8,
		},
		{
			name:       "chapter 2 with 10 areas",
			chapterNum: 2,
			title:      "The Forgotten Temple",
			areaCount:  10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, tmpDir := setupChapterPartsTest(t)

			opener := "The heroes arrive at the dungeon entrance. The air is cold and damp.\n"
			npcs := "## NPCs in this Chapter\n\n### Guard Captain\n*Neutral good human*\n\nA seasoned warrior protecting the entrance.\n"
			encounters := "## Encounters\n\n### Encounter 1: Ambush\n*Difficulty: Medium*\n\nBandits attack from the shadows.\n\n**Monsters:**\n- 3 **Bandit**\n\n**Rewards:**\n- 100 XP\n"

			areasPart1 := generateTestAreas(1, tt.areaCount/2, tt.areaCount)
			areasPart2 := generateTestAreas(tt.areaCount/2+1, tt.areaCount, tt.areaCount)
			closing := "## Consequences\n\nThe dungeon is cleared. The party finds a map leading to the next location.\n"

			parts := []struct {
				partName string
				content  string
			}{
				{"opener", opener},
				{"npcs", npcs},
				{"encounters", encounters},
				{"areas-1", areasPart1},
				{"areas-2", areasPart2},
				{"closing", closing},
			}

			for _, p := range parts {
				_, err := svc.SaveChapterPart("test-campaign", tt.chapterNum, p.partName, p.content)
				if err != nil {
					t.Fatalf("SaveChapterPart(%s) error: %v", p.partName, err)
				}
			}

			result, err := svc.FinalizeChapter("test-campaign", tt.chapterNum, tt.title)
			if err != nil {
				t.Fatalf("FinalizeChapter() error: %v", err)
			}

			if result.Status != "ok" {
				t.Errorf("Status = %q, want %q", result.Status, "ok")
			}

			expectedFilename := fmt.Sprintf("chapter_%02d.md", tt.chapterNum)
			if result.Chapter != expectedFilename {
				t.Errorf("Chapter = %q, want %q", result.Chapter, expectedFilename)
			}

			if result.Areas != tt.areaCount {
				t.Errorf("Areas = %d, want %d", result.Areas, tt.areaCount)
			}

			if result.NPCs < 1 {
				t.Errorf("NPCs = %d, want >= 1", result.NPCs)
			}

			if result.Encounters < 1 {
				t.Errorf("Encounters = %d, want >= 1", result.Encounters)
			}

			if result.WordCount == 0 {
				t.Error("WordCount = 0, want > 0")
			}

			finalPath := filepath.Join(tmpDir, "test-campaign", "chapters", expectedFilename)
			data, err := os.ReadFile(finalPath)
			if err != nil {
				t.Fatalf("failed to read final chapter file: %v", err)
			}

			if !strings.Contains(string(data), tt.title) {
				t.Error("final chapter file does not contain the chapter title")
			}

			draftDir := filepath.Join(tmpDir, "test-campaign", "chapters", ".draft", fmt.Sprintf("chapter_%02d", tt.chapterNum))
			if _, err := os.Stat(draftDir); !os.IsNotExist(err) {
				t.Error("draft directory should be cleaned up after successful finalize")
			}
		})
	}
}

func generateTestAreas(start, end, total int) string {
	var sb strings.Builder
	sb.WriteString("## Areas\n\n")

	filler := strings.Repeat("word ", 100)

	for i := start; i <= end; i++ {
		next := i + 1
		prev := i - 1
		if next > total {
			next = 1
		}
		if prev < 1 {
			prev = total
		}

		fmt.Fprintf(&sb, "### Area %d: Room %d\n\n", i, i)
		fmt.Fprintf(&sb, "> **Read-Aloud Text:** *%s*\n\n", filler)
		fmt.Fprintf(&sb, "**Description:** %s\n\n", filler)
		sb.WriteString("**Creatures:**\n- 1 **Goblin**\n\n")
		sb.WriteString("**Treasure:**\n- **XP:** 50 XP\n- **Coin:** 10 gp\n\n")
		fmt.Fprintf(&sb, "**Connections:**\n- → Area %d\n- ← Area %d\n\n", next, prev)
		sb.WriteString("**Secrets and Traps:**\n- **Detect:** Perception DC 12\n\n")
		sb.WriteString("**Development:**\n- If they enter combat: they attack.\n\n")
	}

	return sb.String()
}
