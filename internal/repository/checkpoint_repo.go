package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/pauvalls/grimorio/internal/domain"
)

// CheckpointRepository handles checkpoint persistence
type CheckpointRepository interface {
	// Save persists a checkpoint
	Save(ctx context.Context, campaignID string, checkpoint *domain.PipelineCheckpoint) error

	// Load retrieves a checkpoint by batch ID
	Load(ctx context.Context, campaignID string, batchID string) (*domain.PipelineCheckpoint, error)

	// List returns all checkpoints for a campaign, sorted by CreatedAt descending
	List(ctx context.Context, campaignID string) ([]*domain.PipelineCheckpoint, error)

	// Delete removes a checkpoint
	Delete(ctx context.Context, campaignID string, batchID string) error

	// FindBySessionNum finds checkpoint for a specific session
	FindBySessionNum(ctx context.Context, campaignID string, sessionNum int) (*domain.PipelineCheckpoint, error)
}

// filesystemCheckpointRepository implements CheckpointRepository using filesystem
type filesystemCheckpointRepository struct {
	baseDir string
}

// NewCheckpointRepository creates a new checkpoint repository
func NewCheckpointRepository(baseDir string) CheckpointRepository {
	return &filesystemCheckpointRepository{
		baseDir: baseDir,
	}
}

func (r *filesystemCheckpointRepository) Save(
	ctx context.Context,
	campaignID string,
	checkpoint *domain.PipelineCheckpoint,
) error {
	// Compute hash for integrity
	hash, err := computeCheckpointHash(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to compute checkpoint hash: %w", err)
	}
	checkpoint.CheckpointHash = hash

	// Ensure checkpoints directory exists
	checkpointsDir := filepath.Join(r.baseDir, campaignID, "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoints directory: %w", err)
	}

	// Serialize checkpoint
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Write to file
	filePath := filepath.Join(checkpointsDir, fmt.Sprintf("checkpoint-%s.json", checkpoint.BatchID))
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	// Enforce retention policy: keep last 3 checkpoints
	if err := r.enforceRetention(campaignID, 3); err != nil {
		// Non-fatal: log warning but don't fail
	}

	return nil
}

func (r *filesystemCheckpointRepository) Load(
	ctx context.Context,
	campaignID string,
	batchID string,
) (*domain.PipelineCheckpoint, error) {
	filePath := filepath.Join(r.baseDir, campaignID, "checkpoints", fmt.Sprintf("checkpoint-%s.json", batchID))

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint not found for batch %s", batchID)
		}
		return nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint domain.PipelineCheckpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	// Validate integrity
	computedHash, err := computeCheckpointHash(&checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to compute hash for validation: %w", err)
	}
	if computedHash != checkpoint.CheckpointHash {
		return nil, fmt.Errorf("checkpoint integrity check failed: hash mismatch")
	}

	return &checkpoint, nil
}

func (r *filesystemCheckpointRepository) List(
	ctx context.Context,
	campaignID string,
) ([]*domain.PipelineCheckpoint, error) {
	checkpointsDir := filepath.Join(r.baseDir, campaignID, "checkpoints")

	entries, err := os.ReadDir(checkpointsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*domain.PipelineCheckpoint{}, nil
		}
		return nil, fmt.Errorf("failed to read checkpoints directory: %w", err)
	}

	checkpoints := make([]*domain.PipelineCheckpoint, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !hasSuffix(entry.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(checkpointsDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip unreadable files
		}

		var checkpoint domain.PipelineCheckpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue // Skip invalid files
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	// Sort by CreatedAt descending
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].CreatedAt.After(checkpoints[j].CreatedAt)
	})

	return checkpoints, nil
}

func (r *filesystemCheckpointRepository) Delete(
	ctx context.Context,
	campaignID string,
	batchID string,
) error {
	filePath := filepath.Join(r.baseDir, campaignID, "checkpoints", fmt.Sprintf("checkpoint-%s.json", batchID))

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("checkpoint not found for batch %s", batchID)
		}
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}

	return nil
}

func (r *filesystemCheckpointRepository) FindBySessionNum(
	ctx context.Context,
	campaignID string,
	sessionNum int,
) (*domain.PipelineCheckpoint, error) {
	checkpoints, err := r.List(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	for _, checkpoint := range checkpoints {
		if checkpoint.SessionNum == sessionNum {
			return checkpoint, nil
		}
	}

	return nil, fmt.Errorf("no checkpoint found for session %d", sessionNum)
}

// enforceRetention keeps only the last N checkpoints
func (r *filesystemCheckpointRepository) enforceRetention(campaignID string, maxCheckpoints int) error {
	checkpoints, err := r.List(context.Background(), campaignID)
	if err != nil {
		return err
	}

	if len(checkpoints) <= maxCheckpoints {
		return nil
	}

	// Delete oldest checkpoints
	for i := maxCheckpoints; i < len(checkpoints); i++ {
		if err := r.Delete(context.Background(), campaignID, checkpoints[i].BatchID); err != nil {
			return err
		}
	}

	return nil
}

// computeCheckpointHash computes SHA256 hash of canon+state snapshots
func computeCheckpointHash(checkpoint *domain.PipelineCheckpoint) (string, error) {
	hasher := sha256.New()

	// Serialize canon snapshot
	canonData, err := json.Marshal(checkpoint.CanonSnapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal canon snapshot: %w", err)
	}

	// Serialize state snapshot
	stateData, err := json.Marshal(checkpoint.StateSnapshot)
	if err != nil {
		return "", fmt.Errorf("failed to marshal state snapshot: %w", err)
	}

	// Hash both
	if _, err := hasher.Write(canonData); err != nil {
		return "", err
	}
	if _, err := hasher.Write(stateData); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// hasSuffix checks if filename ends with suffix (helper for Go 1.24 compatibility)
func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}
