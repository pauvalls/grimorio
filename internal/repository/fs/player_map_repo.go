package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemPlayerMapRepository implements PlayerMapRepository using the filesystem.
type FilesystemPlayerMapRepository struct {
	baseDir string
}

// NewFilesystemPlayerMapRepository creates a new FilesystemPlayerMapRepository.
func NewFilesystemPlayerMapRepository(baseDir string) *FilesystemPlayerMapRepository {
	return &FilesystemPlayerMapRepository{baseDir: baseDir}
}

func (r *FilesystemPlayerMapRepository) Create(ctx context.Context, campaignID string, pm *domain.PlayerMap) error {
	if err := pm.Validate(); err != nil {
		return err
	}
	dir, err := ensureSubdir(r.baseDir, campaignID, "player_maps")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(pm.ID)+".json")
	return writeJSON(path, pm)
}

func (r *FilesystemPlayerMapRepository) Read(ctx context.Context, campaignID string, mapID string) (*domain.PlayerMap, error) {
	var pm domain.PlayerMap
	path := entityPath(r.baseDir, campaignID, "player_maps", mapID)
	if err := readJSON(path, &pm); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("player map not found: %w", err)
		}
		return nil, err
	}
	return &pm, nil
}

func (r *FilesystemPlayerMapRepository) Update(ctx context.Context, campaignID string, pm *domain.PlayerMap) error {
	if err := pm.Validate(); err != nil {
		return err
	}
	path := entityPath(r.baseDir, campaignID, "player_maps", pm.ID)
	return writeJSON(path, pm)
}

func (r *FilesystemPlayerMapRepository) Delete(ctx context.Context, campaignID string, mapID string) error {
	path := entityPath(r.baseDir, campaignID, "player_maps", mapID)
	return deleteFile(path)
}

func (r *FilesystemPlayerMapRepository) GetByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.PlayerMap, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "player_maps")
	all, err := listJSONFiles[domain.PlayerMap](dir)
	if err != nil {
		return nil, err
	}
	var result []*domain.PlayerMap
	for _, pm := range all {
		if pm.AreaID == areaID {
			result = append(result, pm)
		}
	}
	return result, nil
}
