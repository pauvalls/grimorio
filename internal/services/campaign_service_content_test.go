package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pauvalls/grimorio/internal/repository"
)

func setupContentTestService(t *testing.T) (*CampaignService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	campaignRepo := repository.NewMemoryCampaignRepository()
	actRepo := repository.NewMemoryActRepository()
	charRepo := repository.NewMemoryCharacterRepository()
	npcRepo := repository.NewMemoryNPCRepository()
	questRepo := repository.NewMemoryQuestRepository()
	service := NewCampaignService(
		campaignRepo, actRepo, charRepo, npcRepo, questRepo,
		tmpDir, "",
	)

	// Create a test campaign
	_, err := service.CreateCampaign("test-campaign", "Test Campaign", "Test Setting")
	if err != nil {
		t.Fatalf("Failed to create test campaign: %v", err)
	}

	return service, tmpDir
}

func TestCampaignService_SaveLore(t *testing.T) {
	service, tmpDir := setupContentTestService(t)

	tests := []struct {
		name     string
		campaign string
		content  string
		wantErr  bool
		wantFile bool
	}{
		{
			name:     "save valid lore",
			campaign: "test-campaign",
			content:  "# World History\n\nLong ago...",
			wantErr:  false,
			wantFile: true,
		},
		{
			name:     "save lore to non-existent campaign",
			campaign: "does-not-exist",
			content:  "Some lore",
			wantErr:  true,
			wantFile: false,
		},
		{
			name:     "save empty lore",
			campaign: "test-campaign",
			content:  "",
			wantErr:  false,
			wantFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveLore(tt.campaign, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveLore() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveLore() unexpected error: %v", err)
				return
			}

			if tt.wantFile {
				lorePath := filepath.Join(tmpDir, tt.campaign, "lore.md")
				if _, err := os.Stat(lorePath); os.IsNotExist(err) {
					t.Errorf("SaveLore() did not create file at %s", lorePath)
				}
				content, err := os.ReadFile(lorePath)
				if err != nil {
					t.Errorf("SaveLore() failed to read file: %v", err)
					return
				}
				if string(content) != tt.content {
					t.Errorf("SaveLore() content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestCampaignService_SaveNPCs(t *testing.T) {
	service, tmpDir := setupContentTestService(t)

	tests := []struct {
		name     string
		campaign string
		content  string
		wantErr  bool
		wantFile bool
	}{
		{
			name:     "save valid npcs",
			campaign: "test-campaign",
			content:  "# NPCs\n\n## Gandalf\nA wise wizard...",
			wantErr:  false,
			wantFile: true,
		},
		{
			name:     "save npcs to non-existent campaign",
			campaign: "does-not-exist",
			content:  "Some npcs",
			wantErr:  true,
			wantFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveNPCs(tt.campaign, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveNPCs() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveNPCs() unexpected error: %v", err)
				return
			}

			if tt.wantFile {
				npcsPath := filepath.Join(tmpDir, tt.campaign, "npcs", "npcs_and_factions.md")
				if _, err := os.Stat(npcsPath); os.IsNotExist(err) {
					t.Errorf("SaveNPCs() did not create file at %s", npcsPath)
				}
				content, err := os.ReadFile(npcsPath)
				if err != nil {
					t.Errorf("SaveNPCs() failed to read file: %v", err)
					return
				}
				if string(content) != tt.content {
					t.Errorf("SaveNPCs() content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestCampaignService_SaveEncounters(t *testing.T) {
	service, tmpDir := setupContentTestService(t)

	tests := []struct {
		name     string
		campaign string
		content  string
		wantErr  bool
		wantFile bool
	}{
		{
			name:     "save valid encounters",
			campaign: "test-campaign",
			content:  "# Encounters\n\n## Ambush\nA bandit ambush...",
			wantErr:  false,
			wantFile: true,
		},
		{
			name:     "save encounters to non-existent campaign",
			campaign: "does-not-exist",
			content:  "Some encounters",
			wantErr:  true,
			wantFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveEncounters(tt.campaign, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveEncounters() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveEncounters() unexpected error: %v", err)
				return
			}

			if tt.wantFile {
				encPath := filepath.Join(tmpDir, tt.campaign, "encounters", "encounters.md")
				if _, err := os.Stat(encPath); os.IsNotExist(err) {
					t.Errorf("SaveEncounters() did not create file at %s", encPath)
				}
				content, err := os.ReadFile(encPath)
				if err != nil {
					t.Errorf("SaveEncounters() failed to read file: %v", err)
					return
				}
				if string(content) != tt.content {
					t.Errorf("SaveEncounters() content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestCampaignService_SaveBestiary(t *testing.T) {
	service, tmpDir := setupContentTestService(t)

	tests := []struct {
		name     string
		campaign string
		content  string
		wantErr  bool
		wantFile bool
	}{
		{
			name:     "save valid bestiary",
			campaign: "test-campaign",
			content:  "# Bestiary\n\n## Goblin\nA small creature...",
			wantErr:  false,
			wantFile: true,
		},
		{
			name:     "save bestiary to non-existent campaign",
			campaign: "does-not-exist",
			content:  "Some monsters",
			wantErr:  true,
			wantFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveBestiary(tt.campaign, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveBestiary() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveBestiary() unexpected error: %v", err)
				return
			}

			if tt.wantFile {
				bestPath := filepath.Join(tmpDir, tt.campaign, "bestiary", "bestiary.md")
				if _, err := os.Stat(bestPath); os.IsNotExist(err) {
					t.Errorf("SaveBestiary() did not create file at %s", bestPath)
				}
				content, err := os.ReadFile(bestPath)
				if err != nil {
					t.Errorf("SaveBestiary() failed to read file: %v", err)
					return
				}
				if string(content) != tt.content {
					t.Errorf("SaveBestiary() content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}

func TestCampaignService_SaveMaps(t *testing.T) {
	service, tmpDir := setupContentTestService(t)

	tests := []struct {
		name     string
		campaign string
		content  string
		wantErr  bool
		wantFile bool
	}{
		{
			name:     "save valid maps",
			campaign: "test-campaign",
			content:  "# Maps\n\n## Dungeon\nA dark dungeon...",
			wantErr:  false,
			wantFile: true,
		},
		{
			name:     "save maps to non-existent campaign",
			campaign: "does-not-exist",
			content:  "Some maps",
			wantErr:  true,
			wantFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveMaps(tt.campaign, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("SaveMaps() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("SaveMaps() unexpected error: %v", err)
				return
			}

			if tt.wantFile {
				mapsPath := filepath.Join(tmpDir, tt.campaign, "maps", "maps.md")
				if _, err := os.Stat(mapsPath); os.IsNotExist(err) {
					t.Errorf("SaveMaps() did not create file at %s", mapsPath)
				}
				content, err := os.ReadFile(mapsPath)
				if err != nil {
					t.Errorf("SaveMaps() failed to read file: %v", err)
					return
				}
				if string(content) != tt.content {
					t.Errorf("SaveMaps() content mismatch: got %q, want %q", string(content), tt.content)
				}
			}
		})
	}
}
