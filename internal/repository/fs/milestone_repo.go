package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemMilestoneXPRepository implements MilestoneXPRepository using the filesystem.
type FilesystemMilestoneXPRepository struct {
	baseDir string
}

// NewFilesystemMilestoneXPRepository creates a new FilesystemMilestoneXPRepository.
func NewFilesystemMilestoneXPRepository(baseDir string) *FilesystemMilestoneXPRepository {
	return &FilesystemMilestoneXPRepository{baseDir: baseDir}
}

func (r *FilesystemMilestoneXPRepository) Create(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error {
	if err := table.Validate(); err != nil {
		return err
	}
	dir, err := ensureSubdir(r.baseDir, campaignID, "milestones")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, domain.SanitizeFilename(table.ChapterID)+".json")
	return writeJSON(path, table)
}

func (r *FilesystemMilestoneXPRepository) Read(ctx context.Context, campaignID string, chapterID string) (*domain.ChapterXPTable, error) {
	var table domain.ChapterXPTable
	path := entityPath(r.baseDir, campaignID, "milestones", chapterID)
	if err := readJSON(path, &table); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("milestone table not found: %w", err)
		}
		return nil, err
	}
	return &table, nil
}

func (r *FilesystemMilestoneXPRepository) Update(ctx context.Context, campaignID string, table *domain.ChapterXPTable) error {
	if err := table.Validate(); err != nil {
		return err
	}
	path := entityPath(r.baseDir, campaignID, "milestones", table.ChapterID)
	return writeJSON(path, table)
}

func (r *FilesystemMilestoneXPRepository) Delete(ctx context.Context, campaignID string, chapterID string) error {
	path := entityPath(r.baseDir, campaignID, "milestones", chapterID)
	return deleteFile(path)
}

func (r *FilesystemMilestoneXPRepository) GetTotalXP(ctx context.Context, campaignID string, partyID string) (int, error) {
	dir := filepath.Join(campaignDir(r.baseDir, campaignID), "milestones")
	all, err := listJSONFiles[domain.ChapterXPTable](dir)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, table := range all {
		for _, m := range table.Milestones {
			total += m.CumulativeXP
		}
	}
	return total, nil
}
