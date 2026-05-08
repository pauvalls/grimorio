package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemActRepository implements ActRepository using filesystem
type FilesystemActRepository struct {
	baseDir string
}

// NewFilesystemActRepository creates a new filesystem act repository
func NewFilesystemActRepository(baseDir string) *FilesystemActRepository {
	return &FilesystemActRepository{baseDir: baseDir}
}

func (r *FilesystemActRepository) campaignDir(name string) string {
	return filepath.Join(r.baseDir, name)
}

func (r *FilesystemActRepository) ensureSubdir(campaign, subdir string) error {
	dir := filepath.Join(r.campaignDir(campaign), subdir)
	return os.MkdirAll(dir, 0755)
}

func (r *FilesystemActRepository) Save(act *domain.Act) error {
	if err := act.Validate(); err != nil {
		return err
	}
	act.UpdatedAt = time.Now()
	if act.CreatedAt.IsZero() {
		act.CreatedAt = time.Now()
	}

	if err := r.ensureSubdir(act.CampaignID, "areas"); err != nil {
		return err
	}

	filename := fmt.Sprintf("chapter_%02d_%s.md", act.Number, domain.SanitizeFilename(act.Title))

	// Add header if content doesn't already start with a heading
	content := act.Content
	if !strings.HasPrefix(strings.TrimSpace(content), "# ") {
		header := fmt.Sprintf("# Capítulo %d: %s\n\n", act.Number, act.Title)
		content = header + content
	}

	path := filepath.Join(r.campaignDir(act.CampaignID), "areas", filename)
	return os.WriteFile(path, []byte(content), 0644)
}

func (r *FilesystemActRepository) Read(campaignID string, number int) (*domain.Act, error) {
	files, err := r.listFiles(campaignID, "areas")
	if err != nil {
		return nil, err
	}

	prefix := fmt.Sprintf("chapter_%02d_", number)
	for _, file := range files {
		if strings.HasPrefix(file, prefix) {
			path := filepath.Join(r.campaignDir(campaignID), "areas", file)
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}
			return &domain.Act{
				CampaignID: campaignID,
				Number:     number,
				Content:    string(content),
			}, nil
		}
	}
	return nil, fmt.Errorf("chapter %d not found in campaign %s", number, campaignID)
}

func (r *FilesystemActRepository) List(campaignID string) ([]domain.Act, error) {
	files, err := r.listFiles(campaignID, "areas")
	if err != nil {
		return nil, err
	}

	var acts []domain.Act
	for _, file := range files {
		if !strings.HasSuffix(file, ".md") {
			continue
		}
		path := filepath.Join(r.campaignDir(campaignID), "areas", file)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var number int
		_, _ = fmt.Sscanf(file, "chapter_%02d_", &number)

		acts = append(acts, domain.Act{
			CampaignID: campaignID,
			Number:     number,
			Content:    string(content),
		})
	}
	return acts, nil
}

func (r *FilesystemActRepository) Delete(campaignID string, number int) error {
	files, err := r.listFiles(campaignID, "areas")
	if err != nil {
		return err
	}

	prefix := fmt.Sprintf("chapter_%02d_", number)
	for _, file := range files {
		if strings.HasPrefix(file, prefix) {
			path := filepath.Join(r.campaignDir(campaignID), "areas", file)
			return os.Remove(path)
		}
	}
	return nil
}

func (r *FilesystemActRepository) listFiles(campaign, subdir string) ([]string, error) {
	dir := filepath.Join(r.campaignDir(campaign), subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// Ensure interface is implemented
var _ ActRepository = (*FilesystemActRepository)(nil)
