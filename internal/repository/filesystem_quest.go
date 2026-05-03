package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemQuestRepository implements QuestRepository using filesystem
type FilesystemQuestRepository struct {
	baseDir string
}

// NewFilesystemQuestRepository creates a new filesystem quest repository
func NewFilesystemQuestRepository(baseDir string) *FilesystemQuestRepository {
	return &FilesystemQuestRepository{baseDir: baseDir}
}

func (r *FilesystemQuestRepository) campaignDir(name string) string {
	return filepath.Join(r.baseDir, name)
}

func (r *FilesystemQuestRepository) ensureSubdir(campaign, subdir string) error {
	dir := filepath.Join(r.campaignDir(campaign), subdir)
	return os.MkdirAll(dir, 0755)
}

func (r *FilesystemQuestRepository) Save(quest *domain.Quest) error {
	if err := quest.Validate(); err != nil {
		return err
	}
	quest.UpdatedAt = time.Now()
	if quest.CreatedAt.IsZero() {
		quest.CreatedAt = time.Now()
	}
	if quest.ID == "" {
		quest.ID = fmt.Sprintf("quest_%d", time.Now().Unix())
	}

	if err := r.ensureSubdir(quest.CampaignID, "quests"); err != nil {
		return err
	}

	filename := quest.ID + ".json"
	path := filepath.Join(r.campaignDir(quest.CampaignID), "quests", filename)
	bytes, err := json.MarshalIndent(quest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal quest: %w", err)
	}
	return os.WriteFile(path, bytes, 0644)
}

func (r *FilesystemQuestRepository) Read(campaignID, id string) (*domain.Quest, error) {
	var quest domain.Quest
	path := filepath.Join(r.campaignDir(campaignID), "quests", id+".json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("quest not found: %w", err)
	}
	if err := json.Unmarshal(bytes, &quest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal quest: %w", err)
	}
	return &quest, nil
}

func (r *FilesystemQuestRepository) List(campaignID string) ([]domain.Quest, error) {
	dir := filepath.Join(r.campaignDir(campaignID), "quests")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Quest{}, nil
		}
		return nil, err
	}

	var quests []domain.Quest
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var quest domain.Quest
		path := filepath.Join(dir, entry.Name())
		bytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(bytes, &quest); err != nil {
			continue
		}
		quests = append(quests, quest)
	}
	return quests, nil
}

func (r *FilesystemQuestRepository) ListByCharacter(campaignID, characterID string) ([]domain.Quest, error) {
	all, err := r.List(campaignID)
	if err != nil {
		return nil, err
	}

	var filtered []domain.Quest
	for _, quest := range all {
		if quest.CharacterID != nil && *quest.CharacterID == characterID {
			filtered = append(filtered, quest)
		}
	}
	return filtered, nil
}

func (r *FilesystemQuestRepository) ListByStatus(campaignID string, status domain.QuestStatus) ([]domain.Quest, error) {
	all, err := r.List(campaignID)
	if err != nil {
		return nil, err
	}

	var filtered []domain.Quest
	for _, quest := range all {
		if quest.Status == status {
			filtered = append(filtered, quest)
		}
	}
	return filtered, nil
}

func (r *FilesystemQuestRepository) Delete(campaignID, id string) error {
	path := filepath.Join(r.campaignDir(campaignID), "quests", id+".json")
	return os.Remove(path)
}

// Ensure interface is implemented
var _ QuestRepository = (*FilesystemQuestRepository)(nil)
