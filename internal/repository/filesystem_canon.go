package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pauvalls/grimorio/internal/domain"
)

// FilesystemCanonRepository implements CanonRepository using filesystem JSON persistence
type FilesystemCanonRepository struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFilesystemCanonRepository creates a new filesystem canon repository
func NewFilesystemCanonRepository(baseDir string) *FilesystemCanonRepository {
	return &FilesystemCanonRepository{baseDir: baseDir}
}

func (r *FilesystemCanonRepository) canonDir(campaignID string) string {
	return filepath.Join(r.baseDir, campaignID, "canon")
}

func (r *FilesystemCanonRepository) Save(campaignID string, doc *domain.CanonDocument) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := doc.Validate(); err != nil {
		return fmt.Errorf("invalid canon document: %w", err)
	}

	dir := r.canonDir(campaignID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create canon directory: %w", err)
	}

	// Atomic write: temp file then rename
	tmpPath := filepath.Join(dir, ".canon.json.tmp")
	finalPath := filepath.Join(dir, "canon.json")

	bytes, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal canon: %w", err)
	}

	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write canon temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize canon file: %w", err)
	}

	return nil
}

func (r *FilesystemCanonRepository) Load(campaignID string) (*domain.CanonDocument, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.canonDir(campaignID), "canon.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("canon not found for campaign: %s", campaignID)
		}
		return nil, fmt.Errorf("failed to read canon file: %w", err)
	}

	var doc domain.CanonDocument
	if err := json.Unmarshal(bytes, &doc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal canon: %w", err)
	}

	if doc.SchemaVersion == "" {
		return nil, fmt.Errorf("canon missing schema_version")
	}
	if doc.SchemaVersion != domain.SchemaVersionV2 {
		return nil, fmt.Errorf("unsupported canon schema version: %s (expected %s)", doc.SchemaVersion, domain.SchemaVersionV2)
	}

	return &doc, nil
}

func (r *FilesystemCanonRepository) Exists(campaignID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, err := os.Stat(filepath.Join(r.canonDir(campaignID), "canon.json"))
	return err == nil
}

// Ensure interface is implemented
var _ CanonRepository = (*FilesystemCanonRepository)(nil)

// FilesystemNarrativeStateRepository implements NarrativeStateRepository using filesystem JSON persistence
type FilesystemNarrativeStateRepository struct {
	baseDir string
	mu      sync.RWMutex
}

// NewFilesystemNarrativeStateRepository creates a new filesystem narrative state repository
func NewFilesystemNarrativeStateRepository(baseDir string) *FilesystemNarrativeStateRepository {
	return &FilesystemNarrativeStateRepository{baseDir: baseDir}
}

func (r *FilesystemNarrativeStateRepository) stateDir(campaignID string) string {
	return filepath.Join(r.baseDir, campaignID, "canon")
}

func (r *FilesystemNarrativeStateRepository) Save(campaignID string, state *domain.NarrativeState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := state.Validate(); err != nil {
		return fmt.Errorf("invalid narrative state: %w", err)
	}

	dir := r.stateDir(campaignID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Atomic write: temp file then rename
	tmpPath := filepath.Join(dir, ".narrative_state.json.tmp")
	finalPath := filepath.Join(dir, "narrative_state.json")

	bytes, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal narrative state: %w", err)
	}

	if err := os.WriteFile(tmpPath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write state temp file: %w", err)
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize state file: %w", err)
	}

	return nil
}

func (r *FilesystemNarrativeStateRepository) Load(campaignID string) (*domain.NarrativeState, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path := filepath.Join(r.stateDir(campaignID), "narrative_state.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("narrative state not found for campaign: %s", campaignID)
		}
		return nil, fmt.Errorf("failed to read narrative state file: %w", err)
	}

	var state domain.NarrativeState
	if err := json.Unmarshal(bytes, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal narrative state: %w", err)
	}

	if state.SchemaVersion == "" {
		return nil, fmt.Errorf("narrative state missing schema_version")
	}
	if state.SchemaVersion != domain.SchemaVersionV2 {
		return nil, fmt.Errorf("unsupported narrative state schema version: %s (expected %s)", state.SchemaVersion, domain.SchemaVersionV2)
	}

	return &state, nil
}

func (r *FilesystemNarrativeStateRepository) Exists(campaignID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, err := os.Stat(filepath.Join(r.stateDir(campaignID), "narrative_state.json"))
	return err == nil
}

// Ensure interface is implemented
var _ NarrativeStateRepository = (*FilesystemNarrativeStateRepository)(nil)
