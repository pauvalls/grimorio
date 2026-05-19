package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemAreaRepositoryV3 implements AreaRepositoryV3 using the filesystem.
type FilesystemAreaRepositoryV3 struct {
	baseDir string
}

// NewFilesystemAreaRepositoryV3 creates a new FilesystemAreaRepositoryV3.
func NewFilesystemAreaRepositoryV3(baseDir string) *FilesystemAreaRepositoryV3 {
	return &FilesystemAreaRepositoryV3{baseDir: baseDir}
}

func (r *FilesystemAreaRepositoryV3) Create(ctx context.Context, campaignID string, area *domain.Area) error {
	if err := area.Validate(); err != nil {
		return err
	}
	dir, err := ensureSubdir(r.baseDir, campaignID, "areas_v3")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(area.ID)+".json")
	return writeJSON(path, area)
}

func (r *FilesystemAreaRepositoryV3) Read(ctx context.Context, campaignID string, areaID string) (*domain.Area, error) {
	var area domain.Area
	path := entityPath(r.baseDir, campaignID, "areas_v3", areaID)
	if err := readJSON(path, &area); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("area not found: %w", err)
		}
		return nil, err
	}
	return &area, nil
}

func (r *FilesystemAreaRepositoryV3) Update(ctx context.Context, campaignID string, area *domain.Area) error {
	if err := area.Validate(); err != nil {
		return err
	}
	path := entityPath(r.baseDir, campaignID, "areas_v3", area.ID)
	return writeJSON(path, area)
}

func (r *FilesystemAreaRepositoryV3) Delete(ctx context.Context, campaignID string, areaID string) error {
	path := entityPath(r.baseDir, campaignID, "areas_v3", areaID)
	return deleteFile(path)
}

func (r *FilesystemAreaRepositoryV3) GetByChapter(ctx context.Context, campaignID string, chapterID string) ([]*domain.Area, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "areas_v3")
	all, err := listJSONFiles[domain.Area](dir)
	if err != nil {
		return nil, err
	}
	var result []*domain.Area
	for _, area := range all {
		if area.ChapterID == chapterID {
			result = append(result, area)
		}
	}
	return result, nil
}
