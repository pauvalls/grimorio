package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemMonsterRepository implements MonsterRepository using the filesystem.
type FilesystemMonsterRepository struct {
	baseDir string
}

// NewFilesystemMonsterRepository creates a new FilesystemMonsterRepository.
func NewFilesystemMonsterRepository(baseDir string) *FilesystemMonsterRepository {
	return &FilesystemMonsterRepository{baseDir: baseDir}
}

func (r *FilesystemMonsterRepository) Save(ctx context.Context, campaignID string, monster *domain.Monster) error {
	dir, err := ensureSubdir(r.baseDir, campaignID, "monsters")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(monster.Name)+".json")
	return writeJSON(path, monster)
}

func (r *FilesystemMonsterRepository) Read(ctx context.Context, campaignID string, name string) (*domain.Monster, error) {
	var monster domain.Monster
	path := entityPath(r.baseDir, campaignID, "monsters", name)
	if err := readJSON(path, &monster); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("monster not found: %w", err)
		}
		return nil, err
	}
	return &monster, nil
}

func (r *FilesystemMonsterRepository) List(ctx context.Context, campaignID string) ([]*domain.Monster, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "monsters")
	return listJSONFiles[domain.Monster](dir)
}

func (r *FilesystemMonsterRepository) Delete(ctx context.Context, campaignID string, name string) error {
	path := entityPath(r.baseDir, campaignID, "monsters", name)
	return deleteFile(path)
}
