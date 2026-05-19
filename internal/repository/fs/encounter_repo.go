package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemEncounterRepository implements EncounterRepository using the filesystem.
type FilesystemEncounterRepository struct {
	baseDir string
}

// NewFilesystemEncounterRepository creates a new FilesystemEncounterRepository.
func NewFilesystemEncounterRepository(baseDir string) *FilesystemEncounterRepository {
	return &FilesystemEncounterRepository{baseDir: baseDir}
}

func (r *FilesystemEncounterRepository) Save(ctx context.Context, campaignID string, encounter *domain.Encounter) error {
	dir, err := ensureSubdir(r.baseDir, campaignID, "encounters")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(encounter.Name)+".json")
	return writeJSON(path, encounter)
}

func (r *FilesystemEncounterRepository) Read(ctx context.Context, campaignID string, name string) (*domain.Encounter, error) {
	var enc domain.Encounter
	path := entityPath(r.baseDir, campaignID, "encounters", name)
	if err := readJSON(path, &enc); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("encounter not found: %w", err)
		}
		return nil, err
	}
	return &enc, nil
}

func (r *FilesystemEncounterRepository) List(ctx context.Context, campaignID string) ([]*domain.Encounter, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "encounters")
	return listJSONFiles[domain.Encounter](dir)
}

func (r *FilesystemEncounterRepository) Delete(ctx context.Context, campaignID string, name string) error {
	path := entityPath(r.baseDir, campaignID, "encounters", name)
	return deleteFile(path)
}
