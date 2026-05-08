package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pauvalls/grimorio/internal/domain"
)

// MemoryFactionRepository implements FactionReputationRepository in memory
type MemoryFactionRepository struct {
	mu       sync.RWMutex
	matrices map[string]*domain.FactionReputationMatrix
}

// NewMemoryFactionRepository creates a new in-memory faction repository
func NewMemoryFactionRepository() *MemoryFactionRepository {
	return &MemoryFactionRepository{
		matrices: make(map[string]*domain.FactionReputationMatrix),
	}
}

func (r *MemoryFactionRepository) Save(campaignID string, matrix *domain.FactionReputationMatrix) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.matrices[campaignID] = matrix
	return nil
}

func (r *MemoryFactionRepository) Load(campaignID string) (*domain.FactionReputationMatrix, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matrix, exists := r.matrices[campaignID]
	if !exists {
		return &domain.FactionReputationMatrix{
			CampaignID: campaignID,
			Entries:    []domain.ReputationEntry{},
		}, nil
	}
	return matrix, nil
}

// Ensure interface is implemented
var _ FactionReputationRepository = (*MemoryFactionRepository)(nil)

// FilesystemFactionRepository implements FactionReputationRepository using filesystem JSON persistence
type FilesystemFactionRepository struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFilesystemFactionRepository creates a new filesystem faction repository
func NewFilesystemFactionRepository(baseDir string) *FilesystemFactionRepository {
	return &FilesystemFactionRepository{baseDir: baseDir}
}

func (r *FilesystemFactionRepository) factionDir(campaignID string) string {
	return filepath.Join(r.baseDir, campaignID, "factions")
}

func (r *FilesystemFactionRepository) Save(campaignID string, matrix *domain.FactionReputationMatrix) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	dir := r.factionDir(campaignID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create faction directory: %w", err)
	}

	tmpPath := filepath.Join(dir, ".reputation_matrix.json.tmp")
	finalPath := filepath.Join(dir, "reputation_matrix.json")

	bytes, err := json.MarshalIndent(matrix, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal reputation matrix: %w", err)
	}

	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write reputation matrix temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize reputation matrix file: %w", err)
	}

	return nil
}

func (r *FilesystemFactionRepository) Load(campaignID string) (*domain.FactionReputationMatrix, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.factionDir(campaignID), "reputation_matrix.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &domain.FactionReputationMatrix{
				CampaignID: campaignID,
				Entries:    []domain.ReputationEntry{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read reputation matrix file: %w", err)
	}

	var matrix domain.FactionReputationMatrix
	if err := json.Unmarshal(bytes, &matrix); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reputation matrix: %w", err)
	}

	return &matrix, nil
}

// Ensure interface is implemented
var _ FactionReputationRepository = (*FilesystemFactionRepository)(nil)
