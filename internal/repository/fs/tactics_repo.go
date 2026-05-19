package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemTacticsRepository implements TacticsRepository using the filesystem.
type FilesystemTacticsRepository struct {
	baseDir string
}

// NewFilesystemTacticsRepository creates a new FilesystemTacticsRepository.
func NewFilesystemTacticsRepository(baseDir string) *FilesystemTacticsRepository {
	return &FilesystemTacticsRepository{baseDir: baseDir}
}

func (r *FilesystemTacticsRepository) Create(ctx context.Context, campaignID string, tactics *domain.Tactics) error {
	if err := tactics.Validate(); err != nil {
		return err
	}
	dir, err := ensureSubdir(r.baseDir, campaignID, "tactics")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(tactics.MonsterID)+".json")
	return writeJSON(path, tactics)
}

func (r *FilesystemTacticsRepository) Read(ctx context.Context, campaignID string, tacticsID string) (*domain.Tactics, error) {
	var tactics domain.Tactics
	path := entityPath(r.baseDir, campaignID, "tactics", tacticsID)
	if err := readJSON(path, &tactics); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tactics not found: %w", err)
		}
		return nil, err
	}
	return &tactics, nil
}

func (r *FilesystemTacticsRepository) ListByEncounter(ctx context.Context, campaignID string, encounterID string) ([]*domain.Tactics, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "tactics")
	all, err := listJSONFiles[domain.Tactics](dir)
	if err != nil {
		return nil, err
	}
	var result []*domain.Tactics
	for _, t := range all {
		if t.EncounterID == encounterID {
			result = append(result, t)
		}
	}
	return result, nil
}
