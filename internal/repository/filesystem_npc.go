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

// FilesystemNPCRepository implements NPCRepository using filesystem
type FilesystemNPCRepository struct {
	baseDir string
}

// NewFilesystemNPCRepository creates a new filesystem NPC repository
func NewFilesystemNPCRepository(baseDir string) *FilesystemNPCRepository {
	return &FilesystemNPCRepository{baseDir: baseDir}
}

func (r *FilesystemNPCRepository) campaignDir(name string) string {
	return filepath.Join(r.baseDir, name)
}

func (r *FilesystemNPCRepository) ensureSubdir(campaign, subdir string) error {
	dir := filepath.Join(r.campaignDir(campaign), subdir)
	return os.MkdirAll(dir, 0755)
}

func (r *FilesystemNPCRepository) Save(npc *domain.NPC) error {
	if npc.CampaignID == "" {
		return domain.NewValidationError("campaign_id", "campaign ID is required")
	}
	if npc.Name == "" {
		return domain.NewValidationError("name", "NPC name is required")
	}
	npc.UpdatedAt = time.Now()
	if npc.CreatedAt.IsZero() {
		npc.CreatedAt = time.Now()
	}

	if err := r.ensureSubdir(npc.CampaignID, "npcs"); err != nil {
		return err
	}

	filename := domain.SanitizeFilename(npc.Name) + ".json"
	path := filepath.Join(r.campaignDir(npc.CampaignID), "npcs", filename)
	bytes, err := json.MarshalIndent(npc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal NPC: %w", err)
	}
	return os.WriteFile(path, bytes, 0644)
}

func (r *FilesystemNPCRepository) Read(campaignID, name string) (*domain.NPC, error) {
	var npc domain.NPC
	filename := domain.SanitizeFilename(name) + ".json"
	path := filepath.Join(r.campaignDir(campaignID), "npcs", filename)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("NPC not found: %w", err)
	}
	if err := json.Unmarshal(bytes, &npc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal NPC: %w", err)
	}
	return &npc, nil
}

func (r *FilesystemNPCRepository) List(campaignID string) ([]domain.NPC, error) {
	dir := filepath.Join(r.campaignDir(campaignID), "npcs")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.NPC{}, nil
		}
		return nil, err
	}

	var npcs []domain.NPC
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var npc domain.NPC
		path := filepath.Join(dir, entry.Name())
		bytes, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(bytes, &npc); err != nil {
			continue
		}
		npcs = append(npcs, npc)
	}
	return npcs, nil
}

func (r *FilesystemNPCRepository) Delete(campaignID, name string) error {
	filename := domain.SanitizeFilename(name) + ".json"
	path := filepath.Join(r.campaignDir(campaignID), "npcs", filename)
	return os.Remove(path)
}

// Ensure interface is implemented
var _ NPCRepository = (*FilesystemNPCRepository)(nil)
