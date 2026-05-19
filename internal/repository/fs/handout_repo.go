package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemHandoutRepositoryV3 implements HandoutRepositoryV3 using the filesystem.
type FilesystemHandoutRepositoryV3 struct {
	baseDir string
}

// NewFilesystemHandoutRepositoryV3 creates a new FilesystemHandoutRepositoryV3.
func NewFilesystemHandoutRepositoryV3(baseDir string) *FilesystemHandoutRepositoryV3 {
	return &FilesystemHandoutRepositoryV3{baseDir: baseDir}
}

func (r *FilesystemHandoutRepositoryV3) Create(ctx context.Context, campaignID string, handout *domain.Handout) error {
	dir, err := ensureSubdir(r.baseDir, campaignID, "handouts")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(handout.ID)+".json")
	return writeJSON(path, handout)
}

func (r *FilesystemHandoutRepositoryV3) Read(ctx context.Context, campaignID string, handoutID string) (*domain.Handout, error) {
	var h domain.Handout
	path := entityPath(r.baseDir, campaignID, "handouts", handoutID)
	if err := readJSON(path, &h); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("handout not found: %w", err)
		}
		return nil, err
	}
	return &h, nil
}

func (r *FilesystemHandoutRepositoryV3) Update(ctx context.Context, campaignID string, handout *domain.Handout) error {
	path := entityPath(r.baseDir, campaignID, "handouts", handout.ID)
	return writeJSON(path, handout)
}

func (r *FilesystemHandoutRepositoryV3) Delete(ctx context.Context, campaignID string, handoutID string) error {
	path := entityPath(r.baseDir, campaignID, "handouts", handoutID)
	return deleteFile(path)
}

func (r *FilesystemHandoutRepositoryV3) GetByQuest(ctx context.Context, campaignID string, questID string) ([]*domain.Handout, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "handouts")
	all, err := listJSONFiles[domain.Handout](dir)
	if err != nil {
		return nil, err
	}
	var result []*domain.Handout
	for _, h := range all {
		for _, ref := range h.QuestRefs {
			if ref == questID {
				result = append(result, h)
				break
			}
		}
	}
	return result, nil
}

func (r *FilesystemHandoutRepositoryV3) GetByArea(ctx context.Context, campaignID string, areaID string) ([]*domain.Handout, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "handouts")
	all, err := listJSONFiles[domain.Handout](dir)
	if err != nil {
		return nil, err
	}
	var result []*domain.Handout
	for _, h := range all {
		for _, ref := range h.AreaRefs {
			if ref == areaID {
				result = append(result, h)
				break
			}
		}
	}
	return result, nil
}
