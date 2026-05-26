package repository

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pauvalls/grimorio/internal/domain"
)

// AuditLogRepository handles audit log persistence
type AuditLogRepository interface {
	// Append adds a new audit entry (append-only)
	Append(ctx context.Context, entry *domain.AuditLogEntry) error

	// GetLog retrieves log entries for a campaign
	GetLog(ctx context.Context, campaignID string, daysBack int) ([]*domain.AuditLogEntry, error)

	// PurgeOld removes entries older than retention period
	PurgeOld(ctx context.Context, campaignID string, olderThan time.Duration) (int, error)
}

// filesystemAuditRepository implements AuditLogRepository using JSONL files
type filesystemAuditRepository struct {
	baseDir string
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(baseDir string) AuditLogRepository {
	return &filesystemAuditRepository{
		baseDir: baseDir,
	}
}

func (r *filesystemAuditRepository) Append(
	ctx context.Context,
	entry *domain.AuditLogEntry,
) error {
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid audit entry: %w", err)
	}

	// Ensure campaign directory exists
	campaignDir := filepath.Join(r.baseDir, entry.CampaignID)
	if err := os.MkdirAll(campaignDir, 0755); err != nil {
		return fmt.Errorf("failed to create campaign directory: %w", err)
	}

	// Open audit log file in append mode
	logPath := filepath.Join(campaignDir, "audit_log.jsonl")
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	// Serialize entry as JSON
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Append line (JSONL format: one JSON object per line)
	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("failed to write audit entry: %w", err)
	}

	return nil
}

func (r *filesystemAuditRepository) GetLog(
	ctx context.Context,
	campaignID string,
	daysBack int,
) ([]*domain.AuditLogEntry, error) {
	logPath := filepath.Join(r.baseDir, campaignID, "audit_log.jsonl")

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return []*domain.AuditLogEntry{}, nil
	}

	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}
	defer file.Close()

	entries := []*domain.AuditLogEntry{}
	cutoff := time.Now().AddDate(0, 0, -daysBack)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry domain.AuditLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // Skip malformed lines
		}

		// Filter by date
		if entry.Timestamp.After(cutoff) {
			entries = append(entries, &entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read audit log: %w", err)
	}

	return entries, nil
}

func (r *filesystemAuditRepository) PurgeOld(
	ctx context.Context,
	campaignID string,
	olderThan time.Duration,
) (int, error) {
	logPath := filepath.Join(r.baseDir, campaignID, "audit_log.jsonl")

	// Check if file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return 0, nil
	}

	// Read all entries
	entries, err := r.GetLog(ctx, campaignID, 9999) // Get all entries
	if err != nil {
		return 0, err
	}

	// Filter out old entries
	cutoff := time.Now().Add(-olderThan)
	retained := []*domain.AuditLogEntry{}
	deletedCount := 0

	for _, entry := range entries {
		if entry.Timestamp.After(cutoff) {
			retained = append(retained, entry)
		} else {
			deletedCount++
		}
	}

	if deletedCount == 0 {
		return 0, nil
	}

	// Rewrite file with retained entries (JSONL requires full rewrite for deletion)
	tempPath := logPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create temp file: %w", err)
	}

	for _, entry := range retained {
		data, err := json.Marshal(entry)
		if err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return 0, fmt.Errorf("failed to marshal entry: %w", err)
		}

		if _, err := tempFile.WriteString(string(data) + "\n"); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return 0, fmt.Errorf("failed to write entry: %w", err)
		}
	}

	tempFile.Close()

	// Atomically replace original file
	if err := os.Rename(tempPath, logPath); err != nil {
		os.Remove(tempPath)
		return 0, fmt.Errorf("failed to replace audit log: %w", err)
	}

	return deletedCount, nil
}

// MemoryAuditLogRepository is an in-memory implementation for tests
type MemoryAuditLogRepository struct {
	entries map[string][]*domain.AuditLogEntry // campaignID -> entries
}

// NewMemoryAuditLogRepository creates a new in-memory audit log repository
func NewMemoryAuditLogRepository() AuditLogRepository {
	return &MemoryAuditLogRepository{
		entries: make(map[string][]*domain.AuditLogEntry),
	}
}

func (r *MemoryAuditLogRepository) Append(ctx context.Context, entry *domain.AuditLogEntry) error {
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid audit entry: %w", err)
	}
	r.entries[entry.CampaignID] = append(r.entries[entry.CampaignID], entry)
	return nil
}

func (r *MemoryAuditLogRepository) GetLog(ctx context.Context, campaignID string, daysBack int) ([]*domain.AuditLogEntry, error) {
	entries, ok := r.entries[campaignID]
	if !ok {
		return []*domain.AuditLogEntry{}, nil
	}
	if daysBack <= 0 {
		return entries, nil
	}
	cutoff := time.Now().AddDate(0, 0, -daysBack)
	result := make([]*domain.AuditLogEntry, 0)
	for _, e := range entries {
		if e.Timestamp.After(cutoff) {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *MemoryAuditLogRepository) PurgeOld(ctx context.Context, campaignID string, olderThan time.Duration) (int, error) {
	entries, ok := r.entries[campaignID]
	if !ok {
		return 0, nil
	}
	cutoff := time.Now().Add(-olderThan)
	result := make([]*domain.AuditLogEntry, 0)
	deletedCount := 0
	for _, e := range entries {
		if e.Timestamp.After(cutoff) {
			result = append(result, e)
		} else {
			deletedCount++
		}
	}
	r.entries[campaignID] = result
	return deletedCount, nil
}
