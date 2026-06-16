package services

import (
	"testing"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
	"github.com/pauvalls/grimorio/internal/repository"
)

func TestQuestService_CreateQuest(t *testing.T) {
	repo := repository.NewMemoryQuestRepository()
	questService := NewQuestService(repo)

	tests := []struct {
		name        string
		campaignID  string
		title       string
		questType   domain.QuestType
		hook        string
		description string
		stakes      string
		wantErr     bool
	}{
		{
			name:       "create valid quest",
			campaignID: "test-campaign",
			title:      "Find the Sword",
			questType:  domain.QuestTypeMain,
			hook:       "A mysterious stranger approaches...",
			stakes:     "The kingdom's fate",
			wantErr:    false,
		},
		{
			name:       "create quest without optional fields",
			campaignID: "test-campaign",
			title:      "Simple Quest",
			wantErr:    false,
		},
		{
			name:    "missing campaign",
			title:   "Bad Quest",
			wantErr: true,
		},
		{
			name:       "missing title",
			campaignID: "test-campaign",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quest, err := questService.CreateQuest(tt.campaignID, tt.title, tt.questType, tt.hook, tt.description, tt.stakes, nil)
			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateQuest() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("CreateQuest() unexpected error: %v", err)
				return
			}
			if quest.Title != tt.title {
				t.Errorf("CreateQuest() title = %v, want %v", quest.Title, tt.title)
			}
			if quest.Status != domain.QuestStatusActive {
				t.Errorf("CreateQuest() status = %v, want active", quest.Status)
			}
			if quest.ID == "" {
				t.Errorf("CreateQuest() ID should not be empty")
			}
		})
	}
}

func TestQuestService_UpdateQuestStatus(t *testing.T) {
	repo := repository.NewMemoryQuestRepository()
	questService := NewQuestService(repo)

	// Create a quest
	quest, err := questService.CreateQuest("status-test", "Test Quest", domain.QuestTypeMain, "", "", "", nil)
	if err != nil {
		t.Fatalf("Failed to create test quest: %v", err)
	}

	tests := []struct {
		name    string
		status  domain.QuestStatus
		notes   string
		wantErr bool
	}{
		{
			name:    "complete quest",
			status:  domain.QuestStatusCompleted,
			notes:   "Quest completed successfully",
			wantErr: false,
		},
		{
			name:    "fail quest",
			status:  domain.QuestStatusFailed,
			wantErr: false,
		},
		{
			name:    "put on hold",
			status:  domain.QuestStatusOnHold,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := questService.UpdateQuestStatus("status-test", quest.ID, tt.status, tt.notes)
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateQuestStatus() expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("UpdateQuestStatus() unexpected error: %v", err)
				return
			}

			// Verify
			updated, err := questService.GetQuest("status-test", quest.ID)
			if err != nil {
				t.Errorf("GetQuest() unexpected error: %v", err)
				return
			}

			if updated.Status != tt.status {
				t.Errorf("UpdateQuestStatus() status = %v, want %v", updated.Status, tt.status)
			}
		})
	}
}

func TestQuestService_ListQuests(t *testing.T) {
	repo := repository.NewMemoryQuestRepository()
	questService := NewQuestService(repo)

	// Create some quests. Sleep between writes so MemoryQuestRepository's
	// time.Now().UnixNano()-based ID generator produces distinct IDs on
	// hosts with coarse timer resolution (Windows).
	_, _ = questService.CreateQuest("list-test", "Quest 1", domain.QuestTypeMain, "", "", "", nil)
	time.Sleep(10 * time.Millisecond)
	_, _ = questService.CreateQuest("list-test", "Quest 2", domain.QuestTypeSide, "", "", "", nil)

	quests, err := questService.ListQuests("list-test")
	if err != nil {
		t.Errorf("ListQuests() unexpected error: %v", err)
		return
	}

	if len(quests) != 2 {
		t.Errorf("ListQuests() got %d quests, want 2", len(quests))
	}
}

func TestQuestService_ListActiveQuests(t *testing.T) {
	repo := repository.NewMemoryQuestRepository()
	questService := NewQuestService(repo)

	// Create quests with different statuses. Sleep between writes so the
	// nanosecond-based ID generator yields distinct IDs on Windows.
	q1, _ := questService.CreateQuest("active-test", "Active Quest", domain.QuestTypeMain, "", "", "", nil)
	time.Sleep(10 * time.Millisecond)
	q2, _ := questService.CreateQuest("active-test", "Completed Quest", domain.QuestTypeSide, "", "", "", nil)
	if err := questService.UpdateQuestStatus("active-test", q2.ID, domain.QuestStatusCompleted, ""); err != nil {
		t.Fatal(err)
	}

	active, err := questService.ListActiveQuests("active-test")
	if err != nil {
		t.Errorf("ListActiveQuests() unexpected error: %v", err)
		return
	}

	if len(active) != 1 {
		t.Errorf("ListActiveQuests() got %d quests, want 1", len(active))
		return
	}

	if active[0].ID != q1.ID {
		t.Errorf("ListActiveQuests() got quest %v, want %v", active[0].ID, q1.ID)
	}
}

func TestQuestService_Objectives(t *testing.T) {
	repo := repository.NewMemoryQuestRepository()
	questService := NewQuestService(repo)

	// Create a quest
	quest, err := questService.CreateQuest("obj-test", "Objective Quest", domain.QuestTypeMain, "", "", "", nil)
	if err != nil {
		t.Fatalf("Failed to create test quest: %v", err)
	}

	// Add objectives
	err = questService.AddObjective("obj-test", quest.ID, "Find the map")
	if err != nil {
		t.Errorf("AddObjective() unexpected error: %v", err)
		return
	}

	err = questService.AddObjective("obj-test", quest.ID, "Travel to the location")
	if err != nil {
		t.Errorf("AddObjective() unexpected error: %v", err)
		return
	}

	// Complete first objective
	quest, _ = questService.GetQuest("obj-test", quest.ID)
	if len(quest.Objectives) != 2 {
		t.Fatalf("Expected 2 objectives, got %d", len(quest.Objectives))
	}

	err = questService.CompleteObjective("obj-test", quest.ID, quest.Objectives[0].ID)
	if err != nil {
		t.Errorf("CompleteObjective() unexpected error: %v", err)
		return
	}

	// Verify
	updated, _ := questService.GetQuest("obj-test", quest.ID)
	if !updated.Objectives[0].Completed {
		t.Errorf("CompleteObjective() objective should be completed")
	}
	if updated.Objectives[1].Completed {
		t.Errorf("CompleteObjective() second objective should not be completed")
	}
}
