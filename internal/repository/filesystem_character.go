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

// FilesystemCharacterRepository implements CharacterRepository using filesystem
type FilesystemCharacterRepository struct {
	baseDir string
}

// NewFilesystemCharacterRepository creates a new filesystem character repository
func NewFilesystemCharacterRepository(baseDir string) *FilesystemCharacterRepository {
	return &FilesystemCharacterRepository{baseDir: baseDir}
}

func (r *FilesystemCharacterRepository) campaignDir(name string) string {
	return filepath.Join(r.baseDir, name)
}

func (r *FilesystemCharacterRepository) ensureSubdir(campaign, subdir string) error {
	dir := filepath.Join(r.campaignDir(campaign), subdir)
	return os.MkdirAll(dir, 0755)
}

func (r *FilesystemCharacterRepository) Save(character *domain.Character) error {
	if err := character.Validate(); err != nil {
		return err
	}
	character.UpdatedAt = time.Now()
	if character.CreatedAt.IsZero() {
		character.CreatedAt = time.Now()
	}

	if err := r.ensureSubdir(character.CampaignID, "characters"); err != nil {
		return err
	}

	filename := domain.SanitizeFilename(character.Name) + ".json"
	path := filepath.Join(r.campaignDir(character.CampaignID), "characters", filename)
	bytes, err := json.MarshalIndent(character, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}
	return os.WriteFile(path, bytes, 0644)
}

func (r *FilesystemCharacterRepository) Read(campaignID, name string) (*domain.Character, error) {
	var character domain.Character
	filename := domain.SanitizeFilename(name) + ".json"
	path := filepath.Join(r.campaignDir(campaignID), "characters", filename)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("character not found: %w", err)
	}
	if err := json.Unmarshal(bytes, &character); err != nil {
		return nil, fmt.Errorf("failed to unmarshal character: %w", err)
	}
	return &character, nil
}

func (r *FilesystemCharacterRepository) List(campaignID string) ([]domain.Character, error) {
	dir := filepath.Join(r.campaignDir(campaignID), "characters")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Character{}, nil
		}
		return nil, err
	}

	var characters []domain.Character
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var character domain.Character
		path := filepath.Join(dir, entry.Name())
		bytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(bytes, &character); err != nil {
			continue
		}
		characters = append(characters, character)
	}
	return characters, nil
}

func (r *FilesystemCharacterRepository) Delete(campaignID, name string) error {
	filename := domain.SanitizeFilename(name) + ".json"
	path := filepath.Join(r.campaignDir(campaignID), "characters", filename)
	return os.Remove(path)
}

// Ensure interface is implemented
var _ CharacterRepository = (*FilesystemCharacterRepository)(nil)
