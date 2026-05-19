package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemMapRepository implements MapRepository using the filesystem.
type FilesystemMapRepository struct {
	baseDir string
}

// NewFilesystemMapRepository creates a new FilesystemMapRepository.
func NewFilesystemMapRepository(baseDir string) *FilesystemMapRepository {
	return &FilesystemMapRepository{baseDir: baseDir}
}

func (r *FilesystemMapRepository) Save(ctx context.Context, campaignID string, m *domain.Map) error {
	dir, err := ensureSubdir(r.baseDir, campaignID, "maps")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(m.Name)+".json")
	return writeJSON(path, m)
}

func (r *FilesystemMapRepository) Read(ctx context.Context, campaignID string, name string) (*domain.Map, error) {
	var m domain.Map
	path := entityPath(r.baseDir, campaignID, "maps", name)
	if err := readJSON(path, &m); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("map not found: %w", err)
		}
		return nil, err
	}
	return &m, nil
}

func (r *FilesystemMapRepository) List(ctx context.Context, campaignID string) ([]*domain.Map, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "maps")
	return listJSONFiles[domain.Map](dir)
}

func (r *FilesystemMapRepository) Delete(ctx context.Context, campaignID string, name string) error {
	path := entityPath(r.baseDir, campaignID, "maps", name)
	return deleteFile(path)
}
